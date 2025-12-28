#  EmojiDB: The Total Emoji Encrypted Database

EmojiDB is a high-performance, embedded database designed for maximum security and visual fun. Every record, every header, and even your schema definition is strictly 100% Emoji encoded.

## 🚀 Getting Started: Standard Workflow

EmojiDB follows a stage-by-stage progression from absolute security to data persistence.

### Stage 1: Security Initialization 🔐
Before creating a database, you must initialize the Master Security Layer. This generates a one-time `secure.pem` file which acts as your "set of emojis" for recovery and authorization.

```go
db, _ := core.Open("mydata.db", "my-secret-key")
err := db.Secure() // Creates secure.pem
```

### Stage 2: Opening & Persistence 📦
EmojiDB automatically manages your database inside a consolidated `emojidb/` folder:
- `emojidb/[dbname].db`: The actual encrypted data.
- `emojidb/[dbname].safety`: The crash-recovery buffer.
- `emojidb/[dbname].schema.json`: Readable schema definitions (Prisma-style).

```go
// core.Open automatically handles folder creation and artifact routing
db, err := core.Open("my.db", "showcase-secret-2025")
defer db.Close()
```

### Stage 3: Schema Management 📐
Like Prisma, EmojiDB uses persistent schemas. These are saved as plain JSON in `emojidb/[dbname].schema.json` for easy inspection.

```go
fields := []core.Field{
    {Name: "id", Type: core.FieldTypeInt, Unique: true},
    {Name: "name", Type: core.FieldTypeString},
}

// Initial definition
db.DefineSchema("users", fields)
```

### Stage 4: Schema Evolution (The Prisma Way) 🔄
You can update your schema and check for conflicts before applying.

```go
// 1. Check for conflicts (Prisma-like Pull/Diff)
report := db.DiffSchema("users", newFields)
if report.Destructive {
    fmt.Println("Warning: Field removal detected!")
}

// 2. Sync if compatible (Prisma-like Push)
err := db.SyncSchema("users", newFields)
```

### Stage 5: Data Operations ⚡
EmojiDB is extremely fast (~45ms for 1500 operations).

```go
// Insert
db.Insert("users", core.Row{"id": 1, "name": "Alice"})

// Query
results, _ := query.NewQuery(db, "users").Filter(...).Execute()
```

## 🛠️ Features

- **Total Emoji Encoding**: 😵🤮😇🤒😷 - your raw data never touches the disk.
- **AES-GCM Encryption**: Military-grade security on every clump.
- **Master Key Recovery**: Use `secure.pem` emoji sequences to rotate your database secret.
- **Unique Constraints**: O(1) performance for uniqueness checks.
- **Safety Engine**: Parallelized batch recovery for zero data loss.

## 🏁 Performance Benchmarks
*Tested with 1500 records + Unique Keys + Full Disk Re-encryption*

| Operation | Timing |
| :--- | :--- |
| **Ingest 1500 Rows** | ~9.1ms |
| **Flush to Disk** | ~3.9ms |
| **Rotate Master Key** | ~7.2ms |
| **TOTAL SHOWCASE** | **~45.3ms** |
## 🌐 Node.js Integration (Standalone SDK)

EmojiDB provides a standalone Node.js SDK that **automatically downloads** the required Go engine for your platform. No Go installation required!

### 1. Installation
```bash
npm install @ikwerre-dev/emojidb
```

### 2. Standalone Usage
```javascript
import EmojiDB from '@ikwerre-dev/emojidb';
const db = new EmojiDB(); // Auto-detects OS/Arch and downloads from GitHub

async function start() {
    await db.connect();
    await db.open('prod.db', 'my-secret');
    
    await db.insert('users', { id: 1, name: 'Alice' });
    const users = await db.query('users', { id: 1 });
}
```

### 📦 Publishing Pre-built Binaries
For the standalone mode to work, you must upload pre-compiled engines to your **GitHub Releases** with the following naming convention:
- `emojidb-darwin-arm64` (Mac M1/M2)
- `emojidb-linux-x64` (Linux Servers)
- `emojidb-win32-x64.exe` (Windows)

---
*EmojiDB: Zero-dependency, military-grade security.*