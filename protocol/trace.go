// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "time"

type TraceKind string

const (
	TraceNetwork TraceKind = "network"
)

type Trace struct {
	Meta      map[string]string
	Payload   []byte
	Timestamp time.Time
}
type TraceHandler func(Trace)
