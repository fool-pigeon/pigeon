package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
)

type Context struct {
	path string
	env  *env.Env
}

func readRequest(name string, ctx *Context) (request.Request, error) {
	// Read File
	b, err := fsx.ReadFile(name)
	if err != nil {
		return nil, err
	}
	// Expand Env
	b = []byte(env.Expand(string(b), ctx.env))

	var meta request.RequestMeta
	if err := toml.Unmarshal(b, &meta); err != nil {
		return nil, err
	}
	switch meta.Protocol {
	default:
		return nil, fmt.Errorf("the protocol %q not implement yet.", meta.Protocol)
	case request.P_HTTP:
		var r httpx.RequestData
		if err := toml.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		return &r, nil

		// TODO readRequest more
	}
}

func isTomlExt(path string) bool {
	return filepath.Ext(path) == ".toml"
}

func errResult(err error) request.Result {
	return request.Result{Err: err}
}
func execTask(name string, ctx *Context, resCh chan request.Result) {
	r, err := readRequest(name, ctx)
	if err != nil {
		resCh <- errResult(err)
		return
	}
	res, err := r.Run()
	if err != nil {
		resCh <- errResult(err)
		return
	}
	resCh <- res
}
func Run(paths []string, handle func(request.Result), tag ...string) error {
	// load startEnv in current directory
	startEnv, err := env.LoadFromPath(".", tag...)
	if err != nil {
		return err
	}
	ctxs := []Context{{".", startEnv}}

	var wg sync.WaitGroup
	resCh := make(chan request.Result)

	for _, path := range paths {
		path = filepath.Clean(path)

		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		// exec if path is file
		if info.Mode().IsRegular() && isTomlExt(path) {
			wg.Go(func() {
				execTask(path, &ctxs[0], resCh)
			})
			continue
		}

		// recursive directory
		filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			parentDir := filepath.Dir(p)
			// pop the ctx if top ctx's path is not parent directory
			for len(ctxs) > 1 && filepath.Clean(ctxs[len(ctxs)-1].path) != filepath.Clean(parentDir) {
				ctxs = ctxs[:len(ctxs)-1]
			}
			// the top of ctxs means the parent's ctx
			ctx := ctxs[len(ctxs)-1]
			switch {
			case d.Type().IsRegular() && isTomlExt(p):
				// exec if file
				wg.Go(func() {
					execTask(p, &ctx, resCh)
				})
			case d.IsDir():
				// override load for update env
				ctx.path = p
				env.LoadFile(ctx.env, filepath.Join(p, env.FileName(tag...)))
				// into the next walk
				ctxs = append(ctxs, ctx)

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

	return nil
}
