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
	Id ulid.ULID `json:"id"`
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

type taskRes struct {
	core.RunResult
	ID ulid.ULID

	HTTPRequest  TaskHTTPExt
	HTTPResponse TaskHTTPExt
}
type TaskResViewMeta struct {
	protocol.ResultCover
	ID   ulid.ULID
	Path string
}
type TaskResView struct {
	taskRes
	Spent string
	// response body size
	ContentSize string

	HTTPRequest  *TaskHTTPExtView
	HTTPResponse *TaskHTTPExtView
}

func (res *taskRes) toViewMeta() (*TaskResViewMeta, error) {
	if res.Res == nil {
		return nil, ErrNilValue
	}
	return &TaskResViewMeta{
		ResultCover: res.Res.ResultCover,
		ID:          res.ID,
		Path:        res.Path,
	}, nil
}
func (res *taskRes) toView() (*TaskResView, error) {
	if res.Res == nil {
		return nil, ErrNilValue
	}

	var view TaskResView
	var err error
	view.HTTPRequest, err = res.HTTPRequest.toView()
	if err != nil {
		return nil, err
	}
	view.HTTPResponse, err = res.HTTPResponse.toView()
	if err != nil {
		return nil, err
	}
	view.taskRes = *res
	view.Spent = res.Res.Duration.Round(time.Millisecond / 10).String()
	view.ContentSize = humanSize(len([]byte(view.HTTPResponse.Body)))

	return &view, nil
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
		Id: id, Path: path,
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
			if cancelledId, ok := event.Data.(string); ok && cancelledId == task.Id.String() {
				cancel()
			}
		})
		defer ungister()
	}

	var rootConf core.RootConf
	if err := fsx.ReadToml(task.RunPaths.RootConf, &rootConf); err != nil {
		alert.warn(fmt.Sprintf("%s read error,using empty root conf", task.RunPaths.RootConf))
	}

	resID, err := uid.NewULIDMonotonic()
	if err != nil {
		return err
	}
	input := core.RunInput{
		Start:    task.RunPaths.Start,
		Tasks:    task.RunPaths.Tasks,
		RootConf: rootConf,
		ResultHandler: func(res core.RunResult) {
			// TODO
			if res.Err != nil {
				alert.error(fmt.Sprint(res.Path, res.Err))
				return
			}
			if res.Res == nil {
				alert.error(fmt.Sprint(res.Path, ErrNilValue))
				return
			}

			tRes := taskRes{RunResult: res, ID: resID}
			tr.parseRunResRaw(&res, &tRes)
			if err := tr.storeTaskRes(tRes, task.Id); err != nil {
				// TODO
				return
			}

		},
		TraceHandler: func(t *protocol.Trace) {
			// TODO
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

	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			tr.evt.Emit(TaskSECancelled, task.Id)
			return fmt.Errorf("cancelled task:%w", ctx.Err())
		}
	}
	return ctx.Err()
}
func (tr *TaskService) parseRunResRaw(res *core.RunResult, tRes *taskRes) {
	switch v := res.Res.Raw.(type) {
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
func (tr *TaskService) storeTaskRes(res taskRes, taskID ulid.ULID) error {
	return tr.db.getDB().Update(func(txn *badger.Txn) error {
		viewMeta, err := res.toViewMeta()
		if err != nil {
			return err
		}
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyPrefixTaskRes), taskID[:], res.ID[:]), viewMeta); err != nil {
			return err
		}
		if err := txn.Set(makeDBKey([]byte(dbKeyPrefixReqRes), res.Res.Request.ID[:], res.ID[:]), nil); err != nil {
			return err
		}
		if err := dbTxnSet(txn, makeDBKey([]byte(dbKeyPrefixRes), res.ID[:]), res); err != nil {
			return err
		}

		// TODO emit event view meta to front

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
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(data); err != nil {
		return *new(T), err
	}
	return data, nil
}
func gzipCompress(b []byte) ([]byte, error) {
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
