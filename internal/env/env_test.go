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
func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("load a normal .env file", func(t *testing.T) {
		e, _ := env.LoadFromFile(createFile(t, tmpDir, "a=1\nb=2"))
		assert.Equal(t, e.All(), map[string]any{"a": "1", "b": "2"})
	})
	t.Run("load a .env has comment", func(t *testing.T) {
		e, _ := env.LoadFromFile(createFile(t, tmpDir, "# line comment\na=1 # comment"))
		assert.Equal(t, e.All(), map[string]any{"a": "1"})
	})
	t.Run("load .env has the same keys", func(t *testing.T) {
		e, _ := env.LoadFromFile(createFile(t, tmpDir, "a=1\na=2"))
		assert.Equal(t, e.All(), map[string]any{"a": "2"})
	})
}

func TestExpand(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "hhh${abc}"
		e := env.New()
		e.Set("abc", 114)
		assert.Equal(t, "hhh114", e.Expand(s))
	})
	t.Run("no ${var} in string", func(t *testing.T) {
		s := "hhh"
		e := env.New()
		e.Set("abc", 114)
		assert.Equal(t, "hhh", e.Expand(s))
	})
	t.Run("the key-value not exist in Env", func(t *testing.T) {
		s := "hhh${abc}"
		e := env.New()
		assert.Equal(t, "hhh${abc}", e.Expand(s))
	})
	t.Run("Variable of secondary path", func(t *testing.T) {
		s := "hhh${a.b.c}"
		e := env.New()
		e.Set("a.b.c", 114)
		e.Print()
		assert.Equal(t, "hhh114", e.Expand(s))
	})
	t.Run("bad syntax; ${", func(t *testing.T) {
		s := "hhh${abc"
		e := env.New()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${abc", e.Expand(s))
	})
	t.Run("bad syntax; ${}", func(t *testing.T) {
		s := "hhh${}"
		e := env.New()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${}", e.Expand(s))
	})
}
