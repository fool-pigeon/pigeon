// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dispatch

import (
	"context"

	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/transport/pgrest"
)

type Dispatcher struct {
	client *protocol.ActualClient
}

func New(client *protocol.ActualClient) *Dispatcher {
	return &Dispatcher{client}
}

// Send implements [protocol.Sender].
func (d *Dispatcher) Send(ctx context.Context, req *protocol.Request, emitTrace protocol.TraceHandler) (*protocol.Result, error) {
	switch req.Protocol {
	case protocol.ProtocolREST, protocol.ProtocolGQL:
		return pgrest.New(d.client).Send(ctx, req, emitTrace)
	default:
		return nil, protocol.ErrNotSupport
	}
}
