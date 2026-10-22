// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type TraceKind string

const (
	TraceNetwork TraceKind = "network"
)

type TraceHandler func(Trace)
type Trace struct {
	ID   uint64
	Kind TraceKind

	ReqID     ulid.ULID
	Timestamp time.Time

	Msg     string
	Meta    map[string]string
	Payload []byte
}

func NewTrace(id uint64, reqID ulid.ULID, kind TraceKind) *Trace {
	return &Trace{ID: id, ReqID: reqID, Kind: kind, Timestamp: time.Now()}
}
func (t *Trace) SetMsg(msg string) *Trace {
	t.Msg = msg
	return t
}
func (t *Trace) SetMeta(meta map[string]string) *Trace {
	t.Meta = meta
	return t
}
func (t *Trace) SetPayload(payload []byte) *Trace {
	t.Payload = payload
	return t
}
