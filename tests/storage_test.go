package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
)

func TestPersistence(t *testing.T) {
	dbPath := "test_persist.db"
	defer os.Remove(dbPath)

	db, err := core.Open(dbPath, "secret", true)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	fields := []core.Field{{Name: "id", Type: core.FieldTypeInt}}
	db.DefineSchema("items", fields)

	table := db.Tables["items"]
	table.HotHeap.MaxRows = 1

	err = db.Insert("items", core.Row{"id": 100})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	db.Close()

	// Re-open
	db2, err := core.Open(dbPath, "secret", true)
	if err != nil {
		t.Fatalf("failed re-open: %v", err)
	}
	defer db2.Close()

	db2.DefineSchema("items", fields)

	table2, ok := db2.Tables["items"]
	if !ok {
		t.Fatal("table items not found")
	}

	if len(table2.SealedClumps) != 1 {
		t.Errorf("expected 1, got %d", len(table2.SealedClumps))
	}

	found := false
	for _, row := range table2.SealedClumps[0].Rows {
		val, ok := row["id"].(float64) // JSON unmarshal uses float64
		if ok && val == 100 {
			found = true
			break
		}
	}
	if !found {
		t.Error("data not found")
	}
}
