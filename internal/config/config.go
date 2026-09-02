package config

import "github.com/fool-pigeon/pigeon/internal/request/httpx"

const (
	DefaultConfigName = "config"

	fileExt = ".toml"
)

type Clients struct {
	Http httpx.ClientData `toml:"http"`
}
type CommentedConfig struct {
	Clients Clients `toml:"clients,commented"`
}
type Config struct {
	Clients Clients `toml:"clients"`
}

func New() Config {
	return Config{}
}
func FileName(suffix ...string) string {
	if len(suffix) > 0 && suffix[0] != "" {
		return DefaultConfigName + "-" + suffix[0] + fileExt
	}
	return DefaultConfigName + fileExt
}
