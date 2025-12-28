package core

import (
	"errors"
)

type FieldType int

const (
	FieldTypeInt FieldType = iota
	FieldTypeString
	FieldTypeFloat
	FieldTypeBool
)

type Field struct {
	Name string
	Type FieldType
}

type Schema struct {
	Version int
	Fields  []Field
}

func (db *Database) DefineSchema(tableName string, fields []Field) error {
	db.Mu.Lock()
	defer db.Mu.Unlock()

	if _, ok := db.Schemas[tableName]; ok {
		return errors.New("schema already defined for table: " + tableName)
	}

	schema := &Schema{
		Version: 1,
		Fields:  fields,
	}

	db.Schemas[tableName] = schema

	if table, ok := db.Tables[tableName]; ok {
		table.Schema = schema
	} else {
		db.Tables[tableName] = &Table{
			Db:      db,
			Name:    tableName,
			Schema:  schema,
			HotHeap: NewHotHeap(1000),
		}
		if orphans, ok := db.Orphans[tableName]; ok {
			db.Tables[tableName].SealedClumps = orphans
			delete(db.Orphans, tableName)
		}
	}

	return nil
}
