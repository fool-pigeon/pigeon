// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "slices"

type ConfigWS struct {
	Subprotocols         []string
	CompressionMode      string // "disabled" / "context_takeover" / "no_context_takeover"
	CompressionThreshold int
	OriginPatterns       []string
	InsecureSkipVerify   bool
}

func (c *ConfigWS) Copy() *ConfigWS {
	if c == nil {
		return nil
	}
	dst := *c
	dst.Subprotocols = slices.Clone(c.Subprotocols)
	dst.OriginPatterns = slices.Clone(c.OriginPatterns)
	return &dst
}
