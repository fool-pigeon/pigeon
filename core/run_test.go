// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"resty.dev/v3"

	"github.com/fool-pigeon/pigeon/dispatch"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
)

// writeFile writes data to path and records any error via assert.
func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	assert.NoError(t, os.WriteFile(path, data, 0o644))
}

// mkdirAll creates dir and records any error via assert.
func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(dir, 0o755))
}

// newCollection creates a temp collection root with a .pg tag dir and an empty
// root .conf, so initReqCtx and Run succeed against it.
func newCollection(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, fsx.RootTag))
	writeFile(t, filepath.Join(root, fsx.Conf), nil)
	return root
}

// restReqBody returns the TOML for a minimal REST GET request against url.
func restReqBody(url string) string {
	return fmt.Sprintf("protocol = %q\nurl = %q\nmethod = %q\n\n[config_rest]\n",
		protocol.ProtocolREST, url, protocol.MethodGet)
}

// newDispatcher returns a Dispatcher backed by a fresh resty client, closed on cleanup.
func newDispatcher(t *testing.T) *dispatch.Dispatcher {
	t.Helper()
	client := resty.New()
	t.Cleanup(func() { client.Close() })
	return dispatch.New(client)
}

// newReqCtx returns a ReqCtx with an empty Conf and a fresh Env.
func newReqCtx() *ReqCtx {
	return &ReqCtx{Conf: &protocol.ReqConf{}, Env: NewEnv()}
}

// countCancelCtx reports a nil Err until n calls, then context.Canceled. It
// exercises the second ctx.Err guard in loadReq/sendReq deterministically.
type countCancelCtx struct {
	context.Context
	n int
}

func (c *countCancelCtx) Err() error {
	if c.n <= 0 {
		return context.Canceled
	}
	c.n--
	return nil
}

// runCollector gathers RunResultHandler calls; safe for concurrent use.
type runCollector struct {
	mu   sync.Mutex
	path []string
	errs []error
}

func (c *runCollector) handle(res *RunResult, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if res != nil {
		c.path = append(c.path, res.Path)
	}
	c.errs = append(c.errs, err)
}

func TestIsTomlExt(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"toml", "a.toml", true},
		{"uppercase", "a.TOML", false},
		{"no ext", "a", false},
		{"other ext", "a.json", false},
		{"dir with toml name", "dir.toml/file", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, isTomlExt(c.path))
		})
	}
}

func TestLoadReq(t *testing.T) {
	conf := &protocol.ReqConf{Timeout: "1s"}
	env := NewEnv()
	assert.NoError(t, env.Set("HOST", "example.com"))

	validBody := "protocol = \"rest\"\nurl = \"http://${HOST}/x\"\nmethod = \"GET\"\n\n[config_rest]\n"

	t.Run("nil reqCtx", func(t *testing.T) {
		_, err := loadReq(context.Background(), "x", nil)
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("nil env", func(t *testing.T) {
		_, err := loadReq(context.Background(), "x", &ReqCtx{Conf: conf})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("nil conf", func(t *testing.T) {
		_, err := loadReq(context.Background(), "x", &ReqCtx{Env: env})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("file not exist", func(t *testing.T) {
		_, err := loadReq(context.Background(), filepath.Join(t.TempDir(), "nope.toml"), &ReqCtx{Conf: conf, Env: env})
		assert.Error(t, err)
	})
	t.Run("invalid toml", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "bad.toml")
		writeFile(t, p, []byte("= bad"))
		_, err := loadReq(context.Background(), p, &ReqCtx{Conf: conf, Env: env})
		assert.Error(t, err)
	})
	t.Run("expand env and merge", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte(validBody))
		req, err := loadReq(context.Background(), p, &ReqCtx{Conf: conf, Env: env})
		assert.NoError(t, err)
		if !assert.NotNil(t, req) {
			return
		}
		assert.Equal(t, "http://example.com/x", req.URL)
		// Merge filled the empty Timeout from the parent conf.
		assert.Equal(t, protocol.DurationStr("1s"), req.Timeout)
	})
	t.Run("canceled ctx before read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte(validBody))
		_, err := loadReq(ctx, p, &ReqCtx{Conf: conf, Env: env})
		assert.ErrorIs(t, err, context.Canceled)
	})
	t.Run("canceled ctx after unmarshal", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte(validBody))
		ctx := &countCancelCtx{Context: context.Background(), n: 1}
		_, err := loadReq(ctx, p, &ReqCtx{Conf: conf, Env: env})
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestSendReq(t *testing.T) {
	t.Run("nil dispatcher", func(t *testing.T) {
		_, err := sendReq(context.Background(), reqSendInput{reqCtx: newReqCtx()})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("nil reqCtx", func(t *testing.T) {
		_, err := sendReq(context.Background(), reqSendInput{dispatcher: newDispatcher(t)})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("nil env", func(t *testing.T) {
		_, err := sendReq(context.Background(), reqSendInput{dispatcher: newDispatcher(t), reqCtx: &ReqCtx{Conf: &protocol.ReqConf{}}})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("nil conf", func(t *testing.T) {
		_, err := sendReq(context.Background(), reqSendInput{dispatcher: newDispatcher(t), reqCtx: &ReqCtx{Env: NewEnv()}})
		assert.ErrorIs(t, err, ErrNilValue)
	})
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer srv.Close()

		dir := t.TempDir()
		p := filepath.Join(dir, "req.toml")
		writeFile(t, p, []byte(restReqBody(srv.URL)))

		res, err := sendReq(context.Background(), reqSendInput{
			dispatcher: newDispatcher(t),
			reqCtx:     newReqCtx(),
			start:      dir,
			path:       p,
		})
		assert.NoError(t, err)
		if !assert.NotNil(t, res) {
			return
		}
		assert.Equal(t, p, res.AbsPath)
		assert.Equal(t, "req.toml", res.Path)
		if assert.NotNil(t, res.Res) {
			assert.True(t, res.Res.Success)
		}
	})
	t.Run("load error", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "req.toml")
		writeFile(t, p, []byte("= bad"))
		_, err := sendReq(context.Background(), reqSendInput{
			dispatcher: newDispatcher(t),
			reqCtx:     newReqCtx(),
			start:      dir,
			path:       p,
		})
		assert.Error(t, err)
	})
	t.Run("dispatch error", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "req.toml")
		writeFile(t, p, []byte("protocol = \"grpc\"\nurl = \"http://x\"\nmethod = \"GET\"\n"))
		_, err := sendReq(context.Background(), reqSendInput{
			dispatcher: newDispatcher(t),
			reqCtx:     newReqCtx(),
			start:      dir,
			path:       p,
		})
		assert.ErrorIs(t, err, protocol.ErrNotSupport)
	})
	t.Run("canceled ctx after load", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "req.toml")
		writeFile(t, p, []byte(restReqBody("http://x")))
		ctx := &countCancelCtx{Context: context.Background(), n: 2}
		_, err := sendReq(ctx, reqSendInput{
			dispatcher: newDispatcher(t),
			reqCtx:     newReqCtx(),
			start:      dir,
			path:       p,
		})
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestSendReqWithStop(t *testing.T) {
	newInput := func(t *testing.T, path string) reqSendInput {
		return reqSendInput{
			dispatcher: newDispatcher(t),
			reqCtx:     newReqCtx(),
			start:      t.TempDir(),
			path:       path,
		}
	}

	t.Run("canceled ctx skips handler", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte(restReqBody("http://x")))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		called := false
		sendReqWithStop(ctx, func(res *RunResult, err error) { called = true }, newInput(t, p))
		assert.False(t, called)
	})
	t.Run("success calls handler", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte(restReqBody(srv.URL)))

		var got *RunResult
		var gotErr error
		sendReqWithStop(context.Background(), func(res *RunResult, err error) {
			got, gotErr = res, err
		}, newInput(t, p))
		assert.NoError(t, gotErr)
		if assert.NotNil(t, got) {
			assert.NotNil(t, got.Res)
		}
	})
	t.Run("error falls back to path", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "req.toml")
		writeFile(t, p, []byte("protocol = \"grpc\"\nurl = \"http://x\"\nmethod = \"GET\"\n"))

		var got *RunResult
		var gotErr error
		sendReqWithStop(context.Background(), func(res *RunResult, err error) {
			got, gotErr = res, err
		}, newInput(t, p))
		assert.ErrorIs(t, gotErr, protocol.ErrNotSupport)
		if assert.NotNil(t, got) {
			assert.Equal(t, p, got.Path)
			assert.Nil(t, got.Res)
		}
	})
}

func TestRunPathsCheck(t *testing.T) {
	t.Run("start not exist", func(t *testing.T) {
		rp := RunPaths{Start: filepath.Join(t.TempDir(), "nope")}
		assert.ErrorIs(t, rp.Check(), os.ErrNotExist)
	})
	t.Run("start is file", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "f")
		writeFile(t, p, nil)
		assert.ErrorIs(t, RunPaths{Start: p}.Check(), fsx.ErrNotDir)
	})
	t.Run("empty rootconf ok", func(t *testing.T) {
		assert.NoError(t, RunPaths{Start: t.TempDir()}.Check())
	})
	t.Run("dot rootconf ok", func(t *testing.T) {
		assert.NoError(t, RunPaths{Start: t.TempDir(), RootConf: "."}.Check())
	})
	t.Run("rootconf equals start ok", func(t *testing.T) {
		start := t.TempDir()
		assert.NoError(t, RunPaths{Start: start, RootConf: start}.Check())
	})
	t.Run("rootconf not exist", func(t *testing.T) {
		rp := RunPaths{Start: t.TempDir(), RootConf: filepath.Join(t.TempDir(), "nope")}
		assert.ErrorIs(t, rp.Check(), os.ErrNotExist)
	})
	t.Run("rootconf is dir", func(t *testing.T) {
		assert.ErrorIs(t, RunPaths{Start: t.TempDir(), RootConf: t.TempDir()}.Check(), fsx.ErrNotFile)
	})
	t.Run("rootconf is file ok", func(t *testing.T) {
		rootConf := filepath.Join(t.TempDir(), "c.toml")
		writeFile(t, rootConf, nil)
		assert.NoError(t, RunPaths{Start: t.TempDir(), RootConf: rootConf}.Check())
	})
}

func TestRunPathsToAbs(t *testing.T) {
	root := t.TempDir()
	t.Run("relative paths joined", func(t *testing.T) {
		rp := RunPaths{Start: "sub", RootConf: "conf.toml", Tasks: []string{"a.toml", "/abs/b.toml"}}
		got := rp.ToAbs(root)
		assert.Equal(t, filepath.Join(root, "sub"), got.Start)
		assert.Equal(t, filepath.Join(root, "conf.toml"), got.RootConf)
		assert.Equal(t, []string{filepath.Join(root, "sub", "a.toml"), "/abs/b.toml"}, got.Tasks)
	})
	t.Run("absolute paths kept", func(t *testing.T) {
		absStart := filepath.Join(root, "sub")
		absConf := filepath.Join(root, "c.toml")
		rp := RunPaths{Start: absStart, RootConf: absConf, Tasks: []string{absStart}}
		got := rp.ToAbs("/other")
		assert.Equal(t, absStart, got.Start)
		assert.Equal(t, absConf, got.RootConf)
		assert.Equal(t, []string{absStart}, got.Tasks)
	})
	t.Run("empty tasks", func(t *testing.T) {
		got := RunPaths{Start: "sub", RootConf: "c.toml"}.ToAbs(root)
		assert.Empty(t, got.Tasks)
	})
}

func TestRunPathsToRel(t *testing.T) {
	root := t.TempDir()
	start := filepath.Join(root, "sub")
	rp := RunPaths{
		Start:    start,
		RootConf: filepath.Join(root, "c.toml"),
		Tasks:    []string{filepath.Join(start, "a.toml")},
	}
	got := rp.ToRel(root)
	assert.Equal(t, "sub", got.Start)
	assert.Equal(t, "c.toml", got.RootConf)
	assert.Equal(t, []string{"a.toml"}, got.Tasks)
}

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	t.Run("init req ctx error", func(t *testing.T) {
		// A start dir without a .pg tag fails before any task runs.
		assert.ErrorIs(t, Run(context.Background(), RunInput{Start: t.TempDir()}), fsx.ErrRootNotExist)
	})

	t.Run("missing task reported", func(t *testing.T) {
		root := newCollection(t)
		missing := filepath.Join(root, "missing.toml")
		resCh := make(chan *RunResult, 1)
		errCh := make(chan error, 1)
		handler := func(res *RunResult, err error) { resCh <- res; errCh <- err }

		assert.NoError(t, Run(context.Background(), RunInput{
			Start:         root,
			Tasks:         []string{missing},
			ResultHandler: handler,
		}))
		res := <-resCh
		if assert.NotNil(t, res) {
			assert.Equal(t, missing, res.Path)
		}
		assert.Error(t, <-errCh)
	})

	t.Run("regular files", func(t *testing.T) {
		root := newCollection(t)
		a := filepath.Join(root, "a.toml")
		b := filepath.Join(root, "b.toml")
		writeFile(t, a, []byte(restReqBody(srv.URL)))
		writeFile(t, b, []byte(restReqBody(srv.URL)))

		col := &runCollector{}
		assert.NoError(t, Run(context.Background(), RunInput{
			Start:         root,
			Tasks:         []string{a, b},
			ResultHandler: col.handle,
		}))
		assert.ElementsMatch(t, []string{"a.toml", "b.toml"}, col.path)
		assert.Len(t, col.errs, 2)
		for _, err := range col.errs {
			assert.NoError(t, err)
		}
	})

	t.Run("walk dir", func(t *testing.T) {
		root := newCollection(t)
		sub := filepath.Join(root, "sub")
		deep := filepath.Join(sub, "deep")
		mkdirAll(t, deep)
		writeFile(t, filepath.Join(sub, "s.toml"), []byte(restReqBody(srv.URL)))
		writeFile(t, filepath.Join(deep, "d.toml"), []byte(restReqBody(srv.URL)))
		// Non-toml files are ignored by the walk.
		writeFile(t, filepath.Join(sub, "note.txt"), []byte("x"))

		col := &runCollector{}
		assert.NoError(t, Run(context.Background(), RunInput{
			Start:         root,
			Tasks:         []string{sub},
			ResultHandler: col.handle,
		}))
		assert.ElementsMatch(t, []string{
			filepath.Join("sub", "s.toml"),
			filepath.Join("sub", "deep", "d.toml"),
		}, col.path)
		for _, err := range col.errs {
			assert.NoError(t, err)
		}
	})

	t.Run("walk skips root tag", func(t *testing.T) {
		root := newCollection(t)
		writeFile(t, filepath.Join(root, "a.toml"), []byte(restReqBody(srv.URL)))
		// A request inside the .pg tag dir must not be executed.
		writeFile(t, filepath.Join(root, fsx.RootTag, "ignored.toml"), []byte(restReqBody(srv.URL)))

		col := &runCollector{}
		assert.NoError(t, Run(context.Background(), RunInput{
			Start:         root,
			Tasks:         []string{root},
			ResultHandler: col.handle,
		}))
		assert.Equal(t, []string{"a.toml"}, col.path)
		for _, err := range col.errs {
			assert.NoError(t, err)
		}
	})

	t.Run("client init called", func(t *testing.T) {
		var called bool
		assert.NoError(t, Run(context.Background(), RunInput{
			Start: newCollection(t),
			ClientInit: func(client *protocol.ActualClient) {
				called = true
			},
		}))
		assert.True(t, called)
	})

	t.Run("concurrency", func(t *testing.T) {
		t.Run("positive", func(t *testing.T) {
			root := newCollection(t)
			for _, name := range []string{"a.toml", "b.toml"} {
				writeFile(t, filepath.Join(root, name), []byte(restReqBody(srv.URL)))
			}
			col := &runCollector{}
			assert.NoError(t, Run(context.Background(), RunInput{
				Start:         root,
				Tasks:         []string{filepath.Join(root, "a.toml"), filepath.Join(root, "b.toml")},
				RootConf:      RootConf{MaxRunningConcurrency: 1},
				ResultHandler: col.handle,
			}))
			assert.Len(t, col.path, 2)
		})
		t.Run("negative clamps to zero", func(t *testing.T) {
			root := newCollection(t)
			a := filepath.Join(root, "a.toml")
			writeFile(t, a, []byte(restReqBody(srv.URL)))
			col := &runCollector{}
			assert.NoError(t, Run(context.Background(), RunInput{
				Start:         root,
				Tasks:         []string{a},
				RootConf:      RootConf{MaxRunningConcurrency: -1},
				ResultHandler: col.handle,
			}))
			assert.Equal(t, []string{"a.toml"}, col.path)
		})
	})

	t.Run("canceled ctx", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.ErrorIs(t, Run(ctx, RunInput{Start: newCollection(t)}), context.Canceled)
	})
}
