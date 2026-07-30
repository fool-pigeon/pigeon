package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/homoarea/pigeon/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// create tmpDir for test
	tmpDir, err := os.MkdirTemp("", "pg-test-dir-*")
	tmpDir, _ = filepath.Abs(tmpDir)
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal("create tmpDir failed", err)
	}
	dir1, dir2 := filepath.Join(tmpDir, "dir1"), filepath.Join(tmpDir, "dir2")
	os.Mkdir(dir1, os.ModePerm)
	os.Mkdir(dir2, os.ModePerm)
	// create .pg/ and .pg
	os.Mkdir(filepath.Join(dir1, ".pg"), os.ModePerm)
	os.Create(filepath.Join(dir2, ".pg"))
	os.Create(filepath.Join(tmpDir, ".pg"))
	// .
	// ├── dir1
	// │   └── .pg/ # this is dir,no match
	// ├── dir2
	// │   └── .pg  # match this
	// └── .pg  	# match this

	require := require.New(t)
	// test dir1
	r, err := config.RootDir(dir1)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(tmpDir, r, "match the root from dir1")
	paths, err := config.PathsToRoot(dir1)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal([]string{dir1, tmpDir}, paths)
	paths, err = config.PathsFromRoot(dir1)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal([]string{tmpDir, dir1}, paths)

}
