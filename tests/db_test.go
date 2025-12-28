package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
)

func TestOpen(t *testing.T) {
	dbPath := "test_open.db"
	defer os.Remove(dbPath)

	db, err := core.Open(dbPath, "secret", true)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	if db.Path != dbPath {
		t.Errorf("expected path %s, got %s", dbPath, db.Path)
	}
	if !db.Config.Encrypt {
		t.Errorf("expected encryption enabled")
	}
}

func TestDefineSchema(t *testing.T) {
	dbPath := "test_schema.db"
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

	err = db.DefineSchema("users", fields)
	if err != nil {
		t.Fatalf("failed to define: %v", err)
	}

	schema, ok := db.Schemas["users"]
	if !ok {
		t.Fatal("schema not found")
	}

	if len(schema.Fields) != 2 {
		t.Errorf("expected 2, got %d", len(schema.Fields))
	}

	err = db.DefineSchema("users", fields)
	if err == nil {
		t.Error("expected error redefine")
	}
}
