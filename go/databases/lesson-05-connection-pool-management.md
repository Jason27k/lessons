# Lesson 5: Connection Pool Management

## The Hidden Pool

Every time you've called `db.Query`, `db.Exec`, or `db.QueryRow`, you weren't managing connections yourself. Behind the scenes, `database/sql` maintains a **connection pool** — a set of reusable database connections that are checked out, used, and returned automatically.

You've been using it this whole time without knowing it. This lesson is about understanding what it does and how to tune it.

---

## 1. How the Pool Works

When you call `sql.Open`, **no connection is created**. You already learned that `db.Ping()` forces the first actual connection. But what happens after that?

```
Your code                     Connection Pool                  Database
─────────                     ───────────────                  ────────
db.Query(...)  ──────────►   Check out a connection  ────────►  Execute SQL
                              (create one if needed)
               ◄──────────   Return connection       ◄────────  Results
                              (keep it idle for reuse)
```

The pool handles three things:
1. **Reusing connections** — instead of opening/closing for every query
2. **Limiting connections** — so you don't overwhelm the database
3. **Cleaning up** — closing connections that have been idle too long

---

## 2. Pool Configuration

`*sql.DB` exposes four knobs:

### `SetMaxOpenConns(n int)`

Maximum number of connections open to the database at once (both in-use and idle).

```go
db.SetMaxOpenConns(25)
```

- **Default: 0 (unlimited)** — the pool will create as many connections as needed
- If all connections are in use, the next caller **blocks** until one becomes available
- For SQLite: set this to **1** because SQLite doesn't handle concurrent writers well
- For Postgres/MySQL: 25-50 is a common starting point

### `SetMaxIdleConns(n int)`

Maximum number of connections kept idle in the pool, ready for reuse.

```go
db.SetMaxIdleConns(5)
```

- **Default: 2**
- Idle connections avoid the overhead of establishing new ones
- Too many idle connections waste memory; too few means frequent reconnection
- Must be ≤ `MaxOpenConns` (if higher, it's silently reduced)

### `SetConnMaxLifetime(d time.Duration)`

Maximum total time a connection can exist before being closed.

```go
db.SetConnMaxLifetime(30 * time.Minute)
```

- **Default: 0 (no limit)** — connections live forever
- Useful when the database or a load balancer drops connections after a timeout
- The pool lazily closes expired connections — it doesn't interrupt active queries

### `SetConnMaxIdleTime(d time.Duration)`

Maximum time a connection can sit idle before being closed.

```go
db.SetConnMaxIdleTime(5 * time.Minute)
```

- **Default: 0 (no limit)**
- Cleans up connections after traffic spikes — if you had 50 connections during a burst but only need 5 now, the rest get closed
- Different from `ConnMaxLifetime`: this is about idle time, not total age

---

## 3. Practical Configuration

For SQLite (what you're using now):

```go
db.SetMaxOpenConns(1)       // SQLite only supports one writer at a time
db.SetMaxIdleConns(1)       // Only need one idle connection
db.SetConnMaxLifetime(0)    // No need — local file, no network timeouts
```

For a network database (Postgres, MySQL) in production:

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(30 * time.Minute)
db.SetConnMaxIdleTime(5 * time.Minute)
```

These are starting points — you'd tune based on actual load and monitoring.

---

## 4. `db.Stats()` — Monitoring the Pool

`db.Stats()` returns a `sql.DBStats` struct with real-time pool information:

```go
stats := db.Stats()
fmt.Printf("Open connections:    %d\n", stats.OpenConnections)
fmt.Printf("In use:              %d\n", stats.InUse)
fmt.Printf("Idle:                %d\n", stats.Idle)
fmt.Printf("Wait count:          %d\n", stats.WaitCount)
fmt.Printf("Wait duration:       %s\n", stats.WaitDuration)
fmt.Printf("Max idle closed:     %d\n", stats.MaxIdleClosed)
fmt.Printf("Max lifetime closed: %d\n", stats.MaxLifetimeClosed)
fmt.Printf("Max idle time closed:%d\n", stats.MaxIdleTimeClosed)
```

| Field | What it tells you |
|-------|-------------------|
| `OpenConnections` | Total connections currently open (in-use + idle) |
| `InUse` | Connections currently executing queries |
| `Idle` | Connections sitting in the pool waiting to be reused |
| `WaitCount` | How many times a goroutine had to **wait** for a connection (pool was full) |
| `WaitDuration` | Total time spent waiting for connections |
| `MaxIdleClosed` | Connections closed because `MaxIdleConns` was exceeded |
| `MaxLifetimeClosed` | Connections closed because `ConnMaxLifetime` expired |
| `MaxIdleTimeClosed` | Connections closed because `ConnMaxIdleTime` expired |

**`WaitCount` and `WaitDuration` are the most important.** If these are high, your pool is too small — goroutines are waiting for connections instead of doing work.

---

## 5. Pool Exhaustion

Pool exhaustion happens when all connections are in use and a new query needs one. The caller blocks until a connection is returned.

### Common causes:

**Forgetting to close `rows`:**

```go
// BUG: rows.Close() is never called — this connection is leaked
rows, err := db.Query("SELECT * FROM accounts")
if err != nil {
    log.Fatal(err)
}
for rows.Next() {
    // process rows...
}
// Missing: defer rows.Close() or rows.Close()
// The connection is NEVER returned to the pool
```

Every open `*sql.Rows` holds a connection. If you don't close it, that connection is gone forever. With `MaxOpenConns` set, you'll eventually run out and everything blocks.

**Long-running transactions:**

```go
tx, err := db.Begin()
// This transaction holds a connection until Commit or Rollback
// If you forget to end it, the connection is stuck
```

**Fix:** always `defer rows.Close()` and `defer tx.Rollback()`.

---

## 6. Connection Lifecycle Summary

```
sql.Open()          → Pool created, zero connections
db.Ping()           → First connection established
db.Query()          → Connection checked out (or created)
rows.Close()        → Connection returned to pool (idle)
                    → Idle connection reused by next query
                    → Idle too long? Closed by pool
                    → Too old? Closed by pool
db.Close()          → All connections closed, pool shut down
```

---

## Key Rules

1. **Always set `MaxOpenConns`** — the unlimited default can overwhelm a database
2. **For SQLite, set `MaxOpenConns(1)`** — SQLite serializes writes anyway
3. **Always close `rows` and end transactions** — leaked connections exhaust the pool
4. **Monitor `WaitCount`** — if it's growing, your pool is too small
5. **`sql.Open` doesn't connect** — the pool creates connections on demand

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Leaving `MaxOpenConns` at 0 | Unlimited connections can overwhelm the database |
| Setting `MaxIdleConns` > `MaxOpenConns` | The extra idle slots are wasted (silently capped) |
| Not closing `*sql.Rows` | Leaks a connection — pool slowly exhausts |
| Opening multiple `sql.DB` instances | Each has its own pool — defeats the purpose of pooling |
| Calling `sql.Open` per request | The pool IS the reuse mechanism — open once, share everywhere |

---

## Your Turn

### Exercise: Pool Configuration & Monitoring

Write a program that:

1. Opens a SQLite database and configures the pool:
   - `MaxOpenConns`: 1
   - `MaxIdleConns`: 1
   - `ConnMaxLifetime`: 1 hour
2. Creates an `accounts` table (same as lesson 4) and seeds Alice (1000) and Bob (500)
3. Writes a function `printStats(db *sql.DB)` that prints `OpenConnections`, `InUse`, `Idle`, and `WaitCount` from `db.Stats()`
4. Calls `printStats` at these points and observes how the numbers change:
   - After `sql.Open` (before any queries)
   - After `db.Ping()`
   - During an open `db.Query` (before closing rows)
   - After closing the rows
5. **Bonus:** Demonstrate pool exhaustion — start a query (hold `rows` open without closing), then try to run another query. With `MaxOpenConns(1)`, the second query will block. Use a goroutine with a timeout to show it hanging.

**Hint for the bonus:** you can use `context.WithTimeout` with `db.QueryContext` to make the second query time out instead of blocking forever:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
_, err := db.QueryContext(ctx, "SELECT 1")
// err will be context.DeadlineExceeded if it couldn't get a connection
```

---

<details>
<summary>Full Answer</summary>

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func printStats(label string, db *sql.DB) {
	stats := db.Stats()
	fmt.Printf("[%s]\n", label)
	fmt.Printf("  Open: %d  InUse: %d  Idle: %d  WaitCount: %d\n\n",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
}

func main() {
	db, err := sql.Open("sqlite", "pool-demo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Configure the pool
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Hour)

	// 1. After sql.Open — no connections yet
	printStats("After sql.Open", db)

	// 2. After Ping — one connection established
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	printStats("After db.Ping", db)

	// Setup
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		name    TEXT NOT NULL UNIQUE,
		balance REAL NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	db.Exec("INSERT OR IGNORE INTO accounts (name, balance) VALUES (?, ?)", "Alice", 1000)
	db.Exec("INSERT OR IGNORE INTO accounts (name, balance) VALUES (?, ?)", "Bob", 500)

	// 3. During open query — connection is in use
	rows, err := db.Query("SELECT name, balance FROM accounts")
	if err != nil {
		log.Fatal(err)
	}
	printStats("During open rows", db)

	// Process and close rows
	for rows.Next() {
		var name string
		var balance float64
		if err := rows.Scan(&name, &balance); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %s: $%.2f\n", name, balance)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	rows.Close()

	// 4. After closing rows — connection returned to pool
	printStats("After rows.Close", db)

	// Bonus: demonstrate pool exhaustion
	fmt.Println("--- Pool Exhaustion Demo ---")
	leakedRows, err := db.Query("SELECT 1")
	if err != nil {
		log.Fatal(err)
	}
	// Intentionally NOT closing leakedRows

	printStats("Holding rows open", db)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("Attempting second query with 2s timeout...")
	_, err = db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		fmt.Printf("Second query failed: %v\n", err)
	}

	printStats("After timeout", db)

	// Clean up the leaked rows
	leakedRows.Close()
	printStats("After closing leaked rows", db)
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| Connection pool | `database/sql` manages a pool of reusable connections automatically |
| `SetMaxOpenConns` | Limits total connections — set this, don't leave it unlimited |
| `SetMaxIdleConns` | How many idle connections to keep ready |
| `SetConnMaxLifetime` | Close connections after this total age |
| `SetConnMaxIdleTime` | Close connections after this idle duration |
| `db.Stats()` | Monitor pool health — watch `WaitCount` and `WaitDuration` |
| Pool exhaustion | Caused by leaked rows or unclosed transactions |
| SQLite | Set `MaxOpenConns(1)` — it serializes writes |
| `sql.Open` | Creates the pool, not a connection — connections are lazy |
