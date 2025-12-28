package core

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ikwerre-dev/emojidb/crypto"
	"github.com/ikwerre-dev/emojidb/storage"
)

type Config struct {
	MemoryLimitMB   int
	ClumpSizeMB     int
	FlushIntervalMS int
}

type Database struct {
	Mu         sync.RWMutex
	Path       string
	Key        string
	File       *os.File
	SafetyFile *os.File
	Config     *Config
	Schemas    map[string]*Schema
	Tables     map[string]*Table
	Orphans    map[string][]*SealedClump
	SyncSafety bool
}

type Table struct {
	Mu            sync.RWMutex
	Db            *Database
	Name          string
	Schema        *Schema
	HotHeap       *HotHeap
	SealedClumps  []*SealedClump
	UniqueIndices map[string]map[interface{}]struct{}
}

func Open(path, key string) (*Database, error) {
	if key == "" {
		return nil, errors.New("database key is required")
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	safetyPath := path + ".safety"
	sFile, err := os.OpenFile(safetyPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		file.Close()
		return nil, err
	}

	db := &Database{
		Path:       path,
		Key:        key,
		File:       file,
		SafetyFile: sFile,
		Config:     &Config{},
		Schemas:    make(map[string]*Schema),
		Tables:     make(map[string]*Table),
		Orphans:    make(map[string][]*SealedClump),
		SyncSafety: true,
	}

	if err := db.Load(); err != nil {
		file.Close()
		sFile.Close()
		return nil, err
	}
	return db, nil
}

func (db *Database) Load() error {
	handleClump := func(tableName string, data []byte) error {
		var clump SealedClump
		if err := json.Unmarshal(data, &clump); err != nil {
			return err
		}
		db.Mu.Lock()
		table, ok := db.Tables[tableName]
		if ok {
			table.SealedClumps = append(table.SealedClumps, &clump)
		} else {
			db.Orphans[tableName] = append(db.Orphans[tableName], &clump)
		}
		db.Mu.Unlock()
		return nil
	}

	return storage.Load(db.File, &db.Mu, db.Key, crypto.Decrypt, handleClump)
}

func (db *Database) PersistClump(tableName string, clump *SealedClump) error {
	return storage.PersistClump(db.File, &db.Mu, tableName, clump, db.Key, crypto.Encrypt, crypto.EncodeToEmojis)
}

func (db *Database) Secure() error {
	path := filepath.Join(filepath.Dir(db.Path), "secure.pem")
	if _, err := os.Stat(path); err == nil {
		return errors.New("security already initialized")
	}

	rawKey := make([]byte, 32)
	crypto.RandRead(rawKey)
	emojiKey := crypto.EncodeToEmojis(rawKey)

	return os.WriteFile(path, []byte(emojiKey), 0600)
}

func (db *Database) ChangeKey(newKey string, masterKey string) error {
	path := filepath.Join(filepath.Dir(db.Path), "secure.pem")
	actualMaster, err := os.ReadFile(path)
	if err != nil {
		return errors.New("security not initialized or secure.pem missing")
	}

	if string(actualMaster) != masterKey {
		return errors.New("invalid master key provided")
	}

	db.Mu.Lock()
	defer db.Mu.Unlock()

	// 1. Truncate and reset file
	err = db.File.Truncate(0)
	if err != nil {
		return err
	}
	_, err = db.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// 2. Write new header
	if err := storage.WriteHeader(db.File); err != nil {
		return err
	}

	// 3. Temporarily set new key to re-persist
	oldKey := db.Key
	db.Key = newKey

	// 4. Re-persist all existing sealed clumps with new key
	for tableName, table := range db.Tables {
		table.Mu.RLock()
		for _, clump := range table.SealedClumps {
			if err := storage.PersistClump(db.File, &db.Mu, tableName, clump, db.Key, crypto.Encrypt, crypto.EncodeToEmojis); err != nil {
				db.Key = oldKey // Rollback key if failed
				table.Mu.RUnlock()
				return err
			}
		}
		table.Mu.RUnlock()
	}

	return nil
}

func (db *Database) DumpAsJSON(tableName string) (string, error) {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return "", errors.New("table not found")
	}

	table.Mu.RLock()
	defer table.Mu.RUnlock()

	var allRows []Row
	// Include clumps
	for _, clump := range table.SealedClumps {
		allRows = append(allRows, clump.Rows...)
	}
	// Include hot heap
	if table.HotHeap != nil {
		allRows = append(allRows, table.HotHeap.Rows...)
	}

	data, err := json.MarshalIndent(allRows, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (db *Database) Close() error {
	db.Mu.Lock()
	defer db.Mu.Unlock()
	if db.SafetyFile != nil {
		db.SafetyFile.Close()
	}
	if db.File != nil {
		return db.File.Close()
	}
	return nil
}
