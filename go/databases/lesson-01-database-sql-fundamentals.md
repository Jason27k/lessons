# Lesson 1: database/sql Fundamentals

## The Problem

You can build HTTP servers, but every request starts from scratch — no memory of previous requests. You need a database. Go's standard library has `database/sql`, but it works differently than most languages: it's an **abstraction layer**, not a database driver. You need to understand the split between the interface and the driver, or you'll be confused by things like "why does `sql.Open` not actually connect?"

We'll use **SQLite** for this topic — no server to install, the database is just a file. Everything you learn applies to Postgres, MySQL, etc. — only the driver import and connection string change.

---

## 1. The Driver Model

Go's `database/sql` package defines **interfaces** — it doesn't know how to talk to any specific database. You pair it with a **driver** that implements those interfaces.

```
Your Code  →  database/sql (stdlib)  →  Driver (third-party)  →  Database
                 interfaces              implementation
```

For SQLite, the driver is `github.com/mattn/go-sqlite3` (CGo-based) or `modernc.org/sqlite` (pure Go). We'll use `modernc.org/sqlite` since it doesn't require a C compiler.

Install it:

```bash
go get modernc.org/sqlite
```

---

## 2. Opening a Connection

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "modernc.org/sqlite" // Register the driver — the _ means "import for side effects only"
)

func main() {
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    fmt.Println("connected (sort of)")
}
```

### The blank import: `_ "modernc.org/sqlite"`

This is Go's driver registration pattern. The driver's `init()` function calls `sql.Register("sqlite", ...)` to register itself. You never use the package directly — you only interact through `database/sql`.

Every database driver works this way:
- `_ "github.com/lib/pq"` → registers `"postgres"`
- `_ "github.com/go-sql-driver/mysql"` → registers `"mysql"`
- `_ "modernc.org/sqlite"` → registers `"sqlite"`

### `sql.Open` does NOT connect

This is the biggest gotcha. `sql.Open` only:
1. Validates the driver name
2. Saves the connection string
3. Returns a `*sql.DB` (a connection **pool**, not a single connection)

No network call happens. No file gets opened. To verify the database is actually reachable, use `Ping`:

```go
db, err := sql.Open("sqlite", "app.db")
if err != nil {
    log.Fatal(err) // only fails if driver name is wrong
}

// Actually test the connection
if err := db.Ping(); err != nil {
    log.Fatal(err) // fails if DB is unreachable
}
```

---

## 3. `*sql.DB` Is a Pool, Not a Connection

This is the key mental model. `*sql.DB`:
- Manages a **pool** of connections behind the scenes
- Is safe for concurrent use (you share one `*sql.DB` across goroutines)
- Opens and closes actual connections as needed
- Should be created **once** and passed around (not opened per-request)

```go
// WRONG — opening a new DB per request
func handler(w http.ResponseWriter, r *http.Request) {
    db, _ := sql.Open("sqlite", "app.db")
    defer db.Close()
    // ... query ...
}

// RIGHT — open once, share everywhere
func main() {
    db, _ := sql.Open("sqlite", "app.db")
    defer db.Close()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // use db here — it's safe for concurrent access
    })
    http.ListenAndServe(":8080", nil)
}
```

---

## 4. Creating a Table

Use `db.Exec` for statements that don't return rows (CREATE, INSERT, UPDATE, DELETE):

```go
_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id    INTEGER PRIMARY KEY AUTOINCREMENT,
        name  TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE
    )
`)
if err != nil {
    log.Fatal(err)
}
```

`Exec` returns a `sql.Result` (which we ignore with `_` here). We'll use it in the next lesson for `LastInsertId` and `RowsAffected`.

---

## 5. Quick Insert and Query (Preview)

Just to see the full cycle — we'll cover these properly in Lesson 2:

```go
// Insert
_, err = db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
if err != nil {
    log.Fatal(err)
}

// Query single row
var name, email string
err = db.QueryRow("SELECT name, email FROM users WHERE id = ?", 1).Scan(&name, &email)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s (%s)\n", name, email)
```

The `?` placeholders are how you pass parameters safely (no SQL injection). The driver substitutes them. **Never** use `fmt.Sprintf` to build SQL strings.

---

## 6. The Full Lifecycle

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "modernc.org/sqlite"
)

func main() {
    // 1. Open (doesn't connect)
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close() // 5. Close when main exits

    // 2. Verify connection
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("database ready")

    // 3. Create schema
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id    INTEGER PRIMARY KEY AUTOINCREMENT,
            name  TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Use it
    _, err = db.Exec("INSERT OR IGNORE INTO users (name, email) VALUES (?, ?)",
        "Alice", "alice@example.com")
    if err != nil {
        log.Fatal(err)
    }

    var name string
    err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("found:", name)
}
```

Run it twice — `IF NOT EXISTS` and `INSERT OR IGNORE` make it idempotent. The `app.db` file appears in your directory after the first run.

---

## Key Rules

1. **`sql.Open` doesn't connect** — use `Ping()` to verify
2. **`*sql.DB` is a pool** — create once, share everywhere, it's concurrency-safe
3. **Blank import registers the driver** — `_ "modernc.org/sqlite"` runs the driver's `init()`
4. **Always `defer db.Close()`** — though for long-running servers it only matters at shutdown
5. **Use `?` placeholders** — never string-format SQL values

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Opening `*sql.DB` per request | Creates a new pool each time; connections leak, performance tanks |
| Assuming `sql.Open` verifies the connection | It doesn't — you won't know the DB is down until the first query |
| Importing the driver without `_` | Compiler error ("imported and not used") — you don't call the driver directly |
| Using `fmt.Sprintf` to inject values into SQL | SQL injection vulnerability — always use parameterized queries |
| Closing `*sql.DB` after every query | It's a pool — closing it shuts down everything |

---

## Your Turn

### Exercise: Database Setup

Write a program in `main.go` that:

1. Opens a SQLite database called `learn.db`
2. Pings it to verify the connection
3. Creates a `books` table with columns: `id` (integer primary key autoincrement), `title` (text, not null), `author` (text, not null), `year` (integer)
4. Inserts 3 books (use `INSERT OR IGNORE` so it's safe to run multiple times — you'll need a UNIQUE constraint on `title` for this)
5. Queries one book by ID and prints it
6. Queries all books and prints them (hint: use `db.Query`, then loop with `rows.Next()` and `rows.Scan`)

Don't worry about getting the multi-row query perfect — we'll cover it properly in Lesson 2. Just give it a shot.

---

<details>
<summary>Full Answer</summary>

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "learn.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("database ready")

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS books (
            id     INTEGER PRIMARY KEY AUTOINCREMENT,
            title  TEXT NOT NULL UNIQUE,
            author TEXT NOT NULL,
            year   INTEGER
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    books := []struct {
        title  string
        author string
        year   int
    }{
        {"The Go Programming Language", "Donovan & Kernighan", 2015},
        {"Concurrency in Go", "Katherine Cox-Buday", 2017},
        {"Let's Go", "Alex Edwards", 2022},
    }

    for _, b := range books {
        _, err := db.Exec("INSERT OR IGNORE INTO books (title, author, year) VALUES (?, ?, ?)",
            b.title, b.author, b.year)
        if err != nil {
            log.Fatal(err)
        }
    }

    // Single row
    var title, author string
    var year int
    err = db.QueryRow("SELECT title, author, year FROM books WHERE id = ?", 1).Scan(&title, &author, &year)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Book 1: %s by %s (%d)\n", title, author, year)

    // All rows
    rows, err := db.Query("SELECT id, title, author, year FROM books")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("\nAll books:")
    for rows.Next() {
        var id, yr int
        var t, a string
        if err := rows.Scan(&id, &t, &a, &yr); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("  %d. %s by %s (%d)\n", id, t, a, yr)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `database/sql` | Abstraction layer — defines interfaces, doesn't talk to any DB directly |
| Driver import | `_ "modernc.org/sqlite"` — blank import runs `init()` to register the driver |
| `sql.Open` | Returns a `*sql.DB` pool — does NOT connect |
| `db.Ping` | Actually tests the connection |
| `*sql.DB` | Connection pool, concurrency-safe, create once and share |
| `db.Exec` | For statements that don't return rows (CREATE, INSERT, UPDATE, DELETE) |
| `db.QueryRow` | For SELECT returning one row — chain with `.Scan()` |
| `db.Query` | For SELECT returning multiple rows — loop with `rows.Next()` |
| `?` placeholders | Safe parameter substitution — never use string formatting |
