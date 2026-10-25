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

func TestBodyTypes(t *testing.T) {
	t.Run("contains_all_kinds", func(t *testing.T) {
		want := []BodyKind{BodyNone, BodyURLEncoded, BodyMultipart, BodyRaw, BodyGQL, BodyJSON, BodyXML, BodyBinary}
		assert.Equal(t, want, BodyTypes)
	})

	t.Run("no_duplicates", func(t *testing.T) {
		seen := make(map[BodyKind]struct{}, len(BodyTypes))
		for _, k := range BodyTypes {
			_, dup := seen[k]
			assert.Falsef(t, dup, "duplicate kind %q", k)
			seen[k] = struct{}{}
		}
	})
}

func TestConfigRESTCopy(t *testing.T) {
	seed := func() *ConfREST {
		return &ConfREST{
			Tests:             map[string]string{"a": "1", "b": "2"},
			Path:              map[string]string{"p": "/x"},
			Retry:             Retry{Count: 3, WaitTime: "1s", MaxWaitTime: "10s", IsAllowNonIdempotent: true},
			UnescapeQuery:     true,
			UnescapeJSON:      true,
			ResponseBodyLimit: 1024,
		}
	}

	t.Run("equal values", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		require.NotSame(t, src, dst)
		assert.Equal(t, src.Tests, dst.Tests)
		assert.Equal(t, src.Path, dst.Path)
		assert.Equal(t, src.Retry, dst.Retry)
		assert.Equal(t, src.UnescapeQuery, dst.UnescapeQuery)
		assert.Equal(t, src.UnescapeJSON, dst.UnescapeJSON)
		assert.Equal(t, src.ResponseBodyLimit, dst.ResponseBodyLimit)
	})

	t.Run("maps do not alias", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		dst.Tests["a"] = "changed"
		dst.Path["p"] = "/y"
		assert.Equal(t, "1", src.Tests["a"])
		assert.Equal(t, "/x", src.Path["p"])
	})

	t.Run("nil maps", func(t *testing.T) {
		src := &ConfREST{}
		dst := src.Copy()
		require.NotSame(t, src, dst)
		require.NotNil(t, dst.Tests)
		require.NotNil(t, dst.Path)
		assert.Empty(t, dst.Tests)
		assert.Empty(t, dst.Path)
	})

	t.Run("mutating source after copy", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		src.Tests["new"] = "v"
		_, ok := dst.Tests["new"]
		assert.False(t, ok)
	})
}
