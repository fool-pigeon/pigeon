// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/protocol"
	"github.com/pelletier/go-toml/v2"
)

var DefaultRootConf = RootConf{MaxRunningConcurrency: 5}

type RootConf struct {
	MaxRunningConcurrency int

	Client protocol.Client
	Conf   Conf
}

func (rootConf *RootConf) initReqCtx(startPath string) (*ReqCtx, error) {

	if rootConf == nil {
		return nil, fmt.Errorf("initialReqCtx rootConf:%w", ErrNilValue)
	}

	initialEnv := NewEnv()
	if err := initialEnv.LoadNestedMap(rootConf.Conf.EnvVars); err != nil {
		return nil, err
	}
	initialConf := rootConf.Conf.ReqConfig
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
		if err != nil {
			return nil, err
		}
		if info.Mode().Type().IsRegular() {
			confBytes, err := fsx.ReadFile(confPath)
			if err != nil {
				return nil, err
			}
			loadConf(confBytes, initialEnv, &initialConf)
		}
	}

	return &ReqCtx{&initialConf, initialEnv}, nil
}

func loadConf(confBytes []byte, env *Env, reqConf *protocol.ReqConfig) error {
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
	reqConf.Merge(&conf.ReqConfig)
	return nil
}

type Conf struct {
	protocol.ReqConfig
	EnvVars map[string]any
}
