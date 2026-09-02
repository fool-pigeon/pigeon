package httpx

import (
	"errors"
	"net/http"
	"regexp"
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
)
