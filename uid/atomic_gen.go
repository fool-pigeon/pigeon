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

func NewU64Gen() *AtomicGen {
	return &AtomicGen{}
}
func (g *AtomicGen) Next() uint64 {
	return g.seq.Add(1)
}
func (g *AtomicGen) NextBytes() []byte {
	return Uint64ToBytes(g.Next())
}
func Uint64ToBytes(v uint64) []byte {
	k := make([]byte, 8)
	binary.BigEndian.PutUint64(k, v)
	return k
}
