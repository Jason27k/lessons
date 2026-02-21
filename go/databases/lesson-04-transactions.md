# Lesson 4: Transactions

## The Problem

So far, every database operation you've done has been independent. Insert a row, update a row, delete a row — each one succeeds or fails on its own. But what happens when you need **multiple operations to succeed or fail together**?

Example: transferring money between accounts. You need to subtract from one account AND add to another. If the subtraction succeeds but the addition fails, money disappears. Transactions guarantee that either **both** happen or **neither** happens.

---

## 1. What Is a Transaction?

A transaction groups multiple SQL statements into a single atomic unit. The database guarantees **ACID** properties:

| Property | Meaning |
|----------|---------|
| **Atomicity** | All statements succeed, or all are rolled back — no partial results |
| **Consistency** | The database moves from one valid state to another |
| **Isolation** | Concurrent transactions don't see each other's intermediate states |
| **Durability** | Once committed, the data survives crashes |

Without a transaction, each `db.Exec` is its own implicit transaction — it auto-commits immediately.

---

## 2. Basic Transaction Flow

```go
// 1. Begin
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

// 2. Do work (use tx, NOT db)
_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", 100, 2)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

// 3. Commit
if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

### The three steps

1. **`db.Begin()`** — starts a transaction, returns a `*sql.Tx`
2. **Use `tx` for all operations** — `tx.Exec`, `tx.QueryRow`, `tx.Query` (NOT `db.Exec` etc.)
3. **`tx.Commit()`** or **`tx.Rollback()`** — finalize or undo everything

**Critical:** once you call `Begin`, you must eventually call either `Commit` or `Rollback`. If you don't, the transaction stays open and the connection is stuck.

---

## 3. The Defer-Rollback Pattern

Manually calling `Rollback` at every error point is repetitive and error-prone. The standard Go pattern is to defer the rollback:

```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback() // no-op if tx.Commit() was already called

_, err = tx.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "Book A", "Author A", 2020)
if err != nil {
    log.Fatal(err) // deferred Rollback runs
}

_, err = tx.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "Book B", "Author B", 2021)
if err != nil {
    log.Fatal(err) // deferred Rollback runs
}

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
// Commit succeeded — deferred Rollback is a no-op
```

### Why this works

- `tx.Rollback()` after `tx.Commit()` returns `sql.ErrTxDone` — effectively a no-op
- If the function returns early (error, panic, etc.), the deferred `Rollback` cleans up
- You never forget to rollback — the defer handles every exit path

**This is the pattern you should always use.** Put `defer tx.Rollback()` immediately after `db.Begin()`.

---

## 4. Using `tx` vs `db`

Once you start a transaction, **all operations must go through `tx`**, not `db`:

```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// RIGHT — uses tx
_, err = tx.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "Book A", "Author A", 2020)

// WRONG — uses db, runs OUTSIDE the transaction
_, err = db.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "Book B", "Author B", 2021)
```

If you use `db.Exec` inside a transaction block, that statement runs on a different connection, outside the transaction. It won't be rolled back if the transaction fails.

`*sql.Tx` has the same methods as `*sql.DB`: `Exec`, `Query`, `QueryRow`, `Prepare`. The API is identical — you just call them on `tx` instead of `db`.

---

## 5. Transactions in Functions

When a function needs a transaction, there are two approaches:

### Approach 1: Function manages its own transaction

```go
func transferMoney(db *sql.DB, from, to int, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from)
    if err != nil {
        return err
    }

    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### Approach 2: Function receives a transaction

```go
func addBookToTx(tx *sql.Tx, title, author string, year int) error {
    _, err := tx.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
        title, author, year)
    return err
}
```

The caller controls the transaction boundary, and this function is one piece of it. This is useful when multiple operations across different functions need to be in the same transaction.

---

## 6. When Transactions Matter

### You need a transaction when:

**Multiple writes that must be atomic:**
```go
// Transfer: both must succeed or both must fail
tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = 1")
tx.Exec("UPDATE accounts SET balance = balance + 100 WHERE id = 2")
```

**Read-then-write consistency:**
```go
// Check stock, then reserve — no one else should change stock in between
var stock int
tx.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&stock)
if stock > 0 {
    tx.Exec("UPDATE products SET stock = stock - 1 WHERE id = ?", productID)
    tx.Exec("INSERT INTO orders (product_id, user_id) VALUES (?, ?)", productID, userID)
}
```

**Batch inserts that must all succeed or all fail:**
```go
for _, book := range books {
    _, err := tx.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
        book.Title, book.Author, book.Year)
    if err != nil {
        return err // deferred Rollback undoes all previous inserts
    }
}
```

### You don't need a transaction for:

- A single INSERT, UPDATE, or DELETE (it's already atomic on its own)
- Read-only queries where stale data is acceptable
- Independent operations that don't depend on each other

---

## 7. Transaction Isolation Levels

Isolation determines what concurrent transactions can see. SQLite has limited isolation options (it serializes writes), but for network databases like Postgres you'll encounter:

| Level | Dirty Reads | Non-Repeatable Reads | Phantom Reads |
|-------|-------------|---------------------|---------------|
| Read Uncommitted | Yes | Yes | Yes |
| Read Committed | No | Yes | Yes |
| Repeatable Read | No | No | Yes |
| Serializable | No | No | No |

In Go, you can set the isolation level with `db.BeginTx`:

```go
tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
    Isolation: sql.LevelSerializable,
})
```

For SQLite, you won't need this. Just know it exists for when you move to Postgres or MySQL.

---

## Key Rules

1. **`defer tx.Rollback()` immediately after `db.Begin()`** — the standard pattern
2. **Use `tx`, not `db`** for all operations inside a transaction
3. **Always end with `Commit` or `Rollback`** — never leave a transaction open
4. **Use transactions for multiple related writes** — single operations are already atomic
5. **Return `tx.Commit()`** as the last line — its error matters

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Using `db.Exec` inside a transaction | Runs outside the transaction on a different connection |
| Forgetting to `Commit` or `Rollback` | Connection stays locked; pool eventually exhausts |
| Not deferring `Rollback` | Early returns or panics leave the transaction open |
| Wrapping a single INSERT in a transaction | Unnecessary — single statements are already atomic |
| Ignoring the error from `tx.Commit()` | The commit can fail (e.g., constraint violation at commit time) |

---

## Your Turn

### Exercise: Transactional Book Operations

Write a program that:

1. Creates an `accounts` table with columns: `id` (integer primary key autoincrement), `name` (text, not null, unique), `balance` (real, not null)
2. Seeds two accounts: "Alice" with balance 1000 and "Bob" with balance 500 (use `INSERT OR IGNORE`)
3. Writes a function `transfer(db *sql.DB, from, to string, amount float64) error` that:
   - Begins a transaction
   - Uses `defer tx.Rollback()`
   - Checks the sender's balance (if insufficient, return an error **without** committing)
   - Subtracts from the sender
   - Adds to the receiver
   - Commits
4. In `main`, perform a valid transfer (Alice sends Bob 200), print both balances
5. Attempt an invalid transfer (Bob sends Alice 10000), handle the error, print both balances to show nothing changed
6. **Bonus:** Write a function `batchInsertBooks(db *sql.DB, books []Book) error` that inserts all books in a single transaction — if any insert fails, none of them should persist

---

<details>
<summary>Full Answer</summary>

```go
package main

import (
    "database/sql"
    "errors"
    "fmt"
    "log"

    _ "modernc.org/sqlite"
)

type Book struct {
    ID     int
    Title  string
    Author string
    Year   int
}

var ErrInsufficientFunds = errors.New("insufficient funds")

func transfer(db *sql.DB, from, to string, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Check sender's balance
    var balance float64
    err = tx.QueryRow("SELECT balance FROM accounts WHERE name = ?", from).Scan(&balance)
    if err != nil {
        return err
    }

    if balance < amount {
        return ErrInsufficientFunds
    }

    // Subtract from sender
    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE name = ?", amount, from)
    if err != nil {
        return err
    }

    // Add to receiver
    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE name = ?", amount, to)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func batchInsertBooks(db *sql.DB, books []Book) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare("INSERT INTO books (title, author, year) VALUES (?, ?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, b := range books {
        _, err := stmt.Exec(b.Title, b.Author, b.Year)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func printBalances(db *sql.DB) {
    rows, err := db.Query("SELECT name, balance FROM accounts")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

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
}

func main() {
    db, err := sql.Open("sqlite", "learn.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    // Create tables
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id      INTEGER PRIMARY KEY AUTOINCREMENT,
            name    TEXT NOT NULL UNIQUE,
            balance REAL NOT NULL
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS books (
            id     INTEGER PRIMARY KEY AUTOINCREMENT,
            title  TEXT NOT NULL UNIQUE,
            author TEXT NOT NULL,
            year   INTEGER NOT NULL
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Seed accounts
    db.Exec("INSERT OR IGNORE INTO accounts (name, balance) VALUES (?, ?)", "Alice", 1000)
    db.Exec("INSERT OR IGNORE INTO accounts (name, balance) VALUES (?, ?)", "Bob", 500)

    fmt.Println("Initial balances:")
    printBalances(db)

    // Valid transfer
    err = transfer(db, "Alice", "Bob", 200)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nAfter Alice sends Bob $200:")
    printBalances(db)

    // Invalid transfer
    err = transfer(db, "Bob", "Alice", 10000)
    if errors.Is(err, ErrInsufficientFunds) {
        fmt.Printf("\nTransfer failed: %v\n", err)
    } else if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Balances unchanged:")
    printBalances(db)

    // Bonus: batch insert
    books := []Book{
        {Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015},
        {Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Year: 2017},
        {Title: "Learning Go", Author: "Jon Bodner", Year: 2021},
    }

    err = batchInsertBooks(db, books)
    if err != nil {
        log.Printf("Batch insert failed: %v", err)
    } else {
        fmt.Println("\nBatch insert succeeded")
    }
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `db.Begin()` | Starts a transaction, returns `*sql.Tx` |
| `tx.Commit()` | Makes all changes permanent |
| `tx.Rollback()` | Undoes all changes since `Begin` |
| Defer pattern | `defer tx.Rollback()` right after `Begin` — safe cleanup on any exit path |
| Use `tx` not `db` | Operations on `db` run outside the transaction |
| Atomicity | All operations in a transaction succeed or fail together |
| Single operations | Already atomic — don't need a transaction |
| `db.BeginTx` | For setting isolation level and passing context (advanced) |
