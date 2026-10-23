// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigWSCopy(t *testing.T) {
	seed := func() *ConfigWS {
		return &ConfigWS{
			Subprotocols:         []string{"foo", "bar"},
			CompressionMode:      "context_takeover",
			CompressionThreshold: 512,
			OriginPatterns:       []string{"*.foo.com", "bar.com"},
			InsecureSkipVerify:   true,
		}
	}

	t.Run("equal values", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		require.NotSame(t, src, dst)
		assert.Equal(t, src.Subprotocols, dst.Subprotocols)
		assert.Equal(t, src.CompressionMode, dst.CompressionMode)
		assert.Equal(t, src.CompressionThreshold, dst.CompressionThreshold)
		assert.Equal(t, src.OriginPatterns, dst.OriginPatterns)
		assert.Equal(t, src.InsecureSkipVerify, dst.InsecureSkipVerify)
	})

	t.Run("slices not aliased", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		dst.Subprotocols[0] = "changed"
		dst.OriginPatterns[0] = "changed"
		assert.Equal(t, "foo", src.Subprotocols[0])
		assert.Equal(t, "*.foo.com", src.OriginPatterns[0])
	})

	t.Run("empty slices", func(t *testing.T) {
		dst := (&ConfigWS{}).Copy()
		require.NotNil(t, dst)
		assert.Empty(t, dst.Subprotocols)
		assert.Empty(t, dst.OriginPatterns)
	})

	t.Run("nil slices stay nil", func(t *testing.T) {
		src := &ConfigWS{CompressionMode: "disabled"}
		dst := src.Copy()
		assert.Nil(t, dst.Subprotocols)
		assert.Nil(t, dst.OriginPatterns)
		assert.Equal(t, "disabled", dst.CompressionMode)
	})
}
