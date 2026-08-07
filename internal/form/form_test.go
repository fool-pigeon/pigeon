package form_test

import (
	"testing"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/homoarea/pigeon/internal/form"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/stretchr/testify/require"
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
func TestHttpForm(t *testing.T) {
	t.Run("test HTTP", func(t *testing.T) {
		t.Run("test input URL", func(t *testing.T) {
			simulateKeys(t, "e?k1=v1", keys.Enter)
			r := httpx.Request{}
			if err := form.HttpURL(&r); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, "http://e?k1=v1", r.RawURL())
		})
		t.Run("test select method", func(t *testing.T) {
			simulateKeys(t, keys.Down, keys.Enter, keys.Tab)
			r := httpx.Request{}
			if err := form.HttpMethod(&r); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, "POST", r.Method)
		})
	})
}
