// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tmpPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

func TestIsFileExist(t *testing.T) {
	t.Run("regular file", func(t *testing.T) {
		p := tmpPath(t, "f")
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
		ok, err := IsFileExist(p)
		require.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("missing", func(t *testing.T) {
		ok, err := IsFileExist(tmpPath(t, "nope"))
		require.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("dir is not file", func(t *testing.T) {
		p := tmpPath(t, "d")
		require.NoError(t, os.Mkdir(p, 0o755))
		ok, err := IsFileExist(p)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestIsDirExist(t *testing.T) {
	t.Run("dir", func(t *testing.T) {
		p := tmpPath(t, "d")
		require.NoError(t, os.Mkdir(p, 0o755))
		ok, err := IsDirExist(p)
		require.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("missing", func(t *testing.T) {
		ok, err := IsDirExist(tmpPath(t, "nope"))
		require.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("file is not dir", func(t *testing.T) {
		p := tmpPath(t, "f")
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
		ok, err := IsDirExist(p)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestWriteAndReadFile(t *testing.T) {
	p := tmpPath(t, "f")
	require.NoError(t, WriteFile(p, []byte("foo")))

	info, err := os.Stat(p)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

	b, err := ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, []byte("foo"), b)

	s, err := ReadFileAsString(p)
	require.NoError(t, err)
	assert.Equal(t, "foo", s)
}

func TestCreateFileIfNotExists(t *testing.T) {
	t.Run("creates", func(t *testing.T) {
		p := tmpPath(t, "f")
		require.NoError(t, CreateFileIfNotExists(p))
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	})
	t.Run("existing untouched", func(t *testing.T) {
		p := tmpPath(t, "f")
		require.NoError(t, os.WriteFile(p, []byte("keep"), 0o644))
		require.NoError(t, CreateFileIfNotExists(p))
		b, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Equal(t, []byte("keep"), b)
	})
	t.Run("missing parent", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "x", "f")
		assert.Error(t, CreateFileIfNotExists(p))
	})
}

func TestTomlRoundTrip(t *testing.T) {
	type conf struct {
		Name string `toml:"name"`
		Port int    `toml:"port"`
	}
	p := tmpPath(t, "c.toml")
	want := conf{Name: "foo", Port: 8080}

	require.NoError(t, WriteToml(p, want))
	var got conf
	require.NoError(t, ReadToml(p, &got))
	assert.Equal(t, want, got)
}

func TestTomlErrors(t *testing.T) {
	t.Run("read missing", func(t *testing.T) {
		var v map[string]any
		assert.Error(t, ReadToml(tmpPath(t, "nope.toml"), &v))
	})
	t.Run("read invalid", func(t *testing.T) {
		p := tmpPath(t, "bad.toml")
		require.NoError(t, os.WriteFile(p, []byte("= bad"), 0o644))
		var v map[string]any
		assert.Error(t, ReadToml(p, &v))
	})
	t.Run("write unsupported", func(t *testing.T) {
		assert.Error(t, WriteToml(tmpPath(t, "c.toml"), make(chan int)))
	})
}

func TestMkdir(t *testing.T) {
	t.Run("mkdir", func(t *testing.T) {
		p := tmpPath(t, "d")
		require.NoError(t, Mkdir(p))
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	})
	t.Run("mkdir existing errors", func(t *testing.T) {
		p := tmpPath(t, "d")
		require.NoError(t, Mkdir(p))
		assert.Error(t, Mkdir(p))
	})
	t.Run("mkdir all", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "a", "b", "c")
		require.NoError(t, MkdirAll(p))
		ok, err := IsDirExist(p)
		require.NoError(t, err)
		assert.True(t, ok)
	})
}
