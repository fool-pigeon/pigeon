package env

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const DefaultFileName = ".env"

type Env = koanf.Koanf

func New() *Env {
	return koanf.New(".")
}
func FileName(suffix ...string) string {
	if len(suffix) > 0 && suffix[0] != "" {
		return DefaultFileName + "-" + suffix[0]
	}
	return DefaultFileName
}
func LoadFile(e *Env, name string) error {
	if err := e.Load(file.Provider(name), dotenv.Parser()); err != nil {
		return err
	}
	return nil
}
func LoadFromFile(name string) (*Env, error) {
	e := New()
	if err := e.Load(file.Provider(name), dotenv.Parser()); err != nil {
		return nil, err
	}
	return e, nil
}
func LoadFromPath(path string, tag ...string) (*Env, error) {
	// FIXME : env should load the root env, .pg/.env
	paths, err := fsx.PathsToRoot(path)
	if err != nil {
		return nil, err
	}

	// add the root tag dir for load the root env
	fName := FileName(tag...)
	paths = append(paths, filepath.Join(paths[len(paths)-1], fsx.RootTagDirName))
	slices.Reverse(paths)
	e := New()
	for _, path := range paths {
		name := filepath.Join(path, fName)
		isExist, err := fsx.IsFileExist(name)
		if err != nil {
			return nil, err
		}
		if !isExist {
			continue
		}
		LoadFile(e, name)
	}
	return e, nil
}

// Expand replaces ${var} in the string based on Env
// replace it if var in Env,else keep original value
func Expand(s string, e *Env) string {
	var sb strings.Builder
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) && s[j+1] == '{' {
			if sb.Cap() == 0 {
				sb.Grow(2 * len(s))
			}
			sb.WriteString(s[i:j])
			name, w := getVarName(s[j+1:])
			if name == "" && w > 0 {
				// Encountered invalid syntax; keep origin
				sb.WriteString(s[j : j+1+w])
			} else if !e.Exists(name) {
				// var is not exist in Env; keep origin
				sb.WriteString(s[j : j+1+w])
			} else {
				sb.WriteString(e.String(name))
			}
			j += w
			i = j + 1
		}
	}
	if sb.Cap() == 0 {
		return s
	}
	return sb.String() + s[i:]
}

// s like {var}
func getVarName(s string) (string, int) {
	// Scan to closing brace
	for i := 1; i < len(s); i++ {
		if s[i] == '}' {
			if i == 1 {
				return "", 2 // Bad syntax;  "${}"
			}
			return s[1:i], i + 1
		}
	}
	return "", 1 // Bad syntax;  "${"
}

func Store(e *Env, name string) error {
	return fsx.StoreKoanf(e, dotenv.Parser(), name)
}
