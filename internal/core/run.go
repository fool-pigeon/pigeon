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

	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
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

func loadReq(ctx context.Context, path string, env *env.Env, config *config.Config) (request.Request, error) {
	// Read Flie
	b, err := fsx.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Expand Env
	b = []byte(env.Expand(string(b)))
	var meta request.ReqMeta
	if err := toml.Unmarshal(b, &meta); err != nil {
		return nil, err
	}
	switch meta.Protocol {
	default:
		return nil, fmt.Errorf("the protocol %q not implement yet.", meta.Protocol)
	case request.P_HTTP:
		client, err := config.HttpClient()
		if err != nil {
			return nil, err
		}

		var r httpx.ReqData
		if err := toml.Unmarshal(b, &r); err != nil {
			return nil, err
		}

		r.SetClient(client)

		return &r, nil

		// TODO readRequest more
	}
}

type ReqCtx struct {
	Config *config.Config
	Env    *env.Env
}

// exec Request Params
type execRParam struct {
	ReqCtx
	// root dir path
	root string
	// the file path
	path string
}

func execReq(ctx context.Context, ep execRParam) RunResult {
	var res RunResult
	absPath, err := filepath.Abs(ep.path)
	if err != nil {
		res.Err = errors.Join(res.Err, err)
	}
	res.AbsPath = absPath

	// ep.root is abs path, should use absPath to rel
	relPath, err := filepath.Rel(ep.root, absPath)
	if err != nil {
		res.Err = errors.Join(res.Err, err)
	}
	res.Path = relPath

	if res.Err != nil {
		return res
	}

	r, err := loadReq(ctx, absPath, ep.Env, ep.Config)
	if err != nil {
		res.Err = errors.Join(res.Err, err)
		return res
	}
	if err := ctx.Err(); err != nil {
		res.Err = errors.Join(res.Err, err)
		return res
	}

	res.ReqRes = r.RunWithContext(ctx)
	return res
}
func execReqWithContext(ctx context.Context, ep execRParam, resCh chan RunResult) {
	done := make(chan RunResult)
	var res RunResult
	go func() {
		defer close(done)
		res = execReq(ctx, ep)
	}()

	select {
	case <-done:
		if ctx.Err() != nil {
			return
		}
		resCh <- res
	case <-ctx.Done():
		return
	}
}

// run paths needed
// paths are relative paths from collection's root path
type RunPaths struct {
	Root   string   `toml:"root" json:"root"`
	Tasks  []string `toml:"tasks" json:"tasks"`
	Config string   `toml:"config" json:"config"`
	Env    string   `toml:"env" json:"env"`
}

// paths params for run needed
type RunParams struct {
	// readonly
	ReqCtx
	// root dir's abs path
	Root string
	// file for directly and directory for walk
	Tasks []string
}
type RunResult struct {
	// Path relative to the root collection
	Path    string
	AbsPath string
	ReqRes  request.Result
	Err     error
}

func Run(ctx context.Context, runParams RunParams, handle func(res RunResult)) error {
	if runParams.Config == nil {
		return fmt.Errorf("config:%w", ErrNilValue)
	}
	if runParams.Env == nil {
		return fmt.Errorf("env:%w", ErrNilValue)
	}

	var wg sync.WaitGroup
	resCh := make(chan RunResult)

	for _, taskPath := range runParams.Tasks {
		taskPath = filepath.Clean(taskPath)

		info, err := os.Lstat(taskPath)
		if err != nil {
			go func() {
				resCh <- RunResult{Path: taskPath, Err: err}
			}()
			continue
		}
		// exec if path is file
		if info.Mode().IsRegular() && isTomlExt(taskPath) {
			wg.Go(func() {
				execReqWithContext(ctx, execRParam{root: runParams.Root, path: taskPath, ReqCtx: runParams.ReqCtx}, resCh)
			})
			continue
		}

		// recursive directory
		filepath.WalkDir(taskPath, func(p string, d fs.DirEntry, err error) error {
			if d.Name() == fsx.RootTagDirName {
				return filepath.SkipDir
			}
			if d.Type().IsRegular() && isTomlExt(p) {
				wg.Go(func() {
					execReqWithContext(ctx, execRParam{root: runParams.Root, path: p, ReqCtx: runParams.ReqCtx}, resCh)
				})
			}
			return nil
		})
	}

	go func() {
		defer close(resCh)
		wg.Wait()
	}()

	for res := range resCh {
		handle(res)
	}
	return ctx.Err()
}
