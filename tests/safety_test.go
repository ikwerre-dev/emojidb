package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
	"github.com/ikwerre-dev/emojidb/safety"
)

func TestSafetyEngine(t *testing.T) {
	dbPath := "test_safety.db"
	safetyPath := "safety.db"
	defer os.Remove(dbPath)
	defer os.Remove(safetyPath)

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
	db.Insert("users", core.Row{"id": 1, "name": "alice"})

	filter := func(r core.Row) bool {
		val := r["id"]
		switch v := val.(type) {
		case int:
			return v == 1
		case int64:
			return v == 1
		}
		return false
	}

	err = safety.Update(db, "users", safety.FilterFunc(filter), core.Row{"name": "alice_updated"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	q := query.NewQuery(db, "users")
	results, _ := q.Filter(query.FilterFunc(filter)).Execute()
	if len(results) == 0 || results[0]["name"] != "alice_updated" {
		t.Fatalf("update failed in memory: %v", results)
	}

	points, err := safety.ListRecoveryPoints(db)
	if err != nil {
		t.Fatalf("list recovery failed: %v", err)
	}

	if len(points) == 0 {
		t.Fatal("expected recovery point")
	}

	err = safety.Restore(db, points[0], true)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	q2 := query.NewQuery(db, "users")
	results, _ = q2.Filter(func(r core.Row) bool {
		val, ok := r["name"].(string)
		return ok && val == "alice"
	}).Execute()

	if len(results) != 1 {
		t.Errorf("expected 1 restored, got %d", len(results))
	}
}
