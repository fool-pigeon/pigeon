// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	// task:id:{task_ulid} -> [Task] struct
	dbKeyPrefixTaskId = "task:id:"
	// task:res:{task_ulid}:{req_ulid}:{req_res_ulid}
	dbKeyPrefixTaskRes = "task:res:"

	// req:res:{req_ulid}:{res_ulid} -> nil
	dbKeyPrefixReqRes = "req:res"

	// res:id:{res_ulid} -> [TaskRes]
	dbKeyPrefixRes = "res:id:"
)

type Database struct {
	cache *badger.DB
}

func NewDatabase() *Database {
	return &Database{}
}

// ServiceStartup implements [application.ServiceStartup].
func (d *Database) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	const dbName = "badger"

	cacheDbPath := filepath.Join(gui.Paths.Cache(), dbName)
	if err := fsx.MkdirAll(cacheDbPath); err != nil {
		return err
	}
	db, err := badger.Open(badger.DefaultOptions(cacheDbPath))
	if err != nil {
		return err
	}
	d.cache = db

	return nil
}

// ServiceShutdown implements [application.ServiceShutdown].
func (d *Database) ServiceShutdown() error {
	var errs error

	if d.cache != nil {
		errs = errors.Join(errs, d.cache.Close())
	}

	return errs
}
