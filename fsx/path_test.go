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

// mkProject creates <tmp>/a/b/c with a .pg dir at <tmp>.
// Returns (root, mid=a, leaf=b/c) as absolute paths.
func mkProject(t *testing.T) (root, mid, leaf string) {
	t.Helper()
	root = t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, RootTag), 0o755))
	mid = filepath.Join(root, "a")
	leaf = filepath.Join(mid, "b", "c")
	require.NoError(t, os.MkdirAll(leaf, 0o755))
	return
}

func TestRootDir(t *testing.T) {
	root, mid, leaf := mkProject(t)

	t.Run("from leaf", func(t *testing.T) {
		got, err := RootDir(leaf)
		require.NoError(t, err)
		assert.Equal(t, root, got)
	})
	t.Run("from mid", func(t *testing.T) {
		got, err := RootDir(mid)
		require.NoError(t, err)
		assert.Equal(t, root, got)
	})
	t.Run("from root", func(t *testing.T) {
		got, err := RootDir(root)
		require.NoError(t, err)
		assert.Equal(t, root, got)
	})
	t.Run("nearest wins", func(t *testing.T) {
		nearRoot, nearMid, nearLeaf := mkProject(t)
		_ = nearRoot
		require.NoError(t, os.Mkdir(filepath.Join(nearMid, RootTag), 0o755))
		got, err := RootDir(nearLeaf)
		require.NoError(t, err)
		assert.Equal(t, nearMid, got)
	})
}

func TestRootTagDir(t *testing.T) {
	root, _, leaf := mkProject(t)
	got, err := RootTagDir(leaf)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, RootTag), got)
}

func TestPathsToRoot(t *testing.T) {
	root, mid, leaf := mkProject(t)
	got, err := PathsToRoot(leaf)
	require.NoError(t, err)
	assert.Equal(t, []string{leaf, filepath.Dir(leaf), mid, root}, got)
}

func TestPathsFromRoot(t *testing.T) {
	root, mid, leaf := mkProject(t)
	got, err := PathsFromRoot(leaf)
	require.NoError(t, err)
	assert.Equal(t, []string{root, mid, filepath.Dir(leaf), leaf}, got)
}

func TestRootConfDir(t *testing.T) {
	root, _, leaf := mkProject(t)
	want := filepath.Join(root, RootTag, Conf)
	require.NoError(t, os.Mkdir(want, 0o755))

	got, err := RootConfDir(leaf)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestRootRunDir(t *testing.T) {
	root, _, leaf := mkProject(t)
	want := filepath.Join(root, RootTag, RootRun)
	require.NoError(t, os.Mkdir(want, 0o755))

	got, err := RootRunDir(leaf)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestRootSubDirErrors(t *testing.T) {
	t.Run("missing sub dir", func(t *testing.T) {
		_, _, leaf := mkProject(t)
		_, err := RootConfDir(leaf)
		assert.Error(t, err)
	})
	t.Run("sub dir is file", func(t *testing.T) {
		root, _, leaf := mkProject(t)
		require.NoError(t, os.WriteFile(filepath.Join(root, RootTag, Conf), []byte("x"), 0o644))
		_, err := RootConfDir(leaf)
		assert.ErrorIs(t, err, ErrNotDir)
	})
}
