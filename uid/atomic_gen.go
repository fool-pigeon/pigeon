// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package uid

import (
	"encoding/base32"
	"encoding/binary"
	"sync/atomic"
)

// Atomic64Gen generates monotonic uint64 IDs starting from 1.
type Atomic64Gen struct {
	seq atomic.Uint64
}

// NewU64Gen returns a new Atomic64Gen.
func NewU64Gen() *Atomic64Gen {
	return &Atomic64Gen{}
}

// Next atomically increments and returns the next uint64.
func (g *Atomic64Gen) Next() uint64 {
	return g.seq.Add(1)
}

// NextBytes returns the next ID encoded as 8 big-endian bytes.
func (g *Atomic64Gen) NextBytes() []byte {
	return Uint64ToBytes(g.Next())
}

// Uint64ToBytes encodes v as 8 big-endian bytes.
func Uint64ToBytes(v uint64) []byte {
	k := make([]byte, 8)
	binary.BigEndian.PutUint64(k, v)
	return k
}

// b32 is unpadded standard base32 (RFC 4648).
var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// BytesToBase32 encodes b as unpadded base32.
func BytesToBase32(b []byte) string {
	return b32.EncodeToString(b)
}

// Uint64ToBase32 encodes v as unpadded base32 (13 chars).
func Uint64ToBase32(v uint64) string {
	return BytesToBase32(Uint64ToBytes(v))
}

// Atomic32Gen generates monotonic uint32 IDs starting from 1.
type Atomic32Gen struct {
	seq atomic.Uint32
}

// NewU32Gen returns a new Atomic32Gen.
func NewU32Gen() *Atomic32Gen { return &Atomic32Gen{} }

// Next atomically increments and returns the next uint32.
func (g *Atomic32Gen) Next() uint32 { return g.seq.Add(1) }

// NextBytes returns the next ID encoded as 4 big-endian bytes.
func (g *Atomic32Gen) NextBytes() []byte { return Uint32ToBytes(g.Next()) }

// Uint32ToBytes encodes v as 4 big-endian bytes.
func Uint32ToBytes(v uint32) []byte {
	k := make([]byte, 4)
	binary.BigEndian.PutUint32(k, v)
	return k
}

// Uint32ToBase32 encodes v as unpadded base32 (7 chars).
func Uint32ToBase32(v uint32) string {
	return BytesToBase32(Uint32ToBytes(v))
}
