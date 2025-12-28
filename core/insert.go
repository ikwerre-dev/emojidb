package core

import (
	"errors"
)

func (db *Database) Insert(tableName string, record Row) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found: " + tableName)
	}

	return table.Insert(record)
}
