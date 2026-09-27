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

// TODO refact core/run for expansibility

func isTomlExt(path string) bool {
	return filepath.Ext(path) == ".toml"
}

func loadReq(path string, env *env.Env, config *config.Config) (request.Request, error) {
	// Read Flie
	b, err := fsx.ReadFile(path)
	if err != nil {
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

// exec Request Params
type execRParam struct {
	// root dir path
	root string
	// the file path
	path string

	cfg *config.Config
	env *env.Env
}

func execReq(ep execRParam, resCh chan RunResult) {
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
		resCh <- res
		return
	}

	r, err := loadReq(absPath, ep.env, ep.cfg)
	if err != nil {
		res.Err = errors.Join(res.Err, err)
		resCh <- res
		return
	}
	res.ReqRes = r.Run()
	resCh <- res
}
func execReqWithContext(ctx context.Context, ep execRParam, resCh chan RunResult) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		execReq(ep, resCh)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		return
	}
}

// paths params for run needed
type RunPaths struct {
	// root dir's abs path
	Root string
	// file for directly and directory for walk
	Tasks []string
	// config file path
	Config string
	// env file path
	Env string
}
type RunResult struct {
	// Path relative to the root collection
	Path    string
	AbsPath string
	ReqRes  request.Result
	Err     error
}

func Run(ctx context.Context, runParams RunPaths, handle func(res RunResult)) {
	// load config
	cfg, err := config.LoadFromFile(runParams.Config)
	// if err,then use the new config
	if err != nil {
		cfg = config.New()
	}
	// load e in current directory
	e, err := env.LoadFromFile(runParams.Env)
	// if err,then use the new empty env
	if err != nil {
		e = env.New()
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
				execReqWithContext(ctx, execRParam{root: runParams.Root, path: taskPath, cfg: cfg, env: e}, resCh)
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
					execReqWithContext(ctx, execRParam{root: runParams.Root, path: p, cfg: cfg, env: e}, resCh)
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

}
