// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVarName(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantName string
		wantW    int
	}{
		{"valid", "{foo}", "foo", 5},
		{"single char", "{x}", "x", 3},
		{"dotted", "{a.b}", "a.b", 5},
		{"empty braces", "{}", "", 2},
		{"unclosed", "{", "", 1},
		{"unclosed with chars", "{foo", "", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, w := getVarName(c.in)
			assert.Equal(t, c.wantName, name)
			assert.Equal(t, c.wantW, w)
		})
	}
}

func TestNewEnv(t *testing.T) {
	e := NewEnv()
	require.NotNil(t, e)
	require.NotNil(t, e.Data)
	assert.Empty(t, e.All())
}

func TestEnvSetAndString(t *testing.T) {
	e := NewEnv()
	require.NoError(t, e.Set("k", "v"))
	assert.Equal(t, "v", e.String("k"))
	assert.Equal(t, "", e.String("missing"))
}

func TestEnvLoadNestedMap(t *testing.T) {
	t.Run("nested", func(t *testing.T) {
		e := NewEnv()
		require.NoError(t, e.LoadNestedMap(map[string]any{
			"a": map[string]any{"b": "c"},
		}))
		assert.Equal(t, "c", e.String("a.b"))
	})
	t.Run("overrides existing", func(t *testing.T) {
		e := NewEnv()
		require.NoError(t, e.Set("x", "old"))
		require.NoError(t, e.LoadNestedMap(map[string]any{"x": "new"}))
		assert.Equal(t, "new", e.String("x"))
	})
}

func TestEnvAll(t *testing.T) {
	e := NewEnv()
	require.NoError(t, e.Set("x", 1))
	assert.Equal(t, map[string]any{"x": 1}, e.All())
}

func TestEnvCopy(t *testing.T) {
	e := NewEnv()
	require.NoError(t, e.Set("x", 1))
	c := e.Copy()
	require.NotSame(t, e, c)
	require.NoError(t, c.Set("y", 2))
	_, has := e.All()["y"]
	assert.False(t, has)
	assert.Equal(t, 1, e.All()["x"])
}

func TestEnvExpand(t *testing.T) {
	newEnv := func(kv map[string]any) *Env {
		e := NewEnv()
		require.NoError(t, e.LoadNestedMap(kv))
		return e
	}

	cases := []struct {
		name string
		env  map[string]any
		in   string
		want string
	}{
		{"empty", nil, "", ""},
		{"no var", nil, "plain", "plain"},
		{"single", map[string]any{"X": "1"}, "${X}", "1"},
		{"embedded", map[string]any{"X": "1"}, "a${X}b", "a1b"},
		{"at start", map[string]any{"X": "1"}, "${X}b", "1b"},
		{"at end", map[string]any{"X": "1"}, "a${X}", "a1"},
		{"adjacent", map[string]any{"X": "1", "Y": "2"}, "${X}${Y}", "12"},
		{"multiple", map[string]any{"X": "1", "Y": "2"}, "${X}-${Y}", "1-2"},
		{"missing kept", nil, "a${NOPE}b", "a${NOPE}b"},
		{"empty braces kept", nil, "${}", "${}"},
		{"unclosed kept", nil, "${", "${"},
		{"unclosed tail kept", map[string]any{"X": "1"}, "a${X}${", "a1${"},
		{"lone dollar", nil, "$", "$"},
		{"dollar no brace", nil, "$X", "$X"},
		{"dotted key", map[string]any{"a": map[string]any{"b": "c"}}, "${a.b}", "c"},
		{"value not recursed", map[string]any{"X": "${Y}", "Y": "2"}, "${X}", "${Y}"},
		{"value empty", map[string]any{"X": ""}, "a${X}b", "ab"},
		{"double dollar", map[string]any{"X": "1"}, "$${X}", "$1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, newEnv(c.env).Expand(c.in))
		})
	}
}
