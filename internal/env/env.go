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

// TODO: refact env for new project structure
const DefaultFileName = ".env"

type EnvData = koanf.Koanf

// Env is a wrap for [koanf.Koanf].
// the new value will replace the old value when load and has the same key path.
// Env load from a dotenv file,its type can be seen as map[ string ] string
// use [Env.Expand] to replace the ${var} in string.
type Env struct {
	Data *EnvData
}

func (e *Env) All() map[string]any {
	return e.Data.All()
}
func (e *Env) Print() {
	e.Data.Print()
}
func (e *Env) Set(key string, val any) error {
	return e.Data.Set(key, val)
}
func (e *Env) String(key string) string {
	return e.Data.String(key)
}
func (e *Env) LoadFile(name string) error {
	if err := e.Data.Load(file.Provider(name), dotenv.Parser()); err != nil {
		return err
	}
	return nil
}
func (e *Env) Store(name string) error {
	return fsx.StoreKoanf(e.Data, dotenv.Parser(), name)
}

// Expand replaces ${var} in the string based on Env
// replace it if var in Env,else keep original value
func (e *Env) Expand(s string) string {
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
			} else if !e.Data.Exists(name) {
				// var is not exist in Env; keep origin
				sb.WriteString(s[j : j+1+w])
			} else {
				sb.WriteString(e.Data.String(name))
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

// get var name from string like {var}
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

func New() *Env {
	return &Env{koanf.New(".")}
}
func FileName(tag ...string) string {
	if len(tag) > 0 && tag[0] != "" {
		return DefaultFileName + "-" + tag[0]
	}
	return DefaultFileName
}
func LoadFromFile(name string) (*Env, error) {
	e := New()
	if err := e.LoadFile(name); err != nil {
		return nil, err
	}
	return e, nil
}
func LoadFromPath(path string, tag ...string) (*Env, error) {
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
		e.LoadFile(name)
	}
	return e, nil
}
