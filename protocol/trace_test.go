// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvAnyToString(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"int", int(-42), "-42"},
		{"int8", int8(-8), "-8"},
		{"int16", int16(-16), "-16"},
		{"int32", int32(-32), "-32"},
		{"int64", int64(-64), "-64"},
		{"uint", uint(42), "42"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"uintptr", uintptr(7), "7"},
		{"float32", float32(1.5), "1.5"},
		{"float64", 3.14, "3.14"},
		{"complex64", complex64(complex(1, 2)), "(1+2i)"},
		{"complex128", complex(1, 2), "(1+2i)"},
		{"bytes", []byte("abc"), "abc"},
		{"error", errors.New("boom"), "boom"},
		{"addr", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}, "tcp 127.0.0.1:8080"},
		{"slice_fallback", []string{"a", "b"}, "[a b]"},
		{"map_fallback", map[string]int{"a": 1}, "map[a:1]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, convAnyToString(c.in))
		})
	}
}

func TestNewTrace(t *testing.T) {
	req, res := ulid.Make(), ulid.Make()
	before := time.Now()
	tr := NewTrace(42, req, res, TraceNetwork)
	after := time.Now()

	require.NotNil(t, tr)
	assert.Equal(t, uint64(42), tr.ID)
	assert.Equal(t, TraceNetwork, tr.Kind)
	assert.Equal(t, req, tr.ReqID)
	assert.Equal(t, res, tr.ResID)
	assert.WithinRange(t, tr.Timestamp, before, after)
	assert.Empty(t, tr.Msg)
	assert.Nil(t, tr.Payload)
	assert.Nil(t, tr.Meta)
}

func TestTraceSetters(t *testing.T) {
	newTrace := func() *Trace {
		return NewTrace(1, ulid.Make(), ulid.Make(), TraceNetwork)
	}

	t.Run("set_msg", func(t *testing.T) {
		tr := newTrace()
		assert.Same(t, tr, tr.SetMsg("hi"))
		assert.Equal(t, "hi", tr.Msg)
	})

	t.Run("set_payload", func(t *testing.T) {
		tr := newTrace()
		payload := []byte("p")
		assert.Same(t, tr, tr.SetPayload(payload))
		assert.Equal(t, payload, tr.Payload)
	})

	t.Run("set_meta", func(t *testing.T) {
		tr := newTrace()
		meta := map[string]any{"k": "v"}
		assert.Same(t, tr, tr.SetMeta(meta))
		assert.Equal(t, meta, tr.Meta)
	})
}

func TestTraceToView(t *testing.T) {
	newTrace := func() *Trace {
		return NewTrace(1, ulid.Make(), ulid.Make(), TraceNetwork)
	}

	t.Run("copies_body", func(t *testing.T) {
		tr := newTrace().SetMsg("hello").SetPayload([]byte("p"))
		v := tr.ToView()
		assert.Same(t, &tr.TraceBody, v.TraceBody)
	})

	t.Run("converts_meta", func(t *testing.T) {
		tr := newTrace().SetMeta(map[string]any{
			"str":    "value",
			"num":    42,
			"flag":   true,
			"nested": []string{"a", "b"},
		})
		v := tr.ToView()
		assert.Equal(t, map[string]string{
			"str":    "value",
			"num":    "42",
			"flag":   "true",
			"nested": "[a b]",
		}, v.Meta)
	})

	t.Run("nil_meta_becomes_empty", func(t *testing.T) {
		v := newTrace().ToView()
		require.NotNil(t, v.Meta)
		assert.Empty(t, v.Meta)
	})

	t.Run("does_not_alias_source_meta", func(t *testing.T) {
		tr := newTrace().SetMeta(map[string]any{"k": "v"})
		v := tr.ToView()
		v.Meta["k"] = "changed"
		assert.Equal(t, "v", tr.Meta["k"])
	})
}
