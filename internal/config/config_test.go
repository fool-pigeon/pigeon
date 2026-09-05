package config_test

import (
	"path/filepath"
	"testing"

	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileName(t *testing.T) {
	t.Run("dafault", func(t *testing.T) {
		require.Equal(t, "config.toml", config.FileName())
	})
	t.Run("with a suffix", func(t *testing.T) {
		require.Equal(t, "config-dev.toml", config.FileName("dev"))
	})
}
func TestInstanceClient(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		t.Run("success instance", func(t *testing.T) {
			c := config.New()
			c.HttpData.SetBaseURL("url")
			httpClient, _ := c.HttpClient()
			assert.Equal(t, "url", httpClient.BaseURL())
			h2, _ := c.HttpClient()
			assert.Equal(t, httpClient, h2, "get the same client")
		})
		t.Run("failed to parse time", func(t *testing.T) {
			c := config.New()
			c.HttpData.SetTimeout("hello")
			_, err := c.HttpClient()
			assert.Error(t, err)
		})
	})

}
func TestLoad(t *testing.T) {
	exampleCfg := config.New()
	baseURL := "http://127.0.0.1"
	exampleCfg.HttpData.SetBaseURL(baseURL)
	exampleFileContent, err := toml.Marshal(exampleCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("load from file", func(t *testing.T) {
		fName := filepath.Join(t.TempDir(), "file")
		fsx.WriteFile(fName, exampleFileContent)
		cfg, _ := config.LoadFromFile(fName)
		assert.Equal(t, baseURL, cfg.HttpData.BaseURL)
	})
	t.Run("load from path", func(t *testing.T) {
		tmpDir := t.TempDir()
		tagDir := filepath.Join(tmpDir, ".pg")
		fsx.Mkdir(tagDir)
		fName := filepath.Join(tagDir, "config.toml")
		fsx.WriteFile(fName, exampleFileContent)
		cfg, _ := config.LoadFromPath(tmpDir)
		assert.Equal(t, baseURL, cfg.HttpData.BaseURL)
	})
}
