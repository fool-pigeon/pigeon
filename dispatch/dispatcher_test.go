// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dispatch

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http/httptrace"
	"testing"
	"time"

	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAddr string

func (a stubAddr) Network() string { return "tcp" }
func (a stubAddr) String() string  { return string(a) }

// stubConn satisfies net.Conn methods used by GotConn.
type stubConn struct{ net.Conn }

func (stubConn) LocalAddr() net.Addr  { return stubAddr("127.0.0.1:1") }
func (stubConn) RemoteAddr() net.Addr { return stubAddr("127.0.0.1:2") }

func capture(t *testing.T, reqID ulid.ULID, resID ulid.ULID, call func(*httptrace.ClientTrace)) []*protocol.Trace {
	t.Helper()
	var got []*protocol.Trace
	tr := getHTTPTrace(&uid.Atomic64Gen{}, reqID, resID, func(ev *protocol.Trace) {
		got = append(got, ev)
	})
	call(tr)
	return got
}

func TestGetHTTPTrace(t *testing.T) {
	t.Run("all hooks set", func(t *testing.T) {
		assert := assert.New(t)
		tr := getHTTPTrace(&uid.Atomic64Gen{}, ulid.Make(), ulid.Make(), func(*protocol.Trace) {})

		assert.NotNil(tr.GetConn, "GetConn")
		assert.NotNil(tr.GotConn, "GotConn")
		assert.NotNil(tr.PutIdleConn, "PutIdleConn")
		assert.NotNil(tr.GotFirstResponseByte, "GotFirstResponseByte")
		assert.NotNil(tr.DNSStart, "DNSStart")
		assert.NotNil(tr.DNSDone, "DNSDone")
		assert.NotNil(tr.ConnectStart, "ConnectStart")
		assert.NotNil(tr.ConnectDone, "ConnectDone")
		assert.NotNil(tr.TLSHandshakeStart, "TLSHandshakeStart")
		assert.NotNil(tr.TLSHandshakeDone, "TLSHandshakeDone")
		assert.NotNil(tr.WroteHeaderField, "WroteHeaderField")
		assert.NotNil(tr.WroteHeaders, "WroteHeaders")
		assert.NotNil(tr.WroteRequest, "WroteRequest")
	})

	reqID := ulid.Make()
	resID := ulid.Make()
	tests := []struct {
		name    string
		call    func(*httptrace.ClientTrace)
		wantMsg string
		check   func(*testing.T, *protocol.Trace)
	}{
		{
			name:    "GetConn",
			call:    func(c *httptrace.ClientTrace) { c.GetConn("example.com:443") },
			wantMsg: "GetConn",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Equal(t, "example.com:443", tr.Meta["hostPort"])
			},
		},
		{
			name: "GotConn",
			call: func(c *httptrace.ClientTrace) {
				c.GotConn(httptrace.GotConnInfo{
					Conn: stubConn{}, Reused: true, WasIdle: true,
					IdleTime: 5 * time.Second,
				})
			},
			wantMsg: "GotConn",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal(true, tr.Meta["reused"])
				assert.Equal(true, tr.Meta["wasIdle"])
				assert.Equal(5*time.Second, tr.Meta["idleTime"])
				local, ok := tr.Meta["local"].(net.Addr)
				require.True(t, ok, "local is net.Addr")
				assert.Equal("127.0.0.1:1", local.String())
				remote, ok := tr.Meta["remote"].(net.Addr)
				require.True(t, ok, "remote is net.Addr")
				assert.Equal("127.0.0.1:2", remote.String())
			},
		},
		{
			name:    "PutIdleConn",
			call:    func(c *httptrace.ClientTrace) { c.PutIdleConn(nil) },
			wantMsg: "PutIdleConn",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Nil(t, tr.Meta["err"])
			},
		},
		{
			name:    "GotFirstResponseByte",
			call:    func(c *httptrace.ClientTrace) { c.GotFirstResponseByte() },
			wantMsg: "GotFirstResponseByte",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Nil(t, tr.Meta, "Meta should be nil")
			},
		},
		{
			name:    "DNSStart",
			call:    func(c *httptrace.ClientTrace) { c.DNSStart(httptrace.DNSStartInfo{Host: "example.com"}) },
			wantMsg: "DNSStart",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Equal(t, "example.com", tr.Meta["host"])
			},
		},
		{
			name: "DNSDone",
			call: func(c *httptrace.ClientTrace) {
				c.DNSDone(httptrace.DNSDoneInfo{
					Addrs:     []net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}},
					Coalesced: true,
				})
			},
			wantMsg: "DNSDone",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal([]net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}}, tr.Meta["addrs"])
				assert.Equal(true, tr.Meta["coalesced"])
			},
		},
		{
			name:    "ConnectStart",
			call:    func(c *httptrace.ClientTrace) { c.ConnectStart("tcp", "127.0.0.1:80") },
			wantMsg: "ConnectStart",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal("tcp", tr.Meta["network"])
				assert.Equal("127.0.0.1:80", tr.Meta["addr"])
			},
		},
		{
			name:    "ConnectDone",
			call:    func(c *httptrace.ClientTrace) { c.ConnectDone("tcp", "127.0.0.1:80", nil) },
			wantMsg: "ConnectDone",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal("127.0.0.1:80", tr.Meta["addr"])
				assert.Nil(tr.Meta["err"])
			},
		},
		{
			name:    "TLSHandshakeStart",
			call:    func(c *httptrace.ClientTrace) { c.TLSHandshakeStart() },
			wantMsg: "TLSHandshakeStart",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Nil(t, tr.Meta, "Meta should be nil")
			},
		},
		{
			name: "TLSHandshakeDone/success",
			call: func(c *httptrace.ClientTrace) {
				c.TLSHandshakeDone(tls.ConnectionState{
					Version:            tls.VersionTLS13,
					HandshakeComplete:  true,
					DidResume:          true,
					CipherSuite:        tls.TLS_AES_128_GCM_SHA256,
					NegotiatedProtocol: "h2",
					ServerName:         "example.com",
				}, nil)
			},
			wantMsg: "TLSHandshakeDone",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal(uint16(tls.VersionTLS13), tr.Meta["version"])
				assert.Equal(true, tr.Meta["handshakeComplete"])
				assert.Equal(true, tr.Meta["didResume"])
				assert.Equal(uint16(tls.TLS_AES_128_GCM_SHA256), tr.Meta["cipherSuite"])
				assert.Equal("h2", tr.Meta["negotiatedProtocol"])
				assert.Equal("example.com", tr.Meta["serverName"])
			},
		},
		{
			name: "TLSHandshakeDone/error",
			call: func(c *httptrace.ClientTrace) {
				c.TLSHandshakeDone(tls.ConnectionState{}, errors.New("boom"))
			},
			wantMsg: "TLSHandshakeDone",
			check: func(t *testing.T, tr *protocol.Trace) {
				require.Len(t, tr.Meta, 1, "err-only path")
				err, ok := tr.Meta["err"].(error)
				require.True(t, ok)
				assert.Equal(t, "boom", err.Error())
			},
		},
		{
			name: "WroteHeaderField",
			call: func(c *httptrace.ClientTrace) {
				c.WroteHeaderField("Accept", []string{"application/json"})
			},
			wantMsg: "WroteHeaderField",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert := assert.New(t)
				assert.Equal("Accept", tr.Meta["key"])
				assert.Equal([]string{"application/json"}, tr.Meta["value"])
			},
		},
		{
			name:    "WroteHeaders",
			call:    func(c *httptrace.ClientTrace) { c.WroteHeaders() },
			wantMsg: "WroteHeaders",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Nil(t, tr.Meta, "Meta should be nil")
			},
		},
		{
			name:    "WroteRequest",
			call:    func(c *httptrace.ClientTrace) { c.WroteRequest(httptrace.WroteRequestInfo{}) },
			wantMsg: "WroteRequest",
			check: func(t *testing.T, tr *protocol.Trace) {
				assert.Nil(t, tr.Meta["err"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capture(t, reqID, resID, tt.call)
			require.Len(t, got, 1)
			tr := got[0]

			assert := assert.New(t)
			assert.Equal(protocol.TraceNetwork, tr.Kind)
			assert.Equal(reqID, tr.ReqID)
			assert.Equal(tt.wantMsg, tr.Msg)
			assert.False(tr.Timestamp.IsZero(), "Timestamp")

			tt.check(t, tr)
		})
	}

	t.Run("trace IDs increase monotonically", func(t *testing.T) {
		got := capture(t, reqID, resID, func(c *httptrace.ClientTrace) {
			c.GetConn("a:1")
			c.GetConn("b:2")
			c.GetConn("c:3")
		})
		require.Len(t, got, 3)
		assert.Greater(t, got[1].ID, got[0].ID)
		assert.Greater(t, got[2].ID, got[1].ID)
	})
}

func TestSend(t *testing.T) {
	reqID := ulid.Make()
	resID := ulid.Make()
	t.Run("unsupported protocol", func(t *testing.T) {
		d := New(nil)
		req := &protocol.Request{ReqMeta: protocol.ReqMeta{ID: reqID, Protocol: "bogus"}}
		res, err := d.Send(context.Background(), req, resID, &uid.Atomic64Gen{}, nil)
		require.ErrorIs(t, err, protocol.ErrNotSupport)
		assert.Nil(t, res)
	})

	t.Run("unsupported protocol emits nothing", func(t *testing.T) {
		d := New(nil)
		req := &protocol.Request{ReqMeta: protocol.ReqMeta{ID: reqID, Protocol: "bogus"}}
		var got []*protocol.Trace
		_, err := d.Send(context.Background(), req, resID, &uid.Atomic64Gen{}, func(tr *protocol.Trace) {
			got = append(got, tr)
		})
		require.ErrorIs(t, err, protocol.ErrNotSupport)
		assert.Empty(t, got)
	})
}
