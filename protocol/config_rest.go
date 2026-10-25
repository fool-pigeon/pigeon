// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "maps"

type BodyKind string

const (
	BodyNone       BodyKind = ""
	BodyURLEncoded BodyKind = "url-encoded"
	BodyMultipart  BodyKind = "multipart"
	// raw string type
	BodyRaw  BodyKind = "raw"
	BodyGQL  BodyKind = "graphql"
	BodyJSON BodyKind = "json"
	BodyXML  BodyKind = "xml"
	// file path,input file content
	BodyBinary BodyKind = "binary"
)

var BodyTypes = []BodyKind{BodyNone, BodyURLEncoded, BodyMultipart, BodyRaw, BodyGQL, BodyJSON, BodyXML, BodyBinary}

type (
	Body struct {
		Kind       BodyKind            `json:"kind" toml:"kind"`
		Raw        string              `json:"raw" toml:"raw"`
		Multipart  []MultiField        `json:"multipart" toml:"multipart"`
		URLEncoded map[string][]string `json:"urlEncoded" toml:"url_encoded"`
		BinaryPath string              `json:"binaryPath" toml:"binary_path"`
	}
	MultiField struct {
		Name        string `json:"name" toml:"name"`
		Value       string `json:"value" toml:"value"`
		FileName    string `json:"fileName" toml:"file_name"`
		FilePath    string `json:"filePath" toml:"file_path"`
		ContentType string `json:"contentType" toml:"content_type"`
	}
)
type Retry struct {
	Count                int         `json:"count" toml:"count"`
	WaitTime             DurationStr `json:"waitTime" toml:"wait_time"`
	MaxWaitTime          DurationStr `json:"maxWaitTime" toml:"max_wait_time"`
	IsAllowNonIdempotent bool        `json:"isAllowNonIdempotent" toml:"is_allow_non_idempotent"`
}

type ConfREST struct {
	// test expr script
	Tests             map[string]string `json:"tests" toml:"tests"`
	Path              map[string]string `json:"path" toml:"path"`
	Retry             Retry             `json:"retry" toml:"retry"`
	UnescapeQuery     bool              `json:"unescapeQuery" toml:"unescape_query"`
	UnescapeJSON      bool              `json:"unescapeJSON" toml:"unescape_json"`
	ResponseBodyLimit int64             `json:"responseBodyLimit" toml:"response_body_limit"`
}

func (c *ConfREST) Copy() *ConfREST {
	if c == nil {
		return nil
	}
	dst := *c
	dst.Tests = make(map[string]string, len(c.Tests))
	dst.Path = make(map[string]string, len(c.Path))
	maps.Copy(dst.Tests, c.Tests)
	maps.Copy(dst.Path, c.Path)
	return &dst
}
