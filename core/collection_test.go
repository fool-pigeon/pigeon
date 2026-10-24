// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fool-pigeon/pigeon/fsx"
)

func TestInitCollection(t *testing.T) {
	t.Run("creates layout", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, InitCollection(dir))

		tagDir := filepath.Join(dir, fsx.RootTag)
		for _, d := range []string{
			tagDir,
			filepath.Join(tagDir, fsx.RootConf),
			filepath.Join(tagDir, fsx.RootRun),
		} {
			ok, err := fsx.IsDirExist(d)
			require.NoError(t, err)
			assert.Truef(t, ok, "missing dir %s", d)
		}

		confFile := filepath.Join(tagDir, fsx.RootConf, fsx.DefaultRootConfFile)
		ok, err := fsx.IsFileExist(confFile)
		require.NoError(t, err)
		assert.True(t, ok)

		var conf RootConf
		require.NoError(t, fsx.ReadToml(confFile, &conf))
	})

	t.Run("dir not exist", func(t *testing.T) {
		err := InitCollection(filepath.Join(t.TempDir(), "nope"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("path is file", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "f")
		require.NoError(t, os.WriteFile(p, nil, 0o644))
		err := InitCollection(p)
		assert.ErrorIs(t, err, fsx.ErrNotDir)
	})

	t.Run("already initialized errors", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, InitCollection(dir))
		assert.ErrorIs(t, InitCollection(dir), os.ErrExist)
	})
}
