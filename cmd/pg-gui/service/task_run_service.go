// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

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
	TaskCancel TaskSEvent = "task-service:cancel"
	// go send data:
	// id:string
	//
	// it will send the cancelled task's id
	TaskCancelled TaskSEvent = "task-service:cancelled"
	// go Send data:
	// [TaskResultView]
	TaskResultView TaskSEvent = "task-service:result-view"
)

type Task struct {
	Id ulid.ULID `json:"id"`
	// For Task:this is file path for calc collection's root path
	Path string `json:"path"`
	// file content of run file
	RunPaths core.RunPaths
}

type TaskRes struct {
	core.RunResult
	ID ulid.ULID
}

func (res *TaskRes) toTaskViewMeta() TaskResViewMeta {
	return TaskResViewMeta{
		ResultCover: protocol.ResultCover{Success: res.Res.Success, Tests: res.Res.Tests},
		ID:          res.ID,
		Path:        res.Path,
	}
}
func (res *TaskRes) toTaskView() TaskResView {
	view := TaskResView{
		TaskRes: *res,
		Spent:   res.Res.Duration.Round(time.Millisecond / 10).String(),
	}
	switch v := res.Res.Raw.(type) {
	case *pgrest.Response:
		// Parse the result for request
		if req := v.Request; req != nil {
			// get request body
			body, _ := req.Context().Value(HttpBodyKey{}).([]byte)
			view.ReqMeta.Body = string(body)
			// set request header
			view.ReqMeta.Header = req.Header
			if r := req.RawRequest; r != nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
				// set raw request text
				if b, err := httputil.DumpRequestOut(r, true); err == nil {
					view.ReqMeta.Raw = string(b)
				}
			}
		}
		// Parse the result for response
		if resp := v.RawResponse; resp != nil {
			b := v.Bytes()
			view.ContentSize = humanSize(len(b))
			resp.Body = io.NopCloser(bytes.NewReader(b))
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				// set raw response text
				view.ResMeta.Raw = string(b)
			}
		}
		view.ResMeta.Body = v.String()
		view.ResMeta.Header = v.Header()
	}

	return view
}

type TaskResViewMeta struct {
	protocol.ResultCover
	ID   ulid.ULID
	Path string
}
type HttpMeta struct {
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`
	Raw    string              `json:"raw"`
}
type TaskResView struct {
	TaskRes
	Spent       string
	ContentSize string

	ReqMeta HttpMeta
	ResMeta HttpMeta
}

// service for running run tasks or collection
type TaskService struct {
	evt *application.EventManager
}

func NewTaskRunService(app *application.App) *TaskService {
	return &TaskService{app.Event}
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
		ungister := tr.evt.On(TaskCancel, func(event *application.CustomEvent) {
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

	// evt := tr.app.Event
	input := core.RunInput{
		Start:    task.RunPaths.Start,
		Tasks:    task.RunPaths.Tasks,
		RootConf: rootConf,
		ResultHandler: func(res core.RunResult) {
			// TODO
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
			tr.evt.Emit(TaskCancelled, task.Id)
			return fmt.Errorf("cancelled task:%w", ctx.Err())
		}
	}
	return ctx.Err()
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
