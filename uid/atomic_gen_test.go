// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package uid

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomic64Gen(t *testing.T) {
	t.Run("Next", func(t *testing.T) {
		t.Run("FirstCallReturnsOne", func(t *testing.T) {
			g := NewU64Gen()
			assert.Equal(t, uint64(1), g.Next())
		})
		t.Run("SuccessiveCallsIncrement", func(t *testing.T) {
			g := NewU64Gen()
			for i := uint64(1); i <= 5; i++ {
				assert.Equal(t, i, g.Next())
			}
		})
	})

	t.Run("NextBytes", func(t *testing.T) {
		t.Run("ReturnsEightBytes", func(t *testing.T) {
			g := NewU64Gen()
			assert.Len(t, g.NextBytes(), 8)
		})
		t.Run("BigEndianEncoding", func(t *testing.T) {
			g := NewU64Gen()
			assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 1}, g.NextBytes())
		})
		t.Run("ConsumesSequence", func(t *testing.T) {
			g := NewU64Gen()
			_ = g.NextBytes() // seq -> 1
			assert.Equal(t, uint64(2), g.Next())
		})
		t.Run("EncodesCurrentSequence", func(t *testing.T) {
			g := NewU64Gen()
			_ = g.Next()        // seq -> 1
			id := g.NextBytes() // seq -> 2
			assert.Equal(t, Uint64ToBytes(2), id)
		})
	})

	t.Run("LexicographicOrder", func(t *testing.T) {
		g := NewU64Gen()
		var prev []byte
		for i := 0; i < 1000; i++ {
			cur := g.NextBytes()
			if prev != nil {
				assert.Negative(t, bytes.Compare(prev, cur),
					"index %d: %v not less than %v", i, prev, cur)
			}
			prev = cur
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		const n = 1000
		t.Run("NextIsUnique", func(t *testing.T) {
			g := NewU64Gen()
			out := make([]uint64, n)
			var wg sync.WaitGroup
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func(i int) {
					defer wg.Done()
					out[i] = g.Next()
				}(i)
			}
			wg.Wait()
			seen := make(map[uint64]struct{}, n)
			for _, v := range out {
				_, dup := seen[v]
				assert.False(t, dup, "duplicate %d", v)
				seen[v] = struct{}{}
			}
		})
		t.Run("NextBytesIsUnique", func(t *testing.T) {
			g := NewU64Gen()
			out := make([][]byte, n)
			var wg sync.WaitGroup
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func(i int) {
					defer wg.Done()
					out[i] = g.NextBytes()
				}(i)
			}
			wg.Wait()
			seen := make(map[string]struct{}, n)
			for _, v := range out {
				_, dup := seen[string(v)]
				assert.False(t, dup, "duplicate %v", v)
				seen[string(v)] = struct{}{}
			}
		})
	})
}

func TestAtomic32Gen(t *testing.T) {
	t.Run("Next", func(t *testing.T) {
		t.Run("FirstCallReturnsOne", func(t *testing.T) {
			g := NewU32Gen()
			assert.Equal(t, uint32(1), g.Next())
		})
		t.Run("SuccessiveCallsIncrement", func(t *testing.T) {
			g := NewU32Gen()
			for i := uint32(1); i <= 5; i++ {
				assert.Equal(t, i, g.Next())
			}
		})
	})

	t.Run("NextBytes", func(t *testing.T) {
		t.Run("ReturnsFourBytes", func(t *testing.T) {
			g := NewU32Gen()
			assert.Len(t, g.NextBytes(), 4)
		})
		t.Run("BigEndianEncoding", func(t *testing.T) {
			g := NewU32Gen()
			assert.Equal(t, []byte{0, 0, 0, 1}, g.NextBytes())
		})
		t.Run("ConsumesSequence", func(t *testing.T) {
			g := NewU32Gen()
			_ = g.NextBytes() // seq -> 1
			assert.Equal(t, uint32(2), g.Next())
		})
		t.Run("EncodesCurrentSequence", func(t *testing.T) {
			g := NewU32Gen()
			_ = g.Next()        // seq -> 1
			id := g.NextBytes() // seq -> 2
			assert.Equal(t, Uint32ToBytes(2), id)
		})
	})

	t.Run("LexicographicOrder", func(t *testing.T) {
		g := NewU32Gen()
		var prev []byte
		for i := 0; i < 1000; i++ {
			cur := g.NextBytes()
			if prev != nil {
				assert.Negative(t, bytes.Compare(prev, cur),
					"index %d: %v not less than %v", i, prev, cur)
			}
			prev = cur
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		const n = 1000
		t.Run("NextIsUnique", func(t *testing.T) {
			g := NewU32Gen()
			out := make([]uint32, n)
			var wg sync.WaitGroup
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func(i int) {
					defer wg.Done()
					out[i] = g.Next()
				}(i)
			}
			wg.Wait()
			seen := make(map[uint32]struct{}, n)
			for _, v := range out {
				_, dup := seen[v]
				assert.False(t, dup, "duplicate %d", v)
				seen[v] = struct{}{}
			}
		})
		t.Run("NextBytesIsUnique", func(t *testing.T) {
			g := NewU32Gen()
			out := make([][]byte, n)
			var wg sync.WaitGroup
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func(i int) {
					defer wg.Done()
					out[i] = g.NextBytes()
				}(i)
			}
			wg.Wait()
			seen := make(map[string]struct{}, n)
			for _, v := range out {
				_, dup := seen[string(v)]
				assert.False(t, dup, "duplicate %v", v)
				seen[string(v)] = struct{}{}
			}
		})
	})
}

func TestUint64ToBytes(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		assert.Len(t, Uint64ToBytes(0), 8)
	})
	t.Run("BigEndian", func(t *testing.T) {
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 1}, Uint64ToBytes(1))
	})
	t.Run("MaxValue", func(t *testing.T) {
		assert.Equal(t,
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			Uint64ToBytes(^uint64(0)))
	})
}

func TestUint32ToBytes(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		assert.Len(t, Uint32ToBytes(0), 4)
	})
	t.Run("BigEndian", func(t *testing.T) {
		assert.Equal(t, []byte{0, 0, 0, 1}, Uint32ToBytes(1))
	})
	t.Run("MaxValue", func(t *testing.T) {
		assert.Equal(t,
			[]byte{0xFF, 0xFF, 0xFF, 0xFF},
			Uint32ToBytes(^uint32(0)))
	})
}

func TestBytesToBase32(t *testing.T) {
	t.Run("NoPadding", func(t *testing.T) {
		assert.Len(t, BytesToBase32([]byte{0, 0, 0, 1}), 7)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		assert.Empty(t, BytesToBase32(nil))
	})
	t.Run("ZeroBytes", func(t *testing.T) {
		assert.Equal(t, "AAAAAAA", BytesToBase32(make([]byte, 4)))
		assert.Equal(t, "AAAAAAAAAAAAA", BytesToBase32(make([]byte, 8)))
	})
	t.Run("RoundTrip", func(t *testing.T) {
		inputs := [][]byte{
			{},
			{0, 0, 0, 1},
			{0xFF, 0xFF, 0xFF, 0xFF},
			{1, 2, 3, 4, 5, 6, 7, 8},
		}
		for _, in := range inputs {
			got, err := b32.DecodeString(BytesToBase32(in))
			assert.NoError(t, err)
			assert.Equal(t, in, got)
		}
	})
}

func TestUint32ToBase32(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		assert.Len(t, Uint32ToBase32(0), 7)
	})
	t.Run("MatchesBytesEncoding", func(t *testing.T) {
		for _, v := range []uint32{0, 1, 2, 31, 32, 1000, ^uint32(0)} {
			assert.Equal(t, BytesToBase32(Uint32ToBytes(v)), Uint32ToBase32(v))
		}
	})
}

func TestUint64ToBase32(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		assert.Len(t, Uint64ToBase32(0), 13)
	})
	t.Run("MatchesBytesEncoding", func(t *testing.T) {
		for _, v := range []uint64{0, 1, 2, 31, 32, 1 << 40, ^uint64(0)} {
			assert.Equal(t, BytesToBase32(Uint64ToBytes(v)), Uint64ToBase32(v))
		}
	})
}
