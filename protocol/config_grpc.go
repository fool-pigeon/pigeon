// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

type ConfGRPC struct {
	ReadMaxBytes      int  `json:"readMaxBytes" toml:"read_max_bytes"`
	SendMaxBytes      int  `json:"sendMaxBytes" toml:"send_max_bytes"`
	UseProtoJSON      bool `json:"useProtoJSON" toml:"use_proto_json"`
	EnableHTTPGet     bool `json:"enableHTTPGet" toml:"enable_http_get"`
	HTTPGetMaxURLSize int  `json:"httpGetMaxURLSize" toml:"http_get_max_url_size"`
	SendGzip          bool `json:"sendGzip" toml:"send_gzip"`
}

func (c *ConfGRPC) Copy() *ConfGRPC {
	if c == nil {
		return nil
	}
	dst := *c
	return &dst
}
