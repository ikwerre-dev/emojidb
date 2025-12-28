package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
	"github.com/ikwerre-dev/emojidb/safety"
)

func TestFullShowcase(t *testing.T) {
	wd, _ := os.Getwd()
	dbPath := filepath.Join(wd, "showcase.db")
	dumpPath := filepath.Join(wd, "dump.json")
	safetyPath := dbPath + ".safety"

	key := "showcase-secret-2025"

	totalStart := time.Now()
	var timings []struct {
		name string
		took time.Duration
	}

	fmt.Println("\nSTARTING EMOJIDB FULL SHOWCASE (Multi-Table + Unique Keys)")
	fmt.Println("==========================================================")

	// cleanup
	os.Remove(dbPath)
	os.Remove(safetyPath)
	os.Remove(dumpPath)
	os.Remove(filepath.Join(wd, "secure.pem"))

	// 1. Open Database
	start := time.Now()
	fmt.Printf("1. Opening Database (Encryption Mandatory)\n")
	db, err := core.Open(dbPath, key)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer db.Close()
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Open Database", time.Since(start)})

	// 2. Define Schemas (Multi-Table + Unique Keys)
	start = time.Now()
	fmt.Println("2. Defining Schemas: 'products' and 'orders' (Unique Keys enabled)")

	productFields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt, Unique: true},
		{Name: "name", Type: core.FieldTypeString},
		{Name: "price", Type: core.FieldTypeInt},
		{Name: "category", Type: core.FieldTypeString},
	}
	err = db.DefineSchema("products", productFields)
	if err != nil {
		t.Fatalf("Failed products schema: %v", err)
	}

	orderFields := []core.Field{
		{Name: "order_id", Type: core.FieldTypeInt, Unique: true},
		{Name: "product_id", Type: core.FieldTypeInt},
		{Name: "customer", Type: core.FieldTypeString},
	}
	err = db.DefineSchema("orders", orderFields)
	if err != nil {
		t.Fatalf("Failed orders schema: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Define 2 Schemas", time.Since(start)})

	// 3. Ingestion (1500 records)
	start = time.Now()
	fmt.Println("3. Ingesting 1500 records with Unique Key validation")
	for i := 1; i <= 1500; i++ {
		category := "tech"
		if i%3 == 0 {
			category = "food"
		} else if i%5 == 0 {
			category = "home"
		}

		row := core.Row{
			"id":       i,
			"name":     fmt.Sprintf("Product %d", i),
			"price":    i * 10,
			"category": category,
		}
		err := db.Insert("products", row)
		if err != nil {
			t.Fatalf("Failed insert at %d: %v", i, err)
		}
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Ingest 1500 Rows", time.Since(start)})

	// 4. Test Unique Constraint Violation
	start = time.Now()
	fmt.Println("4. Testing Unique Constraint Violation (expected failure)")
	dupRow := core.Row{"id": 1, "name": "Dup", "price": 0, "category": "none"}
	err = db.Insert("products", dupRow)
	if err == nil {
		t.Fatal("Expected unique constraint violation error, got nil")
	}
	fmt.Printf("   Caught expected error: %v\n", err)
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Unique Constraint Test", time.Since(start)})

	// 5. Bulk Update (50 records)
	start = time.Now()
	fmt.Println("5. Safety Engine: Bulk Updating 50 records")
	err = safety.Update(db, "products", func(r core.Row) bool {
		id, ok := r["id"].(int)
		return ok && id >= 1100 && id < 1150
	}, core.Row{"category": "updated_bulk"})
	if err != nil {
		t.Fatalf("Bulk update failed: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Bulk Update (50)", time.Since(start)})

	// 6. Optimized Single Updates (5 records)
	start = time.Now()
	fmt.Println("6. Optimized Single Updating 5 records (Deferred Sync)")
	db.SyncSafety = false // Disable per-op sync
	for i := 1200; i < 1205; i++ {
		targetID := i
		safety.Update(db, "products", func(r core.Row) bool {
			id, _ := r["id"].(int)
			return id == targetID
		}, core.Row{"name": "Updated Single"})
	}
	safety.CommitSafety(db) // Sync once
	db.SyncSafety = true    // Re-enable
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Single Update (5) - OPT", time.Since(start)})

	// 7. Bulk Delete (50 records)
	start = time.Now()
	fmt.Println("7. Safety Engine: Bulk Deleting 50 records")
	err = safety.Delete(db, "products", func(r core.Row) bool {
		id, ok := r["id"].(int)
		return ok && id >= 1300 && id < 1350
	})
	if err != nil {
		t.Fatalf("Bulk delete failed: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Bulk Delete (50)", time.Since(start)})

	// 8. Optimized Single Deletes (5 records)
	start = time.Now()
	fmt.Println("8. Optimized Single Deleting 5 records (Deferred Sync)")
	db.SyncSafety = false // Disable per-op sync
	for i := 1400; i < 1405; i++ {
		targetID := i
		safety.Delete(db, "products", func(r core.Row) bool {
			id, _ := r["id"].(int)
			return id == targetID
		})
	}
	safety.CommitSafety(db) // Sync once
	db.SyncSafety = true    // Re-enable
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Single Delete (5) - OPT", time.Since(start)})

	// 9. Multi-Table Check: Insert into 'orders'
	start = time.Now()
	fmt.Println("9. Multi-Table: Ingesting into 'orders'")
	db.Insert("orders", core.Row{"order_id": 101, "product_id": 1, "customer": "Alice"})
	db.Insert("orders", core.Row{"order_id": 102, "product_id": 2, "customer": "Bob"})
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Multi-Table Append", time.Since(start)})

	// 10. Persistence (Flush)
	start = time.Now()
	fmt.Println("10. Flushing 'products' to Disk (Total Emoji Encoding)")
	db.Flush("products")
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Flush to Disk", time.Since(start)})

	// 11. Inspect File
	start = time.Now()
	fmt.Println("11. Inspecting Disk Content (Should be 100% Emojis)")
	content, _ := os.ReadFile(dbPath)
	if len(content) > 60 {
		fmt.Printf("   Preview: %s...\n", string(content[:60]))
	} else {
		fmt.Printf("   Content: %s\n", string(content))
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Inspect File", time.Since(start)})

	// 12. Query Engine
	start = time.Now()
	fmt.Println("12. Running Fluent Query: Category = 'updated_bulk'")
	results, err := query.NewQuery(db, "products").Filter(func(r core.Row) bool {
		cat, _ := r["category"].(string)
		return cat == "updated_bulk"
	}).Execute()

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("   Query Result: found %d matches\n", len(results))
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Execute Query", time.Since(start)})

	// 13. JSON Dump to File
	start = time.Now()
	fmt.Println("13. Dumping 'products' to dump.json")
	jsonDump, err := db.DumpAsJSON("products")
	if err != nil {
		t.Fatalf("Dump failed: %v", err)
	}

	err = os.WriteFile(dumpPath, []byte(jsonDump), 0644)
	if err != nil {
		t.Fatalf("Failed to write dump: %v", err)
	}
	fmt.Printf("   Saved to: %s\n", dumpPath)
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"JSON Export", time.Since(start)})

	// 14. Advanced Security: Generating secure.pem
	start = time.Now()
	fmt.Println("14. Advanced Security: Generating secure.pem (Master Key)")
	err = db.Secure()
	if err != nil {
		t.Fatalf("Secure operation failed: %v", err)
	}
	pemPath := filepath.Join(filepath.Dir(dbPath), "secure.pem")
	masterKeyBytes, err := os.ReadFile(pemPath)
	if err != nil {
		t.Fatalf("Failed to read secure.pem: %v", err)
	}
	masterKey := string(masterKeyBytes)
	fmt.Printf("   Master Key Generated: %s...\n", string(masterKey[:20]))
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Generate Secure PEM", time.Since(start)})

	// 15. Advanced Security: Authorizing Key Rotation
	start = time.Now()
	fmt.Println("15. Advanced Security: Authorizing Key Rotation (Full Disk Re-encryption)")
	newSecret := "rotated-secret-2026"
	err = db.ChangeKey(newSecret, masterKey)
	if err != nil {
		t.Fatalf("Key rotation failed: %v", err)
	}
	fmt.Printf("   Success: Database re-encrypted with new key\n")
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Rotate Master Key", time.Since(start)})

	fmt.Println("\nTEST SUMMARY")
	fmt.Println("==========================================================")
	for _, t := range timings {
		ms := float64(t.took.Nanoseconds()) / 1e6
		fmt.Printf("%-30s : %.3fms\n", t.name, ms)
	}
	fmt.Println("----------------------------------------------------------")
	totalMs := float64(time.Since(totalStart).Nanoseconds()) / 1e6
	fmt.Printf("%-30s : %.3fms\n", "TOTAL TIME", totalMs)
	fmt.Println("==========================================================")
}
