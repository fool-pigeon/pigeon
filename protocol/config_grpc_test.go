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

func TestConfigGRPCCopy(t *testing.T) {
	seed := func() *ConfigGRPC {
		return &ConfigGRPC{
			ReadMaxBytes:      1024,
			SendMaxBytes:      2048,
			UseProtoJSON:      true,
			EnableHTTPGet:     true,
			HTTPGetMaxURLSize: 4096,
			SendGzip:          true,
		}
	}

	t.Run("equal values", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		require.NotSame(t, src, dst)
		assert.Equal(t, src, dst)
	})

	t.Run("mutating copy keeps source", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		dst.ReadMaxBytes = 1
		dst.SendGzip = false
		assert.Equal(t, 1024, src.ReadMaxBytes)
		assert.True(t, src.SendGzip)
	})

	t.Run("zero value", func(t *testing.T) {
		dst := (&ConfigGRPC{}).Copy()
		require.NotNil(t, dst)
		assert.Equal(t, &ConfigGRPC{}, dst)
	})
}
