// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"context"
)

type Runner interface {
	Run(...context.Context) Result
	RunWithContext(ctx context.Context) Result
	ReqName() string
	ReqProtocol() ProtocolKind
}

type ProtocolKind = string

const (
	// kinds of protocol
	ProtocolREST ProtocolKind = "rest"
	ProtocolGRPC ProtocolKind = "grpc"
	ProtocolWS   ProtocolKind = "ws"
	ProtocolGQL  ProtocolKind = "gql"
)

var Protocols = []ProtocolKind{ProtocolREST, ProtocolGRPC, ProtocolWS, ProtocolGQL}

type ReqMeta struct {
	Name     string       `toml:"name" json:"name"`
	Protocol ProtocolKind `toml:"protocol" comment:"'http'" json:"protocol"`
}

// ReqName implements [Runner].
func (r *ReqMeta) ReqName() string {
	return r.Name
}

// ReqProtocol implements [Runner].
func (r *ReqMeta) ReqProtocol() ProtocolKind {
	return r.Protocol
}

type AuthKind string

const (
	AuthInherited AuthKind = ""
	AuthNone      AuthKind = "none"
	AuthBasic     AuthKind = "basic"
	AuthDigest    AuthKind = "digest"
	AuthToken     AuthKind = "token"
)

var AuthTypes = []AuthKind{AuthInherited, AuthNone, AuthBasic, AuthDigest, AuthToken}

type (
	Auth struct {
		Type  AuthKind
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
type Request struct {
	ReqMeta
	URL    string
	Header map[string][]string
	Query  map[string][]string
	Desc   string
	// Auth   string
	Auth    Auth   `toml:"auth" comment:"token,basic,digest(only-client)" json:"auth"`
	Timeout string `toml:"timeout" comment:"time>=0, e.g.:'1m30s' time_unit:ns|us|ms|s|m|h" json:"timeout"`

	ConfigREST *ConfigREST
	ConfigGRPC *ConfigGRPC
	ConfigWS   *ConfigWS
}

type ConfigREST struct {
	Body              Body
	Retry             Retry  `toml:"retry" json:"retry"`
	CookieString      string `toml:"cookies" json:"cookieString"`
	UnescapeQuery     bool   `toml:"unescape_query" json:"unescapeQuery"`
	UnescapeJson      bool   `toml:"unescape_json" json:"unescapeJson"`
	ResponseBodyLimit int64  `toml:"response_body_limit" comment:"if limit<=0,it means no limit" json:"responseBodyLimit"`
}
type Retry struct {
	Count                int    `toml:"count" json:"count"`
	WaitTime             string `toml:"wait_time" json:"waitTime"`
	MaxWaitTime          string `toml:"max_wait_time" json:"maxWaitTime"`
	IsAllowNonIdempotent bool   `toml:"is_allow_non_idempotent" json:"isAllowNonIdempotent"`
}

type BodyKind string

const (
	BodyNone       BodyKind = ""
	BodyUrlEncoded BodyKind = "url-encoded"
	BodyMultipart  BodyKind = "multipart"
	// raw string type
	BodyRaw  BodyKind = "raw"
	BodyGQL  BodyKind = "graphql"
	BodyJSON BodyKind = "json"
	BodyXML  BodyKind = "xml"
	// file path,input file content
	BodyBinary BodyKind = "binary"
)

var BodyTypes = []BodyKind{BodyNone, BodyUrlEncoded, BodyMultipart, BodyRaw, BodyGQL, BodyJSON, BodyXML, BodyBinary}

type (
	Body struct {
		Type       BodyKind
		Raw        string
		Multipart  []MultiField
		UrlEncoded map[string][]string
		BinaryPath string
	}
	MultiField struct {
		Name        string
		Value       string
		FileName    string
		FilePath    string
		ContentType string
	}
)
type ConfigGRPC struct {
}
type ConfigWS struct {
}
