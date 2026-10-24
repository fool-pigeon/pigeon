// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type TraceKind string

const (
	TraceNetwork  TraceKind = "network"
	TraceReceived TraceKind = "received"
)

type TraceHandler func(*Trace)

type TraceBody struct {
	ID   uint64    `json:"id"`
	Kind TraceKind `json:"kind"`

	ReqID     ulid.ULID `json:"reqID"`
	ResID     ulid.ULID `json:"resID"`
	Timestamp time.Time `json:"timestamp"`

	Msg     string `json:"msg"`
	Payload []byte `json:"payload"`
}

type Trace struct {
	TraceBody
	Meta map[string]any
}
type TraceView struct {
	*TraceBody
	Meta map[string]string `json:"meta"`
}

func NewTrace(id uint64, reqID ulid.ULID, resID ulid.ULID, kind TraceKind) *Trace {
	return &Trace{TraceBody: TraceBody{ID: id, ReqID: reqID, ResID: resID, Kind: kind, Timestamp: time.Now()}}
}
func (t *Trace) SetMsg(msg string) *Trace {
	t.Msg = msg
	return t
}
func (t *Trace) SetMeta(meta map[string]any) *Trace {
	t.Meta = meta
	return t
}
func (t *Trace) SetPayload(payload []byte) *Trace {
	t.Payload = payload
	return t
}
func (t *Trace) ToView() TraceView {
	meta := make(map[string]string, len(t.Meta))
	for k, v := range t.Meta {
		if str, ok := v.(string); ok {
			meta[k] = str
		} else {
			meta[k] = convAnyToString(v)
		}
	}
	return TraceView{&t.TraceBody, meta}
}
func convAnyToString(a any) string {
	switch v := a.(type) {
	default:
		return fmt.Sprint(v)
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32: // include rune
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8: // include byte
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uintptr:
		return strconv.FormatUint(uint64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case complex64:
		return strconv.FormatComplex(complex128(v), 'g', -1, 64)
	case complex128:
		return strconv.FormatComplex(v, 'g', -1, 128)
	case []byte:
		return string(v)

	case error:
		return v.Error()

	case net.Addr:
		var sb strings.Builder
		sb.WriteString(v.Network())
		sb.WriteRune(' ')
		sb.WriteString(v.String())
		return sb.String()
	}
}
