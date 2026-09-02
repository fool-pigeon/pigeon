package env_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/stretchr/testify/assert"
)

// returns filename(default timestamp),the file's content is the string s
func createFile(t *testing.T, parentDir string, s string, name ...string) string {
	t.Helper()
	fName := time.Now().String()
	if len(name) > 0 {
		fName = name[0]
	}
	fName = filepath.Join(parentDir, fName)
	fsx.WriteFile(fName, []byte(s))

	return fName
}
func createDir(t *testing.T, parentDir string, name ...string) string {
	t.Helper()
	dir := time.Now().String()
	if len(name) > 0 {
		dir = name[0]
	}
	dir = filepath.Join(parentDir, dir)
	fsx.Mkdir(dir)

	return dir
}
func TestFileName(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, ".env", env.FileName())
	})
	t.Run("has a suffix tag", func(t *testing.T) {
		assert.Equal(t, ".env-tag", env.FileName("tag"))
	})
	t.Run("more tags", func(t *testing.T) {
		assert.Equal(t, ".env-tag", env.FileName("tag", "tag1", "tag2"))
	})
}
func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("load a normal .env file", func(t *testing.T) {
		e := env.New()
		env.LoadFile(e, createFile(t, tmpDir, "a=1\nb=2"))
		assert.Equal(t, e.All(), map[string]any{"a": "1", "b": "2"})
	})
	t.Run("load a .env has comment", func(t *testing.T) {
		e := env.New()
		env.LoadFile(e, createFile(t, tmpDir, "# line comment\na=1 # comment"))
		assert.Equal(t, e.All(), map[string]any{"a": "1"})
	})
	t.Run("load .env has the same keys", func(t *testing.T) {
		e := env.New()
		env.LoadFile(e, createFile(t, tmpDir, "a=1\na=2"))
		assert.Equal(t, e.All(), map[string]any{"a": "2"})
	})
}

func TestLoadFromPath(t *testing.T) {
	t.Run("not exist .pg/", func(t *testing.T) {
		_, err := env.LoadFromPath(t.TempDir())
		assert.ErrorIs(t, fsx.ErrRootNotExist, err)
	})
	t.Run("normal load", func(t *testing.T) {
		rootDir := t.TempDir()
		tagDir := createDir(t, rootDir, ".pg")
		sonDir := createDir(t, rootDir, "son")
		createFile(t, tagDir, "a=1", ".env")
		createFile(t, rootDir, "b=2", ".env")
		createFile(t, sonDir, "c=3", ".env")
		e, _ := env.LoadFromPath(sonDir)
		assert.Equal(t, e.All(), map[string]any{"a": "1", "b": "2", "c": "3"})
	})
	t.Run("load that some path not has the .env", func(t *testing.T) {
		rootDir := t.TempDir()
		tagDir := createDir(t, rootDir, ".pg")
		sonDir := createDir(t, rootDir, "son")
		createFile(t, tagDir, "a=1", ".env")
		createFile(t, sonDir, "c=3", ".env")
		e, _ := env.LoadFromPath(sonDir)
		assert.Equal(t, e.All(), map[string]any{"a": "1", "c": "3"})
	})
	t.Run("load with override same key", func(t *testing.T) {
		rootDir := t.TempDir()
		tagDir := createDir(t, rootDir, ".pg")
		sonDir := createDir(t, rootDir, "son")
		createFile(t, tagDir, "a=1", ".env")
		createFile(t, rootDir, "a=2", ".env")
		createFile(t, sonDir, "a=3", ".env")
		e, _ := env.LoadFromPath(sonDir)
		assert.Equal(t, e.All(), map[string]any{"a": "3"})
	})
	t.Run("load with tag", func(t *testing.T) {
		rootDir := t.TempDir()
		tagDir := createDir(t, rootDir, ".pg")
		sonDir := createDir(t, rootDir, "son")
		createFile(t, tagDir, "a=1", ".env")
		createFile(t, rootDir, "b=2", ".env")
		createFile(t, sonDir, "c=3", ".env")
		createFile(t, tagDir, "a=a", ".env-dev")
		createFile(t, rootDir, "b=b", ".env-dev")
		createFile(t, sonDir, "c=c", ".env-dev")
		e, _ := env.LoadFromPath(sonDir, "dev")
		assert.Equal(t, e.All(), map[string]any{"a": "a", "b": "b", "c": "c"})
	})
}
func TestExpand(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "hhh${abc}"
		e := env.New()
		e.Set("abc", 114)
		assert.Equal(t, "hhh114", env.Expand(s, e))
	})
	t.Run("no ${var} in string", func(t *testing.T) {
		s := "hhh"
		e := env.New()
		e.Set("abc", 114)
		assert.Equal(t, "hhh", env.Expand(s, e))
	})
	t.Run("the key-value not exist in Env", func(t *testing.T) {
		s := "hhh${abc}"
		e := env.New()
		assert.Equal(t, "hhh${abc}", env.Expand(s, e))
	})
	t.Run("Variable of secondary path", func(t *testing.T) {
		s := "hhh${a.b.c}"
		e := env.New()
		e.Set("a.b.c", 114)
		e.Print()
		assert.Equal(t, "hhh114", env.Expand(s, e))
	})
	t.Run("bad syntax; ${", func(t *testing.T) {
		s := "hhh${abc"
		e := env.New()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${abc", env.Expand(s, e))
	})
	t.Run("bad syntax; ${}", func(t *testing.T) {
		s := "hhh${}"
		e := env.New()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${}", env.Expand(s, e))
	})
}
