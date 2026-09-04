package config

import (
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
)

const (
	// the default file name of config without file ext
	ConfigFile = "config"

	fileExt = ".toml"
)

// config store the data of clients
// config as the entrypoint of clients
// every client has its own mutex for concurrency
type Config struct {
	HttpData httpx.ClientData `toml:"http"`

	http *httpx.Client
}

// return a http client with its own Mutex
func (c *Config) Http() (*httpx.Client, error) {
	if c.http != nil {
		return c.http, nil
	}

	http, err := c.HttpData.InstancedClient()
	if err != nil {
		return nil, err
	}
	c.http = http

	return http, nil
}

func New() *Config {
	return &Config{}
}

// load config from file in .pg/
//
// return error if not find the root tag dir .pg/
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

// generate config's file name.
// for example:
//
//	FileName() == "config.toml"
//	FileName("dev") == "config-dev.toml"
func FileName(tag ...string) string {
	if len(tag) > 0 && tag[0] != "" {
		return ConfigFile + "-" + tag[0] + fileExt
	}
	return ConfigFile + fileExt
}
