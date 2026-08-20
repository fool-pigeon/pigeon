package config

const (
	DefaultConfigName = "config"

	fileExt = ".toml"
)

func FileName(suffix ...string) string {
	if len(suffix) > 0 {
		return DefaultConfigName + "-" + suffix[0] + fileExt
	}
	return DefaultConfigName + fileExt
}

type Config struct {
}
