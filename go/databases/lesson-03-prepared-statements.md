# Lesson 3: Prepared Statements & SQL Injection

## The Problem

You've been using `?` placeholders in every query so far — but do you know *why*? It's not just a convenience. It's the line between a secure application and one that hands the database to an attacker. In this lesson, you'll see exactly how SQL injection works, why parameterized queries prevent it, and when to use `db.Prepare` for explicit prepared statements.

---

## 1. SQL Injection — The Attack

Imagine building a query with string formatting:

```go
// NEVER DO THIS
func getBookByTitle(db *sql.DB, userInput string) (Book, error) {
    query := fmt.Sprintf("SELECT id, title, author, year FROM books WHERE title = '%s'", userInput)
    var b Book
    err := db.QueryRow(query).Scan(&b.ID, &b.Title, &b.Author, &b.Year)
    return b, err
}
```

If a user passes `"Heated Rivalry"`, the query becomes:
```sql
SELECT id, title, author, year FROM books WHERE title = 'Heated Rivalry'
```

Fine. But what if they pass `"'; DROP TABLE books; --"`?

```sql
SELECT id, title, author, year FROM books WHERE title = ''; DROP TABLE books; --'
```

The `'` closes the string literal. `DROP TABLE books` runs. `--` comments out the rest. Your table is gone.

Other attacks:
- `' OR '1'='1` — returns all rows (bypasses authentication)
- `' UNION SELECT username, password, 1, 1 FROM users --` — reads other tables
- `'; UPDATE users SET role='admin' WHERE username='attacker' --` — privilege escalation

This isn't theoretical. SQL injection has been the #1 web vulnerability for decades.

---

## 2. Why Parameterized Queries Are Safe

When you use `?` placeholders:

```go
db.QueryRow("SELECT id, title, author, year FROM books WHERE title = ?", userInput)
```

The driver sends the query and the parameter **separately** to the database. The database:
1. Parses the SQL structure: `SELECT ... WHERE title = <param>`
2. Treats the parameter as a **value**, never as SQL syntax

Even if `userInput` is `"'; DROP TABLE books; --"`, the database searches for a book literally titled `'; DROP TABLE books; --`. The `'` and `;` have no special meaning — they're just characters in a string value.

**The key insight:** parameterized queries separate **code** (SQL structure) from **data** (values). String formatting mixes them together, which is where the vulnerability lives.

---

## 3. The Rule

**Never use `fmt.Sprintf`, string concatenation, or any string formatting to put values into SQL.** Always use `?` placeholders.

```go
// WRONG — SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM books WHERE year > %d", year)
query := "SELECT * FROM books WHERE title = '" + title + "'"

// RIGHT — parameterized
db.Query("SELECT * FROM books WHERE year > ?", year)
db.QueryRow("SELECT * FROM books WHERE title = ?", title)
db.Exec("INSERT INTO books (title) VALUES (?)", title)
```

Even for integer values — don't format them into the string. Use placeholders consistently.

### What about table/column names?

Placeholders only work for **values**. You can't parameterize table names, column names, or SQL keywords:

```go
// This does NOT work
db.Query("SELECT * FROM ? WHERE ? = ?", tableName, columnName, value)

// You must use string formatting for structural parts
// BUT: validate against a whitelist, never use raw user input
allowedColumns := map[string]bool{"title": true, "author": true, "year": true}
if !allowedColumns[sortColumn] {
    return fmt.Errorf("invalid column: %s", sortColumn)
}
query := fmt.Sprintf("SELECT * FROM books ORDER BY %s", sortColumn)
```

---

## 4. `db.Prepare` — Explicit Prepared Statements

So far you've been using **inline parameters** — passing `?` values directly to `Exec`, `QueryRow`, or `Query`. Behind the scenes, the driver may or may not prepare the statement. With `db.Prepare`, you explicitly create a prepared statement that you can reuse:

```go
stmt, err := db.Prepare("INSERT INTO books (title, author, year) VALUES (?, ?, ?)")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

// Reuse the same prepared statement multiple times
stmt.Exec("Book One", "Author A", 2020)
stmt.Exec("Book Two", "Author B", 2021)
stmt.Exec("Book Three", "Author C", 2022)
```

### How it works

1. `db.Prepare` sends the SQL to the database, which parses and plans it once
2. Each `stmt.Exec` or `stmt.Query` only sends the parameter values
3. The database reuses the parsed plan — no re-parsing

### When to use `db.Prepare`

Use it when you're executing the **same statement many times** in a loop:

```go
stmt, err := db.Prepare("INSERT INTO books (title, author, year) VALUES (?, ?, ?)")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

for _, b := range booksToInsert {
    _, err := stmt.Exec(b.Title, b.Author, b.Year)
    if err != nil {
        log.Fatal(err)
    }
}
```

Without `Prepare`, each `db.Exec` call would parse the SQL again. With `Prepare`, the SQL is parsed once and reused.

### When NOT to use `db.Prepare`

For one-off queries, inline parameters are simpler and just as safe:

```go
// This is fine — no need for Prepare
db.QueryRow("SELECT title FROM books WHERE id = ?", id)
```

Don't over-use `Prepare`. It adds complexity (you have to `Close` the statement) for no benefit on single-use queries.

---

## 5. The Connection Pinning Tradeoff

Here's a subtlety. When you call `db.Prepare`, the prepared statement is created on a **specific database connection**. But `*sql.DB` is a connection pool — subsequent calls might use a different connection.

Go handles this by **re-preparing** the statement on new connections as needed. This is mostly transparent, but it means:

- A `*sql.Stmt` can hold onto a connection (or multiple connections)
- If you create many prepared statements and don't close them, you can exhaust the pool
- Always `defer stmt.Close()` when you're done

In practice, this rarely matters for SQLite (single-file database). It matters more with network databases like Postgres where connections are expensive.

---

## 6. Placeholder Syntax by Database

Different databases use different placeholder syntax:

| Database | Placeholder | Example |
|----------|------------|---------|
| SQLite | `?` | `WHERE id = ?` |
| MySQL | `?` | `WHERE id = ?` |
| PostgreSQL | `$1`, `$2`, ... | `WHERE id = $1` |
| SQL Server | `@p1`, `@p2`, ... | `WHERE id = @p1` |

The `?` style is positional — first `?` gets the first argument, second `?` gets the second, etc. PostgreSQL's `$N` style lets you reference the same parameter multiple times:

```sql
-- PostgreSQL: use $1 twice
SELECT * FROM books WHERE title = $1 OR author = $1
```

With `?` style, you'd have to pass the value twice:

```go
db.Query("SELECT * FROM books WHERE title = ? OR author = ?", search, search)
```

Since we're using SQLite, stick with `?`.

---

## Key Rules

1. **Never format values into SQL strings** — always use `?` placeholders
2. **Placeholders are for values only** — not table names, column names, or SQL keywords
3. **Use `db.Prepare`** when executing the same statement many times in a loop
4. **Don't use `db.Prepare`** for one-off queries — inline parameters are simpler
5. **Always `defer stmt.Close()`** — prepared statements hold connection resources
6. **Validate structural inputs** against a whitelist if you must dynamically build SQL

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| `fmt.Sprintf("... WHERE id = %d", id)` | SQL injection — even integers should use placeholders for consistency |
| Using `?` for table/column names | Doesn't work — the database expects values in placeholder positions |
| Creating a prepared statement for a one-off query | Unnecessary complexity — use inline parameters |
| Forgetting `stmt.Close()` | Leaks connections from the pool |
| Assuming prepared statements are always faster | The overhead of prepare + close can exceed the cost of re-parsing for single-use queries |

---

## Your Turn

### Exercise: Batch Insert with Prepared Statements

Write a program that:

1. Defines a slice of 5+ books (use a `[]Book` or `[]struct{...}`)
2. Uses `db.Prepare` to create a prepared INSERT statement
3. Loops over the books and inserts each one using the prepared statement
4. Tracks and prints how many were successfully inserted (using `LastInsertId` or a counter)
5. After inserting, queries all books and prints them
6. **Bonus:** Write a function `searchBooks(db *sql.DB, term string) ([]Book, error)` that searches for books where the title OR author contains the search term (use `LIKE` with `?` — hint: the `%` wildcards go in the argument, not the SQL: `db.Query("... LIKE ?", "%"+term+"%")`)

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

type Book struct {
    ID     int
    Title  string
    Author string
    Year   int
}

func searchBooks(db *sql.DB, term string) ([]Book, error) {
    rows, err := db.Query(
        "SELECT id, title, author, year FROM books WHERE title LIKE ? OR author LIKE ?",
        "%"+term+"%", "%"+term+"%",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var b Book
        if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
            return nil, err
        }
        books = append(books, b)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return books, nil
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

    books := []Book{
        {Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015},
        {Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Year: 2017},
        {Title: "Let's Go", Author: "Alex Edwards", Year: 2022},
        {Title: "Let's Go Further", Author: "Alex Edwards", Year: 2022},
        {Title: "Learning Go", Author: "Jon Bodner", Year: 2021},
    }

    // Prepare the statement once
    stmt, err := db.Prepare("INSERT OR IGNORE INTO books (title, author, year) VALUES (?, ?, ?)")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    inserted := 0
    for _, b := range books {
        result, err := stmt.Exec(b.Title, b.Author, b.Year)
        if err != nil {
            log.Printf("Failed to insert %q: %v", b.Title, err)
            continue
        }
        affected, _ := result.RowsAffected()
        if affected > 0 {
            inserted++
        }
    }
    fmt.Printf("Inserted %d/%d books\n", inserted, len(books))

    // Query all
    rows, err := db.Query("SELECT id, title, author, year FROM books")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("\nAll books:")
    for rows.Next() {
        var b Book
        if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("  %d. %s by %s (%d)\n", b.ID, b.Title, b.Author, b.Year)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }

    // Search
    fmt.Println("\nSearch for 'Edwards':")
    results, err := searchBooks(db, "Edwards")
    if err != nil {
        log.Fatal(err)
    }
    for _, b := range results {
        fmt.Printf("  %d. %s by %s (%d)\n", b.ID, b.Title, b.Author, b.Year)
    }
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| SQL injection | Mixing user input into SQL strings lets attackers run arbitrary SQL |
| Parameterized queries | `?` placeholders send values separately — the database treats them as data, never syntax |
| `db.Prepare` | Parses SQL once, reuse with different values — good for loops |
| `stmt.Close()` | Always close prepared statements to release connection resources |
| Inline parameters | `db.Query("... WHERE id = ?", id)` — simpler for one-off queries, equally safe |
| Placeholders are for values only | Can't parameterize table names, column names, or keywords — validate those against a whitelist |
| `LIKE` with parameters | Put `%` in the argument: `db.Query("... LIKE ?", "%"+term+"%")` |
