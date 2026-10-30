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
	ID       ulid.ULID    `toml:"id" json:"id"`
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
		Kind  AuthKind      `json:"kind" toml:"kind"`
		Token AuthDataToken `json:"token" toml:"token"`
		Creds AuthDataCreds `json:"creds" toml:"creds"`
	}
	AuthDataToken struct {
		Header string `json:"header" toml:"header"`
		Scheme string `json:"scheme" toml:"scheme"`
		Value  string `json:"value" toml:"value"`
	}
	AuthDataCreds struct {
		Username string `json:"username" toml:"username"`
		Password string `json:"password" toml:"password"`
	}
)

// the config of request and can be merge
//
// readonly,so that it does not cause data races except [ReqConf.Merge]
type ReqConf struct {
	// do not inherited
	Desc string `json:"desc" toml:"desc"`

	Auth    Auth                `json:"auth" toml:"auth"`
	Query   map[string][]string `json:"query" toml:"query"`
	Header  map[string][]string `json:"header" toml:"header"`
	Timeout DurationStr         `json:"timeout" toml:"timeout"`

	ConfREST *ConfREST `json:"confREST" toml:"conf_rest"`
	ConfGRPC *ConfGRPC `json:"confGRPC" toml:"conf_grpc"`
	ConfWS   *ConfWS   `json:"confWS" toml:"conf_ws"`
}

func (r *ReqConf) Copy() *ReqConf {
	dst := *r

	dst.Query = cloneSliceMap(r.Query)
	dst.Header = cloneSliceMap(r.Header)

	if r.ConfREST != nil {
		dst.ConfREST = r.ConfREST.Copy()
	}
	if r.ConfGRPC != nil {
		dst.ConfGRPC = r.ConfGRPC.Copy()
	}
	if r.ConfWS != nil {
		dst.ConfWS = r.ConfWS.Copy()
	}

	return &dst
}

// inherit required value from parent if the value is default
func (r *ReqConf) Merge(p *ReqConf) {
	if p == nil {
		return
	}
	r.mergeAuth(p.Auth)
	r.mergeHeader(p.Header)
	r.mergeQuery(p.Query)
	r.mergeTimeout(p.Timeout)
	r.mergeConfigREST(p.ConfREST)
	r.mergeConfigGRPC(p.ConfGRPC)
	r.mergeConfigWS(p.ConfWS)
}
func (r *ReqConf) mergeAuth(auth Auth) {
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
func (r *ReqConf) mergeQuery(query map[string][]string) {
	if len(query) <= 0 {
		return
	}
	if len(r.Query) <= 0 {
		r.Query = query
		return
	}
	mergeMap(r.Query, query)
}
func (r *ReqConf) mergeHeader(header map[string][]string) {
	if len(header) <= 0 {
		return
	}
	if len(r.Header) <= 0 {
		r.Header = header
		return
	}
	mergeMap(r.Header, header)
}
func (r *ReqConf) mergeTimeout(durationStr DurationStr) {
	if r.Timeout != "" {
		return
	}
	r.Timeout = durationStr
}
func (r *ReqConf) mergeConfigREST(c *ConfREST) {
	if c == nil {
		return
	}
	if r.ConfREST == nil {
		r.ConfREST = c
		return
	}
	r.ConfREST = r.ConfREST.Copy()
	cur := r.ConfREST
	// merge Tests
	if len(c.Tests) > 0 {
		if len(cur.Tests) <= 0 {
			cur.Tests = c.Tests
		} else {
			mergeMap(cur.Tests, c.Tests)
		}
	}
	// merge Path
	if len(c.Path) > 0 {
		if len(cur.Path) <= 0 {
			cur.Path = c.Path
		} else {
			mergeMap(cur.Path, c.Path)
		}
	}
	// merge Retry
	retry := c.Retry
	if cout := retry.Count; cout >= 0 && cur.Retry.Count == 0 {
		cur.Retry.Count = cout
	}
	if wt := retry.WaitTime; wt != "" && cur.Retry.WaitTime == "" {
		cur.Retry.WaitTime = wt
	}
	if mwt := retry.MaxWaitTime; mwt != "" && cur.Retry.MaxWaitTime == "" {
		cur.Retry.MaxWaitTime = mwt
	}
	if isAllow := retry.IsAllowNonIdempotent; isAllow && !cur.Retry.IsAllowNonIdempotent {
		cur.Retry.IsAllowNonIdempotent = isAllow
	}
	// merge other
	if uq := c.UnescapeQuery; uq && !cur.UnescapeQuery {
		cur.UnescapeQuery = uq
	}
	if uj := c.UnescapeJSON; uj && !cur.UnescapeJSON {
		cur.UnescapeJSON = uj
	}
	if rbl := c.ResponseBodyLimit; rbl > 0 && cur.ResponseBodyLimit <= 0 {
		cur.ResponseBodyLimit = rbl
	}
}
func (r *ReqConf) mergeConfigGRPC(c *ConfGRPC) {
	if c == nil {
		return
	}
	if r.ConfGRPC == nil {
		r.ConfGRPC = c
		return
	}
	cur := r.ConfGRPC.Copy()
	r.ConfGRPC = cur
	// ReadMaxBytes
	if cur.ReadMaxBytes == 0 {
		cur.ReadMaxBytes = c.ReadMaxBytes
	}
	// SendMaxBytes
	if cur.SendMaxBytes == 0 {
		cur.SendMaxBytes = c.SendMaxBytes
	}
	// UseProtoJSON
	if !cur.UseProtoJSON {
		cur.UseProtoJSON = c.UseProtoJSON
	}
	// EnableHTTPGet
	if !cur.EnableHTTPGet {
		cur.EnableHTTPGet = c.EnableHTTPGet
	}
	// HTTPGetMaxURLSize
	if cur.HTTPGetMaxURLSize == 0 {
		cur.HTTPGetMaxURLSize = c.HTTPGetMaxURLSize
	}
	// SendGzip
	if !cur.SendGzip {
		cur.SendGzip = c.SendGzip
	}
}
func (r *ReqConf) mergeConfigWS(c *ConfWS) {
	if c == nil {
		return
	}
	if r.ConfWS == nil {
		r.ConfWS = c
		return
	}
	cur := r.ConfWS.Copy()
	r.ConfWS = cur
	// Subprotocols
	if len(c.Subprotocols) > 0 {
		if len(cur.Subprotocols) <= 0 {
			cur.Subprotocols = c.Subprotocols
		} else {
			for _, v := range c.Subprotocols {
				if !slices.Contains(cur.Subprotocols, v) {
					cur.Subprotocols = append(cur.Subprotocols, v)
				}
			}
		}
	}
	// CompressionMode
	if cur.CompressionMode == "" {
		cur.CompressionMode = c.CompressionMode
	}
	// CompressionThreshold
	if cur.CompressionThreshold == 0 {
		cur.CompressionThreshold = c.CompressionThreshold
	}
	// OriginPatterns
	if len(c.OriginPatterns) > 0 {
		if len(cur.OriginPatterns) <= 0 {
			cur.OriginPatterns = c.OriginPatterns
		} else {
			for _, v := range c.OriginPatterns {
				if !slices.Contains(cur.OriginPatterns, v) {
					cur.OriginPatterns = append(cur.OriginPatterns, v)
				}
			}
		}
	}
	// InsecureSkipVerify
	if !cur.InsecureSkipVerify {
		cur.InsecureSkipVerify = c.InsecureSkipVerify
	}
}

type Request struct {
	ReqMeta
	ReqConf

	URL    string `json:"url" toml:"url"`
	Method Method `json:"method" toml:"method"`

	PayloadREST Body `json:"payloadREST" toml:"payload_rest"`

	PreReqExpr string `json:"preReqExpr" toml:"pre_req_expr"`
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
	if r.ConfREST == nil {
		r.ConfREST = &ConfREST{}
	}
	pathMap := make(map[string]string)
	for _, matched := range pathParamRe.FindAllStringSubmatch(r.URL, -1) {
		name := matched[1]
		pathMap[name] = ""
		if r.ConfREST.Path == nil {
			continue
		}
		if v, ok := r.ConfREST.Path[name]; ok {
			pathMap[name] = v
		}
	}
	r.ConfREST.Path = pathMap
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
