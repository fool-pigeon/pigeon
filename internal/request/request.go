// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package request

import (
	"context"
	"time"
)

type Request interface {
	Run(...context.Context) Result
	RunWithContext(ctx context.Context) Result
	ReqName() string
	ReqProtocol() string
}

type TestRes struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Err    error  `json:"err"`
}

type ResultMeta struct {
	Success bool   `json:"success"`
	ReqName string `json:"reqName"`
	// the mini and most important info of request and will view when cause errror
	ReqInfo string `json:"reqInfo"`
	// e.g. HTTP/1.1,HTTP/2,gPRC,
	Proto string `json:"proto"`
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

func (r *ResultMeta) SetReqName(ReqName string) *ResultMeta {
	r.ReqName = ReqName
	return r
}
func (r *ResultMeta) SetReqInfo(ReqInfo string) *ResultMeta {
	r.ReqInfo = ReqInfo
	return r
}

func (r *ResultMeta) SetProto(proto string) *ResultMeta {
	r.Proto = proto
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

// As the View Result of [Request.Run]
type Result struct {
	ResultMeta
	// the raw content of response,e.g. [httpx.Response]
	Raw any `json:"raw"`
}

type ResultView[T any] struct {
	ResultMeta
	// the raw content of response,e.g. [httpx.Response]
	Raw T `json:"raw"`
}

func (result *Result) SetRaw(Raw any) *Result {
	result.Raw = Raw
	return result
}

func NewResult() Result {
	return Result{}
}
func ErrResult(reqName string, err error) Result {
	return Result{ResultMeta: ResultMeta{ReqName: reqName, Err: err}}
}

type ReqMeta struct {
	Name     string `toml:"name" json:"name"`
	Protocol string `toml:"protocol" comment:"'http'" json:"protocol"`
}

func View[T any](r Result) ResultView[T] {
	v, _ := r.Raw.(T)
	return ResultView[T]{ResultMeta: r.ResultMeta, Raw: v}
}

// ReqName implements [Request].
func (r *ReqMeta) ReqName() string {
	return r.Name
}

// ReqProtocol implements [Request].
func (r *ReqMeta) ReqProtocol() string {
	return r.Protocol
}

const (
	// kinds of protocol
	P_HTTP = "http"
	P_GRPC = "grpc"
	P_WS   = "ws"
	P_GQL  = "gql"
)

var (
	Protocols = []string{P_HTTP, P_GRPC, P_WS, P_GQL}
)
