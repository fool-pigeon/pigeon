package fsx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/homoarea/pigeon/internal/fsx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestIsFileExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "pg-fsx-test-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	require := require.New(t)
	// test a file
	f, err := os.CreateTemp(tmpDir, "tmp_file")
	isFileExist, err := fsx.IsFileExist(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(true, isFileExist)

	// test a dir
	d, err := os.MkdirTemp(tmpDir, "tmp_dir")
	if err != nil {
		t.Fatal(err)
	}
	isFileExist, err = fsx.IsFileExist(d)
	require.Equal(false, isFileExist)

	// test a file has no perm
	d = filepath.Join(tmpDir, "no_perm_dir")
	err = os.Mkdir(filepath.Join(tmpDir, "no_perm_dir"), os.FileMode(0755))
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.OpenFile(filepath.Join(d, "no_perm"), os.O_CREATE|os.O_RDWR, os.FileMode(0755))
	if err != nil {
		t.Fatal(err)
	}
	// change permssion of dir,the file can't be accessed
	os.Chmod(d, os.FileMode(0000))
	isFileExist, err = fsx.IsFileExist(f.Name())
	require.ErrorIs(err, os.ErrPermission)
	// reset for cleaning tmpDir
	os.Chmod(d, os.FileMode(0755))
}

func TestStoreKoanf(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "pg-fsx-test-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	name := filepath.Join(tmpDir, "test_file")
	k := koanf.New(".")
	k.Set("hello", "world")
	if err := fsx.StoreKoanf(k, dotenv.Parser(), name); err != nil {
		t.Fatal(err)
	}

	k1 := koanf.New(".")
	if err := k1.Load(file.Provider(name), dotenv.Parser()); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, k, k1)
}
