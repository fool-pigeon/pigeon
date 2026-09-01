package env

import (
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var Envs = koanf.New(".")

const DefaultFileName = ".env"

type Env = koanf.Koanf

func FileName(suffix ...string) string {
	if len(suffix) > 0 && suffix[0] != "" {
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
func LoadFromPath(path string, tag ...string) (*Env, error) {
	paths, err := fsx.PathsFromRoot(path)
	if err != nil {
		return nil, err
	}

	fName := FileName(tag...)
	k := koanf.New(".")
	for _, path := range paths {
		name := filepath.Join(path, fName)
		isExist, err := fsx.IsFileExist(name)
		if err != nil {
			return nil, err
		}
		if !isExist {
			continue
		}
		LoadFromFile(k, name)
	}
	return k, nil
}

func Store(e *Env, name string) error {
	return fsx.StoreKoanf(e, dotenv.Parser(), name)
}
