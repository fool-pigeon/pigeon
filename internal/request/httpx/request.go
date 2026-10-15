// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/jaswdr/faker/v2"
	"resty.dev/v3"
)

type (
	// type field of HTTP multipart/form-data
	MultiField struct {
		Name        string `toml:"name" json:"name"`
		Value       string `toml:"value" json:"value"`
		FileName    string `toml:"file_name" json:"fileName"`
		FilePath    string `toml:"file_path" json:"filePath"`
		ContentType string `toml:"content_type" json:"contentType"`
	}
	Body struct {
		// raw data
		Raw string `toml:"raw,multiline" comment:"raw data" json:"raw"`
		// multipart/form-data
		Multipart []MultiField `toml:"multipart,multiline" comment:"list of multipart/form-data" json:"multipart"`
		// x-www-form-url-encoded
		Form map[string][]string `toml:"form,multiline" comment:"key-values of x-www-form-url-encoded" json:"form"`
		// file/binary path
		Binary string `toml:"binary,multiline" comment:"the file/binary path" json:"binary"`
		// json,xml,text,or origin string value
		ContentType string `toml:"content_type,omitempty" comment:"'json'|'xml'|'text',else keep origin" json:"contentType"`
	}

	rDataError struct {
		Msg string
		Err error
	}
)

func (c *rDataError) Error() string {
	return fmt.Sprintf("[http.request_data] %s:%v", c.Msg, c.Err)
}

func (c *rDataError) Unwrap() error {
	return c.Err
}

type (
	Response = resty.Response
	Request  = resty.Request

	// [ReqData] is a wrap for [resty.Request].
	// It store request's data into toml file.
	// The request will be [resty.Request] when called [ReqData.InstancedRequest]
	ReqData struct {
		request.ReqMeta
		Method  Method              `toml:"mothod" comment:"GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS,TRACE,CONNECT" json:"method"`
		URL     string              `toml:"url" json:"url"`
		Path    map[string]string   `toml:"path,omitempty,multiline" json:"path"`
		Query   map[string][]string `toml:"query,omitempty,multiline" json:"query"`
		Header  map[string][]string `toml:"header,omitempty,multiline" json:"header"`
		Body    Body                `toml:"body,omitempty" json:"body"`
		Desc    string              `toml:"desc,omitempty,multiline" json:"desc"`
		Setting Setting             `toml:"setting,omitempty" json:"setting"`

		// the expr script before send request,could set random data for request
		PreReq string `toml:"pre_req,omitempty,multiline" json:"preReq"`

		// the exprs whose the result value type is bool
		// key is test's name
		// value is expr expression
		Tests map[string]string `toml:"tests,omitempty,multiline" json:"tests"`

		client *Client
		req    *Request
	}
)

// Run implements [request.Request].
func (r *ReqData) Run(ctxs ...context.Context) request.Result {
	var ctx context.Context
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	} else {
		ctx = context.Background()
	}

	info := fmt.Sprintf("%s %s", r.Method, r.URL)
	if r.req == nil {
		req, err := r.InstancedRequest()
		if err != nil {
			errRes := request.ErrResult(r.Name, err)
			errRes.SetReqInfo(info)
			return errRes
		}
		r.req = req
	}
	// send request
	res, err := r.req.SetContext(ctx).Send()
	if err != nil {
		errRes := request.ErrResult(r.Name, err)
		errRes.SetReqInfo(info).SetReqProto(request.P_HTTP)
		return errRes
	}
	// set raw response
	result := request.NewResult()
	result.SetRaw(res).SetReqName(r.ReqName()).SetReqProto(request.P_HTTP).SetReqInfo(info).SetProto(res.Proto()).SetStatus(res.Status())
	result.SetSuccess(res.StatusCode() >= 200 && res.StatusCode() < 300)
	result.SetDuration(res.Duration())

	// run tests
	if len(r.Tests) != 0 {
		for name, code := range r.Tests {
			tRes := request.TestRes{Name: name}
			env := map[string]any{
				"res":  res,
				"fake": faker.New(),
			}

			program, err := expr.Compile(code, expr.Env(env), expr.AsBool())
			if err != nil {
				tRes.Err = err
				result.SetSuccess(false)
				result.AppendTestRes(tRes)
				continue
			}
			output, err := expr.Run(program, env)
			if err != nil {
				tRes.Err = err
				result.SetSuccess(false)
				result.AppendTestRes(tRes)
				continue
			}

			tRes.Passed = output.(bool)
			result.SetSuccess(result.Success && tRes.Passed)
			result.AppendTestRes(tRes)
		}
	}

	return result
}

// RunWithContext implements [request.Request].
func (r *ReqData) RunWithContext(ctx context.Context) request.Result {
	res := r.Run(ctx)
	if err := ctx.Err(); err != nil {
		errRes := request.ErrResult(r.ReqName(), err)
		errRes.SetReqInfo(fmt.Sprintf("%s %s", r.Method, r.URL)).SetReqProto(request.P_HTTP)
		return errRes
	}
	return res
}

func NewReqData(requestName string) *ReqData {
	return &ReqData{ReqMeta: request.ReqMeta{
		Name:     requestName,
		Protocol: request.P_HTTP,
	}}
}

// Setters

func (r *ReqData) SetMethod(m Method) *ReqData {
	r.Method = m
	return r
}
func (r *ReqData) SetURL(url string) *ReqData {
	r.URL = url
	return r
}
func (r *ReqData) SetPathParams(params map[string]string) *ReqData {
	r.Path = params
	return r
}
func (r *ReqData) SetPathParam(key string, val string) *ReqData {
	if r.Path == nil {
		r.Path = make(map[string]string)
	}
	r.Path[key] = val
	return r
}
func (r *ReqData) SetQueryParams(params map[string][]string) *ReqData {
	r.Query = params
	return r
}
func (r *ReqData) SetHeaders(header map[string][]string) *ReqData {
	r.Header = header
	return r
}
func (r *ReqData) SetBody(body Body) *ReqData {
	r.Body = body
	return r
}
func (r *ReqData) SetBodyRaw(raw string) *ReqData {
	r.Body.Raw = raw
	return r
}
func (r *ReqData) SetBodyMultipart(multipart ...MultiField) *ReqData {
	r.Body.Multipart = multipart
	return r
}
func (r *ReqData) SetBodyForm(form map[string][]string) *ReqData {
	r.Body.Form = form
	return r
}
func (r *ReqData) SetBodyBinary(path string) *ReqData {
	r.Body.Binary = path
	return r
}
func (r *ReqData) SetBodyContentType(ct string) *ReqData {
	r.Body.ContentType = ct
	return r
}
func (r *ReqData) SetDesc(desc string) *ReqData {
	r.Desc = desc
	return r
}
func (r *ReqData) SetSetting(setting Setting) *ReqData {
	r.Setting = setting
	return r
}
func (r *ReqData) SetTimeout(timeStr string) *ReqData {
	r.Setting.Timeout = timeStr
	return r
}
func (r *ReqData) SetRetry(retry Retry) *ReqData {
	r.Setting.Retry = retry
	return r
}
func (r *ReqData) SetAuth(auth Auth) *ReqData {
	r.Setting.Auth = auth
	return r
}
func (r *ReqData) SetCookieString(cookieStr string) *ReqData {
	r.Setting.CookieString = cookieStr
	return r
}
func (r *ReqData) SetUnescapeQuery(unescape bool) *ReqData {
	r.Setting.UnescapeQuery = unescape
	return r
}
func (r *ReqData) SetUnescapeJson(unescape bool) *ReqData {
	r.Setting.UnescapeJson = unescape
	return r
}
func (r *ReqData) SetResponseBodyLimit(limit int64) *ReqData {
	r.Setting.ResponseBodyLimit = limit
	return r
}
func (r *ReqData) SetPreReq(preReq string) *ReqData {
	r.PreReq = preReq
	return r
}
func (r *ReqData) SetTests(tests map[string]string) *ReqData {
	r.Tests = tests
	return r
}
func (r *ReqData) SetTest(name string, expression string) *ReqData {
	if r.Tests == nil {
		r.Tests = make(map[string]string)
	}
	r.Tests[name] = expression
	return r
}

func (r *ReqData) SetClient(client *Client) *ReqData {
	r.client = client
	return r
}
func (r *ReqData) Req() *Request {
	return r.req
}

func (r *ReqData) ParseRawURL(rawURL string) error {
	if rawURL == "" {
		return &rDataError{Msg: "rawURL", Err: ErrEmptyValue}
	}

	// if not has http scheme
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	s := strings.SplitN(rawURL, "?", 2)
	// set URL
	r.URL = s[0]
	// Set PathParams
	pathMap := make(map[string]string)
	for _, matched := range pathParamRe.FindAllStringSubmatch(r.URL, -1) {
		name := matched[1]
		pathMap[name] = ""
		if r.Path == nil {
			continue
		}
		if v, ok := r.Path[name]; ok {
			pathMap[name] = v
		}
	}
	r.Path = pathMap
	// set QueryParams
	if len(s) > 1 {
		query, err := url.ParseQuery(s[1])
		if err != nil {
			return err
		}
		r.Query = query
	} else {
		r.Query = map[string][]string{}
	}

	return nil
}

func (r *ReqData) RawURL() string {
	if len(r.Query) > 0 {
		return r.URL + "?" + toQueryString(r.Query)
	}
	return r.URL
}

// instanced the [Request] from [ReqData],if r.client is nil,then create a new client to use
func (r *ReqData) InstancedRequest() (*Request, error) {
	client := r.client
	if client == nil {
		cData := NewClientData()
		c, err := cData.InstancedClient()
		if err != nil {
			return nil, err
		}
		client = c
	}

	req := client.R().
		SetURL(r.URL).
		SetMethod(r.Method).
		SetPathParams(r.Path).
		SetQueryParamsFromValues(r.Query)

	// Filter out empty string keys
	// set raw headers
	delete(r.Header, "")
	maps.Copy(req.Header, r.Header)

	if err := r.convertBody(req); err != nil {
		return nil, err
	}

	if err := r.convertSetting(req); err != nil {
		return nil, err
	}

	// exec PreReq script
	if r.PreReq != "" {
		env := map[string]any{
			// req is a pointer which we can modify its data
			"req":  req,
			"fake": faker.New(),
		}
		code := r.PreReq
		program, err := expr.Compile(code, expr.Env(env))
		if err != nil {
			return nil, err
		}

		if _, err := expr.Run(program, env); err != nil {
			return nil, err
		}

	}

	r.req = req
	return req, nil
}
func (r *ReqData) convertBody(req *Request) error {

	switch b := r.Body; {
	case b.Raw != "":
		req.SetBody(r.Body.Raw)

	case len(b.Multipart) > 0:
		var buf bytes.Buffer
		mWriter := multipart.NewWriter(&buf)
		for _, fd := range b.Multipart {
			isSetContentType := fd.ContentType != ""
			isSetFile := fd.FilePath != ""
			var part io.Writer
			var err error
			// set part
			switch {
			case isSetContentType:
				h := make(textproto.MIMEHeader)
				h.Set(hdrContentTypeKey, fd.ContentType)
				if isSetFile {
					h.Set(hdrContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fd.Name, fd.FileName))
				} else {
					h.Set(hdrContentDisposition, fmt.Sprintf(`form-data; name="%s"`, fd.Name))
				}
				if part, err = mWriter.CreatePart(h); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			case isSetFile:
				if part, err = mWriter.CreateFormFile(fd.Name, fd.FileName); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			default:
				if part, err = mWriter.CreateFormField(fd.Name); err != nil {
					return fmt.Errorf("set part %s failed:%w", fd.Name, err)
				}
			}
			// write content
			if !isSetFile {
				part.Write([]byte(fd.Value))
			} else {
				fileBytes, err := fsx.ReadFile(fd.FilePath)
				if err != nil {
					return fmt.Errorf("read %s failed:%w", fd.FilePath, err)
				}
				part.Write(fileBytes)
			}
		}
		if err := mWriter.Close(); err != nil {
			return fmt.Errorf("close multipart writer failed:%w", err)
		}
		req.SetContentType(mWriter.FormDataContentType()).
			SetBody(buf.Bytes())

	case len(b.Form) > 0:
		req.SetFormDataFromValues(r.Body.Form)

	case b.Binary != "":
		b, err := fsx.ReadFile(r.Body.Binary)
		if err != nil {
			return err
		}
		req.SetBody(b)

	}

	// set Content-Type
	switch ct := r.Body.ContentType; ct {
	case textKey:
		req.SetContentType(plainTextContentType)
	case jsonKey:
		req.SetContentType(jsonContentType)
	case xmlKey:
		req.SetContentType(xmlContentType)
	default:
		if ct != "" {
			req.SetContentType(ct)
		}
	}

	return nil
}
func (r *ReqData) convertSetting(req *Request) error {
	// timeout
	if timeout := r.Setting.Timeout; timeout != "" {
		dur, err := time.ParseDuration(timeout)
		if err != nil {
			return &cDataError{Msg: "timeout", Err: err}
		}
		if dur > 0 {
			req.SetTimeout(dur)
		}
	}
	// resty
	if err := r.convertRetry(req); err != nil {
		return err
	}
	// auth
	if err := r.convertAuth(req); err != nil {
		return err
	}
	// cookie
	if s := r.Setting.CookieString; s != "" {
		cookies, err := http.ParseCookie(s)
		if err != nil {
			return &cDataError{Msg: "cookies", Err: err}
		}
		req.SetCookies(cookies)
	}
	// unescape query
	req.SetQueryParamsUnescape(r.Setting.UnescapeQuery)
	// unescape json
	req.SetJSONEscapeHTML(!r.Setting.UnescapeJson)
	// response body limit
	req.SetResponseBodyLimit(r.Setting.ResponseBodyLimit)
	return nil
}
func (r *ReqData) convertRetry(req *Request) error {
	retry := r.Setting.Retry
	if count := retry.Count; count > 0 {
		req.SetRetryCount(count)
	}
	if wtStr := retry.WaitTime; wtStr != "" {
		wt, err := time.ParseDuration(wtStr)
		if err != nil {
			return &rDataError{Msg: "retry.wait_time", Err: err}
		}
		if wt > 0 {
			req.SetRetryWaitTime(wt)
		}
	}
	if mwtStr := retry.MaxWaitTime; mwtStr != "" {
		mwt, err := time.ParseDuration(mwtStr)
		if err != nil {
			return &rDataError{Msg: "retry.max_wait_time", Err: err}
		}
		if mwt > 0 {
			req.SetRetryMaxWaitTime(mwt)
		}
	}
	if isAllow := retry.IsAllowNonIdempotent; isAllow {
		req.SetRetryAllowNonIdempotent(isAllow)
	}
	return nil
}
func (r *ReqData) convertAuth(req *Request) error {
	switch auth := r.Setting.Auth; {
	case auth.Baisc != AuthCreds{}:
		creds := auth.Baisc
		req.SetBasicAuth(creds.Username, creds.Password)
	case auth.Token != AuthToken{}:
		token := auth.Token
		req.SetAuthToken(token.Value)
		if token.Scheme != "" {
			req.SetAuthScheme(token.Scheme)
		}
		if token.Header != "" {
			req.SetHeaderAuthorizationKey(token.Header)
		}
	}

	return nil
}

// into URL query string without “URL encoded” form
// ("bar=baz&foo=quux&{=}") sorted by key.
func toQueryString(v map[string][]string) string {
	if len(v) == 0 {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, len(v))

	var i int
	for k := range v {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	for _, k := range keys {
		vs := v[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}
