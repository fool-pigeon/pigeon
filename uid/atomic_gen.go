// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package uid

import (
	"encoding/binary"
	"sync/atomic"
)

type AtomicGen struct {
	seq atomic.Uint64
}

func (g *AtomicGen) Next() uint64 {
	return g.seq.Add(1)
}
func (g *AtomicGen) NextID() []byte {
	k := make([]byte, 8)
	binary.BigEndian.PutUint64(k, g.Next())
	return k
}
