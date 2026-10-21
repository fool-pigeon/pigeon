// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fool-pigeon/pigeon/core"
	"github.com/fool-pigeon/pigeon/fsx"
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
func TestExpand(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "hhh${abc}"
		e := core.NewEnv()
		e.Set("abc", 114)
		assert.Equal(t, "hhh114", e.Expand(s))
	})
	t.Run("no ${var} in string", func(t *testing.T) {
		s := "hhh"
		e := core.NewEnv()
		e.Set("abc", 114)
		assert.Equal(t, "hhh", e.Expand(s))
	})
	t.Run("the key-value not exist in Env", func(t *testing.T) {
		s := "hhh${abc}"
		e := core.NewEnv()
		assert.Equal(t, "hhh${abc}", e.Expand(s))
	})
	t.Run("Variable of secondary path", func(t *testing.T) {
		s := "hhh${a.b.c}"
		e := core.NewEnv()
		e.Set("a.b.c", 114)
		e.Print()
		assert.Equal(t, "hhh114", e.Expand(s))
	})
	t.Run("bad syntax; ${", func(t *testing.T) {
		s := "hhh${abc"
		e := core.NewEnv()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${abc", e.Expand(s))
	})
	t.Run("bad syntax; ${}", func(t *testing.T) {
		s := "hhh${}"
		e := core.NewEnv()
		e.Set("abc", 114)
		e.Print()
		assert.Equal(t, "hhh${}", e.Expand(s))
	})
	t.Run("with quotation mark", func(t *testing.T) {
		s := `'${hello}'='${hello}'`
		e := core.NewEnv()
		e.Set("hello", "world")
		assert.Equal(t, "'world'='world'", e.Expand(s))
	})
}
