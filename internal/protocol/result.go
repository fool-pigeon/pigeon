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

type ResultMeta struct {
	Success bool `json:"success"`
	// this means the tag of status,such as "2xx","4xx","5xx","err","failed"
	Status string `json:"status"`
	// the end-to-end duration between send request and received response
	Duration time.Duration `json:"duration"`
	// results of test
	Tests []TestRes `json:"tests"`
	// this is default nil if not get error in context
	Err error `json:"err"`
}

func (r *ResultMeta) SetSuccess(Success bool) *ResultMeta {
	r.Success = Success
	return r
}

func (r *ResultMeta) SetStatus(Status string) *ResultMeta {
	r.Status = Status
	return r
}

func (r *ResultMeta) SetDuration(duration time.Duration) *ResultMeta {
	r.Duration = duration
	return r
}

func (r *ResultMeta) SetTests(Tests []TestRes) *ResultMeta {
	r.Tests = Tests
	return r
}
func (r *ResultMeta) SetErr(Err error) *ResultMeta {
	r.Err = Err
	return r
}

func (r *ResultMeta) AppendTestRes(tRes TestRes) {
	r.Tests = append(r.Tests, tRes)
}

// As the View Result of [Runner.Run]
type Result struct {
	ResultMeta
	Request Request
	// the raw content of response,e.g. [httpx.Response]
	Raw any `json:"raw"`
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
	// the raw content of response,e.g. [httpx.Response]
	Raw T `json:"raw"`
}

func View[T any](r Result) ResultView[T] {
	v, _ := r.Raw.(T)
	return ResultView[T]{ResultMeta: r.ResultMeta, Raw: v}
}

type TraceKind string

const (
	TraceNetwork TraceKind = "network"
)

type Trace struct {
}
