package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"path/filepath"
	"sync"

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

	cfg, err := config.LoadFromFile(filepath.Join(r.tag, fsx.ConfigDefaultName))
	if err != nil {
		cfg = config.New()
	}
	r.cfg = cfg
	if err := r.initReqClient(); err != nil {
		return "", err
	}
	return root, nil
}

// set config from path
func (r *ReqCtxService) LoadConfig(path string) error {
	cfg, err := config.LoadFromFile(path)
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
	evt, err := env.LoadFromFile(path)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.env = evt
	return nil
}
func (r *ReqCtxService) initReqClient() error {
	if r.cfg == nil {
		return fmt.Errorf("ReqCtxService - config:%w", ErrNilValue)
	}
	httpClient, err := r.cfg.HttpClient()
	if err != nil {
		return err
	}
	httpClient.SetRequestMiddlewares(
		resty.MiddlewareRequestCreate,
		func(c *resty.Client, req *resty.Request) error {
			var bodyBytes []byte
			if r := req.RawRequest; r != nil {
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
	request.ResultMeta
	URL      string   `json:"URL"`
	Method   string   `json:"method"`
	Request  HttpMeta `json:"request"`
	Response HttpMeta `json:"response"`
}

func (r *ReqCtxService) SendHttp(data *httpx.ReqData) (HttpRes, error) {
	var result HttpRes
	if data == nil {
		return result, fmt.Errorf("request data:%w", ErrNilValue)
	}
	httpClent, err := r.cfg.HttpClient()
	if err != nil {
		return result, err
	}
	// inject client
	data.SetClient(httpClent)

	ctx, cancel := context.WithCancel(context.Background())
	ungister := r.app.Event.On(gui.CancelRequest, func(event *application.CustomEvent) { cancel() })
	defer ungister()

	// Run with context
	res := request.View[*httpx.Response](data.RunWithContext(ctx))
	// set meta of result
	result.ResultMeta = res.ResultMeta

	// Parse the result for request
	if req := res.Raw.Request; req != nil {
		// get request body
		body := req.Context().Value(HttpBodyKey{}).([]byte)
		result.Request.Body = string(body)
		// set request header
		result.Request.Header = req.Header
		if r := req.RawRequest; r != nil {
			// set Method
			result.Method = r.Method
			// set request's URL after parsing in result
			if r.URL != nil {
				result.URL = r.URL.String()
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			// set raw request text
			if b, err := httputil.DumpRequestOut(r, true); err == nil {
				result.Request.Raw = string(b)
			}
		}
	}
	// Parse the result for response
	if resp := res.Raw.RawResponse; resp != nil {
		resp.Body = io.NopCloser(bytes.NewReader(res.Raw.Bytes()))
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			// set raw response text
			result.Response.Raw = string(b)
		}
	}
	result.Response.Body = res.Raw.String()
	result.Response.Header = res.Raw.Header()

	return result, nil
}
