// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/oklog/ulid/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	// task:id:{task_ulid} -> [Task]
	dbKeyPrefixTaskID = "task:id"
	// task:res:{task_ulid}:{res_ulid} -> [TaskResViewMeta]
	dbKeyPrefixTaskRes = "task:res"

	// req:res:{req_ulid}:{res_ulid} -> nil
	dbKeyPrefixReqRes = "req:res"
	// res:id:{res_ulid} -> [TaskRes]
	dbKeyPrefixResID = "res:id"
	// res:trace:{res_ulid}:{trace_ulid} -> [protocol.Trace]
	dbKeyPrefixResTrace = "res:trace"
)

type DataSource struct {
	cache *badger.DB
}

// ServiceStartup implements [application.ServiceStartup].
func (d *DataSource) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
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
func (d *DataSource) ServiceShutdown() error {
	var errs error

	if d.cache != nil {
		errs = errors.Join(errs, d.cache.Close())
	}

	return errs
}
func makeDBKey(parts ...[]byte) []byte {
	if len(parts) <= 0 {
		return nil
	}
	keyLen := len(parts) - 1
	for _, p := range parts {
		keyLen += len(p)
	}
	key := make([]byte, 0, keyLen)
	for i, p := range parts {
		if i != 0 {
			key = append(key, ':')
		}
		key = append(key, p...)
	}
	return key
}

type BaseRepoULID[T any] struct {
	prefix string
	db     *badger.DB
}

func (repo *BaseRepoULID[T]) makeKey(id ulid.ULID) []byte {
	return makeDBKey([]byte(repo.prefix), id[:])
}
func (repo *BaseRepoULID[T]) GetByID(id ulid.ULID) (*T, error) {
	var data T
	// Database query
	err := repo.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(repo.makeKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return gob.NewDecoder(bytes.NewReader(val)).Decode(&data)
		})
	})
	return &data, err
}
func (repo *BaseRepoULID[T]) Set(id ulid.ULID, entry *T) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
			return err
		}
		return txn.Set(repo.makeKey(id), buf.Bytes())
	})
}
func (repo *BaseRepoULID[T]) Delete(id ulid.ULID) error {
	// Delete
	return repo.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(repo.makeKey(id))
	})
}

type CacheDB struct {
	source *DataSource

	taskRepo *BaseRepoULID[Task]
	resRepo  *BaseRepoULID[taskRes]
}

func NewCacheDB(source *DataSource) *CacheDB {
	return &CacheDB{source: source}
}
func (db *CacheDB) getDB() *badger.DB {
	return db.source.cache
}
func (db *CacheDB) GetTaskRepo() *BaseRepoULID[Task] {
	if db.taskRepo == nil {
		db.taskRepo = &BaseRepoULID[Task]{prefix: dbKeyPrefixTaskID, db: db.getDB()}
	}
	return db.taskRepo
}
func (db *CacheDB) GetTaskResRepo() *BaseRepoULID[taskRes] {
	if db.resRepo == nil {
		db.resRepo = &BaseRepoULID[taskRes]{prefix: dbKeyPrefixResID, db: db.getDB()}
	}
	return db.resRepo
}
