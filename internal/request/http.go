package request

import (
	"net/http"
	"net/url"

	"resty.dev/v3"
)

type HttpSettings struct {
	EncodeUrl     bool `toml:"encode_url,omitempty"`
	Timeout       uint `toml:"timeout,omitempty"`
	AutoRedirects bool `toml:"auto_redirects"`
	MaxRedirects  uint `toml:"max_redirects,omitempty"`
}
type HttpRequest struct {
	RequestMeta
	Method      string       `toml:"mothod"`
	Url         string       `toml:"url"`
	QueryParams [][2]string  `toml:"query_params,omitempty,multiline"`
	PathParams  [][2]string  `toml:"path_params,omitempty,multiline"`
	Headers     [][2]string  `toml:"headers,omitempty,multiline"`
	BodyType    string       `toml:"body_type,omitempty"`
	Body        string       `toml:"body,omitempty,multiline"`
	Settings    HttpSettings `toml:"settings"`
	Docs        string       `toml:"docs,omitempty,multiline"`
}

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
)

func params2Map(params [][2]string) map[string]string {
	res := make(map[string]string)
	for _, v := range params {
		res[v[0]] = v[1]
	}
	return res
}

func params2Values(params [][2]string) url.Values {
	values := make(url.Values)
	for _, v := range params {
		values.Add(v[0], v[1])
	}
	return values
}

func (r *HttpRequest) Send() (*resty.Response, error) {
	client := resty.New()
	defer client.Close()
	res, err := client.R().
		SetMethod(r.Method).
		SetURL(r.Url).
		SetQueryParamsFromValues(params2Values(r.QueryParams)).
		SetPathParams(params2Map(r.PathParams)).
		SetHeaders(params2Map(r.Headers)).
		SetBody(r.Body).
		Send()

	return res, err
}
