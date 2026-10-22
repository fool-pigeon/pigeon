// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"context"

	"github.com/fool-pigeon/pigeon/uid"
)

type Sender interface {
	Send(ctx context.Context, req *Request, idGen *uid.AtomicGen, emitTrace TraceHandler) (*Result, error)
}
