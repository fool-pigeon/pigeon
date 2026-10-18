// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/protocol/httpx"
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
	http     clientWrap[*httpx.Client]

	MaxRunningConcurrency uint32 `toml:"max_running_concurrency"`
}

func New() *Config {
	return &Config{
		MaxRunningConcurrency: 5,
	}
}
func (c *Config) checkAndReset() {
	if c.MaxRunningConcurrency <= 0 {
		c.MaxRunningConcurrency = 5
	}
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

func LoadFromFile(name string) (*Config, error) {
	c := New()
	if err := fsx.ReadToml(name, c); err != nil {
		return nil, err
	}
	c.checkAndReset()
	return c, nil
}
