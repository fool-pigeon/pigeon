package httpx

import (
	"errors"
	"net/http"

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
	Docs        string              `toml:"docs,omitempty,multiline"`
}
type Body struct {
	Type string `toml:"type,omitempty"`
	Data any    `toml:"data,omitempty,multiline"`
}
type MultipartField struct {
	Name        string   `toml:"name,omitempty"`
	Values      []string `toml:"values,omitempty"`
	FileName    string   `toml:"file_name,omitempty"`
	FilePath    string   `toml:"file_path,omitempty"`
	ContentType string   `toml:"content_type,omitempty"`
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
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
	BodyTypes = []string{}
)

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
		// TODO: impl handle the body
		switch bt {
		case BodyJSON, BodyXML, BodyText, BodyRaw:
			req.SetBody(r.Body)
		case BodyMultiForm:
			return nil, errors.New("the multiform not impl yet.")
		}
	}
	res, err := req.Send()
	return res, err
}
