package fsx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsFileExist(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("test normal file", func(t *testing.T) {
		f, err := os.CreateTemp(tmpDir, "tmp_file")
		isFileExist, err := fsx.IsFileExist(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, true, isFileExist)
	})

	t.Run("test a file not exist", func(t *testing.T) {
		name := filepath.Join(tmpDir, "the_file_not_exist")
		isExist, err := fsx.IsFileExist(name)
		require.Equal(t, false, isExist)
		require.Nil(t, err)
	})

	t.Run("test a dir", func(t *testing.T) {
		d, err := os.MkdirTemp(tmpDir, "tmp_dir")
		if err != nil {
			t.Fatal(err)
		}
		isFileExist, err := fsx.IsFileExist(d)
		require.Equal(t, false, isFileExist)
	})

	t.Run("test a file has no perm", func(t *testing.T) {
		d := filepath.Join(tmpDir, "no_perm_dir")
		err := os.Mkdir(filepath.Join(tmpDir, "no_perm_dir"), os.FileMode(0o755))
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(filepath.Join(d, "no_perm"), os.O_CREATE|os.O_RDWR, os.FileMode(0o755))
		if err != nil {
			t.Fatal(err)
		}
		// change permssion of dir,the file can't be accessed
		os.Chmod(d, os.FileMode(0o000))
		defer os.Chmod(d, os.FileMode(0o755))

		isFileExist, err := fsx.IsFileExist(f.Name())
		require.ErrorIs(t, err, os.ErrPermission)
		require.Equal(t, false, isFileExist)
	})
}
func TestIsDirExist(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("test normal dir", func(t *testing.T) {
		dir, err := os.MkdirTemp(tmpDir, "tmp_file")
		if err != nil {
			t.Fatal(err)
		}
		isExist, err := fsx.IsDirExist(dir)
		require.Equal(t, true, isExist)
		require.Nil(t, err)
	})

	t.Run("test a dir not exist", func(t *testing.T) {
		name := filepath.Join(tmpDir, "the_dir_not_exist")
		isExist, err := fsx.IsDirExist(name)
		require.Equal(t, false, isExist)
		require.Nil(t, err)
	})

	t.Run("test a file", func(t *testing.T) {
		f, err := os.CreateTemp(tmpDir, "tmp_dir")
		if err != nil {
			t.Fatal(err)
		}
		isExist, err := fsx.IsDirExist(f.Name())
		require.Equal(t, false, isExist)
		require.Nil(t, err)
	})

	t.Run("test a dir has no perm", func(t *testing.T) {
		parent, err := os.MkdirTemp(tmpDir, "no_perm_dir")
		if err != nil {
			t.Fatal(err)
		}
		d, err := os.MkdirTemp(parent, "son")
		if err != nil {
			t.Fatal(err)
		}
		// change permssion of dir,the file can't be accessed
		os.Chmod(parent, os.FileMode(0o000))
		defer os.Chmod(parent, os.FileMode(0o755))

		isExist, err := fsx.IsDirExist(d)
		require.ErrorIs(t, err, os.ErrPermission)
		require.Equal(t, false, isExist)
	})
}
func TestRWFile(t *testing.T) {
	t.Run("read and write a file", func(t *testing.T) {
		tmpDir := t.TempDir()
		fName := filepath.Join(tmpDir, "file")
		data := []byte("data")
		fsx.WriteFile(fName, data)
		b, _ := fsx.ReadFile(fName)
		assert.Equal(t, data, b)
		bs, _ := fsx.ReadFileAsString(fName)
		assert.Equal(t, string(data), bs)
	})
	t.Run("read a empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		fName := filepath.Join(tmpDir, "not-exist")
		fsx.WriteFile(fName, []byte{})
		b, _ := fsx.ReadFile(fName)
		assert.Equal(t, []byte{}, b)
		bs, _ := fsx.ReadFileAsString(fName)
		assert.Equal(t, "", bs)
	})
	t.Run("read a file not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		fName := filepath.Join(tmpDir, "not-exist")
		_, err := fsx.ReadFile(fName)
		assert.ErrorIs(t, err, fsx.ErrNotExist)
	})
}
func TestStoreKoanf(t *testing.T) {
	tmpDir := t.TempDir()

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
	tmpDir := t.TempDir()

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
	t.Log(string(b))
	t.Log(m)
	require := require.New(t)
	require.Equal(data, m)
	err = toml.Unmarshal([]byte(`k1='v1`), &m)
	require.Error(err)
	t.Log(err)

	mfs := []httpx.MultiField{
		{Name: "name1", Value: "v1"},
	}
	mf := mfs[0]
	b, err = toml.Marshal(mf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))

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
	t.Log(d)
	b, err = toml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}
