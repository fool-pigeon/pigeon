package fsx_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/homoarea/pigeon/internal/fsx"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"
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
	err = os.Mkdir(filepath.Join(tmpDir, "no_perm_dir"), os.FileMode(0o755))
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.OpenFile(filepath.Join(d, "no_perm"), os.O_CREATE|os.O_RDWR, os.FileMode(0o755))
	if err != nil {
		t.Fatal(err)
	}
	// change permssion of dir,the file can't be accessed
	os.Chmod(d, os.FileMode(0o000))
	isFileExist, err = fsx.IsFileExist(f.Name())
	require.ErrorIs(err, os.ErrPermission)
	// reset for cleaning tmpDir
	os.Chmod(d, os.FileMode(0o755))
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

func TestRDWRToml(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "pg-fsx-test-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	type Data struct {
		Version int      `toml:"version"`
		K1      string   `toml:"k1"`
		K2      string   `toml:"K2"`
		Tags    []string `toml:"tags"`
	}

	data := Data{
		Version: 1,
		K1:      "hello",
		K2:      "world",
		Tags:    []string{"go", "pg"},
	}
	name := filepath.Join(tmpDir, "toml_file")
	if err := fsx.WriteToml(name, &data); err != nil {
		t.Fatal(err)
	}
	var v Data
	if err := fsx.ReadToml(name, &v); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, data, v)
}

func TestToml(t *testing.T) {
	data := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	b, err := toml.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := toml.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
	fmt.Println(m)
	require := require.New(t)
	require.Equal(data, m)
	err = toml.Unmarshal([]byte(`k1='v1`), &m)
	require.Error(err)
	fmt.Println(err)

	mfs := []httpx.MultipartField{
		{Name: "name1", Values: []string{"v1", "v2"}},
	}
	mf := mfs[0]
	b, err = toml.Marshal(mf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	type Inner struct {
		F1 int    `toml:"f1"`
		F2 string `toml:"f2"`
	}
	d := struct {
		Data []Inner `toml:"data"`
	}{
		[]Inner{
			{F1: 1, F2: "114"},
			{F1: 2, F2: "514"},
		},
	}
	fmt.Println(d)
	b, err = toml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}
