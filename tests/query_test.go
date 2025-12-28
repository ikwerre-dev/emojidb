package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
)

func TestQuery(t *testing.T) {
	dbPath := "test_query.db"
	defer os.Remove(dbPath)

	db, err := core.Open(dbPath, "secret", false)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
		{Name: "age", Type: core.FieldTypeInt},
	}
	db.DefineSchema("users", fields)

	db.Insert("users", core.Row{"id": 1, "name": "alice", "age": 30})
	db.Insert("users", core.Row{"id": 2, "name": "bob", "age": 25})
	db.Insert("users", core.Row{"id": 3, "name": "charlie", "age": 35})

	q := query.NewQuery(db, "users")
	results, err := q.Filter(func(r core.Row) bool {
		age, ok := r["age"].(int)
		return ok && age > 28
	}).Execute()

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}

	q2 := query.NewQuery(db, "users")
	results, _ = q2.Filter(func(r core.Row) bool { return r["name"] == "bob" }).
		Select("name").
		Execute()

	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}

	if len(results[0]) != 1 {
		t.Errorf("expected 1 col, got %d", len(results[0]))
	}

	if _, ok := results[0]["age"]; ok {
		t.Error("age column found")
	}
}
