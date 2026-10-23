// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"context"

	"github.com/fool-pigeon/pigeon/uid"
	"github.com/oklog/ulid/v2"
)

type Sender interface {
	Send(ctx context.Context, req *Request, resID ulid.ULID, idGen *uid.Atomic64Gen, emitTrace TraceHandler) (*Result, error)
}
