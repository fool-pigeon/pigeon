// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodConstants(t *testing.T) {
	pairs := []struct {
		name string
		m    Method
		want string
	}{
		{"get", MethodGet, http.MethodGet},
		{"post", MethodPost, http.MethodPost},
		{"put", MethodPut, http.MethodPut},
		{"delete", MethodDelete, http.MethodDelete},
		{"patch", MethodPatch, http.MethodPatch},
		{"head", MethodHead, http.MethodHead},
		{"options", MethodOptions, http.MethodOptions},
		{"trace", MethodTrace, http.MethodTrace},
		{"connect", MethodConnect, http.MethodConnect},
	}
	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) {
			assert.Equal(t, p.want, string(p.m))
		})
	}
}

func TestMethodsSlice(t *testing.T) {
	want := []Method{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodOptions, MethodConnect, MethodPatch, MethodTrace}
	assert.Equal(t, want, Methods)

	seen := map[Method]bool{}
	for _, m := range Methods {
		assert.Falsef(t, seen[m], "duplicate method %q", m)
		seen[m] = true
	}
}

func TestSentinelErrors(t *testing.T) {
	errs := []error{ErrNilValue, ErrEmptyValue, ErrInvalidValue, ErrNotSupport}
	seen := map[error]bool{}
	for _, e := range errs {
		assert.NotNil(t, e)
		assert.NotEmpty(t, e.Error())
		assert.Falsef(t, seen[e], "duplicate error %q", e)
		seen[e] = true
	}
}

func TestHeaderKeysAreCanonical(t *testing.T) {
	keys := map[string]string{
		"user agent":          HdrUserAgentKey,
		"accept":              HdrAcceptKey,
		"accept encoding":     HdrAcceptEncodingKey,
		"content type":        HdrContentTypeKey,
		"content length":      HdrContentLengthKey,
		"content encoding":    HdrContentEncodingKey,
		"content disposition": HdrContentDisposition,
		"authorization":       HdrAuthorizationKey,
		"www authenticate":    HdrWwwAuthenticateKey,
		"retry after":         HdrRetryAfterKey,
		"cookie":              HdrCookieKey,
	}
	for name, k := range keys {
		t.Run(name, func(t *testing.T) {
			require.NotEmpty(t, k)
			assert.Equal(t, k, http.CanonicalHeaderKey(k), "not canonical")
		})
	}
}

func TestContentTypeAndKeyConstants(t *testing.T) {
	t.Run("content types", func(t *testing.T) {
		assert.Equal(t, "text/plain; charset=utf-8", plainTextContentType)
		assert.Equal(t, "application/json", jsonContentType)
		assert.Equal(t, "application/xml", xmlContentType)
	})
	t.Run("keys", func(t *testing.T) {
		assert.Equal(t, "text", textKey)
		assert.Equal(t, "json", jsonKey)
		assert.Equal(t, "xml", xmlKey)
	})
}

func TestUserAgentValue(t *testing.T) {
	require.NotEmpty(t, hdrUserAgentValue)
	assert.True(t, strings.HasPrefix(hdrUserAgentValue, "pigeon@"))
}

func TestPathParamRe(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "/users/{id}", []string{"id"}},
		{"two", "/users/{id}/posts/{pid}", []string{"id", "pid"}},
		{"root", "{id}", []string{"id"}},
		{"no params", "/users/list", nil},
		{"dollar rejected", "/{a${b}}", nil},
		{"slash rejected", "/{a/b}", nil},
		{"empty braces", "/{}", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ms := pathParamRe.FindAllStringSubmatch(c.in, -1)
			var got []string
			for _, m := range ms {
				got = append(got, m[1])
			}
			assert.Equal(t, c.want, got)
		})
	}
}

func TestClient(t *testing.T) {
	t.Run("nil before instance", func(t *testing.T) {
		c := &Client{}
		assert.Nil(t, c.GetClient())
	})

	t.Run("instances are independent", func(t *testing.T) {
		c := &Client{}
		first, err := c.InstancedClient()
		require.NoError(t, err)
		require.NotNil(t, first)
		assert.Same(t, first, c.GetClient())

		second, err := c.InstancedClient()
		require.NoError(t, err)
		require.NotNil(t, second)
		assert.NotSame(t, first, second)
		assert.Same(t, second, c.GetClient())
	})

	t.Run("sends user agent", func(t *testing.T) {
		var got string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Get(HdrUserAgentKey)
		}))
		defer srv.Close()

		cli, err := new(Client).InstancedClient()
		require.NoError(t, err)
		_, err = cli.R().Get(srv.URL)
		require.NoError(t, err)
		assert.Equal(t, hdrUserAgentValue, got)
	})

	t.Run("accepts proxy and redirect config", func(t *testing.T) {
		c := &Client{
			ProxyURL: "http://127.0.0.1:9999",
			Redirect: Redirect{MaxCount: 3, Domains: []string{"example.com"}},
		}
		cli, err := c.InstancedClient()
		require.NoError(t, err)
		require.NotNil(t, cli)
	})
}
