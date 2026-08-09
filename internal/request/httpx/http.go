package httpx

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/homoarea/pigeon/internal/request"
	"resty.dev/v3"
)

type ErrUnkownBodyType struct {
	BodyType string
}

func (e *ErrUnkownBodyType) Error() string {
	return "the body type " + e.BodyType + " is unkown."
}

var (
	ErrEmptyRawURL = errors.New("RawURL is empty.")
	ErrEmptyHost   = errors.New("Host is empty.")
)

const (
	// HTTP Body Types
	// body is type string
	BodyRaw = "raw"
	// body is type string
	BodyJSON = "application/json"
	// body is type string
	BodyXML = "xml"
	// body is []MultipartField
	BodyMultiForm = "multipart/form-data"
	// body is map[string][]string
	BodyFormUrlEncoded = "x-www-form-urlencoded"
	// body is map[string]string
	BodyBinary = "file/binary"
)

const (
	MethodGet     = http.MethodGet
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodDelete  = http.MethodDelete
	MethodHead    = http.MethodHead
	MethodOptions = http.MethodOptions
	MethodConnect = http.MethodConnect
	MethodPatch   = http.MethodPatch
	MethodTrace   = http.MethodTrace

	// this field is type string
	FieldURL = "URL"
	// this field is type string
	FieldMethod = "Method"
	// this field is type map[string]string
	FieldPathParams = "PathParams"
	// this field is type map[string][]string
	FieldQueryParams = "QueryParams"
	// this field is type map[string][]string
	FieldHeaders = "Headers"
	// this field is type httpx.Body
	FieldBody = "Body"
	// this field is type string
	FieldDesc = "Description"
	// this field is type httpx.Settings
	FieldSettings = "Settings"
)

var (
	Methods   = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}
	BodyTypes = []string{BodyRaw, BodyJSON, BodyXML, BodyMultiForm, BodyFormUrlEncoded, BodyBinary}
	// FieldNames of Request
	Fields = []string{FieldURL, FieldMethod, FieldPathParams, FieldQueryParams, FieldHeaders, FieldBody, FieldDesc, FieldSettings}

	// regex pattern
	pathParamRe = regexp.MustCompile(`(?:^|/){([^${}/]+)}`)
)

type Response = resty.Response

type Settings struct {
	EncodeUrl     bool `toml:"encode_url,omitempty"`
	Timeout       uint `toml:"timeout,omitempty"`
	AutoRedirects bool `toml:"auto_redirects"`
	MaxRedirects  uint `toml:"max_redirects,omitempty"`
}
type Request struct {
	request.RequestMeta
	Method string `toml:"mothod"`
	// URL without query string, like http://localhost:8080/v1/{userId}/post
	Url         string              `toml:"url"`
	PathParams  map[string]string   `toml:"path_params,omitempty,multiline"`
	QueryParams map[string][]string `toml:"query_params,omitempty,multiline"`
	Headers     map[string][]string `toml:"headers,omitempty,multiline"`
	Body        Body                `toml:"body,omitempty"`
	Settings    Settings            `toml:"settings"`
	Desc        string              `toml:"desc,omitempty,multiline"`
}
type Body struct {
	Type string   `toml:"type,omitempty"`
	Data BodyData `toml:"data,omitempty,multiline"`
}
type BodyData struct {
	Raw            string              `toml:"raw,omitempty,multiline"`
	MultiForm      []MultipartField    `toml:"multi_form,omitempty,multiline"`
	FormUrlEncoded map[string][]string `toml:"form_url_encoded,omitempty,multiline"`
	BinaryPath     string              `toml:"binary_path,omitempty,multiline"`
}

// type field of HTTP Multipart Form
type MultipartField struct {
	Name        string   `toml:"name,omitempty"`
	Values      []string `toml:"values,omitempty"`
	FileName    string   `toml:"file_name,omitempty"`
	FilePath    string   `toml:"file_path,omitempty"`
	ContentType string   `toml:"content_type,omitempty"`
}

func mergeMapValues(dist *map[string][]string, src *map[string][]string) {
	for k, v := range *src {
		if (*dist)[k] != nil {
			(*dist)[k] = append((*dist)[k], v...)
		} else {
			(*dist)[k] = (*src)[k]
		}
	}
}

func (r *Request) AddQueryParams(values map[string][]string) *Request {
	mergeMapValues(&r.QueryParams, &values)
	return r
}

func (r *Request) AddHeaders(values map[string][]string) *Request {
	mergeMapValues(&r.Headers, &values)
	return r
}

func (r *Request) ParseRawURL(rawURL string) error {
	if rawURL == "" {
		return ErrEmptyRawURL
	}

	// TODO: add env swap system here
	// here

	// if not has http scheme
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	s := strings.SplitN(rawURL, "?", 2)
	// set URL
	r.Url = s[0]
	// Set PathParams
	if r.PathParams == nil {
		r.PathParams = make(map[string]string)
	}
	for _, matched := range pathParamRe.FindAllStringSubmatch(r.Url, -1) {
		name := matched[1]
		if _, ok := r.PathParams[name]; !ok {
			r.PathParams[name] = ""
		}
	}
	// set QueryParams
	if len(s) > 1 {
		query, err := url.ParseQuery(s[1])
		if err != nil {
			return err
		}
		r.QueryParams = query
	}

	return nil
}

func (r *Request) RawURL() string {
	if len(r.QueryParams) > 0 {
		return r.Url + "?" + toQueryString(r.QueryParams)
	}
	return r.Url
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

func (r *Request) Send() (*Response, error) {
	client := resty.New()
	defer client.Close()
	req := client.R().
		SetURL(r.Url).
		SetMethod(r.Method).
		SetPathParams(r.PathParams).
		SetQueryParamsFromValues(r.QueryParams).
		SetHeaderMultiValues(r.Headers)

	if bt := r.Body.Type; len(bt) > 0 {
		switch bt {
		case BodyJSON:
			req.SetBody(r.Body.Data.Raw).SetContentType("application/json")
		case BodyXML:
			req.SetBody(r.Body.Data.Raw).SetContentType("application/xml")
		case BodyRaw:
			req.SetBody(r.Body.Data.Raw)
		case BodyMultiForm:
			fields := make([]*resty.MultipartField, 0, len(r.Body.Data.MultiForm))
			for _, v := range r.Body.Data.MultiForm {
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
			req.SetFormDataFromValues(r.Body.Data.FormUrlEncoded)
		case BodyBinary:
			b, err := os.ReadFile(r.Body.Data.BinaryPath)
			if err != nil {
				return nil, err
			}
			req.SetBody(b)
		default:
			return nil, &ErrUnkownBodyType{BodyType: bt}
		}
	}
	res, err := req.Send()
	return res, err
}
