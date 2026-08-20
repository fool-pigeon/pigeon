package config_test

import (
	"testing"

	"github.com/homoarea/pigeon/internal/config"
	"github.com/stretchr/testify/require"
)

func TestFileName(t *testing.T) {
	t.Run("dafault", func(t *testing.T) {
		require.Equal(t, "config.toml", config.FileName())
	})
	t.Run("with a suffix", func(t *testing.T) {
		require.Equal(t, "config-dev.toml", config.FileName("dev"))
	})
}
