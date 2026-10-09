// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// [TaskRunService] Event
type TrsEvent = string

const (
	// go receive data:
	// id:string
	//
	// it will cancel the task whose task's id equals that
	TrsCancel TrsEvent = "task-run-service:cancel"
	// go send data:
	// id:string
	//
	// it will send the cancelled task's id
	TrsCancelld TrsEvent = "task-run-service:cancelled"
	// go Send data:
	// [TaskResErr]
	TrsResultErr TrsEvent = "task-run-service:result-error"
	// go Send data:
	// [TaskResHttp]
	TrsResultHttp TrsEvent = "task-run-service:result-http"
	// go Send data:
	// [TaskResAny]
	TrsResultAny TrsEvent = "task-run-service:result-any"
)

type TaskMeta struct {
	// set UUID,id of Task
	Id string `json:"id"`
	// For Task:this is file path for calc collection's root path
	//
	// For TaskRes: this is the runned file path of relative path from root
	Path string `json:"path"`
}
type Task struct {
	TaskMeta
	// file content of run file
	Param core.RunPaths `json:"param"`
}
type TaskRes[T any] struct {
	request.ResultMeta
	Meta TaskMeta `json:"meta"`
	Data T        `json:"data"`
}

type TaskResErr = TaskRes[struct{}]
type TaskResHttp = TaskRes[HttpRes]
type TaskResAny = TaskRes[any]

// service for running run tasks or collection
type TaskRunService struct {
	reqCtx *ReqCtxService

	alertService *AlertService

	app *application.App
}

func NewTaskRunService(reqCtx *ReqCtxService, alertService *AlertService, app *application.App) *TaskRunService {
	return &TaskRunService{reqCtx, alertService, app}
}

func (tr *TaskRunService) RunDir(id string, path string) (Task, error) {
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return Task{}, fmt.Errorf("%s:%w", path, err)
	}
	if !info.IsDir() {
		return Task{}, fmt.Errorf("%s:%w", path, fsx.ErrNotDir)
	}
	root, err := fsx.RootDir(path)
	if err != nil {
		return Task{}, err
	}
	relRoot, err := filepath.Rel(root, path)
	task := Task{
		TaskMeta{Id: id, Path: path},
		core.RunPaths{
			Root:   relRoot,
			Tasks:  []string{"."},
			Config: tr.reqCtx.CfgRelPath(),
			Env:    tr.reqCtx.EnvRelPath(),
		},
	}
	go func() {
		tr.Run(task)
	}()
	return task, nil
}

// param: file path and file content of run file
//
// use the file path to calc the collection's root path
//
// returns id of Task
func (tr *TaskRunService) Run(task Task) error {
	root, err := fsx.RootDir(task.Path)
	if err != nil {
		return fmt.Errorf("%s:%w", task.Path, fsx.ErrRootNotExist)
	}

	p := task.Param
	// config load
	var cfg *config.Config
	switch {
	case p.Config == "":
		cfg = config.New()
	default:
		configFile := filepath.Join(root, p.Config)
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			tr.alertService.Warn(fmt.Sprintf("load %s,failed,use empty config", configFile))
			cfg = config.New()
		}
	}
	// env load
	var e *env.Env
	switch {
	case p.Env == "":
		e = env.New()
	default:
		envFile := filepath.Join(root, p.Env)
		e, err = env.LoadFromFile(envFile)
		if err != nil {
			tr.alertService.Warn(fmt.Sprintf("load %s,failed,use empty env", envFile))
			e = env.New()
		}
	}

	for i := range p.Tasks {
		p.Tasks[i] = filepath.Join(root, p.Root, p.Tasks[i])
	}

	runParams := core.RunParams{
		ReqCtx: core.ReqCtx{Config: cfg, Env: e},
		Root:   filepath.Join(root, p.Root),
		Tasks:  p.Tasks,
	}
	return tr.run(task.Id, runParams)
}
func (tr *TaskRunService) run(id string, params core.RunParams) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if tr.app != nil {
		ungister := tr.app.Event.On(TrsCancel, func(event *application.CustomEvent) {
			if cancelledId, ok := event.Data.(string); ok && cancelledId == id {
				cancel()
			}
		})
		defer ungister()
	}

	evt := tr.app.Event
	core.Run(ctx, params, func(res core.RunResult) {
		meta := TaskMeta{Id: id, Path: res.Path}
		if res.Err != nil {
			evt.Emit(TrsResultErr, TaskResErr{Meta: meta, ResultMeta: request.ResultMeta{Err: res.Err}})
			return
		}
		rRes := res.ReqRes
		rMeta := rRes.ResultMeta
		if rRes.Err != nil {
			evt.Emit(TrsResultErr, TaskResErr{Meta: meta, ResultMeta: rMeta})
			return
		}

		switch raw := rRes.Raw.(type) {
		case *httpx.Response:
			evt.Emit(TrsResultHttp, tr.reqCtx.parseResHttp(request.ResultView[*httpx.Response]{
				ResultMeta: res.ReqRes.ResultMeta,
				Raw:        raw,
			}, meta))

		default:
			evt.Emit(TrsResultAny, TaskResAny{ResultMeta: rMeta, Meta: meta, Data: raw})
		}
	})

	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			tr.app.Event.Emit(TrsCancelld, id)
			return fmt.Errorf("cancelled task:%w", ctx.Err())
		}
	}
	return ctx.Err()
}
