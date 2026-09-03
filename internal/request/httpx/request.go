package httpx

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/jaswdr/faker/v2"
	"resty.dev/v3"
)

const (
	BodyNone BodyType = iota
	BodyRaw
	BodyJSON
	BodyXML
	BodyMultiForm
	BodyFormUrlEncoded
	BodyBinary
)

type (
	BodyType = int
	// type field of HTTP Multipart Form
	MultipartField struct {
		Name        string   `toml:"name"`
		Values      []string `toml:"values"`
		FileName    string   `toml:"file_name"`
		FilePath    string   `toml:"file_path"`
		ContentType string   `toml:"content_type"`
	}
	Body struct {
		Type BodyType `toml:"type" comment:"0:none,1:raw,2:json,3:xml,4:multiform,5:form-url-encoded,6:binary"`

		Raw            string           `toml:"raw,multiline" comment:"raw data,also json,xml"`
		MultiForm      []MultipartField `toml:"multiform,multiline" comment:"list of form-data"`
		FormUrlEncoded url.Values       `toml:"form_url_encoded,multiline" comment:"key-values of form-url-encoded"`
		BinaryPath     string           `toml:"binary_path,multiline" comment:"the file/binary path"`
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

	RequestData struct {
		request.RequestMeta
		Method      string            `toml:"mothod"`
		URL         string            `toml:"url"`
		PathParams  map[string]string `toml:"path_params,omitempty,multiline"`
		QueryParams url.Values        `toml:"query_params,omitempty,multiline"`
		Headers     http.Header       `toml:"headers,omitempty,multiline"`
		Body        Body              `toml:"body,omitempty"`
		Desc        string            `toml:"desc,omitempty,multiline"`
		Setting     Setting           `toml:"setting,omitempty"`

		// the expr script before send request,could set random data for request
		PreReq string `toml:"pre_req,omitempty,multiline"`

		// the exprs whose the result value type is bool
		// key is test's name
		// value is expr expression
		Tests map[string]string `toml:"tests,omitempty,multiline"`

		client *Client
	}
)

// Run implements [request.Request].
func (r *RequestData) Run() request.Result {
	// send request
	res, err := r.Send()
	if err != nil {
		return request.ErrResult(r.Name, err)
	}
	// set raw response
	result := request.NewResult()
	result.SetReqName(r.ReqName()).SetRaw(res).SetStatus(res.Status())

	isSuccess := res.StatusCode() >= 200 && res.StatusCode() < 300

	// run tests
	if len(r.Tests) != 0 {
		for name, code := range r.Tests {
			tRes := request.TestRes{Name: name}
			env := map[string]any{
				"res":   res,
				"faker": faker.New(),
			}

			program, err := expr.Compile(code, expr.Env(env), expr.AsBool())
			if err != nil {
				tRes.Err = err
				isSuccess = false
				result.AppendTestRes(tRes)
				continue
			}
			output, err := expr.Run(program, env)
			if err != nil {
				tRes.Err = err
				isSuccess = false
				result.AppendTestRes(tRes)
				continue
			}

			tRes.Passed = output.(bool)
			isSuccess = isSuccess && tRes.Passed

			result.AppendTestRes(tRes)
		}
	}
	result.SetSuccess(isSuccess)

	return result
}

func NewRequestData(requestName string) RequestData {
	return RequestData{RequestMeta: request.RequestMeta{
		Name:     requestName,
		Protocol: request.P_HTTP,
	}}
}

// Setters

func (r *RequestData) SetMethod(m string) *RequestData {
	r.Method = m
	return r
}
func (r *RequestData) SetURL(url string) *RequestData {
	r.URL = url
	return r
}
func (r *RequestData) SetPathParams(params map[string]string) *RequestData {
	r.PathParams = params
	return r
}
func (r *RequestData) SetQueryParams(params url.Values) *RequestData {
	r.QueryParams = params
	return r
}
func (r *RequestData) SetHeaders(headers http.Header) *RequestData {
	r.Headers = headers
	return r
}
func (r *RequestData) SetBody(body Body) *RequestData {
	r.Body = body
	return r
}
func (r *RequestData) SetDesc(desc string) *RequestData {
	r.Desc = desc
	return r
}
func (r *RequestData) SetSetting(setting Setting) *RequestData {
	r.Setting = setting
	return r
}
func (r *RequestData) SetTimeout(timeStr string) *RequestData {
	r.Setting.Timeout = timeStr
	return r
}
func (r *RequestData) SetRetry(retry Retry) *RequestData {
	r.Setting.Retry = retry
	return r
}
func (r *RequestData) SetAuth(auth Auth) *RequestData {
	r.Setting.Auth = auth
	return r
}
func (r *RequestData) SetCookieString(cookieStr string) *RequestData {
	r.Setting.CookieString = cookieStr
	return r
}
func (r *RequestData) SetUnescapeQuery(unescape bool) *RequestData {
	r.Setting.UnescapeQuery = unescape
	return r
}
func (r *RequestData) SetUnescapeJson(unescape bool) *RequestData {
	r.Setting.UnescapeJson = unescape
	return r
}
func (r *RequestData) SetResponseBodyLimit(limit int64) *RequestData {
	r.Setting.ResponseBodyLimit = limit
	return r
}
func (r *RequestData) SetClient(client *Client) *RequestData {
	r.client = client
	return r
}

func (r *RequestData) ParseRawURL(rawURL string) error {
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
		if r.PathParams == nil {
			continue
		}
		if v, ok := r.PathParams[name]; ok {
			pathMap[name] = v
		}
	}
	r.PathParams = pathMap
	// set QueryParams
	if len(s) > 1 {
		query, err := url.ParseQuery(s[1])
		if err != nil {
			return err
		}
		r.QueryParams = query
	} else {
		r.QueryParams = map[string][]string{}
	}

	return nil
}

func (r *RequestData) RawURL() string {
	if len(r.QueryParams) > 0 {
		return r.URL + "?" + toQueryString(r.QueryParams)
	}
	return r.URL
}

func (r *RequestData) Send() (*Response, error) {
	req, err := r.InstancedRequest()
	if err != nil {
		return nil, err
	}
	return req.Send()
}

// instanced the [Request] from [RequestData],if r.client is nil,then create a new client to use
func (r *RequestData) InstancedRequest() (*Request, error) {
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
		SetPathParams(r.PathParams).
		SetQueryParamsFromValues(r.QueryParams).
		SetHeaderMultiValues(r.Headers)

	if err := r.convertBody(req); err != nil {
		return nil, err
	}

	if err := r.convertSetting(req); err != nil {
		return nil, err
	}

	// exec PreReq script
	if r.PreReq != "" {
		env := map[string]any{
			"req":   &req,
			"faker": faker.New(),
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

	return req, nil
}
func (r *RequestData) convertBody(req *Request) error {

	bt := r.Body.Type
	switch bt {
	default:
		return &rDataError{Msg: "body_type", Err: ErrInvalidValue}

	case BodyNone:
		// do nothing
	case BodyJSON:
		fallthrough
	case BodyXML:
		fallthrough
	case BodyRaw:
		req.SetBody(r.Body.Raw)

	case BodyMultiForm:
		fields := make([]*resty.MultipartField, 0, len(r.Body.MultiForm))
		for _, v := range r.Body.MultiForm {
			fields = append(fields, &resty.MultipartField{
				Name:        v.Name,
				Values:      v.Values,
				FileName:    v.FileName,
				FilePath:    v.FilePath,
				ContentType: v.ContentType,
			})
		}
		req.SetMultipartFields(fields...)
	case BodyFormUrlEncoded:
		req.SetFormDataFromValues(r.Body.FormUrlEncoded)
	case BodyBinary:
		b, err := os.ReadFile(r.Body.BinaryPath)
		if err != nil {
			return err
		}
		req.SetBody(b)

	}

	// set Content-Type for json and xml type
	switch bt {
	case BodyJSON:
		req.SetContentType("application/json")
	case BodyXML:
		req.SetContentType("application/xml")

	}

	return nil
}
func (r *RequestData) convertSetting(req *Request) error {
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
func (r *RequestData) convertRetry(req *Request) error {
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
	if isAllow := retry.IsAllowNonIdempotent; isAllow != false {
		req.SetRetryAllowNonIdempotent(isAllow)
	}
	return nil
}
func (r *RequestData) convertAuth(req *Request) error {
	switch auth := r.Setting.Auth; auth.Type {
	default:
		return &rDataError{Msg: "auth_type", Err: ErrInvalidValue}

	case AuthNone:
		// nothing
	case AuthTypeBasic:
		creds := auth.Creds
		req.SetBasicAuth(creds.Username, creds.Password)
	case AuthTypeToken:
		data := auth.Token
		req.SetAuthToken(data.Token)
		if data.Scheme != "" {
			req.SetAuthScheme(data.Scheme)
		}
		if data.Header != "" {
			req.SetHeaderAuthorizationKey(data.Header)
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
