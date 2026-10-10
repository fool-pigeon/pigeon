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
	"sync"
	"time"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/core"
	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
	"resty.dev/v3"
)

var (
	ErrNilValue = errors.New("the value is nil.")
)

// The key for get body in Context of [httpx.Request]
// For example,get the body value from context of [http.Request] named req:
//
// b := req.Context().Value(HttpBodyKey{}).([]byte)
type HttpBodyKey struct{}

type FileMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
type CtxInfo struct {
	reqCtx core.ReqCtx
	Config FileMeta `json:"config"`
	Env    FileMeta `json:"env"`
}
type ReqCtxService struct {
	//current collection root path
	root string
	// root -> CtxInfo
	ctxs map[string]CtxInfo

	app *application.App
	mu  sync.RWMutex
}

func NewReqCtxService(app *application.App) *ReqCtxService {
	return &ReqCtxService{app: app, ctxs: map[string]CtxInfo{}}
}
func (r *ReqCtxService) ReqCtx() core.ReqCtx {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctxs[r.root].reqCtx
}
func (r *ReqCtxService) CfgInfo() FileMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctxs[r.root].Config
}
func (r *ReqCtxService) EnvInfo() FileMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctxs[r.root].Env
}
func (r *ReqCtxService) Root() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.root
}
func (r *ReqCtxService) SetCfgInfo(cfgFile FileMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.setCfgInfo(cfgFile)
}
func (r *ReqCtxService) setCfgInfo(cfgFile FileMeta) error {
	cfg, err := config.LoadFromFile(cfgFile.Path)
	if err != nil {
		return err
	}
	root := r.root
	ctx := r.ctxs[root]
	ctx.Config = cfgFile
	ctx.reqCtx.Config = cfg
	r.ctxs[root] = ctx
	return nil
}

func (r *ReqCtxService) SetEnvInfo(envFile FileMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.setEnvInfo(envFile)
}
func (r *ReqCtxService) setEnvInfo(envFile FileMeta) error {
	e, err := env.LoadFromFile(envFile.Path)
	if err != nil {
		return err
	}
	root := r.root
	ctx := r.ctxs[root]
	ctx.Env = envFile
	ctx.reqCtx.Env = e
	r.ctxs[root] = ctx
	return nil
}

// Switch root from path and returns root
func (r *ReqCtxService) SetRootFromPath(path string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	root, err := fsx.RootDir(path)
	if err != nil {
		return "", err
	}
	r.root = root

	ctx := r.ctxs[root]
	if ctx.reqCtx.Config == nil {
		ctx.reqCtx.Config = config.New()
	}
	if ctx.reqCtx.Env == nil {
		ctx.reqCtx.Env = env.New()
	}
	r.ctxs[root] = ctx

	// tip:temporary hard coding config path
	r.setCfgInfo(FileMeta{"config.toml", filepath.Join(root, fsx.RootTagDirName, "config.toml")})

	r.initConfig(r.ctxs[root].reqCtx.Config)

	return root, nil
}

// TODO: parse [FIleMeta] from collections path
//
//	func (r *ReqCtxService) ParseCfgs(path string) ([]FileMeta, error) {
//		root, err := fsx.RootDir(path)
//		if err != nil {
//			return nil, err
//		}
//		return res, nil
//	}

func (r *ReqCtxService) ParseEnvs(path string) ([]FileMeta, error) {
	root, err := fsx.RootDir(path)
	if err != nil {
		return nil, err
	}
	envDir := filepath.Join(root, fsx.RootTagDirName, fsx.EnvDirName)
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", envDir, err)
	}

	res := make([]FileMeta, 0)
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			res = append(res, FileMeta{entry.Name(), filepath.Join(envDir, entry.Name())})
		}
	}
	return res, nil
}

func (r *ReqCtxService) initConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("ReqCtxService - config:%w", ErrNilValue)
	}
	httpClient, err := cfg.HttpClient()
	if err != nil {
		return err
	}
	httpClient.SetRequestMiddlewares(
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
	return nil
}

type HttpMeta struct {
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`
	Raw    string              `json:"raw"`
}
type HttpRes struct {
	URL    string `json:"URL"`
	Method string `json:"method"`
	// String of Duration in [request.ResultMeta]
	Spent string `json:"spent"`
	// humaniszed string of ContentLength
	ContentSize string   `json:"size"`
	Request     HttpMeta `json:"request"`
	Response    HttpMeta `json:"response"`
}

func (r *ReqCtxService) SendHttp(data *httpx.ReqData) (TaskResHttp, error) {
	if data == nil {
		return TaskResHttp{}, fmt.Errorf("request data:%w", ErrNilValue)
	}

	// set request context
	reqCtx := r.ReqCtx()
	// inject client
	httpClent, err := reqCtx.Config.HttpClient()
	if err != nil {
		return TaskResHttp{}, err
	}
	data.SetClient(httpClent)
	// Expand env vars
	b, err := toml.Marshal(data)
	if err != nil {
		return TaskResHttp{}, err
	}
	if err := toml.Unmarshal([]byte(reqCtx.Env.Expand(string(b))), data); err != nil {
		return TaskResHttp{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if r.app != nil {
		ungister := r.app.Event.On(gui.CancelRequest, func(event *application.CustomEvent) { cancel() })
		defer ungister()
	}

	// Run with context
	res := request.View[*httpx.Response](data.RunWithContext(ctx))

	return r.parseResHttp(res), nil
}
func (r *ReqCtxService) parseResHttp(res request.ResultView[*httpx.Response], meta ...TaskMeta) TaskResHttp {
	var taskRes TaskResHttp
	if len(meta) > 0 {
		taskRes.Meta = meta[0]
	}
	// set meta of result
	taskRes.ResultMeta = res.ResultMeta
	taskRes.Data.Spent = res.Duration.Round(time.Millisecond / 10).String()

	if res.Raw == nil {
		return taskRes
	}

	// Parse the result for request
	if req := res.Raw.Request; req != nil {
		// get request body
		body, _ := req.Context().Value(HttpBodyKey{}).([]byte)
		taskRes.Data.Request.Body = string(body)
		// set request header
		taskRes.Data.Request.Header = req.Header
		if r := req.RawRequest; r != nil {
			// set Method
			taskRes.Data.Method = r.Method
			// set request's URL after parsing in result
			if r.URL != nil {
				taskRes.Data.URL = r.URL.String()
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			// set raw request text
			if b, err := httputil.DumpRequestOut(r, true); err == nil {
				taskRes.Data.Request.Raw = string(b)
			}
		}
	}
	// Parse the result for response
	if resp := res.Raw.RawResponse; resp != nil {
		b := res.Raw.Bytes()
		taskRes.Data.ContentSize = humanSize(b)
		resp.Body = io.NopCloser(bytes.NewReader(b))
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			// set raw response text
			taskRes.Data.Response.Raw = string(b)
		}
	}
	taskRes.Data.Response.Body = res.Raw.String()
	taskRes.Data.Response.Header = res.Raw.Header()

	return taskRes
}

func humanSize(b []byte) string {
	const base = 1024
	if len(b) < base {
		return fmt.Sprintf("%dB", len(b))
	}
	n := float64(len(b))
	for _, unit := range []string{"KB", "MB", "GB", "TB", "PB"} {
		n /= base
		if n < base {
			return fmt.Sprintf("%.1f%s", n, unit)
		}
	}
	return fmt.Sprintf("%.1fPB", n)
}
