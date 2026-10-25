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

func TestProtocolKindConstants(t *testing.T) {
	pairs := []struct {
		name string
		p    ProtocolKind
		want string
	}{
		{"rest", ProtocolREST, "rest"},
		{"grpc", ProtocolGRPC, "grpc"},
		{"ws", ProtocolWS, "ws"},
		{"gql", ProtocolGQL, "gql"},
	}
	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) { assert.Equal(t, p.want, string(p.p)) })
	}
	assert.Equal(t, []ProtocolKind{ProtocolREST, ProtocolGRPC, ProtocolWS, ProtocolGQL}, Protocols)
}

func TestAuthKindConstants(t *testing.T) {
	pairs := []struct {
		name string
		k    AuthKind
		want string
	}{
		{"inherited", AuthInherited, ""},
		{"none", AuthNone, "none"},
		{"basic", AuthBasic, "basic"},
		{"digest", AuthDigest, "digest"},
		{"token", AuthToken, "token"},
	}
	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) { assert.Equal(t, p.want, string(p.k)) })
	}
	assert.Equal(t, []AuthKind{AuthInherited, AuthNone, AuthBasic, AuthDigest, AuthToken}, AuthKinds)
}

func TestReqConfigCopy(t *testing.T) {
	seed := func() *ReqConfig {
		return &ReqConfig{
			Desc:    "d",
			Auth:    Auth{Kind: AuthToken, Token: AuthDataToken{Header: "h", Scheme: "Bearer", Value: "v"}},
			Query:   map[string][]string{"a": {"1", "2"}},
			Header:  map[string][]string{"h": {"x"}},
			Timeout: "5s",
			ConfigREST: &ConfigREST{
				Tests: map[string]string{"t": "1"},
				Path:  map[string]string{"p": "/"},
			},
			ConfigGRPC: &ConfigGRPC{},
			ConfigWS:   &ConfigWS{},
		}
	}

	t.Run("equal values", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		require.NotSame(t, src, dst)
		assert.Equal(t, src.Desc, dst.Desc)
		assert.Equal(t, src.Auth, dst.Auth)
		assert.Equal(t, src.Timeout, dst.Timeout)
		assert.Equal(t, src.Query, dst.Query)
		assert.Equal(t, src.Header, dst.Header)
		assert.Equal(t, src.ConfigREST.Tests, dst.ConfigREST.Tests)
		assert.Equal(t, src.ConfigREST.Path, dst.ConfigREST.Path)
	})

	t.Run("slices not aliased", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		dst.Query["a"][0] = "changed"
		dst.Header["h"][0] = "changed"
		assert.Equal(t, "1", src.Query["a"][0])
		assert.Equal(t, "x", src.Header["h"][0])
	})

	t.Run("sub configs not aliased", func(t *testing.T) {
		src := seed()
		dst := src.Copy()
		assert.NotSame(t, src.ConfigREST, dst.ConfigREST)
		assert.NotSame(t, src.ConfigGRPC, dst.ConfigGRPC)
		assert.NotSame(t, src.ConfigWS, dst.ConfigWS)
	})
}

func TestReqConfigMerge(t *testing.T) {
	t.Run("nil parent noop", func(t *testing.T) {
		r := &ReqConfig{Desc: "d"}
		r.Merge(nil)
		assert.Equal(t, "d", r.Desc)
	})

	t.Run("auth", func(t *testing.T) {
		t.Run("inherited overridden", func(t *testing.T) {
			r := &ReqConfig{Auth: Auth{Kind: AuthInherited}}
			p := &ReqConfig{Auth: Auth{Kind: AuthBasic, Creds: AuthDataCreds{Username: "u", Password: "p"}}}
			r.Merge(p)
			assert.Equal(t, p.Auth, r.Auth)
		})
		t.Run("explicit kept", func(t *testing.T) {
			r := &ReqConfig{Auth: Auth{Kind: AuthNone}}
			r.Merge(&ReqConfig{Auth: Auth{Kind: AuthBasic}})
			assert.Equal(t, AuthNone, r.Auth.Kind)
		})
	})

	t.Run("header", func(t *testing.T) {
		t.Run("parent empty noop", func(t *testing.T) {
			r := &ReqConfig{Header: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConfig{})
			assert.Equal(t, map[string][]string{"a": {"1"}}, r.Header)
		})
		t.Run("self empty takes parent", func(t *testing.T) {
			r := &ReqConfig{}
			p := map[string][]string{"a": {"1"}}
			r.Merge(&ReqConfig{Header: p})
			assert.Equal(t, p, r.Header)
		})
		t.Run("dst preferred", func(t *testing.T) {
			r := &ReqConfig{Header: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConfig{Header: map[string][]string{"a": {"9"}, "b": {"2"}}})
			assert.Equal(t, map[string][]string{"a": {"1"}, "b": {"2"}}, r.Header)
		})
	})

	t.Run("query", func(t *testing.T) {
		t.Run("parent empty noop", func(t *testing.T) {
			r := &ReqConfig{Query: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConfig{})
			assert.Equal(t, map[string][]string{"a": {"1"}}, r.Query)
		})
		t.Run("dst preferred", func(t *testing.T) {
			r := &ReqConfig{Query: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConfig{Query: map[string][]string{"a": {"9"}, "b": {"2"}}})
			assert.Equal(t, map[string][]string{"a": {"1"}, "b": {"2"}}, r.Query)
		})
	})

	t.Run("timeout", func(t *testing.T) {
		t.Run("empty takes parent", func(t *testing.T) {
			r := &ReqConfig{}
			r.Merge(&ReqConfig{Timeout: "5s"})
			assert.Equal(t, DurationStr("5s"), r.Timeout)
		})
		t.Run("set kept", func(t *testing.T) {
			r := &ReqConfig{Timeout: "1s"}
			r.Merge(&ReqConfig{Timeout: "5s"})
			assert.Equal(t, DurationStr("1s"), r.Timeout)
		})
	})

	t.Run("config rest", func(t *testing.T) {
		t.Run("parent nil noop", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{}}
			r.Merge(&ReqConfig{})
			assert.NotNil(t, r.ConfigREST)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConfig{}
			p := &ConfigREST{Tests: map[string]string{"t": "1"}}
			r.Merge(&ReqConfig{ConfigREST: p})
			assert.Same(t, p, r.ConfigREST)
		})
		t.Run("merge tests", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{Tests: map[string]string{"a": "1"}}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{Tests: map[string]string{"a": "9", "b": "2"}}})
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, r.ConfigREST.Tests)
		})
		t.Run("merge path", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{Path: map[string]string{"a": "1"}}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{Path: map[string]string{"a": "9", "b": "2"}}})
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, r.ConfigREST.Path)
		})
		t.Run("merge retry", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{Retry: Retry{
				Count: 3, WaitTime: "1s", MaxWaitTime: "10s", IsAllowNonIdempotent: true,
			}}})
			assert.Equal(t, 3, r.ConfigREST.Retry.Count)
			assert.Equal(t, DurationStr("1s"), r.ConfigREST.Retry.WaitTime)
			assert.Equal(t, DurationStr("10s"), r.ConfigREST.Retry.MaxWaitTime)
			assert.True(t, r.ConfigREST.Retry.IsAllowNonIdempotent)
		})
		t.Run("retry dst kept", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{Retry: Retry{Count: 5, WaitTime: "2s"}}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{Retry: Retry{Count: 3, WaitTime: "1s"}}})
			assert.Equal(t, 5, r.ConfigREST.Retry.Count)
			assert.Equal(t, DurationStr("2s"), r.ConfigREST.Retry.WaitTime)
		})
		t.Run("merge flags", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{
				UnescapeQuery: true, UnescapeJSON: true, ResponseBodyLimit: 100,
			}})
			assert.True(t, r.ConfigREST.UnescapeQuery)
			assert.True(t, r.ConfigREST.UnescapeJSON)
			assert.Equal(t, int64(100), r.ConfigREST.ResponseBodyLimit)
		})
		t.Run("empty maps take parent", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{
				Tests: map[string]string{"t": "1"},
				Path:  map[string]string{"p": "/x"},
			}})
			assert.Equal(t, map[string]string{"t": "1"}, r.ConfigREST.Tests)
			assert.Equal(t, map[string]string{"p": "/x"}, r.ConfigREST.Path)
		})
		t.Run("retry partial fill", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{Retry: Retry{Count: 5}}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{Retry: Retry{
				Count: 3, WaitTime: "1s", MaxWaitTime: "10s", IsAllowNonIdempotent: true,
			}}})
			assert.Equal(t, 5, r.ConfigREST.Retry.Count)
			assert.Equal(t, DurationStr("1s"), r.ConfigREST.Retry.WaitTime)
			assert.Equal(t, DurationStr("10s"), r.ConfigREST.Retry.MaxWaitTime)
			assert.True(t, r.ConfigREST.Retry.IsAllowNonIdempotent)
		})
		t.Run("flags dst kept", func(t *testing.T) {
			r := &ReqConfig{ConfigREST: &ConfigREST{UnescapeQuery: true, UnescapeJSON: true, ResponseBodyLimit: 100}}
			r.Merge(&ReqConfig{ConfigREST: &ConfigREST{UnescapeQuery: false, UnescapeJSON: false, ResponseBodyLimit: 200}})
			assert.True(t, r.ConfigREST.UnescapeQuery)
			assert.True(t, r.ConfigREST.UnescapeJSON)
			assert.Equal(t, int64(100), r.ConfigREST.ResponseBodyLimit)
		})
	})

	t.Run("config grpc", func(t *testing.T) {
		t.Run("nil parent noop", func(t *testing.T) {
			own := &ConfigGRPC{ReadMaxBytes: 1}
			r := &ReqConfig{ConfigGRPC: own}
			r.Merge(&ReqConfig{})
			assert.Same(t, own, r.ConfigGRPC)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConfig{}
			p := &ConfigGRPC{ReadMaxBytes: 1}
			r.Merge(&ReqConfig{ConfigGRPC: p})
			assert.Same(t, p, r.ConfigGRPC)
		})
		t.Run("empty fields filled", func(t *testing.T) {
			r := &ReqConfig{ConfigGRPC: &ConfigGRPC{}}
			r.Merge(&ReqConfig{ConfigGRPC: &ConfigGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 3,
				SendGzip:          true,
			}})
			assert.Equal(t, &ConfigGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 3,
				SendGzip:          true,
			}, r.ConfigGRPC)
		})
		t.Run("set values kept", func(t *testing.T) {
			own := &ConfigGRPC{
				ReadMaxBytes:      10,
				SendMaxBytes:      20,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 30,
				SendGzip:          true,
			}
			r := &ReqConfig{ConfigGRPC: own}
			r.Merge(&ReqConfig{ConfigGRPC: &ConfigGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      false,
				EnableHTTPGet:     false,
				HTTPGetMaxURLSize: 3,
				SendGzip:          false,
			}})
			assert.Equal(t, own, r.ConfigGRPC)
			// Merge reassigns a copy, never mutating or reusing the original.
			assert.NotSame(t, own, r.ConfigGRPC)
		})
	})
	t.Run("config ws", func(t *testing.T) {
		t.Run("nil parent noop", func(t *testing.T) {
			own := &ConfigWS{CompressionMode: "disabled"}
			r := &ReqConfig{ConfigWS: own}
			r.Merge(&ReqConfig{})
			assert.Same(t, own, r.ConfigWS)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConfig{}
			p := &ConfigWS{CompressionMode: "disabled"}
			r.Merge(&ReqConfig{ConfigWS: p})
			assert.Same(t, p, r.ConfigWS)
		})
		t.Run("empty fields filled", func(t *testing.T) {
			r := &ReqConfig{ConfigWS: &ConfigWS{}}
			r.Merge(&ReqConfig{ConfigWS: &ConfigWS{
				Subprotocols:         []string{"a", "b"},
				CompressionMode:      "context_takeover",
				CompressionThreshold: 100,
				OriginPatterns:       []string{"x"},
				InsecureSkipVerify:   true,
			}})
			assert.Equal(t, &ConfigWS{
				Subprotocols:         []string{"a", "b"},
				CompressionMode:      "context_takeover",
				CompressionThreshold: 100,
				OriginPatterns:       []string{"x"},
				InsecureSkipVerify:   true,
			}, r.ConfigWS)
		})
		t.Run("scalars dst kept", func(t *testing.T) {
			r := &ReqConfig{ConfigWS: &ConfigWS{
				CompressionMode:      "no_context_takeover",
				CompressionThreshold: 50,
				InsecureSkipVerify:   true,
			}}
			r.Merge(&ReqConfig{ConfigWS: &ConfigWS{
				CompressionMode:      "disabled",
				CompressionThreshold: 10,
				InsecureSkipVerify:   false,
			}})
			assert.Equal(t, "no_context_takeover", r.ConfigWS.CompressionMode)
			assert.Equal(t, 50, r.ConfigWS.CompressionThreshold)
			assert.True(t, r.ConfigWS.InsecureSkipVerify)
		})
		t.Run("subprotocols union", func(t *testing.T) {
			r := &ReqConfig{ConfigWS: &ConfigWS{Subprotocols: []string{"a"}}}
			r.Merge(&ReqConfig{ConfigWS: &ConfigWS{Subprotocols: []string{"a", "b"}}})
			assert.Equal(t, []string{"a", "b"}, r.ConfigWS.Subprotocols)
		})
		t.Run("origin patterns union", func(t *testing.T) {
			r := &ReqConfig{ConfigWS: &ConfigWS{OriginPatterns: []string{"a"}}}
			r.Merge(&ReqConfig{ConfigWS: &ConfigWS{OriginPatterns: []string{"a", "b"}}})
			assert.Equal(t, []string{"a", "b"}, r.ConfigWS.OriginPatterns)
		})
	})
}

func TestRequestParseRESTRawURL(t *testing.T) {
	restReq := func() *Request { return &Request{ReqMeta: ReqMeta{Protocol: ProtocolREST}} }

	t.Run("empty url", func(t *testing.T) {
		err := restReq().ParseRESTRawURL("")
		assert.ErrorIs(t, err, ErrEmptyValue)
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		for _, p := range []ProtocolKind{ProtocolGRPC, ProtocolWS} {
			r := &Request{ReqMeta: ReqMeta{Protocol: p}}
			assert.Error(t, r.ParseRESTRawURL("http://x/p"))
		}
	})

	t.Run("rest and gql ok", func(t *testing.T) {
		for _, p := range []ProtocolKind{ProtocolREST, ProtocolGQL} {
			r := &Request{ReqMeta: ReqMeta{Protocol: p}}
			assert.NoError(t, r.ParseRESTRawURL("http://example.com/p"))
		}
	})

	t.Run("adds http scheme", func(t *testing.T) {
		r := restReq()
		require.NoError(t, r.ParseRESTRawURL("example.com/p"))
		assert.Equal(t, "http://example.com/p", r.URL)
	})

	t.Run("keeps https scheme", func(t *testing.T) {
		r := restReq()
		require.NoError(t, r.ParseRESTRawURL("https://example.com/p"))
		assert.Equal(t, "https://example.com/p", r.URL)
	})

	t.Run("extracts path params", func(t *testing.T) {
		r := restReq()
		require.NoError(t, r.ParseRESTRawURL("http://x/users/{id}/posts/{pid}"))
		assert.Equal(t, map[string]string{"id": "", "pid": ""}, r.ConfigREST.Path)
	})

	t.Run("preserves existing path values", func(t *testing.T) {
		r := restReq()
		r.ConfigREST = &ConfigREST{Path: map[string]string{"id": "42", "old": "x"}}
		require.NoError(t, r.ParseRESTRawURL("http://x/users/{id}"))
		assert.Equal(t, map[string]string{"id": "42"}, r.ConfigREST.Path)
	})

	t.Run("parses query", func(t *testing.T) {
		r := restReq()
		require.NoError(t, r.ParseRESTRawURL("http://x/p?a=1&b=2&a=3"))
		assert.Equal(t, map[string][]string{"a": {"1", "3"}, "b": {"2"}}, r.Query)
	})

	t.Run("no query sets empty map", func(t *testing.T) {
		r := restReq()
		require.NoError(t, r.ParseRESTRawURL("http://x/p"))
		require.NotNil(t, r.Query)
		assert.Empty(t, r.Query)
	})
}

func TestRequestRawURL(t *testing.T) {
	t.Run("no query", func(t *testing.T) {
		r := &Request{URL: "http://x/p"}
		assert.Equal(t, "http://x/p", r.RawURL())
	})
	t.Run("empty query map", func(t *testing.T) {
		r := &Request{URL: "http://x/p", ReqConfig: ReqConfig{Query: map[string][]string{}}}
		assert.Equal(t, "http://x/p", r.RawURL())
	})
	t.Run("sorted by key", func(t *testing.T) {
		r := &Request{URL: "http://x/p", ReqConfig: ReqConfig{Query: map[string][]string{
			"b": {"2"}, "a": {"1"},
		}}}
		assert.Equal(t, "http://x/p?a=1&b=2", r.RawURL())
	})
	t.Run("multi values", func(t *testing.T) {
		r := &Request{URL: "http://x/p", ReqConfig: ReqConfig{Query: map[string][]string{
			"a": {"1", "2"},
		}}}
		assert.Equal(t, "http://x/p?a=1&a=2", r.RawURL())
	})
}

func TestToQueryString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", toQueryString(nil))
		assert.Equal(t, "", toQueryString(map[string][]string{}))
	})
	t.Run("sorted", func(t *testing.T) {
		assert.Equal(t, "a=1&b=2", toQueryString(map[string][]string{"b": {"2"}, "a": {"1"}}))
	})
	t.Run("multi values", func(t *testing.T) {
		assert.Equal(t, "a=1&a=2&a=3", toQueryString(map[string][]string{"a": {"1", "2", "3"}}))
	})
	t.Run("values not encoded", func(t *testing.T) {
		assert.Equal(t, "a={x}", toQueryString(map[string][]string{"a": {"{x}"}}))
	})
}
