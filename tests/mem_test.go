package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
)

func TestInsert(t *testing.T) {
	dbPath := "test_insert.db"
	defer os.Remove(dbPath)

	db, err := core.Open(dbPath, "secret", true)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
	}
	db.DefineSchema("users", fields)

	err = db.Insert("users", core.Row{"id": 1, "name": "alice"})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	table := db.Tables["users"]
	if len(table.HotHeap.Rows) != 1 {
		t.Errorf("expected 1, got %d", len(table.HotHeap.Rows))
	}

	err = db.Insert("users", core.Row{"id": 2})
	if err == nil {
		t.Error("expected error missing field")
	}
}

func TestClumpSealing(t *testing.T) {
	dbPath := "test_seal.db"
	defer os.Remove(dbPath)

	db, err := core.Open(dbPath, "secret", true)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	fields := []core.Field{{Name: "id", Type: core.FieldTypeInt}}
	db.DefineSchema("items", fields)

	table := db.Tables["items"]
	table.HotHeap.MaxRows = 3

	db.Insert("items", core.Row{"id": 1})
	db.Insert("items", core.Row{"id": 2})
	db.Insert("items", core.Row{"id": 3})

	if len(table.SealedClumps) != 1 {
		t.Errorf("expected 1, got %d", len(table.SealedClumps))
	}
	if len(table.HotHeap.Rows) != 0 {
		t.Errorf("expected 0, got %d", len(table.HotHeap.Rows))
	}

	db.Insert("items", core.Row{"id": 4})
	if len(table.HotHeap.Rows) != 1 {
		t.Errorf("expected 1, got %d", len(table.HotHeap.Rows))
	}
}
