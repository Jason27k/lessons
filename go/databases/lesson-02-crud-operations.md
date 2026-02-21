# Lesson 2: CRUD Operations

## The Problem

You know how to open a database and create a table. Now you need to actually work with data — insert rows, read them back, update them, delete them. Go's `database/sql` has three main methods for this: `Exec`, `QueryRow`, and `Query`. Each one has specific behavior you need to understand, especially around scanning results into Go variables and structs.

---

## 1. The Three Methods

| Method | Use for | Returns |
|--------|---------|---------|
| `db.Exec` | INSERT, UPDATE, DELETE (no rows back) | `sql.Result` |
| `db.QueryRow` | SELECT that returns **one** row | `*sql.Row` |
| `db.Query` | SELECT that returns **multiple** rows | `*sql.Rows` |

Pick the right one based on whether you need rows back, and how many.

---

## 2. INSERT with `db.Exec`

```go
result, err := db.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "The Go Programming Language", "Donovan & Kernighan", 2015)
if err != nil {
    log.Fatal(err)
}
```

### `sql.Result` — what you get back

`Exec` returns a `sql.Result` interface with two methods:

```go
id, err := result.LastInsertId()   // the auto-generated ID of the inserted row
affected, err := result.RowsAffected() // number of rows affected
```

```go
result, err := db.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
    "Concurrency in Go", "Katherine Cox-Buday", 2017)
if err != nil {
    log.Fatal(err)
}

id, err := result.LastInsertId()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Inserted book with ID: %d\n", id)

affected, err := result.RowsAffected()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rows affected: %d\n", affected)
```

**Note:** `LastInsertId` and `RowsAffected` can each return errors because not all drivers support them. SQLite supports both.

---

## 3. UPDATE and DELETE with `db.Exec`

Same method, different SQL:

```go
// UPDATE
result, err := db.Exec("UPDATE books SET year = ? WHERE id = ?", 2016, 1)
if err != nil {
    log.Fatal(err)
}
affected, _ := result.RowsAffected()
fmt.Printf("Updated %d row(s)\n", affected)

// DELETE
result, err = db.Exec("DELETE FROM books WHERE id = ?", 2)
if err != nil {
    log.Fatal(err)
}
affected, _ = result.RowsAffected()
fmt.Printf("Deleted %d row(s)\n", affected)
```

`RowsAffected` is especially useful here — it tells you if the row actually existed. If you try to delete ID 999 and it returns 0, nothing was deleted.

---

## 4. SELECT One Row with `db.QueryRow`

When you expect exactly one row (or zero):

```go
var title string
var author string
var year int

err := db.QueryRow("SELECT title, author, year FROM books WHERE id = ?", 1).Scan(&title, &author, &year)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%s by %s (%d)\n", title, author, year)
```

### The `Scan` contract

- You pass **pointers** — `Scan` writes into your variables
- The number and order of `Scan` arguments must match the columns in your SELECT
- `Scan` does type conversion where possible (e.g., INTEGER column into `int`, `int64`, or even `string`)

### Handling "not found"

When no row matches, `Scan` returns `sql.ErrNoRows`:

```go
err := db.QueryRow("SELECT title FROM books WHERE id = ?", 999).Scan(&title)
if err == sql.ErrNoRows {
    fmt.Println("book not found")
} else if err != nil {
    log.Fatal(err) // actual database error
}
```

This is important — "not found" is not a fatal error in most applications. You'll typically want to handle it differently (e.g., return a 404 in an HTTP handler).

---

## 5. SELECT Multiple Rows with `db.Query`

When you expect zero or more rows:

```go
rows, err := db.Query("SELECT id, title, author, year FROM books")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id, year int
    var title, author string
    if err := rows.Scan(&id, &title, &author, &year); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d. %s by %s (%d)\n", id, title, author, year)
}
if err := rows.Err(); err != nil {
    log.Fatal(err)
}
```

### The three rules of `db.Query`

1. **Check the error** from `db.Query` before using `rows`
2. **`defer rows.Close()`** — releases the connection back to the pool
3. **Check `rows.Err()`** after the loop — catches errors that stopped iteration early

You already learned these in Lesson 1. They're non-negotiable.

### Why `rows.Close()` matters

`db.Query` grabs a connection from the pool and holds it until `rows.Close()`. If you forget to close, that connection is stuck. Do this enough times and you exhaust the pool — your app hangs waiting for a free connection.

---

## 6. Scanning into Structs

Scanning into individual variables gets tedious. Define a struct and scan into its fields:

```go
type Book struct {
    ID     int
    Title  string
    Author string
    Year   int
}

// Single row
var b Book
err := db.QueryRow("SELECT id, title, author, year FROM books WHERE id = ?", 1).
    Scan(&b.ID, &b.Title, &b.Author, &b.Year)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%+v\n", b)

// Multiple rows
rows, err := db.Query("SELECT id, title, author, year FROM books")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

var books []Book
for rows.Next() {
    var b Book
    if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
        log.Fatal(err)
    }
    books = append(books, b)
}
if err := rows.Err(); err != nil {
    log.Fatal(err)
}

for _, b := range books {
    fmt.Printf("%d. %s by %s (%d)\n", b.ID, b.Title, b.Author, b.Year)
}
```

**Note:** `database/sql` has no auto-mapping from columns to struct fields. You must list the fields in `Scan` in the same order as the columns in your SELECT. This is verbose but explicit — you always know exactly what's being mapped where. (Libraries like `sqlx` add auto-mapping, which we'll cover in Lesson 9.)

---

## 7. Handling NULL Columns

What if a column can be NULL? You can't scan NULL into a `string` or `int` — Go has no null concept for value types.

Use the `sql.Null*` wrapper types:

```go
var year sql.NullInt64

err := db.QueryRow("SELECT year FROM books WHERE id = ?", 1).Scan(&year)
if err != nil {
    log.Fatal(err)
}

if year.Valid {
    fmt.Printf("Year: %d\n", year.Int64)
} else {
    fmt.Println("Year: unknown")
}
```

Available types: `sql.NullString`, `sql.NullInt64`, `sql.NullFloat64`, `sql.NullBool`, `sql.NullTime`.

Each has two fields:
- The value (e.g., `Int64`, `String`)
- `Valid bool` — true if the value is not NULL

In your `books` table, `year` is `NOT NULL`, so you won't hit this. But when you work with schemas that allow NULLs, you'll need these.

---

## Key Rules

1. **`Exec` for writes** — INSERT, UPDATE, DELETE. Returns `sql.Result` with `LastInsertId` and `RowsAffected`
2. **`QueryRow` for one row** — chain `.Scan()`. Check for `sql.ErrNoRows` when the row might not exist
3. **`Query` for many rows** — check error, defer close, check `rows.Err()`
4. **Scan order matches SELECT order** — no auto-mapping in `database/sql`
5. **Use structs** — scanning into individual variables doesn't scale
6. **`sql.Null*` for nullable columns** — Go value types can't represent NULL

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Using `Query` when you expect one row | `Query` requires `rows.Close()` and a loop — `QueryRow` is simpler for single rows |
| Treating `sql.ErrNoRows` as a fatal error | "Not found" is usually expected — handle it gracefully |
| Wrong number of `Scan` arguments | Runtime panic if count doesn't match the columns in your SELECT |
| Wrong `Scan` order | Values end up in the wrong variables — no compile-time check catches this |
| Ignoring `RowsAffected` on UPDATE/DELETE | You won't know if the row actually existed |

---

## Your Turn

### Exercise: Full CRUD

Build on your Lesson 1 code. Write a program that:

1. Uses a `Book` struct with fields `ID`, `Title`, `Author`, `Year`
2. Writes a function `insertBook(db *sql.DB, title, author string, year int) (int64, error)` that inserts a book and returns its ID (using `LastInsertId`)
3. Writes a function `getBook(db *sql.DB, id int) (Book, error)` that returns a single book by ID — handle the "not found" case by returning a specific error
4. Writes a function `allBooks(db *sql.DB) ([]Book, error)` that returns all books
5. Writes a function `updateBookYear(db *sql.DB, id int, year int) (bool, error)` that updates a book's year and returns whether the row existed (using `RowsAffected`)
6. Writes a function `deleteBook(db *sql.DB, id int) (bool, error)` that deletes a book and returns whether the row existed
7. In `main`, demonstrate the full cycle: insert a few books, query them, update one, delete one, query again to show the changes

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

var ErrBookNotFound = errors.New("book not found")

func insertBook(db *sql.DB, title, author string, year int) (int64, error) {
    result, err := db.Exec("INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
        title, author, year)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func getBook(db *sql.DB, id int) (Book, error) {
    var b Book
    err := db.QueryRow("SELECT id, title, author, year FROM books WHERE id = ?", id).
        Scan(&b.ID, &b.Title, &b.Author, &b.Year)
    if err == sql.ErrNoRows {
        return Book{}, ErrBookNotFound
    }
    if err != nil {
        return Book{}, err
    }
    return b, nil
}

func allBooks(db *sql.DB) ([]Book, error) {
    rows, err := db.Query("SELECT id, title, author, year FROM books")
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

func updateBookYear(db *sql.DB, id int, year int) (bool, error) {
    result, err := db.Exec("UPDATE books SET year = ? WHERE id = ?", year, id)
    if err != nil {
        return false, err
    }
    affected, err := result.RowsAffected()
    if err != nil {
        return false, err
    }
    return affected > 0, nil
}

func deleteBook(db *sql.DB, id int) (bool, error) {
    result, err := db.Exec("DELETE FROM books WHERE id = ?", id)
    if err != nil {
        return false, err
    }
    affected, err := result.RowsAffected()
    if err != nil {
        return false, err
    }
    return affected > 0, nil
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

    // Insert
    id1, err := insertBook(db, "The Go Programming Language", "Donovan & Kernighan", 2015)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Inserted book ID: %d\n", id1)

    id2, err := insertBook(db, "Concurrency in Go", "Katherine Cox-Buday", 2017)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Inserted book ID: %d\n", id2)

    id3, err := insertBook(db, "Let's Go", "Alex Edwards", 2022)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Inserted book ID: %d\n", id3)

    // Read one
    book, err := getBook(db, int(id1))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Got book: %+v\n", book)

    // Read one that doesn't exist
    _, err = getBook(db, 999)
    if errors.Is(err, ErrBookNotFound) {
        fmt.Println("Book 999: not found (expected)")
    }

    // Read all
    books, err := allBooks(db)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nAll books:")
    for _, b := range books {
        fmt.Printf("  %d. %s by %s (%d)\n", b.ID, b.Title, b.Author, b.Year)
    }

    // Update
    found, err := updateBookYear(db, int(id1), 2016)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\nUpdated book %d: %v\n", id1, found)

    // Delete
    found, err = deleteBook(db, int(id2))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Deleted book %d: %v\n", id2, found)

    // Final state
    books, err = allBooks(db)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nBooks after changes:")
    for _, b := range books {
        fmt.Printf("  %d. %s by %s (%d)\n", b.ID, b.Title, b.Author, b.Year)
    }
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `db.Exec` | For INSERT, UPDATE, DELETE — returns `sql.Result` |
| `sql.Result` | `.LastInsertId()` and `.RowsAffected()` — both can error |
| `db.QueryRow` | For single-row SELECT — chain `.Scan()` |
| `sql.ErrNoRows` | Returned by `QueryRow.Scan` when no row matches — not a fatal error |
| `db.Query` | For multi-row SELECT — check error, defer close, check `rows.Err()` |
| `Scan` | Writes column values into pointers — order must match SELECT columns |
| Structs | Define a type and scan into its fields — cleaner than loose variables |
| `sql.Null*` | Wrapper types for nullable columns (`NullString`, `NullInt64`, etc.) |
