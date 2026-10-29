// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/fool-pigeon/pigeon/core"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/pgrest"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/oklog/ulid/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
	"resty.dev/v3"
)

var (
	ErrStreamDisconnected = errors.New("stream disconnected")
)

// [TaskService] Event
type TaskSEvent = string

const (
	// receive [Task.ID]
	//
	// it will cancel the task whose task's id equals that
	TaskSECancel TaskSEvent = "task-service:cancel"
	// send [Task.ID]
	//
	// it will send the cancelled task's id
	TaskSECancelled TaskSEvent = "task-service:cancelled"
	// receive [protocol.Request.ID]
	//
	// it will cancel the single request sending whose request's id equals that
	TaskSECancelSingle TaskSEvent = "task-service:cancel-single"
	// send [protocol.Result.ID]
	//
	// it send resID to [TaskSEResIDInitial]:{reqID}
	TaskSEResIDInitial TaskSEvent = "task-service:res-id-initial"
)

// [TaskService] Stream
type TaskSStream = string

const (
	// Send JSON
	//
	// [Task]
	TaskSSTask TaskSStream = "task-service:stream:task"
	// Send JSON
	//
	// [TaskResMeta]
	TaskSSResMeta TaskSStream = "task-service:stream:task-res-meta"
	// Send JSON
	//
	// [protocol.TraceView]
	TaskSSTraceView TaskSStream = "task-service:stream:task-res-trace-view"
)

type Task struct {
	ID ulid.ULID `json:"id"`
	// For Task:this is file path for calc collection's root path
	Path string `json:"path"`
	// file content of run file
	RunPaths core.RunPaths
}

type TaskHTTPExt struct {
	Header map[string][]string
	Body   []byte
	Dump   []byte
}
type TaskHTTPExtView struct {
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`
	Dump   string              `json:"raw"`
}

func (h *TaskHTTPExt) toView() (*TaskHTTPExtView, error) {
	var err error
	var body, dump []byte
	if len(h.Body) > 0 {
		body, err = gzipDecompress(h.Body)
		if err != nil {
			return nil, err
		}
	}
	if len(h.Dump) > 0 {
		dump, err = gzipDecompress(h.Dump)
		if err != nil {
			return nil, err
		}
	}
	return &TaskHTTPExtView{
		Header: h.Header,
		Body:   string(body),
		Dump:   string(dump),
	}, nil
}

type TaskResBody struct {
	core.RunMeta
	protocol.ResMeta
	ErrStr  string `json:"errStr"`
	Content []byte `json:"content"`
}
type taskRes struct {
	TaskResBody
	HTTPRequest  TaskHTTPExt
	HTTPResponse TaskHTTPExt
}
type TaskResMeta struct {
	protocol.ResultCover
	TaskID ulid.ULID `json:"taskID"`
	ResID  ulid.ULID `json:"resID"`
	Path   string    `json:"path"`
}
type TaskResView struct {
	TaskResBody
	Spent string `json:"spent"`
	// response body size
	ContentSize  string           `json:"contentSize"`
	HTTPRequest  *TaskHTTPExtView `json:"httpRequest"`
	HTTPResponse *TaskHTTPExtView `json:"httpResponse"`
}

func (res *taskRes) toMeta(taskID ulid.ULID) TaskResMeta {
	return TaskResMeta{
		TaskID:      taskID,
		ResID:       res.ResMeta.ID,
		Path:        res.RunMeta.Path,
		ResultCover: res.ResMeta.ResultCover,
	}
}
func (res *taskRes) toView() (*TaskResView, error) {
	reqExt, err := res.HTTPRequest.toView()
	if err != nil {
		return nil, err
	}
	resExt, err := res.HTTPResponse.toView()
	if err != nil {
		return nil, err
	}
	return &TaskResView{
		TaskResBody:  res.TaskResBody,
		Spent:        res.ResMeta.Duration.Round(time.Millisecond / 10).String(),
		ContentSize:  humanSize(len(resExt.Body)),
		HTTPRequest:  reqExt,
		HTTPResponse: resExt,
	}, nil
}

// service for running run tasks or collection
type TaskService struct {
	app *application.App
	evt *application.EventManager
	db  *CacheDB

	taskCh        chan *Task
	taskDone      chan struct{}
	resMetaCh     chan *TaskResMeta
	resMetaDone   chan struct{}
	traceViewCh   chan *protocol.TraceView
	traceViewDone chan struct{}
}

func NewTaskRunService(app *application.App, db *CacheDB) *TaskService {
	return &TaskService{
		app: app,
		evt: app.Event,
		db:  db,

		taskCh:        make(chan *Task),
		taskDone:      make(chan struct{}),
		resMetaCh:     make(chan *TaskResMeta, 8),
		resMetaDone:   make(chan struct{}),
		traceViewCh:   make(chan *protocol.TraceView, 16),
		traceViewDone: make(chan struct{}),
	}
}

// ServiceShutdown implements [application.ServiceShutdown].
func (tr *TaskService) ServiceShutdown() error {
	close(tr.taskDone)
	close(tr.taskCh)
	close(tr.resMetaDone)
	close(tr.resMetaCh)
	close(tr.traceViewDone)
	close(tr.traceViewCh)
	return nil
}

func registerJSONOutputStreamHandler[T any](app *application.App, name string, ch chan T, done ...chan struct{}) {
	app.HandleStream(name, func(sc *application.StreamConn) {
		defer sc.Close()
		if len(done) > 0 {
			defer func() {
				done[0] <- struct{}{}
			}()
		}
		for {
			select {
			case <-sc.Context().Done():
				return
			case v := <-ch:
				if err := sc.SendJSON(&v); err != nil {
					return
				}
			}
		}
	})
}

// ServiceStartup implements [application.ServiceStartup].
func (tr *TaskService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	registerJSONOutputStreamHandler(tr.app, TaskSSTask, tr.taskCh, tr.taskDone)
	registerJSONOutputStreamHandler(tr.app, TaskSSResMeta, tr.resMetaCh, tr.resMetaDone)
	registerJSONOutputStreamHandler(tr.app, TaskSSTraceView, tr.traceViewCh, tr.traceViewDone)
	return nil
}

func (tr *TaskService) LoadTaskList() error {
	return tr.db.GetTaskRepo().SeekValues(func(t *Task) error {
		select {
		case <-tr.taskDone:
			return ErrStreamDisconnected
		case tr.taskCh <- t:
			return nil
		}
	}, true)
}
func (tr *TaskService) LoadTaskResMetaList(taskID ulid.ULID) error {
	return tr.db.getDB().View(func(txn *badger.Txn) error {
		return dbTxnSeekValues(txn, makeDBKey([]byte(dbKeyTaskResMeta), taskID[:], nil), func(trm *TaskResMeta) error {
			select {
			case <-tr.resMetaDone:
				return ErrStreamDisconnected
			case tr.resMetaCh <- trm:
				return nil
			}
		})
	})
}
func (tr *TaskService) LoadTraceViewList(resID ulid.ULID) error {
	return tr.db.getDB().View(func(txn *badger.Txn) error {
		return dbTxnSeekValues(txn, makeDBKey([]byte(dbKeyResTraceView), resID[:], nil), func(t *protocol.TraceView) error {
			select {
			case <-tr.traceViewDone:
				return ErrStreamDisconnected
			case tr.traceViewCh <- t:
				return nil
			}
		})
	})
}

func (tr *TaskService) RunDir(path string, rootConf string) error {
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("%s:%w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s:%w", path, fsx.ErrNotDir)
	}
	id, err := uid.NewULID()
	if err != nil {
		return err
	}
	task := Task{
		ID: id, Path: path,
		RunPaths: core.RunPaths{
			Start:    path,
			Tasks:    []string{path},
			RootConf: rootConf,
		},
	}
	return tr.Run(task)
}

// param: file path and file content of run file
//
// all paths are absolute
func (tr *TaskService) Run(task Task) error {
	root, err := fsx.RootDir(task.Path)
	if err != nil {
		return fmt.Errorf("%s:%w", task.Path, fsx.ErrRootNotExist)
	}

	if err := task.RunPaths.Check(); err != nil {
		return err
	}
	task.RunPaths = task.RunPaths.ToAbs(root)

	return tr.run(context.Background(), task)
}

type resultInner struct {
	runResult *core.RunResult
	err       error
}

func (tr *TaskService) RunSingleReq(rootConfPath string, reqPath string, request *protocol.Request) (*TaskResView, error) {
	if request == nil {
		return nil, ErrNilValue
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// register request cancel
	unregister := tr.evt.On(TaskSECancelSingle, func(event *application.CustomEvent) {
		if cancelledID, ok := event.Data.(string); ok && cancelledID == request.ID.String() {
			cancel()
		}
	})
	defer unregister()

	// custom generate resID & send to front
	resID, err := uid.NewULIDMonotonic()
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteString(TaskSEResIDInitial)
	sb.WriteByte(':')
	sb.WriteString(request.ID.String())
	tr.evt.Emit(sb.String(), resID.String())

	var rootConf core.RootConf
	if err := fsx.ReadToml(rootConfPath, &rootConf); err != nil {
		alert.warn(fmt.Sprintf("%s read error,using empty root conf", rootConfPath))
	}

	// send trace view stream
	traceCh := make(chan *protocol.Trace, 2)
	defer close(traceCh)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case trace, ok := <-traceCh:
				if !ok {
					return
				}
				if err := tr.storeTaskResTrace(trace); err != nil {
					alert.error(fmt.Sprint(trace.ResID, trace.ID, err))
					continue
				}
			}
		}
	}()
	var inner resultInner
	input := core.RunInput{
		Start:      filepath.Dir(reqPath),
		Tasks:      []string{filepath.Clean(reqPath)},
		RootConf:   rootConf,
		ClientInit: tr.reqClientInit,
		Handlers: core.RunHandlers{
			ResIDGen: func() ulid.ULID { return resID },
			ResultHandler: func(res *core.RunResult, err error) {
				inner.runResult = res
				inner.err = err
			},
			TraceHandler: func(t *protocol.Trace) {
				if t == nil {
					return
				}
				select {
				case <-ctx.Done():
				case traceCh <- t:
				}
			},
		},
	}
	if err := core.Run(ctx, input); err != nil {
		return nil, err
	}
	// parse result
	if inner.runResult.Res == nil {
		alert.error(fmt.Sprint(inner.runResult.Path, ErrNilValue))
		return nil, nil
	}
	tRes := taskRes{TaskResBody: TaskResBody{RunMeta: inner.runResult.RunMeta, ResMeta: inner.runResult.Res.ResMeta}}
	if err := inner.err; err != nil {
		tRes.ErrStr = err.Error()
	} else {
		tr.parseRunResRaw(inner.runResult, &tRes)
	}
	// write into db
	if err := tr.db.getDB().Update(func(txn *badger.Txn) error {
		if err := txn.Set(makeDBKey([]byte(dbKeyReqRes), tRes.ResMeta.Request.ID[:], tRes.ResMeta.ID[:]), nil); err != nil {
			return err
		}
		return dbTxnSet(txn, makeDBKey([]byte(dbKeyResID), tRes.ResMeta.ID[:]), &tRes)
	}); err != nil {
		return nil, err
	}

	// convert and return
	return tRes.toView()
}

func (tr *TaskService) run(ctx context.Context, task Task) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if tr.evt != nil {
		ungister := tr.evt.On(TaskSECancel, func(event *application.CustomEvent) {
			if cancelledID, ok := event.Data.(string); ok && cancelledID == task.ID.String() {
				cancel()
			}
		})
		defer ungister()
	}

	var rootConf core.RootConf
	if err := fsx.ReadToml(task.RunPaths.RootConf, &rootConf); err != nil {
		alert.warn(fmt.Sprintf("%s read error,using empty root conf", task.RunPaths.RootConf))
	}

	innerCh := make(chan resultInner, 8)
	defer close(innerCh)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case inner, ok := <-innerCh:
				if !ok {
					return
				}
				runResult := inner.runResult
				if runResult.Res == nil {
					alert.error(fmt.Sprint(runResult.Path, ErrNilValue))
					continue
				}
				tRes := taskRes{TaskResBody: TaskResBody{RunMeta: runResult.RunMeta, ResMeta: runResult.Res.ResMeta}}
				if err := inner.err; err != nil {
					tRes.ErrStr = err.Error()
				} else {
					tr.parseRunResRaw(runResult, &tRes)
				}
				if err := tr.storeTaskRes(tRes, task.ID); err != nil {
					alert.error(fmt.Sprint(tRes.RunMeta.Path, err))
					continue
				}
			}
		}
	}()

	traceCh := make(chan *protocol.Trace, 16)
	defer close(traceCh)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case trace, ok := <-traceCh:
				if !ok {
					return
				}
				if err := tr.storeTaskResTrace(trace); err != nil {
					alert.error(fmt.Sprint(trace.ResID, trace.ID, err))
					continue
				}
			}
		}
	}()

	input := core.RunInput{
		Start:      task.RunPaths.Start,
		Tasks:      task.RunPaths.Tasks,
		RootConf:   rootConf,
		ClientInit: tr.reqClientInit,
		Handlers: core.RunHandlers{
			ResultHandler: func(runResult *core.RunResult, err error) {
				select {
				case <-ctx.Done():
				case innerCh <- resultInner{runResult, err}:
				}
			},
			TraceHandler: func(t *protocol.Trace) {
				if t == nil {
					return
				}
				select {
				case <-ctx.Done():
				case traceCh <- t:
				default:
					alert.error("[task-service] run traceCh chan flood")
				}
			},
		},
	}
	if err := core.Run(ctx, input); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			tr.evt.Emit(TaskSECancelled, task.ID)
			return fmt.Errorf("cancelled task:%w", err)
		}
		return err
	}
	return nil
}

// The key for get body in Context of [httpx.Request]
// For example,get the body value from context of [http.Request] named req:
//
// b := req.Context().Value(HTTPBodyKey{}).([]byte)
type HTTPBodyKey struct{}

func (tr *TaskService) reqClientInit(client *protocol.ActualClient) {
	client.SetRequestMiddlewares(
		resty.MiddlewareRequestCreate,
		func(c *resty.Client, req *resty.Request) error {
			var bodyBytes []byte
			if r := req.RawRequest; r != nil && r.Body != nil {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}
				r.Body.Close()
				bodyBytes = b
				// Resume Body, subsequent requests sent normally
				r.Body = io.NopCloser(bytes.NewReader(b))
			}
			req.SetContext(context.WithValue(req.Context(), HTTPBodyKey{}, bodyBytes))
			return nil
		},
	)
}
func (tr *TaskService) parseRunResRaw(runRes *core.RunResult, tRes *taskRes) {
	switch v := runRes.Res.Raw.(type) {
	case *pgrest.Response:
		// Parse the result for request
		if req := v.Request; req != nil {
			// get request body
			body, _ := req.Context().Value(HTTPBodyKey{}).([]byte)
			tRes.HTTPRequest.Body, _ = gzipCompress(body)
			// set request header
			tRes.HTTPRequest.Header = req.Header
			if r := req.RawRequest; r != nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
				if b, err := httputil.DumpRequestOut(r, true); err == nil {
					tRes.HTTPRequest.Dump, _ = gzipCompress(b)
				}
			}
		}
		// Parse the result for response
		if resp := v.RawResponse; resp != nil {
			b := v.Bytes()
			resp.Body = io.NopCloser(bytes.NewReader(b))
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				// view.ResMeta.Raw = string(b)
				tRes.HTTPResponse.Dump, _ = gzipCompress(b)
			}
		}
		tRes.HTTPResponse.Header = v.Header()
		tRes.HTTPResponse.Body, _ = gzipCompress(v.Bytes())
	}
}
func (tr *TaskService) storeTaskRes(tRes taskRes, taskID ulid.ULID) error {
	return tr.db.getDB().Update(func(txn *badger.Txn) error {
		// write txn
		resMeta := tRes.toMeta(taskID)
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyTaskResMeta), taskID[:], tRes.ResMeta.ID[:]), &resMeta); err != nil {
			return err
		}
		if err := txn.Set(makeDBKey([]byte(dbKeyReqRes), tRes.ResMeta.Request.ID[:], tRes.ResMeta.ID[:]), nil); err != nil {
			return err
		}
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyResID), tRes.ResMeta.ID[:]), &tRes); err != nil {
			return err
		}

		// to stream,emit to front
		tr.resMetaCh <- &resMeta

		return nil
	})
}
func (tr *TaskService) storeTaskResTrace(trace *protocol.Trace) error {
	return tr.db.getDB().Update(func(txn *badger.Txn) error {
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyResTraceView), trace.ResID[:], uid.Uint64ToBytes(trace.ID)), trace); err != nil {
			return err
		}

		// to stream,emit to front
		traceView := trace.ToView()
		tr.traceViewCh <- &traceView

		return nil
	})
}

func gzipCompress(b []byte) ([]byte, error) {
	if len(b) <= 0 {
		return b, nil
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func gzipDecompress(b []byte) ([]byte, error) {
	if len(b) <= 0 {
		return b, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
func humanSize(bytesLen int) string {
	const base = 1024
	if bytesLen < base {
		return fmt.Sprintf("%dB", bytesLen)
	}
	n := float64(bytesLen)
	for _, unit := range []string{"KB", "MB", "GB", "TB", "PB"} {
		n /= base
		if n < base {
			return fmt.Sprintf("%.1f%s", n, unit)
		}
	}
	return fmt.Sprintf("%.1fPB", n)
}
