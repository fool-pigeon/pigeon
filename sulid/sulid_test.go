// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sulid

import (
	"sync"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

func TestTimestamp(t *testing.T) {
	t.Run("ReturnsMillisecondUnixTime", func(t *testing.T) {
		before := uint64(time.Now().UnixMilli())
		ts := Timestamp()
		after := uint64(time.Now().UnixMilli())

		if ts < before || ts > after {
			t.Fatalf("Timestamp() = %d, want in [%d, %d]", ts, before, after)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("ReturnsValidULID", func(t *testing.T) {
		id, err := New()
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}

		// 解码验证格式合法。
		parsed, err := ulid.Parse(id.String())
		if err != nil {
			t.Fatalf("generated ULID is not parsable: %v", err)
		}
		if parsed != id {
			t.Fatalf("parsed ULID mismatch: got %v, want %v", parsed, id)
		}
	})

	t.Run("TimestampEmbeddedIsRecent", func(t *testing.T) {
		id, err := New()
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}

		// ULID 时间戳应落在生成前后 1 秒内。
		now := time.Now().UnixMilli()
		got := int64(id.Time())
		if diff := now - got; diff < -1000 || diff > 1000 {
			t.Fatalf("embedded timestamp %d differs from now %d by %d ms", got, now, diff)
		}
	})

	t.Run("PoolRefillBoundary", func(t *testing.T) {
		// poolSize=250, entropyLen=10, 恰好 25 个 ULID 耗尽一个池。
		// 生成 26 个以触发一次重填。
		seen := make(map[ulid.ULID]struct{}, 26)
		for i := 0; i < 26; i++ {
			id, err := New()
			if err != nil {
				t.Fatalf("iteration %d: New() unexpected error: %v", i, err)
			}
			if _, dup := seen[id]; dup {
				t.Fatalf("iteration %d: duplicate ULID %v", i, id)
			}
			seen[id] = struct{}{}
		}
	})
}

func TestNewMonotonic(t *testing.T) {
	t.Run("StrictlyIncreasingWithinSameMillisecond", func(t *testing.T) {
		// 连续生成足够多的 ID，跨越至少一个池边界，单调性应保持。
		const n = 50
		ids := make([]ulid.ULID, n)
		for i := range ids {
			id, err := NewMonotonic()
			if err != nil {
				t.Fatalf("iteration %d: NewMonotonic() unexpected error: %v", i, err)
			}
			ids[i] = id
		}

		for i := 1; i < n; i++ {
			if ids[i].Compare(ids[i-1]) <= 0 {
				t.Fatalf("not strictly increasing at index %d: %v <= %v", i, ids[i], ids[i-1])
			}
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("NewIsConcurrentSafe", func(t *testing.T) {
		const goroutines = 8
		const perGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(goroutines)

		results := make(chan ulid.ULID, goroutines*perGoroutine)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					id, err := New()
					if err != nil {
						t.Errorf("New() unexpected error: %v", err)
						return
					}
					results <- id
				}
			}()
		}
		wg.Wait()
		close(results)

		seen := make(map[ulid.ULID]struct{}, goroutines*perGoroutine)
		for id := range results {
			if _, dup := seen[id]; dup {
				t.Fatalf("duplicate ULID under concurrency: %v", id)
			}
			seen[id] = struct{}{}
		}
		if len(seen) != goroutines*perGoroutine {
			t.Fatalf("generated %d unique ULIDs, want %d", len(seen), goroutines*perGoroutine)
		}
	})

	t.Run("NewMonotonicIsConcurrentSafe", func(t *testing.T) {
		const goroutines = 8
		const perGoroutine = 50

		var wg sync.WaitGroup
		wg.Add(goroutines)

		results := make(chan ulid.ULID, goroutines*perGoroutine)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					id, err := NewMonotonic()
					if err != nil {
						t.Errorf("NewMonotonic() unexpected error: %v", err)
						return
					}
					results <- id
				}
			}()
		}
		wg.Wait()
		close(results)

		seen := make(map[ulid.ULID]struct{}, goroutines*perGoroutine)
		for id := range results {
			if _, dup := seen[id]; dup {
				t.Fatalf("duplicate ULID under concurrent NewMonotonic: %v", id)
			}
			seen[id] = struct{}{}
		}
		if len(seen) != goroutines*perGoroutine {
			t.Fatalf("generated %d unique ULIDs, want %d", len(seen), goroutines*perGoroutine)
		}
	})
}
