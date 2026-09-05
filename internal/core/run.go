package core

import (
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

type Context struct {
	path string
	env  *env.Env
}

func readRequest(name string, ctx *Context, c *config.Config) (request.Request, error) {
	// Read File
	b, err := fsx.ReadFile(name)
	if err != nil {
		return nil, err
	}
	// Expand Env
	b = []byte(ctx.env.Expand(string(b)))

	var meta request.ReqMeta
	if err := toml.Unmarshal(b, &meta); err != nil {
		return nil, err
	}
	switch meta.Protocol {
	default:
		return nil, fmt.Errorf("the protocol %q not implement yet.", meta.Protocol)
	case request.P_HTTP:
		client, err := c.HttpClient()
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

func isTomlExt(path string) bool {
	return filepath.Ext(path) == ".toml"
}

func execTask(name string, ctx *Context, c *config.Config, resCh chan request.Result) {
	r, err := readRequest(name, ctx, c)
	if err != nil {
		resCh <- request.ErrResult(name, fmt.Errorf("ReadError:%w", err))
		return
	}
	res := r.Run()
	resCh <- res
}
func Run(paths []string, handle func(res request.Result), tag ...string) error {
	startDir := "."
	// load config
	cfg, err := config.LoadFromPath(startDir, tag...)
	// if err,then use the new config
	if err != nil {
		cfg = config.New()
	}

	// load startEnv in current directory
	startEnv, err := env.LoadFromPath(startDir, tag...)
	// use empty start env if not exist root tag dir
	// return if cause fatal error
	if errors.Is(err, fsx.ErrRootNotExist) {
		startEnv = env.New()
	} else if err != nil {
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
				execTask(path, &ctxs[0], cfg, resCh)
			})
			continue
		}

		// recursive directory
		filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if d.Name() == fsx.RootTagDirName {
				return filepath.SkipDir
			}

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
					execTask(p, &ctx, cfg, resCh)
				})
			case d.IsDir():
				// override load for update env
				ctx.path = p
				ctx.env.LoadFile(filepath.Join(p, env.FileName(tag...)))
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
