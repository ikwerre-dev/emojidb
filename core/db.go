package core

import (
	"encoding/json"
	"os"
	"sync"
	"github.com/ikwerre-dev/emojidb/crypto"
	"github.com/ikwerre-dev/emojidb/storage"
)

type Config struct {
	MemoryLimitMB   int
	ClumpSizeMB     int
	FlushIntervalMS int
	Encrypt         bool
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
}

type Table struct {
	Mu           sync.RWMutex
	Db           *Database
	Name         string
	Schema       *Schema
	HotHeap      *HotHeap
	SealedClumps []*SealedClump
}

func Open(path, key string, encrypt bool) (*Database, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	sFile, err := os.OpenFile("safety.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		file.Close()
		return nil, err
	}

	db := &Database{
		Path:       path,
		Key:        key,
		File:       file,
		SafetyFile: sFile,
		Config:     &Config{Encrypt: encrypt},
		Schemas:    make(map[string]*Schema),
		Tables:     make(map[string]*Table),
		Orphans:    make(map[string][]*SealedClump),
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

	return storage.Load(db.File, &db.Mu, db.Key, crypto.Decrypt, crypto.DecodeFromEmojis, handleClump)
}

func (db *Database) PersistClump(tableName string, clump *SealedClump) error {
	return storage.PersistClump(db.File, &db.Mu, tableName, clump, db.Config.Encrypt, db.Key, crypto.Encrypt, crypto.EncodeToEmojis)
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
