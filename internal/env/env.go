package env

import (
	"path/filepath"

	"github.com/homoarea/pigeon/internal/fsx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var Envs = koanf.New(".")

const DefaultFileName = ".env"

type Env = koanf.Koanf

func FileName(suffix ...string) string {
	if len(suffix) > 0 {
		return DefaultFileName + "-" + suffix[0]
	}
	return DefaultFileName
}
func LoadFromFile(e *Env, name string) error {
	if err := e.Load(file.Provider(name), dotenv.Parser()); err != nil {
		return err
	}
	return nil
}

func Load(path string) (*Env, error) {
	paths, err := fsx.PathsFromRoot(path)
	if err != nil {
		return nil, err
	}
	k := koanf.New(".")
	for _, path := range paths {
		fileName := filepath.Join(path, DefaultFileName)
		isExist, err := fsx.IsFileExist(fileName)
		if err != nil {
			return nil, err
		}
		if isExist {
			LoadFromFile(k, fileName)
		}
	}
	return k, nil
}

func Store(e *Env, name string) error {
	return fsx.StoreKoanf(e, dotenv.Parser(), name)
}
