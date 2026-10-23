// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fool-pigeon/pigeon/dispatch"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/pelletier/go-toml/v2"
)

var (
	ErrInvalidValue = errors.New("value is invalid")
	ErrNilValue     = errors.New("value is nil.")
)

// TODO refact core/run for expansibility

func isTomlExt(path string) bool {
	return filepath.Ext(path) == ".toml"
}

type ReqCtx struct {
	Conf *protocol.ReqConfig
	Env  *Env
}

// load request from file & expand env vars & merge ReqConfig
func loadReq(ctx context.Context, path string, reqCtx *ReqCtx) (*protocol.Request, error) {
	if reqCtx == nil || reqCtx.Env == nil || reqCtx.Conf == nil {
		return nil, fmt.Errorf("load request reqCtx:%w", ErrNilValue)
	}
	// Read Flie
	b, err := fsx.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Expand Env
	b = []byte(reqCtx.Env.Expand(string(b)))

	var req protocol.Request
	if err := toml.Unmarshal(b, &req); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req.ReqConfig.Merge(reqCtx.Conf)

	return &req, nil
}

type reqSendInput struct {
	dispatcher  *dispatch.Dispatcher
	traceHandle protocol.TraceHandler
	reqCtx      *ReqCtx
	start       string
	path        string
}

func sendReq(ctx context.Context, input reqSendInput) (*RunResult, error) {
	if input.dispatcher == nil {
		return nil, fmt.Errorf("exec request dispatcher:%w", ErrNilValue)
	}
	if input.reqCtx == nil || input.reqCtx.Env == nil || input.reqCtx.Conf == nil {
		return nil, fmt.Errorf("exec request reqCtx:%w", ErrNilValue)
	}

	var res RunResult
	absPath, err := filepath.Abs(input.path)
	if err != nil {
		return nil, err
	}
	res.AbsPath = absPath

	// ep.root is abs path, should use absPath to rel
	relPath, err := filepath.Rel(input.start, absPath)
	if err != nil {
		return nil, err
	}
	res.Path = relPath

	req, err := loadReq(ctx, absPath, input.reqCtx)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resID, err := uid.NewULIDMonotonic()
	if err != nil {
		return nil, err
	}
	response, err := input.dispatcher.Send(ctx, req, resID, uid.NewU64Gen(), input.traceHandle)
	if err != nil {
		return nil, err
	}
	res.Res = response

	return &res, nil
}

// the request will stop when ctx.Done()
func sendReqWithStop(ctx context.Context, resHandle RunResultHandler, input reqSendInput) {
	done := make(chan struct{})
	var res *RunResult
	var err error
	go func() {
		defer close(done)
		res, err = sendReq(ctx, input)
	}()

	select {
	case <-done:
		if ctx.Err() != nil {
			return
		}
		if err != nil || res == nil {
			res = &RunResult{RunMeta: RunMeta{Path: input.path}}
		}
		resHandle(res, err)

	case <-ctx.Done():
		return
	}
}

type RunMeta struct {
	// Path relative to the root collection
	Path    string
	AbsPath string
}
type RunResult struct {
	RunMeta
	Res *protocol.Result
}
type RunResultHandler func(res *RunResult, err error)

// paths params for run needed
type RunInput struct {
	// root dir's abs path
	Start string
	// file for directly and directory for walk
	Tasks []string

	RootConf RootConf

	ResultHandler RunResultHandler
	TraceHandler  protocol.TraceHandler

	ClientInit func(client *protocol.ActualClient)
}

func Run(ctx context.Context, rp RunInput) error {

	maxConcurrency := rp.RootConf.MaxRunningConcurrency
	if maxConcurrency < 0 {
		maxConcurrency = 0
	}
	sem := make(chan struct{}, maxConcurrency)
	defer close(sem)

	// set dispather to send request
	rootConf := rp.RootConf
	client, err := rootConf.Client.InstancedClient()
	if err != nil {
		return err
	}
	if rp.ClientInit != nil {
		rp.ClientInit(client)
	}

	dispatcher := dispatch.New(client)

	// set the ReqCtx
	initialReqCtx, err := rootConf.initReqCtx(rp.Start)
	if err != nil {
		return err
	}

	type reqCtxFrame struct {
		path string
		ctx  *ReqCtx
	}
	reqCtxStack := []*reqCtxFrame{{path: rp.Start, ctx: initialReqCtx}}

	var wg sync.WaitGroup
	runReq := func(path string, reqCtx *ReqCtx) {
		wg.Go(func() {
			if maxConcurrency > 0 {
				sem <- struct{}{}
				defer func() { <-sem }()
			}
			sendReqWithStop(ctx, rp.ResultHandler, reqSendInput{
				dispatcher:  dispatcher,
				traceHandle: rp.TraceHandler,
				reqCtx:      reqCtx,
				start:       rp.Start,
				path:        path,
			})
		})
	}
	for _, taskPath := range rp.Tasks {
		taskPath = filepath.Clean(taskPath)

		info, err := os.Lstat(taskPath)
		if err != nil {
			go func() {
				rp.ResultHandler(&RunResult{RunMeta: RunMeta{Path: taskPath}}, err)
			}()
			continue
		}
		// exec if path is file
		if info.Mode().IsRegular() && isTomlExt(taskPath) {
			runReq(taskPath, initialReqCtx)
			continue
		}

		// recursive directory
		filepath.WalkDir(taskPath, func(p string, d fs.DirEntry, err error) error {

			if d.Name() == fsx.RootTag {
				return filepath.SkipDir
			}

			parentDir := filepath.Dir(p)
			// pop the ctx if top ctx's path is not parent directory
			for len(reqCtxStack) > 1 && filepath.Clean(reqCtxStack[len(reqCtxStack)-1].path) != filepath.Clean(parentDir) {
				reqCtxStack = reqCtxStack[:len(reqCtxStack)-1]
			}
			// the top of reqCtxStack means the parent's parentFrame
			parentFrame := reqCtxStack[len(reqCtxStack)-1]

			switch {
			// exec if is a request file
			case d.Type().IsRegular() && isTomlExt(p):
				wg.Go(func() {
					runReq(p, parentFrame.ctx)
				})

			// add reqCtxFrame to stack
			case d.IsDir():
				confPath := filepath.Join(p, fsx.Conf)
				info, err := os.Lstat(confPath)

				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				if !info.Mode().IsRegular() || errors.Is(err, os.ErrNotExist) {
					reqCtxStack = append(reqCtxStack, &reqCtxFrame{p, parentFrame.ctx})
				}

				// override load for update env
				confBytes, err := fsx.ReadFile(confPath)
				if err != nil {
					return err
				}
				frame := reqCtxFrame{path: p}
				env := parentFrame.ctx.Env.Copy()
				reqConf := parentFrame.ctx.Conf.Copy()
				if err := loadConf(confBytes, env, reqConf); err != nil {
					return err
				}
				frame.ctx = &ReqCtx{reqConf, env}
				// into the next walk
				reqCtxStack = append(reqCtxStack, &frame)
			}
			return nil
		})
	}

	go func() {
		wg.Wait()
	}()

	return ctx.Err()
}

// run paths needed
// paths are relative paths from collection's root path
type RunPaths struct {
	// conf of collection
	RootConf string
	// the start directory
	Start string
	// the paths sub of start directory
	Tasks []string `toml:"tasks" json:"tasks"`
}

// check all path(except tasks) exist and paths' file types are corrent
//
// returns nil if all already corrent
func (rp RunPaths) Check() error {
	info, err := os.Lstat(rp.Start)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fsx.ErrNotDir
	}
	path := rp.RootConf
	// the empty string means that not use path
	if path == "" || path == "." || path == rp.Start {
		return nil
	}
	info, err = os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fsx.ErrNotFile
	}
	return nil
}

// root must be collection's abs root
// this will change all path to abs path
func (rp RunPaths) ToAbs(root string) RunPaths {
	res := RunPaths{Start: rp.Start, RootConf: rp.RootConf}
	if !filepath.IsAbs(rp.Start) {
		res.Start = filepath.Join(root, rp.Start)
	}
	if !filepath.IsAbs(rp.RootConf) {
		res.RootConf = filepath.Join(root, rp.RootConf)
	}
	tasks := make([]string, 0, len(rp.Tasks))
	for _, task := range rp.Tasks {
		absTask := task
		if !filepath.IsAbs(absTask) {
			absTask = filepath.Join(res.Start, task)
		}
		tasks = append(tasks, absTask)
	}
	res.Tasks = tasks
	return res
}

// root must be the collection's abs path
// change all abs paths to all rel paths
func (rp RunPaths) ToRel(root string) RunPaths {
	rel := func(base string, target string) string {
		relPath, _ := filepath.Rel(base, target)
		return relPath
	}
	tasks := make([]string, 0, len(rp.Tasks))
	for _, task := range rp.Tasks {
		tasks = append(tasks, rel(rp.Start, task))
	}
	return RunPaths{
		Start:    rel(root, rp.Start),
		RootConf: rel(root, rp.RootConf),
		Tasks:    tasks,
	}
}
