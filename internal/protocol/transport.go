// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "context"

type Sender interface {
	Send(ctx context.Context, req *Request, emitTrace TraceHandler) (*Result, error)
}
