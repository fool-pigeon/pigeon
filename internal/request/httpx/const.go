package httpx

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/homoarea/pigeon/internal/request"
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
)

var (
	ErrEmptyValue   = errors.New("the value is empty.")
	ErrInvalidValue = errors.New("the value is invalid.")
)
var (
	Methods = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}
	// regex pattern
	pathParamRe = regexp.MustCompile(`(?:^|/){([^${}/]+)}`)

	ExampleBody = Body{
		Type:           BodyNone,
		Raw:            "",
		MultiForm:      []MultipartField{{Name: "", Values: []string{""}, FileName: "", FilePath: "", ContentType: ""}},
		FormUrlEncoded: map[string][]string{"name": {"value"}},
	}
	ExampleAuth    = Auth{}
	ExampleSetting = Setting{
		Auth: ExampleAuth,
	}
	ExampleClientData  = CommentedClientData{}
	ExampleRequestData = CommentedRequestData{
		RequestMeta: request.RequestMeta{Protocol: request.P_HTTP},
		Body:        ExampleBody,
	}
)
