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
	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
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

type ReqCtxService struct {
	cfgRelPath string
	envRelPath string

	cfg *config.Config
	env *env.Env
	// collection root path
	root string
	// collection tag dir for config
	tag string

	app *application.App
	mu  sync.RWMutex
}

func NewReqCtxService(app *application.App) *ReqCtxService {
	return &ReqCtxService{
		cfg:  config.New(),
		env:  env.New(),
		root: "",
		tag:  "",
		app:  app,
	}
}
func (r *ReqCtxService) CfgRelPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfgRelPath
}
func (r *ReqCtxService) EnvRelPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.envRelPath
}
func (r *ReqCtxService) Config() *config.Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg
}
func (r *ReqCtxService) Env() *env.Env {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.env
}
func (r *ReqCtxService) Root() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.root
}
func (r *ReqCtxService) TagDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tag
}

// init root,tag,config from path and returns root
func (r *ReqCtxService) InitFromPath(path string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	root, err := fsx.RootDir(path)
	if err != nil {
		return "", err
	}
	r.root = root
	r.tag = filepath.Join(root, fsx.RootTagDirName)

	cfgPath := filepath.Join(r.tag, fsx.ConfigDefaultName)
	cfg, err := config.LoadFromFile(cfgPath)
	r.cfgRelPath, _ = filepath.Rel(r.root, cfgPath)

	if err != nil {
		cfg = config.New()
	}
	r.cfg = cfg
	if err := r.InitReqClient(); err != nil {
		return "", err
	}
	return root, nil
}

// set config from path
func (r *ReqCtxService) LoadConfig(path string) error {
	cfg, err := config.LoadFromFile(path)
	r.cfgRelPath, _ = filepath.Rel(r.root, path)

	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cfg = cfg
	return nil
}

// set env from path
func (r *ReqCtxService) LoadEnv(path string) error {
	e, err := env.LoadFromFile(path)
	r.envRelPath, _ = filepath.Rel(r.root, path)

	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.env = e
	return nil
}

type FileMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// TODO: parse [FIleMeta] from collections path
func (r *ReqCtxService) ParseCollectionCfgs(path string) ([]FileMeta, error) {
	root, err := fsx.RootDir(path)
	if err != nil {
		return nil, err
	}
	envDir := filepath.Join(root, fsx.RootTagDirName, fsx.ConfigDefaultName)
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
func (r *ReqCtxService) ParseCollectionEnvs(path string) ([]FileMeta, error) {
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

func (r *ReqCtxService) InitReqClient() error {
	return r.initConfig(r.cfg)
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
	httpClent, err := r.cfg.HttpClient()
	if err != nil {
		return TaskResHttp{}, err
	}
	// inject client
	data.SetClient(httpClent)

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
func (r *ReqCtxService) parseResHttp(res request.ResultView[*httpx.Response]) TaskResHttp {
	var taskRes TaskResHttp

	// set meta of result
	taskRes.ResultMeta = res.ResultMeta
	taskRes.Data.Spent = res.Duration.Round(time.Millisecond / 10).String()

	if res.Raw == nil {
		return taskRes
	}

	// Parse the result for request
	if req := res.Raw.Request; req != nil {
		// get request body
		body := req.Context().Value(HttpBodyKey{}).([]byte)
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
