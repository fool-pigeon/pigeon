// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dispatch

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"

	"github.com/fool-pigeon/pigeon/pgrest"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/oklog/ulid/v2"
)

type Dispatcher struct {
	client *protocol.ActualClient
}

func New(client *protocol.ActualClient) *Dispatcher {
	return &Dispatcher{client}
}

// Send implements [protocol.Sender].
func (d *Dispatcher) Send(ctx context.Context, req *protocol.Request, idGen *uid.AtomicGen, emitTrace protocol.TraceHandler) (*protocol.Result, error) {

	if emitTrace != nil {
		ctx = httptrace.WithClientTrace(ctx, getHTTPTrace(idGen, req.ID, emitTrace))
	}

	switch req.Protocol {
	case protocol.ProtocolREST, protocol.ProtocolGQL:
		return pgrest.New(d.client).Send(ctx, req, idGen, emitTrace)
	default:
		return nil, protocol.ErrNotSupport
	}
}
func getHTTPTrace(idGen *uid.AtomicGen, reqID ulid.ULID, emit protocol.TraceHandler) *httptrace.ClientTrace {
	traceNetwork := func() *protocol.Trace { return protocol.NewTrace(idGen.Next(), reqID, protocol.TraceNetwork) }
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			emit(traceNetwork().SetMsg("GetConn").SetMeta(map[string]any{
				"hostPort": hostPort,
			}))
		},
		GotConn: func(info httptrace.GotConnInfo) {
			emit(traceNetwork().SetMsg("GotConn").SetMeta(map[string]any{
				"reused":   info.Reused,
				"wasIdle":  info.WasIdle,
				"idleTime": info.IdleTime,
				"local":    info.Conn.LocalAddr(),
				"remote":   info.Conn.RemoteAddr(),
			}))
		},
		PutIdleConn: func(err error) {
			emit(traceNetwork().SetMsg("PutIdleConn").SetMeta(map[string]any{
				"err": err,
			}))
		},
		GotFirstResponseByte: func() {
			emit(traceNetwork().SetMsg("GotFirstResponseByte"))
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			emit(traceNetwork().SetMsg("DNSStart").SetMeta(map[string]any{
				"host": info.Host,
			}))
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			emit(traceNetwork().SetMsg("DNSDone").SetMeta(map[string]any{
				"addrs":     info.Addrs,
				"err":       info.Err,
				"coalesced": info.Coalesced,
			}))
		},
		ConnectStart: func(network, addr string) {
			emit(traceNetwork().SetMsg("ConnectStart").SetMeta(map[string]any{
				"network": network,
				"addr":    addr,
			}))
		},
		ConnectDone: func(network, addr string, err error) {
			emit(traceNetwork().SetMsg("ConnectDone").SetMeta(map[string]any{
				"network": network,
				"addr":    addr,
				"err":     err,
			}))
		},
		TLSHandshakeStart: func() {
			emit(traceNetwork().SetMsg("TLSHandshakeStart"))
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			t := traceNetwork().SetMsg("TLSHandshakeDone")
			if err != nil {
				emit(t.SetMeta(map[string]any{"err": err}))
				return
			}
			emit(t.SetMeta(map[string]any{
				"version":            state.Version,
				"handshakeComplete":  state.HandshakeComplete,
				"didResume":          state.DidResume,
				"cipherSuite":        state.CipherSuite,
				"negotiatedProtocol": state.NegotiatedProtocol,
				"serverName":         state.ServerName,
			}))
		},
		WroteHeaderField: func(key string, value []string) {
			emit(traceNetwork().SetMsg("WroteHeaderField").SetMeta(map[string]any{
				"key":   key,
				"value": value,
			}))
		},
		WroteHeaders: func() {
			emit(traceNetwork().SetMsg("WroteHeaders"))
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			emit(traceNetwork().SetMsg("WroteRequest").SetMeta(map[string]any{
				"err": info.Err,
			}))
		},
	}
}
