// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fsx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/stretchr/testify/require"
)

func TestRootDir(t *testing.T) {
	t.Run("the tagDir in current dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		tagDir := filepath.Join(tmpDir, fsx.RootTagDirName)
		fsx.Mkdir(tagDir)
		rootDir, err := fsx.RootDir(tmpDir)
		require.Equal(t, tmpDir, rootDir)
		require.Nil(t, err)
	})

	t.Run("find the tag dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		tagDir := filepath.Join(tmpDir, fsx.RootTagDirName)
		fsx.Mkdir(tagDir)
		rootTag, err := fsx.RootTagDir(tmpDir)
		require.Equal(t, tagDir, rootTag)
		require.Nil(t, err)
	})

	t.Run("the tagDir in the parent dir", func(t *testing.T) {
		parentDir := t.TempDir()
		tagDir := filepath.Join(parentDir, fsx.RootTagDirName)
		fsx.Mkdir(tagDir)
		sonDir, _ := os.MkdirTemp(parentDir, "son")
		rootDir, err := fsx.RootDir(sonDir)
		require.Equal(t, parentDir, rootDir)
		require.Nil(t, err)
	})

	t.Run("not exist tag dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		sonDir, _ := os.MkdirTemp(tmpDir, "son")
		rootDir, err := fsx.RootDir(sonDir)
		require.Equal(t, rootDir, "")
		require.ErrorIs(t, err, fsx.ErrRootNotExist)
	})
}
func TestPathsToRoot(t *testing.T) {
	t.Run("not exist tag dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		sonDir, _ := os.MkdirTemp(tmpDir, "son")
		paths, err := fsx.PathsToRoot(sonDir)
		require.Nil(t, paths)
		require.ErrorIs(t, err, fsx.ErrRootNotExist)
	})
	t.Run("tag dir in current dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		fsx.Mkdir(filepath.Join(tmpDir, fsx.RootTagDirName))
		paths, err := fsx.PathsToRoot(tmpDir)
		require.Equal(t, []string{tmpDir}, paths)
		require.Nil(t, err)
	})
	t.Run("tag dir in parent dir", func(t *testing.T) {
		parentDir := t.TempDir()
		fsx.Mkdir(filepath.Join(parentDir, fsx.RootTagDirName))
		sonDir, _ := os.MkdirTemp(parentDir, "son")
		paths, err := fsx.PathsToRoot(sonDir)
		require.Equal(t, []string{sonDir, parentDir}, paths)
		require.Nil(t, err)
	})
}
