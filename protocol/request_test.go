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
	seed := func() *ReqConf {
		return &ReqConf{
			Desc:    "d",
			Auth:    Auth{Kind: AuthToken, Token: AuthDataToken{Header: "h", Scheme: "Bearer", Value: "v"}},
			Query:   map[string][]string{"a": {"1", "2"}},
			Header:  map[string][]string{"h": {"x"}},
			Timeout: "5s",
			ConfREST: &ConfREST{
				Tests: map[string]string{"t": "1"},
				Path:  map[string]string{"p": "/"},
			},
			ConfGRPC: &ConfGRPC{},
			ConfWS:   &ConfWS{},
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
		assert.Equal(t, src.ConfREST.Tests, dst.ConfREST.Tests)
		assert.Equal(t, src.ConfREST.Path, dst.ConfREST.Path)
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
		assert.NotSame(t, src.ConfREST, dst.ConfREST)
		assert.NotSame(t, src.ConfGRPC, dst.ConfGRPC)
		assert.NotSame(t, src.ConfWS, dst.ConfWS)
	})
}

func TestReqConfigMerge(t *testing.T) {
	t.Run("nil parent noop", func(t *testing.T) {
		r := &ReqConf{Desc: "d"}
		r.Merge(nil)
		assert.Equal(t, "d", r.Desc)
	})

	t.Run("auth", func(t *testing.T) {
		t.Run("inherited overridden", func(t *testing.T) {
			r := &ReqConf{Auth: Auth{Kind: AuthInherited}}
			p := &ReqConf{Auth: Auth{Kind: AuthBasic, Creds: AuthDataCreds{Username: "u", Password: "p"}}}
			r.Merge(p)
			assert.Equal(t, p.Auth, r.Auth)
		})
		t.Run("explicit kept", func(t *testing.T) {
			r := &ReqConf{Auth: Auth{Kind: AuthNone}}
			r.Merge(&ReqConf{Auth: Auth{Kind: AuthBasic}})
			assert.Equal(t, AuthNone, r.Auth.Kind)
		})
	})

	t.Run("header", func(t *testing.T) {
		t.Run("parent empty noop", func(t *testing.T) {
			r := &ReqConf{Header: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConf{})
			assert.Equal(t, map[string][]string{"a": {"1"}}, r.Header)
		})
		t.Run("self empty takes parent", func(t *testing.T) {
			r := &ReqConf{}
			p := map[string][]string{"a": {"1"}}
			r.Merge(&ReqConf{Header: p})
			assert.Equal(t, p, r.Header)
		})
		t.Run("dst preferred", func(t *testing.T) {
			r := &ReqConf{Header: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConf{Header: map[string][]string{"a": {"9"}, "b": {"2"}}})
			assert.Equal(t, map[string][]string{"a": {"1"}, "b": {"2"}}, r.Header)
		})
	})

	t.Run("query", func(t *testing.T) {
		t.Run("parent empty noop", func(t *testing.T) {
			r := &ReqConf{Query: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConf{})
			assert.Equal(t, map[string][]string{"a": {"1"}}, r.Query)
		})
		t.Run("dst preferred", func(t *testing.T) {
			r := &ReqConf{Query: map[string][]string{"a": {"1"}}}
			r.Merge(&ReqConf{Query: map[string][]string{"a": {"9"}, "b": {"2"}}})
			assert.Equal(t, map[string][]string{"a": {"1"}, "b": {"2"}}, r.Query)
		})
	})

	t.Run("timeout", func(t *testing.T) {
		t.Run("empty takes parent", func(t *testing.T) {
			r := &ReqConf{}
			r.Merge(&ReqConf{Timeout: "5s"})
			assert.Equal(t, DurationStr("5s"), r.Timeout)
		})
		t.Run("set kept", func(t *testing.T) {
			r := &ReqConf{Timeout: "1s"}
			r.Merge(&ReqConf{Timeout: "5s"})
			assert.Equal(t, DurationStr("1s"), r.Timeout)
		})
	})

	t.Run("config rest", func(t *testing.T) {
		t.Run("parent nil noop", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{}}
			r.Merge(&ReqConf{})
			assert.NotNil(t, r.ConfREST)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConf{}
			p := &ConfREST{Tests: map[string]string{"t": "1"}}
			r.Merge(&ReqConf{ConfREST: p})
			assert.Same(t, p, r.ConfREST)
		})
		t.Run("merge tests", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{Tests: map[string]string{"a": "1"}}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{Tests: map[string]string{"a": "9", "b": "2"}}})
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, r.ConfREST.Tests)
		})
		t.Run("merge path", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{Path: map[string]string{"a": "1"}}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{Path: map[string]string{"a": "9", "b": "2"}}})
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, r.ConfREST.Path)
		})
		t.Run("merge retry", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{Retry: Retry{
				Count: 3, WaitTime: "1s", MaxWaitTime: "10s", IsAllowNonIdempotent: true,
			}}})
			assert.Equal(t, 3, r.ConfREST.Retry.Count)
			assert.Equal(t, DurationStr("1s"), r.ConfREST.Retry.WaitTime)
			assert.Equal(t, DurationStr("10s"), r.ConfREST.Retry.MaxWaitTime)
			assert.True(t, r.ConfREST.Retry.IsAllowNonIdempotent)
		})
		t.Run("retry dst kept", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{Retry: Retry{Count: 5, WaitTime: "2s"}}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{Retry: Retry{Count: 3, WaitTime: "1s"}}})
			assert.Equal(t, 5, r.ConfREST.Retry.Count)
			assert.Equal(t, DurationStr("2s"), r.ConfREST.Retry.WaitTime)
		})
		t.Run("merge flags", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{
				UnescapeQuery: true, UnescapeJSON: true, ResponseBodyLimit: 100,
			}})
			assert.True(t, r.ConfREST.UnescapeQuery)
			assert.True(t, r.ConfREST.UnescapeJSON)
			assert.Equal(t, int64(100), r.ConfREST.ResponseBodyLimit)
		})
		t.Run("empty maps take parent", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{
				Tests: map[string]string{"t": "1"},
				Path:  map[string]string{"p": "/x"},
			}})
			assert.Equal(t, map[string]string{"t": "1"}, r.ConfREST.Tests)
			assert.Equal(t, map[string]string{"p": "/x"}, r.ConfREST.Path)
		})
		t.Run("retry partial fill", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{Retry: Retry{Count: 5}}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{Retry: Retry{
				Count: 3, WaitTime: "1s", MaxWaitTime: "10s", IsAllowNonIdempotent: true,
			}}})
			assert.Equal(t, 5, r.ConfREST.Retry.Count)
			assert.Equal(t, DurationStr("1s"), r.ConfREST.Retry.WaitTime)
			assert.Equal(t, DurationStr("10s"), r.ConfREST.Retry.MaxWaitTime)
			assert.True(t, r.ConfREST.Retry.IsAllowNonIdempotent)
		})
		t.Run("flags dst kept", func(t *testing.T) {
			r := &ReqConf{ConfREST: &ConfREST{UnescapeQuery: true, UnescapeJSON: true, ResponseBodyLimit: 100}}
			r.Merge(&ReqConf{ConfREST: &ConfREST{UnescapeQuery: false, UnescapeJSON: false, ResponseBodyLimit: 200}})
			assert.True(t, r.ConfREST.UnescapeQuery)
			assert.True(t, r.ConfREST.UnescapeJSON)
			assert.Equal(t, int64(100), r.ConfREST.ResponseBodyLimit)
		})
	})

	t.Run("config grpc", func(t *testing.T) {
		t.Run("nil parent noop", func(t *testing.T) {
			own := &ConfGRPC{ReadMaxBytes: 1}
			r := &ReqConf{ConfGRPC: own}
			r.Merge(&ReqConf{})
			assert.Same(t, own, r.ConfGRPC)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConf{}
			p := &ConfGRPC{ReadMaxBytes: 1}
			r.Merge(&ReqConf{ConfGRPC: p})
			assert.Same(t, p, r.ConfGRPC)
		})
		t.Run("empty fields filled", func(t *testing.T) {
			r := &ReqConf{ConfGRPC: &ConfGRPC{}}
			r.Merge(&ReqConf{ConfGRPC: &ConfGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 3,
				SendGzip:          true,
			}})
			assert.Equal(t, &ConfGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 3,
				SendGzip:          true,
			}, r.ConfGRPC)
		})
		t.Run("set values kept", func(t *testing.T) {
			own := &ConfGRPC{
				ReadMaxBytes:      10,
				SendMaxBytes:      20,
				UseProtoJSON:      true,
				EnableHTTPGet:     true,
				HTTPGetMaxURLSize: 30,
				SendGzip:          true,
			}
			r := &ReqConf{ConfGRPC: own}
			r.Merge(&ReqConf{ConfGRPC: &ConfGRPC{
				ReadMaxBytes:      1,
				SendMaxBytes:      2,
				UseProtoJSON:      false,
				EnableHTTPGet:     false,
				HTTPGetMaxURLSize: 3,
				SendGzip:          false,
			}})
			assert.Equal(t, own, r.ConfGRPC)
			// Merge reassigns a copy, never mutating or reusing the original.
			assert.NotSame(t, own, r.ConfGRPC)
		})
	})
	t.Run("config ws", func(t *testing.T) {
		t.Run("nil parent noop", func(t *testing.T) {
			own := &ConfWS{CompressionMode: "disabled"}
			r := &ReqConf{ConfWS: own}
			r.Merge(&ReqConf{})
			assert.Same(t, own, r.ConfWS)
		})
		t.Run("self nil takes parent", func(t *testing.T) {
			r := &ReqConf{}
			p := &ConfWS{CompressionMode: "disabled"}
			r.Merge(&ReqConf{ConfWS: p})
			assert.Same(t, p, r.ConfWS)
		})
		t.Run("empty fields filled", func(t *testing.T) {
			r := &ReqConf{ConfWS: &ConfWS{}}
			r.Merge(&ReqConf{ConfWS: &ConfWS{
				Subprotocols:         []string{"a", "b"},
				CompressionMode:      "context_takeover",
				CompressionThreshold: 100,
				OriginPatterns:       []string{"x"},
				InsecureSkipVerify:   true,
			}})
			assert.Equal(t, &ConfWS{
				Subprotocols:         []string{"a", "b"},
				CompressionMode:      "context_takeover",
				CompressionThreshold: 100,
				OriginPatterns:       []string{"x"},
				InsecureSkipVerify:   true,
			}, r.ConfWS)
		})
		t.Run("scalars dst kept", func(t *testing.T) {
			r := &ReqConf{ConfWS: &ConfWS{
				CompressionMode:      "no_context_takeover",
				CompressionThreshold: 50,
				InsecureSkipVerify:   true,
			}}
			r.Merge(&ReqConf{ConfWS: &ConfWS{
				CompressionMode:      "disabled",
				CompressionThreshold: 10,
				InsecureSkipVerify:   false,
			}})
			assert.Equal(t, "no_context_takeover", r.ConfWS.CompressionMode)
			assert.Equal(t, 50, r.ConfWS.CompressionThreshold)
			assert.True(t, r.ConfWS.InsecureSkipVerify)
		})
		t.Run("subprotocols union", func(t *testing.T) {
			r := &ReqConf{ConfWS: &ConfWS{Subprotocols: []string{"a"}}}
			r.Merge(&ReqConf{ConfWS: &ConfWS{Subprotocols: []string{"a", "b"}}})
			assert.Equal(t, []string{"a", "b"}, r.ConfWS.Subprotocols)
		})
		t.Run("origin patterns union", func(t *testing.T) {
			r := &ReqConf{ConfWS: &ConfWS{OriginPatterns: []string{"a"}}}
			r.Merge(&ReqConf{ConfWS: &ConfWS{OriginPatterns: []string{"a", "b"}}})
			assert.Equal(t, []string{"a", "b"}, r.ConfWS.OriginPatterns)
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
		assert.Equal(t, map[string]string{"id": "", "pid": ""}, r.ConfREST.Path)
	})

	t.Run("preserves existing path values", func(t *testing.T) {
		r := restReq()
		r.ConfREST = &ConfREST{Path: map[string]string{"id": "42", "old": "x"}}
		require.NoError(t, r.ParseRESTRawURL("http://x/users/{id}"))
		assert.Equal(t, map[string]string{"id": "42"}, r.ConfREST.Path)
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
		r := &Request{URL: "http://x/p", ReqConf: ReqConf{Query: map[string][]string{}}}
		assert.Equal(t, "http://x/p", r.RawURL())
	})
	t.Run("sorted by key", func(t *testing.T) {
		r := &Request{URL: "http://x/p", ReqConf: ReqConf{Query: map[string][]string{
			"b": {"2"}, "a": {"1"},
		}}}
		assert.Equal(t, "http://x/p?a=1&b=2", r.RawURL())
	})
	t.Run("multi values", func(t *testing.T) {
		r := &Request{URL: "http://x/p", ReqConf: ReqConf{Query: map[string][]string{
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
