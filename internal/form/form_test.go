package form_test

import (
	"testing"
	"time"

	"atomicgo.dev/keyboard"
)

const inputTimeout = 5 * time.Second

// Cannot be used for parallel testing
// cannot print "e?k1=v1" directly,change to "e","k1=v1"
func simulateKeys(t *testing.T, keySequence ...any) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for _, key := range keySequence {
			_ = keyboard.SimulateKeyPress(key)
		}
	}()

	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(inputTimeout):
			t.Fatal("Has not consume all input keys")
		}
	})
}
