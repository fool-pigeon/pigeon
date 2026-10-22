// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "time"

type TestRes struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Err    error  `json:"err"`
}
type ResultCover struct {
	Success bool `json:"success"`
	// results of test
	Tests []TestRes `json:"tests"`
}
type ResultMeta struct {
	ResultCover
	// this means the tag of status,such as "2xx","4xx","5xx","err","failed"
	Status string `json:"status"`
	// the end-to-end duration between send request and received response
	Duration time.Duration `json:"duration"`
	// this is default nil if not get error in context
	Err error `json:"err"`
}

// As the View Result
type Result struct {
	ResultMeta
	Request Request
	// the raw content of response,e.g. [resty.Response]
	Raw any
}

func (result *Result) SetRaw(Raw any) *Result {
	result.Raw = Raw
	return result
}

func NewResult() Result {
	return Result{}
}

type ResultView[T any] struct {
	ResultMeta
	Request Request
	// the raw content of response,e.g. [httpx.Response]
	Raw T
}

func View[T any](r Result) ResultView[T] {
	v, _ := r.Raw.(T)
	return ResultView[T]{r.ResultMeta, r.Request, v}
}
