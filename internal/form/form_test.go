package form_test

import (
	"testing"
	"time"

	"atomicgo.dev/keyboard"
)

const inputTimeout = 5 * time.Second

func simulateKeys(t *testing.T, keySequence ...any) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		// keyboard.SimulateKeyPress只处理它的第一个参数（它
		// 在其循环中返回），因此密钥将逐个发送。
		for _, key := range keySequence {
			_ = keyboard.SimulateKeyPress(key)
		}
	}()

	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(inputTimeout):
			t.Error("Has not consume all input keys")
		}
	})
}
func TestHttpForm(t *testing.T) {
	t.Run("test fill http body", func(t *testing.T) {

	})
}
