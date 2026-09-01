package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fool-pigeon/pigeon/internal/env"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// init dir
	tmpDir, err := os.MkdirTemp(os.TempDir(), "pg-env-test-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	os.Mkdir(dir1, os.FileMode(0755))
	os.Mkdir(dir2, os.FileMode(0755))
	envName := filepath.Join(tmpDir, env.DefaultFileName)
	env1Name := filepath.Join(dir1, env.DefaultFileName)
	os.WriteFile(envName, []byte("hello=world\ne1=v1"), os.FileMode(0666))
	os.WriteFile(env1Name, []byte("hello=you\ne2=v1"), os.FileMode(0666))
	os.Create(filepath.Join(tmpDir, ".pg"))
	// .
	// ├── dir1
	// │   └── .env # hello=you e2=v1
	// ├── dir2
	// └── .env # hello=world e1=v1
	// └── .pg

	require := require.New(t)
	k1, err := env.LoadFromPath(dir1)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal("you", k1.Get("hello"))
	require.Equal("v1", k1.String("e2"))

	k2, err := env.LoadFromPath(dir2)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal("world", k2.Get("hello"))
	require.Nil(k2.Get("e2"))

	k, err := env.LoadFromPath(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(k, k2)
	require.NotEqual(k1, k2)

	// load the file not exist
	k3 := koanf.New(".")
	if err := env.LoadFromFile(k3, filepath.Join(tmpDir, "file_not_exist")); err != nil {
		require.ErrorIs(err, os.ErrNotExist)
	}

}
