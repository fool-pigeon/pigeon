// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
)

// setupRoot creates <root>/.pg + <root>/a/b, and writes fsx.Conf (".conf")
// into the dirs named in contents. Keys: "root", "mid", "leaf".
// initReqCtx calls os.Lstat on every path, so all three are required.
func setupRoot(t *testing.T, contents map[string]string) (root, mid, leaf string) {
	t.Helper()
	root = t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, fsx.RootTag), 0o755))
	mid = filepath.Join(root, "a")
	leaf = filepath.Join(mid, "b")
	require.NoError(t, os.MkdirAll(leaf, 0o755))
	dirs := map[string]string{"root": root, "mid": mid, "leaf": leaf}
	for name, body := range contents {
		require.NoError(t, os.WriteFile(filepath.Join(dirs[name], fsx.Conf), []byte(body), 0o644))
	}
	return
}

func TestDefaultRootConf(t *testing.T) {
	assert.Equal(t, 5, DefaultRootConf.MaxRunningConcurrency)
}

func TestRootConfUnmarshal(t *testing.T) {
	body := `
max_running_concurrency = 10

[client]
proxy_url = "http://x"

[conf.env_vars]
HOST = "example.com"
`
	var rc RootConf
	require.NoError(t, toml.Unmarshal([]byte(body), &rc))
	assert.Equal(t, 10, rc.MaxRunningConcurrency)
	assert.Equal(t, "http://x", rc.Client.ProxyURL)
	assert.Equal(t, "example.com", rc.Conf.EnvVars["HOST"])
}

func TestInitReqCtx(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var rc *RootConf
		_, err := rc.initReqCtx(t.TempDir())
		assert.ErrorIs(t, err, ErrNilValue)
	})

	t.Run("missing conf errors", func(t *testing.T) {
		_, _, leaf := setupRoot(t, nil)
		_, err := (&RootConf{}).initReqCtx(leaf)
		assert.Error(t, err)
	})

	t.Run("partial confs error", func(t *testing.T) {
		_, _, leaf := setupRoot(t, map[string]string{"root": ""})
		_, err := (&RootConf{}).initReqCtx(leaf)
		assert.Error(t, err)
	})

	t.Run("all confs present", func(t *testing.T) {
		_, _, leaf := setupRoot(t, map[string]string{
			"root": "",
			"mid":  "",
			"leaf": "",
		})
		ctx, err := (&RootConf{}).initReqCtx(leaf)
		require.NoError(t, err)
		assert.NotNil(t, ctx)
	})
}

func TestLoadConf(t *testing.T) {
	t.Run("valid plain", func(t *testing.T) {
		conf := &protocol.ReqConf{}
		require.NoError(t, loadConf([]byte(`timeout = "5s"`), NewEnv(), conf))
		assert.Equal(t, protocol.DurationStr("5s"), conf.Timeout)
	})

	t.Run("invalid toml", func(t *testing.T) {
		conf := &protocol.ReqConf{}
		assert.Error(t, loadConf([]byte("= bad"), NewEnv(), conf))
	})

	t.Run("existing wins on merge", func(t *testing.T) {
		conf := &protocol.ReqConf{Timeout: "1s"}
		require.NoError(t, loadConf([]byte(`timeout = "5s"`), NewEnv(), conf))
		assert.Equal(t, protocol.DurationStr("1s"), conf.Timeout)
	})

	t.Run("expands env vars", func(t *testing.T) {
		conf := &protocol.ReqConf{}
		body := `
timeout = "${TIME}"

[env_vars]
TIME = "114s"
`
		require.NoError(t, loadConf([]byte(body), NewEnv(), conf))
		assert.EqualValues(t, "114s", conf.Timeout)
	})
}
