// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package uid

import (
	"bytes"
	"encoding/binary"
	"sync"
	"testing"
)

func TestAtomicGen(t *testing.T) {
	t.Run("Next", func(t *testing.T) {
		t.Run("FirstCallReturnsOne", func(t *testing.T) {
			g := &AtomicGen{}
			if got := g.Next(); got != 1 {
				t.Fatalf("got %d, want 1", got)
			}
		})
		t.Run("SuccessiveCallsIncrement", func(t *testing.T) {
			g := &AtomicGen{}
			for i := uint64(1); i <= 5; i++ {
				if got := g.Next(); got != i {
					t.Fatalf("call %d: got %d, want %d", i, got, i)
				}
			}
		})
	})

	t.Run("NextID", func(t *testing.T) {
		t.Run("ReturnsEightBytes", func(t *testing.T) {
			g := &AtomicGen{}
			if got := len(g.NextID()); got != 8 {
				t.Fatalf("got %d bytes, want 8", got)
			}
		})
		t.Run("BigEndianEncoding", func(t *testing.T) {
			g := &AtomicGen{}
			want := []byte{0, 0, 0, 0, 0, 0, 0, 1}
			if got := g.NextID(); !bytes.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
		t.Run("ConsumesSequence", func(t *testing.T) {
			g := &AtomicGen{}
			_ = g.NextID() // seq -> 1
			if got := g.Next(); got != 2 {
				t.Fatalf("got %d, want 2", got)
			}
		})
		t.Run("EncodesCurrentSequence", func(t *testing.T) {
			g := &AtomicGen{}
			_ = g.Next()     // seq -> 1
			id := g.NextID() // seq -> 2
			want := make([]byte, 8)
			binary.BigEndian.PutUint64(want, 2)
			if !bytes.Equal(id, want) {
				t.Fatalf("got %v, want %v", id, want)
			}
		})
	})

	t.Run("LexicographicOrder", func(t *testing.T) {
		g := &AtomicGen{}
		var prev []byte
		for i := 0; i < 1000; i++ {
			cur := g.NextID()
			if prev != nil && bytes.Compare(prev, cur) >= 0 {
				t.Fatalf("index %d: %v not less than %v", i, prev, cur)
			}
			prev = cur
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		t.Run("NextIsUnique", func(t *testing.T) {
			g := &AtomicGen{}
			const n = 1000
			out := make([]uint64, n)
			var wg sync.WaitGroup
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					out[i] = g.Next()
				}(i)
			}
			wg.Wait()
			seen := make(map[uint64]struct{}, n)
			for _, v := range out {
				if _, dup := seen[v]; dup {
					t.Fatalf("duplicate %d", v)
				}
				seen[v] = struct{}{}
			}
		})
		t.Run("NextIDIsUnique", func(t *testing.T) {
			g := &AtomicGen{}
			const n = 1000
			out := make([][]byte, n)
			var wg sync.WaitGroup
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					out[i] = g.NextID()
				}(i)
			}
			wg.Wait()
			seen := make(map[string]struct{}, n)
			for _, v := range out {
				if _, dup := seen[string(v)]; dup {
					t.Fatalf("duplicate %v", v)
				}
				seen[string(v)] = struct{}{}
			}
		})
	})
}
