// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/oklog/ulid/v2"
)

type ProtocolKind string

const (
	// kinds of protocol
	ProtocolREST ProtocolKind = "rest"
	ProtocolGRPC ProtocolKind = "grpc"
	ProtocolWS   ProtocolKind = "ws"
	ProtocolGQL  ProtocolKind = "gql"
)

var Protocols = []ProtocolKind{ProtocolREST, ProtocolGRPC, ProtocolWS, ProtocolGQL}

type ReqMeta struct {
	Id       ulid.ULID    `toml:"id" json:"id"`
	Protocol ProtocolKind `toml:"protocol" json:"protocol"`
}

type AuthKind string

const (
	AuthInherited AuthKind = ""
	AuthNone      AuthKind = "none"
	AuthBasic     AuthKind = "basic"
	AuthDigest    AuthKind = "digest"
	AuthToken     AuthKind = "token"
)

var AuthKinds = []AuthKind{AuthInherited, AuthNone, AuthBasic, AuthDigest, AuthToken}

type (
	Auth struct {
		Kind  AuthKind
		Token AuthDataToken
		Creds AuthDataCreds
	}
	AuthDataToken struct {
		Header string `toml:"header,omitempty" json:"header"`
		Scheme string `toml:"scheme,omitempty" json:"scheme"`
		Value  string `toml:"val,omitempty" json:"value"`
	}
	AuthDataCreds struct {
		Username string `toml:"username" json:"username"`
		Password string `toml:"password" json:"password"`
	}
)

// the config of request and can be merge
//
// readonly,so that it does not cause data races except [ReqConfig.Merge]
type ReqConfig struct {
	// do not inherited
	Desc string

	Auth    Auth
	Query   map[string][]string
	Header  map[string][]string
	Timeout DurationStr

	ConfigREST *ConfigREST
	ConfigGRPC *ConfigGRPC
	ConfigWS   *ConfigWS
}

func (r *ReqConfig) Copy() *ReqConfig {
	dst := *r

	dst.Query = cloneSliceMap(r.Query)
	dst.Header = cloneSliceMap(r.Header)

	dst.ConfigREST = r.ConfigREST.Copy()
	dst.ConfigGRPC = r.ConfigGRPC.Copy()
	dst.ConfigWS = r.ConfigWS.Copy()

	return &dst
}

// inherit required value from parent if the value is default
func (r *ReqConfig) Merge(p *ReqConfig) {
	if p == nil {
		return
	}
	r.mergeAuth(p.Auth)
	r.mergeHeader(p.Header)
	r.mergeQuery(p.Query)
	r.mergeTimeout(p.Timeout)
	r.mergeConfigREST(p.ConfigREST)
	r.mergeConfigGRPC(p.ConfigGRPC)
	r.mergeConfigWS(p.ConfigWS)
}
func (r *ReqConfig) mergeAuth(auth Auth) {
	if r.Auth.Kind != AuthInherited {
		return
	}
	r.Auth = auth
}
func cloneSliceMap[M ~map[K][]V, K comparable, V any](src M) M {
	dst := make(M, len(src))
	for k, v := range src {
		dst[k] = slices.Clone(v)
	}
	return dst
}
func mergeMap[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}
func (r *ReqConfig) mergeQuery(query map[string][]string) {
	if len(query) <= 0 {
		return
	}
	if len(r.Query) <= 0 {
		r.Query = query
		return
	}
	mergeMap(r.Query, query)
}
func (r *ReqConfig) mergeHeader(header map[string][]string) {
	if len(header) <= 0 {
		return
	}
	if len(r.Header) <= 0 {
		r.Header = header
		return
	}
	mergeMap(r.Header, header)
}
func (r *ReqConfig) mergeTimeout(durationStr DurationStr) {
	if r.Timeout != "" {
		return
	}
	r.Timeout = durationStr
}
func (r *ReqConfig) mergeConfigREST(c *ConfigREST) {
	if c == nil {
		return
	}
	if r.ConfigREST == nil {
		r.ConfigREST = c
		return
	}
	current := r.ConfigREST
	// merge Tests
	if len(c.Tests) > 0 {
		if len(current.Tests) <= 0 {
			current.Tests = c.Tests
		} else {
			mergeMap(current.Tests, c.Tests)
		}
	}
	// merge Path
	if len(c.Path) > 0 {
		if len(current.Path) <= 0 {
			current.Path = c.Path
		} else {
			mergeMap(current.Path, c.Path)
		}
	}
	// merge Retry
	retry := c.Retry
	if cout := retry.Count; cout >= 0 && current.Retry.Count == 0 {
		current.Retry.Count = cout
	}
	if wt := retry.WaitTime; wt != "" && current.Retry.WaitTime == "" {
		current.Retry.WaitTime = wt
	}
	if mwt := retry.MaxWaitTime; mwt != "" && current.Retry.MaxWaitTime == "" {
		current.Retry.MaxWaitTime = mwt
	}
	if isAllow := retry.IsAllowNonIdempotent; isAllow && !current.Retry.IsAllowNonIdempotent {
		current.Retry.IsAllowNonIdempotent = isAllow
	}
	// merge other
	if uq := c.UnescapeQuery; uq && !current.UnescapeQuery {
		current.UnescapeQuery = uq
	}
	if uj := c.UnescapeJson; uj && !current.UnescapeJson {
		current.UnescapeJson = uj
	}
	if rbl := c.ResponseBodyLimit; rbl > 0 && current.ResponseBodyLimit <= 0 {
		current.ResponseBodyLimit = rbl
	}
}
func (r *ReqConfig) mergeConfigGRPC(c *ConfigGRPC) {
	// TODO:merge configGRPC in ReqConfig
	panic("TODO panic: merge configGRPC in ReqConfig")
}
func (r *ReqConfig) mergeConfigWS(c *ConfigWS) {
	// TODO:merge configWS in ReqConfig
	panic("TODO panic: merge configWS in ReqConfig")
}

type Request struct {
	ReqMeta
	ReqConfig

	URL    string
	Method Method

	PayloadREST Body

	PreReqExpr string
}

func (r *Request) ParseRESTRawURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("rawURL:%w", ErrEmptyValue)
	}
	if r.Protocol != ProtocolREST && r.Protocol != ProtocolGQL {
		return fmt.Errorf("parse raw URL not support %s", r.Protocol)
	}
	// if not has http scheme
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	s := strings.SplitN(rawURL, "?", 2)
	// set URL
	r.URL = s[0]
	// Set PathParams
	if r.ConfigREST == nil {
		r.ConfigREST = &ConfigREST{}
	}
	pathMap := make(map[string]string)
	for _, matched := range pathParamRe.FindAllStringSubmatch(r.URL, -1) {
		name := matched[1]
		pathMap[name] = ""
		if r.ConfigREST.Path == nil {
			continue
		}
		if v, ok := r.ConfigREST.Path[name]; ok {
			pathMap[name] = v
		}
	}
	r.ConfigREST.Path = pathMap
	// set QueryParams
	if len(s) > 1 {
		query, err := url.ParseQuery(s[1])
		if err != nil {
			return err
		}
		r.Query = query
	} else {
		r.Query = map[string][]string{}
	}

	return nil
}

func (r *Request) RawURL() string {
	if len(r.Query) > 0 {
		return r.URL + "?" + toQueryString(r.Query)
	}
	return r.URL
}

// into URL query string without “URL encoded” form
// ("bar=baz&foo=quux&{=}") sorted by key.
func toQueryString(v map[string][]string) string {
	if len(v) == 0 {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, len(v))

	var i int
	for k := range v {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	for _, k := range keys {
		vs := v[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}
