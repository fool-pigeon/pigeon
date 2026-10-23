// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"path/filepath"
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

// [TaskService] Event
type TaskSEvent = string

const (
	// go receive data:
	// id:string
	//
	// it will cancel the task whose task's id equals that
	TaskSECancel TaskSEvent = "task-service:cancel"
	// go send data:
	// id:string
	//
	// it will send the cancelled task's id
	TaskSECancelled TaskSEvent = "task-service:cancelled"
	// go Send data:
	// [TaskSEResView]
	TaskSEResView TaskSEvent = "task-service:res-view"
	// go Send data:
	// [TaskResViewMeta]
	TaskSEResViewMeta TaskSEvent = "task-service:res-view-meta"
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

type taskResMeta struct {
	RunMeta core.RunMeta
	ResMeta protocol.ResMeta
	ErrStr  string
	Content []byte
}
type taskRes struct {
	taskResMeta
	HTTPRequest  TaskHTTPExt
	HTTPResponse TaskHTTPExt
}
type TaskResViewMeta struct {
	protocol.ResultCover
	ID   ulid.ULID
	Path string
}
type TaskResView struct {
	taskResMeta
	Spent string
	// response body size
	ContentSize  string
	HTTPRequest  *TaskHTTPExtView
	HTTPResponse *TaskHTTPExtView
}

func (res *taskRes) toViewMeta() TaskResViewMeta {
	return TaskResViewMeta{
		ID:          res.ResMeta.ID,
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
		taskResMeta:  res.taskResMeta,
		Spent:        res.ResMeta.Duration.Round(time.Millisecond / 10).String(),
		ContentSize:  humanSize(len(resExt.Body)),
		HTTPRequest:  reqExt,
		HTTPResponse: resExt,
	}, nil
}

// service for running run tasks or collection
type TaskService struct {
	evt *application.EventManager
	db  *CacheDB
}

func NewTaskRunService(app *application.App, db *CacheDB) *TaskService {
	return &TaskService{app.Event, db}
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

	return tr.run(task)
}

// The key for get body in Context of [httpx.Request]
// For example,get the body value from context of [http.Request] named req:
//
// b := req.Context().Value(HttpBodyKey{}).([]byte)
type HttpBodyKey struct{}

func (tr *TaskService) run(task Task) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if tr.evt != nil {
		ungister := tr.evt.On(TaskSECancel, func(event *application.CustomEvent) {
			if cancelledId, ok := event.Data.(string); ok && cancelledId == task.ID.String() {
				cancel()
			}
		})
		defer ungister()
	}

	var rootConf core.RootConf
	if err := fsx.ReadToml(task.RunPaths.RootConf, &rootConf); err != nil {
		alert.warn(fmt.Sprintf("%s read error,using empty root conf", task.RunPaths.RootConf))
	}

	type resultInner struct {
		runResult *core.RunResult
		err       error
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
				tRes := taskRes{taskResMeta: taskResMeta{RunMeta: runResult.RunMeta, ResMeta: runResult.Res.ResMeta}}
				if err := inner.err; err != nil {
					tRes.ErrStr = err.Error()
				} else {
					tr.parseRunResRaw(runResult, &tRes)
				}
				if err := tr.storeTaskRes(tRes, task.ID); err != nil {
					// TODO
					continue
				}
			}
		}
	}()

	traceCh := make(chan protocol.Trace, 32)
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
					// TODO
					continue
				}
			}
		}
	}()

	input := core.RunInput{
		Start:    task.RunPaths.Start,
		Tasks:    task.RunPaths.Tasks,
		RootConf: rootConf,
		ResultHandler: func(runResult *core.RunResult, err error) {
			select {
			case innerCh <- resultInner{runResult, err}:
			case <-ctx.Done():
			}
		},
		TraceHandler: func(t *protocol.Trace) {
			if t == nil {
				return
			}
			select {
			case traceCh <- *t:
			default:
				alert.error("[task-service] run traceCh chan flood")
			}
		},
		ClientInit: func(client *protocol.ActualClient) {
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
					req.SetContext(context.WithValue(req.Context(), HttpBodyKey{}, bodyBytes))
					return nil
				},
			)
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
func (tr *TaskService) parseRunResRaw(runRes *core.RunResult, tRes *taskRes) {
	switch v := runRes.Res.Raw.(type) {
	case *pgrest.Response:
		// Parse the result for request
		if req := v.Request; req != nil {
			// get request body
			body, _ := req.Context().Value(HttpBodyKey{}).([]byte)
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
		viewMeta := tRes.toViewMeta()
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyPrefixTaskRes), taskID[:], tRes.ResMeta.ID[:]), viewMeta); err != nil {
			return err
		}
		if err := txn.Set(makeDBKey([]byte(dbKeyPrefixReqRes), tRes.ResMeta.Request.ID[:], tRes.ResMeta.ID[:]), nil); err != nil {
			return err
		}
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyPrefixResID), tRes.ResMeta.ID[:]), tRes); err != nil {
			return err
		}

		// TODO emit event view meta to front

		return nil
	})
}
func (tr *TaskService) storeTaskResTrace(trace protocol.Trace) error {
	return tr.db.getDB().Update(func(txn *badger.Txn) error {
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyPrefixResTrace), trace.ResID[:], uid.Uint64ToBytes(trace.ID)), trace); err != nil {
			return err
		}
		// TODO wrapper and emit task res trace

		return nil
	})
}
func dbTxnSet[T any](txn *badger.Txn, key []byte, data T) error {
	b, err := gobEncode(data)
	if err != nil {
		return err
	}
	return txn.Set(key, b)
}
func gobEncode[T any](data T) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func gobDecode[T any](b []byte) (T, error) {
	var data T
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
		return *new(T), err
	}
	return data, nil
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
