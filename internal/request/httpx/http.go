package httpx

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/homoarea/pigeon/internal/request"
	"resty.dev/v3"
)

type Response = *resty.Response

type Settings struct {
	EncodeUrl     bool `toml:"encode_url,omitempty"`
	Timeout       uint `toml:"timeout,omitempty"`
	AutoRedirects bool `toml:"auto_redirects"`
	MaxRedirects  uint `toml:"max_redirects,omitempty"`
}
type Request struct {
	request.RequestMeta
	Method string `toml:"mothod"`
	// like http://localhost:8080/v1/{userId}/post
	Url         string              `toml:"url"`
	PathParams  map[string]string   `toml:"path_params,omitempty,multiline"`
	QueryParams map[string][]string `toml:"query_params,omitempty,multiline"`
	Headers     map[string][]string `toml:"headers,omitempty,multiline"`
	Body        Body                `toml:"body,omitempty"`
	Settings    Settings            `toml:"settings"`
	Desc        string              `toml:"desc,omitempty,multiline"`
}
type Body struct {
	Type string `toml:"type,omitempty"`
	Data any    `toml:"data,omitempty,multiline"`
}
type FileField struct {
	FileName    string `toml:"file_name,omitempty"`
	FilePath    string `toml:"file_path,omitempty"`
	ContentType string `toml:"content_type,omitempty"`
}
type MultipartField struct {
	Name   string   `toml:"name,omitempty"`
	Values []string `toml:"values,omitempty"`
	FileField
}

const (
	// HTTP body type
	// raw: string
	BodyJSON = "json"
	BodyXML  = "xml"
	BodyText = "text"
	BodyRaw  = "raw"
	// []MultipartField
	BodyMultiForm = "multipart/form-data"
	//[][2]string
	BodyFormUrlEncoded = "x-www-form-urlencoded"
	BodyBinary         = "file/binary"
)

var (
	Methods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodPatch,
		http.MethodTrace,
	}
	BodyTypes = []string{}
)

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
		return errors.New("rawURL is empty.")
	}
	if !strings.HasPrefix(rawURL, "http://") || !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	r.Url = u.Scheme + "://" + u.Host + u.RawPath
	r.QueryParams = u.Query()
	return nil
}
func (r *Request) RawURL() string {
	if r.QueryParams != nil {
		return r.Url + "?" + url.Values(r.QueryParams).Encode()
	}
	return r.Url
}

func (r *Request) Send() (Response, error) {
	client := resty.New()
	defer client.Close()
	req := client.R().
		SetURL(r.Url).
		SetMethod(r.Method)
	if r.QueryParams != nil {
		req.SetQueryParamsFromValues(r.QueryParams)
	}
	if r.PathParams != nil {
		req.SetPathParams(r.PathParams)
	}
	if r.Headers != nil {
		req.SetHeaderMultiValues(r.Headers)
	}
	if bt := r.Body.Type; len(bt) > 0 {
		// TODO: impl http's body handler
		switch bt {
		case BodyJSON, BodyXML, BodyText, BodyRaw:
			req.SetBody(r.Body)
		default:
			return nil, errors.New("the body type handler is not impl yet.")
		}
	}
	res, err := req.Send()
	return res, err
}
