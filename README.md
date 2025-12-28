# EmojiDB

EmojiDB is a memory-first, 100% emoji-encoded embedded database engine written in Go.

## Features
- **Total Emoji Storage**: Every byte on disk (magic, metadata, payload) is stored as emojis.
- **Mandatory Encryption**: AES-GCM encryption is enforced for all stored data.
- **Memory-First**: High-speed ingestion using Hot Heaps and batch clumping.
- **Safety Engine**: Automatic 30-minute recovery window for updates and deletes.
- **Fluent Query API**: Simple, chainable filtering and projection.

## Usage

```go
package main

import (
	"fmt"
	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
)

func main() {
	// Open database (Key is mandatory)
	db, _ := core.Open("data.db", "your-secret-key")
	defer db.Close()

	// Define schema
	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
	}
	db.DefineSchema("users", fields)

	// Insert data
	db.Insert("users", core.Row{"id": 1, "name": "alice"})

	// Force persistence to disk
	db.Flush("users")

	// Query data
	q := query.NewQuery(db, "users")
	results, _ := q.Filter(func(r core.Row) bool {
		return r["name"] == "alice"
	}).Execute()

	fmt.Println(results)
}
```

## Clumping & Performance

EmojiDB uses a **Memory-First, Append-Only** architecture. 

### How Clumping Works
1. **Hot Heap**: New rows are stored in a "Hot Heap" in memory.
2. **Sealing**: When the heap reaches a limit (e.g., 1000 rows or 1MB), it is "sealed."
3. **Persistence**: The sealed heap (a "Clump") is encrypted, emoji-encoded, and appended to the `.db` file.
4. **Efficiency**: Instead of writing every row individually, we write large batches. This minimizes disk I/O and keeps the system fast even with high ingestion rates.

### Testing Performance
You can test the "internet speed" (throughput) by inserting 100,000 rows and measuring the time.
```go
start := time.Now()
for i := 0; i < 100000; i++ {
    db.Insert("users", row)
}
fmt.Printf("Ingestion Rate: %f rows/sec\n", 100000/time.Since(start).Seconds())
```

### Query Testing
Use the fluent API to verify clumping. If data is persisted, queries will scan both the memory (Hot Heap) and the disk (Sealed Clumps).
```go
results, _ := query.NewQuery(db, "users").Filter(f).Execute()
```

## Running in JavaScript

Since EmojiDB is written in Go, you have two main ways to use it in JS:

### 1. WebAssembly (WASM)
Compile EmojiDB to WASM to run directly in the browser or Node.js.
```bash
GOOS=js GOARCH=wasm go build -o emojidb.wasm
```

### 2. Sidecar Service
Run a small Go service as a "sidecar" that provides a REST or JSON-RPC API for your JS frontend. 

## Advanced Security & Key Management

EmojiDB features a multi-layered security system:
- **0600 File Permissions**: Database and safety files are restricted to the owner by default.
- **One-Time Secure PEM**: Run `db.Secure()` to generate a `secure.pem` containing a master emoji-encoded key for recovery and authorization.
- **Key Rotation**: Rotate your database secrets using `db.ChangeKey(newKey, securePemPath)`. This process re-encrypts all data on disk with the new key.

## Dashboard & Connectivity

EmojiDB includes a high-end, minimalistic dashboard built with Next.js for real-time data exploration and management.

### Starting the Bridge Server
To connect the dashboard, run the WebSocket bridge:
```bash
go run main.go
```

### Dashboard UI
1. Navigate to the `dashboard/` directory.
2. Install dependencies: `npm install`.
3. Run the dashboard: `npm run dev`.
4. Connect using your database path and secret key.

## Performance Benchmarks
EmojiDB is optimized for microsecond-precision operations:
- **Bulk Ingestion (1500)**: < 10ms
- **Bulk Delete (50)**: < 5ms
- **Total Showcase Flow**: < 40ms

## Full Feature Showcase
To see a complete end-to-end demonstration of all core and advanced features:
```bash
go test -v tests/full_integration_test.go
```
