// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "maps"

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
		Kind       BodyKind
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
type Retry struct {
	Count                int         `toml:"count" json:"count"`
	WaitTime             DurationStr `toml:"wait_time" json:"waitTime"`
	MaxWaitTime          DurationStr `toml:"max_wait_time" json:"maxWaitTime"`
	IsAllowNonIdempotent bool        `toml:"is_allow_non_idempotent" json:"isAllowNonIdempotent"`
}

type ConfigREST struct {
	// test expr script
	Tests map[string]string

	Path              map[string]string
	Retry             Retry `toml:"retry" json:"retry"`
	UnescapeQuery     bool  `toml:"unescape_query" json:"unescapeQuery"`
	UnescapeJson      bool  `toml:"unescape_json" json:"unescapeJson"`
	ResponseBodyLimit int64 `toml:"response_body_limit" comment:"if limit<=0,it means no limit" json:"responseBodyLimit"`
}

func (c *ConfigREST) Copy() *ConfigREST {
	dst := *c
	maps.Copy(dst.Tests, c.Tests)
	maps.Copy(dst.Path, c.Path)
	return &dst
}
