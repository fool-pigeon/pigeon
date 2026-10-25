// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "slices"

type ConfWS struct {
	Subprotocols []string `json:"subprotocols" toml:"subprotocols"`
	// "disabled" / "context_takeover" / "no_context_takeover"
	CompressionMode      string   `json:"compressionMode" toml:"compression_mode"`
	CompressionThreshold int      `json:"compressionThreshold" toml:"compression_threshold"`
	OriginPatterns       []string `json:"originPatterns" toml:"origin_patterns"`
	InsecureSkipVerify   bool     `json:"insecureSkipVerify" toml:"insecure_skip_verify"`
}

func (c *ConfWS) Copy() *ConfWS {
	if c == nil {
		return nil
	}
	dst := *c
	dst.Subprotocols = slices.Clone(c.Subprotocols)
	dst.OriginPatterns = slices.Clone(c.OriginPatterns)
	return &dst
}
