// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package sulid provides ULID generation backed by crypto/rand with a global
// random pool, mirroring the pool strategy used by github.com/google/uuid.
package sulid

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	entropyLen = 10  // ULID entropy length in bytes
	poolSize   = 250 // global pool size in bytes (25 ULIDs)
)

var (
	poolMu  sync.Mutex
	pool    [poolSize]byte
	poolPos = poolSize // points past the end, marking the pool as empty
)

// readEntropy fills p from the pool, or from crypto/rand when pooling is off.
// Callers must pass a slice of exactly entropyLen bytes.
func readEntropy(p []byte) (int, error) {
	poolMu.Lock()
	defer poolMu.Unlock()
	if poolPos+entropyLen > poolSize {
		if _, err := rand.Read(pool[:]); err != nil {
			return 0, err
		}
		poolPos = 0
	}
	copy(p, pool[poolPos:poolPos+entropyLen])
	poolPos += entropyLen
	return entropyLen, nil
}

// poolEntropy implements io.Reader for ulid.New.
type poolEntropy struct{}

// Read requires len(p) == entropyLen; ulid.New satisfies this.
func (poolEntropy) Read(p []byte) (int, error) { return readEntropy(p) }

var entropy poolEntropy

// returns current timestamp of Unix milliseconds.
func Timestamp() uint64 {
	return ulid.Timestamp(time.Now())
}

// New returns a ULID using the current time and pooled entropy.
func New() (ulid.ULID, error) {
	return ulid.New(Timestamp(), entropy)
}

var (
	monoMu  sync.Mutex
	monoEnt = ulid.Monotonic(entropy, 0)
)

// NewMonotonic returns a ULID strictly increasing within the same millisecond.
// ulid.Monotonic is not concurrency-safe, so callers are serialized by [monoMu].
func NewMonotonic() (ulid.ULID, error) {
	monoMu.Lock()
	defer monoMu.Unlock()
	return ulid.New(Timestamp(), monoEnt)
}
