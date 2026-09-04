package config

import (
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
)

const (
	DefaultConfigName = "config"

	fileExt = ".toml"
)

// config as the entrypoint of clients
// every client has its own mutex for concurrency
type Config struct {
	HttpClientData httpx.ClientData `toml:"http"`

	http *httpx.Client
}

// return a http client with its own Mutex
func (c *Config) Http() (*httpx.Client, error) {
	if c.http != nil {
		return c.http, nil
	}

	http, err := c.HttpClientData.InstancedClient()
	if err != nil {
		return nil, err
	}
	c.http = http

	return http, nil
}

func New() *Config {
	return &Config{}
}
func Load(path string, tag ...string) (*Config, error) {
	parentDir, err := fsx.RootTagDir(path)
	if err != nil {
		return nil, err
	}
	c := New()
	if err := fsx.ReadToml(filepath.Join(parentDir, FileName(tag...)), c); err != nil {
		return nil, err
	}

	return c, nil
}
func FileName(suffix ...string) string {
	if len(suffix) > 0 && suffix[0] != "" {
		return DefaultConfigName + "-" + suffix[0] + fileExt
	}
	return DefaultConfigName + fileExt
}
