// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/pelletier/go-toml/v2"
)

var DefaultRootConf = RootConf{MaxRunningConcurrency: 5}

type RootConf struct {
	MaxRunningConcurrency int `toml:"max_running_concurrency"`

	Client protocol.Client `toml:"client"`
	Conf   Conf            `toml:"conf"`
}

func (rootConf *RootConf) InitReqCtx(startPath string) (*ReqCtx, error) {

	if rootConf == nil {
		return nil, fmt.Errorf("initialReqCtx rootConf:%w", ErrNilValue)
	}

	initialEnv := NewEnv()
	if err := initialEnv.LoadNestedMap(rootConf.Conf.EnvVars); err != nil {
		return nil, err
	}
	initialConf := rootConf.Conf.ReqConf
	initialConfBytes, err := toml.Marshal(initialConf)
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal([]byte(initialEnv.Expand(string(initialConfBytes))), &initialConf); err != nil {
		return nil, err
	}

	paths, err := fsx.PathsFromRoot(startPath)
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		confPath := filepath.Join(p, fsx.Conf)
		info, err := os.Lstat(confPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.Mode().Type().IsRegular() {
			confBytes, err := fsx.ReadFile(confPath)
			if err != nil {
				continue
			}
			loadConf(confBytes, initialEnv, &initialConf)
		}
	}

	return &ReqCtx{&initialConf, initialEnv}, nil
}

func loadConf(confBytes []byte, env *Env, reqConf *protocol.ReqConf) error {
	var conf Conf
	if err := toml.Unmarshal(confBytes, &conf); err != nil {
		return err
	}
	if err := env.LoadNestedMap(conf.EnvVars); err != nil {
		return err
	}
	if err := toml.Unmarshal([]byte(env.Expand(string(confBytes))), &conf); err != nil {
		return err
	}
	reqConf.Merge(&conf.ReqConf)
	return nil
}

type Conf struct {
	protocol.ReqConf
	EnvVars map[string]any `toml:"env_vars"`
}
