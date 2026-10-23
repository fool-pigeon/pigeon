// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type TestRes struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	ErrStr string `json:"err"`
}
type ResultCover struct {
	Success bool `json:"success"`
	// results of test
	Tests []TestRes `json:"tests"`
}
type ResMeta struct {
	ResultCover

	ID ulid.ULID

	// this means the tag of status,such as "2xx","4xx","5xx","err","failed"
	Status string `json:"status"`
	// the end-to-end duration between send request and received response
	Duration time.Duration `json:"duration"`

	Request Request
}

// As the View Result
type Result struct {
	ResMeta
	// the raw content of response,e.g. [resty.Response]
	Raw any
}
