// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

type ConfigGRPC struct {
	ReadMaxBytes      int
	SendMaxBytes      int
	UseProtoJSON      bool
	EnableHTTPGet     bool
	HTTPGetMaxURLSize int
	SendGzip          bool
}

func (c *ConfigGRPC) Copy() *ConfigGRPC {
	dst := *c
	return &dst
}
