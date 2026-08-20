/*
Copyright © 2026 Homoarea <homoarea114@gmail.com>
*/
package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [file1.toml folder/ file2.toml]",
	Aliases: []string{"r"},
	GroupID: GROUP_ID_REQUEST,
	Short:   "alias r, Run collection or request",
	RunE:    runRequest,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRequest(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, ".")
	}
	if err := run(args...); err != nil {
		return err
	}
	return nil
}

type Cache[T any] struct {
	path  string
	state T
}

func readRequest(path string) (request.Request, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta request.RequestMeta
	if err := toml.Unmarshal(b, &meta); err != nil {
		return nil, err
	}
	switch meta.Protocol {
	case request.P_HTTP:
		var r httpx.RequestData
		if err := toml.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		return &r, nil
		// TODO readRequest more
	default:
		return nil, errors.New("the protocol not impl")
	}
}

func execRequest[T any](r request.Request, cache *Cache[T]) (string, error) {
	var msg string
	switch r := r.(type) {
	case *httpx.RequestData:
		rp, err := r.Send()
		if err != nil {
			return "", err
		}
		msg += pterm.Sprintfln("%v - %v (%vms)", r.Protocol, r.Name, rp.Duration().Milliseconds())
		msg += pterm.Sprintfln("%v\nheader:%v\nbody:%v", rp.Status(), rp.Header(), string(rp.Bytes()))
		// TODO execRequest more
	default:
		return "", errors.New("the protocol not impl")
	}
	return msg, nil
}

func run(paths ...string) error {

	isTomlFile := func(path string) bool {
		return filepath.Ext(path) == ".toml"
	}

	execTask := func(path string, cache *Cache[string], msgCh chan string) {
		r, err := readRequest(path)
		if err != nil {
			msgCh <- pterm.Error.Sprintf("%v: %v", path, err)
			return
		}
		msg, err := execRequest(r, cache)
		if err != nil {
			msgCh <- pterm.Error.Sprintf("%v: %v", path, err)
			return
		}
		msgCh <- pterm.Info.Sprint(msg)
	}

	var wg sync.WaitGroup
	buf := []Cache[string]{{path: ".", state: "."}}
	msgCh := make(chan string)

	for _, path := range paths {
		path = filepath.Clean(path)

		info, err := os.Lstat(path)
		if err != nil {
			pterm.Error.Printfln("path:%v,%v", path, err)
		}
		if !info.IsDir() {
			if isTomlFile(path) {
				wg.Go(func() {
					execTask(path, &buf[0], msgCh)
				})
			}
			continue
		}

		// recursive directory
		filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			parentDir := filepath.Dir(p)
			for len(buf) > 1 && buf[len(buf)-1].path != parentDir {
				buf = buf[:len(buf)-1]
			}

			var cache Cache[string]
			if len(buf) > 0 {
				cache = buf[len(buf)-1]
			}

			currentState, _ := filepath.Abs(p)

			switch {
			case d.IsDir():
				buf = append(buf, Cache[string]{path: p, state: currentState})

			case isTomlFile(p):
				wg.Go(func() {
					execTask(p, &cache, msgCh)
				})
			}

			return nil
		})
	}

	go func() {
		wg.Wait()
		defer close(msgCh)
	}()

	for v := range msgCh {
		pterm.Println(v)
	}
	return nil
}
