package config

import (
	"path/filepath"
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
)

const (
	// the default file name of config without file ext
	ConfigFile = "config"

	fileExt = ".toml"
)

type clientWrap[T any] struct {
	once   sync.Once
	err    error
	client T
}

// config store the data of clients
// config as the entrypoint of clients
// every client has its own mutex for concurrency
type Config struct {
	HttpData httpx.ClientData `toml:"http"`

	http clientWrap[*httpx.Client]
}

// return a http client pointer with its own Mutex.
// the client only instance once.
func (c *Config) HttpClient() (*httpx.Client, error) {
	if c.http.client != nil {
		return c.http.client, c.http.err
	}
	c.http.once.Do(func() {
		c.http.client, c.http.err = c.HttpData.InstancedClient()
	})
	return c.http.client, c.http.err
}

func New() *Config {
	return &Config{}
}

func LoadFromFile(name string) (*Config, error) {
	c := New()
	if err := fsx.ReadToml(name, c); err != nil {
		return nil, err
	}
	return c, nil
}

// load config from file in .pg/
//
// return error if not find the root tag dir .pg/
func LoadFromPath(path string, tag ...string) (*Config, error) {
	parentDir, err := fsx.RootTagDir(path)
	if err != nil {
		return nil, err
	}
	return LoadFromFile(filepath.Join(parentDir, FileName(tag...)))
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
