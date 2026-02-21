# Lesson 6: Migrations

## The Problem

So far you've been doing this:

```go
db.Exec(`CREATE TABLE IF NOT EXISTS accounts (...)`)
```

This works for learning, but in real projects it falls apart:

- What if you need to **add a column** to an existing table?
- What if your teammate needs the **same schema changes** you made?
- What if a deploy goes wrong and you need to **undo** a schema change?
- How do you know **which changes have already been applied** to production?

`CREATE TABLE IF NOT EXISTS` only handles one case: "does this table exist?" It can't evolve a schema over time.

**Migrations** solve this. They're versioned, ordered SQL scripts that track what has been applied and can be rolled back.

---

## 1. What Is a Migration?

A migration is a pair of SQL operations:

- **Up** — applies a change (create table, add column, etc.)
- **Down** — reverses that change (drop table, remove column, etc.)

Each migration has a version number. The migration tool tracks which versions have been applied. When you run migrations, it only applies the ones that haven't been run yet.

```
Migration 1: Create users table          (up: CREATE TABLE / down: DROP TABLE)
Migration 2: Add email column to users   (up: ALTER TABLE ADD / down: ALTER TABLE DROP)
Migration 3: Create posts table          (up: CREATE TABLE / down: DROP TABLE)
```

If your database is at version 2, running migrations applies only migration 3. Rolling back undoes migration 3 (or 2, or as far back as you want).

---

## 2. golang-migrate

The standard tool in the Go ecosystem is [`golang-migrate`](https://github.com/golang-migrate/migrate). It has two parts:

1. **CLI tool** — for creating and running migrations from the terminal
2. **Go library** — for running migrations from your Go code

### Installing the CLI

```bash
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

The `-tags 'sqlite'` flag includes SQLite support. For Postgres you'd use `'postgres'` instead (or `'sqlite postgres'` for both).

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

---

## 3. Creating Migrations

Use the CLI to create migration files:

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

This creates two files:

```
migrations/
├── 000001_create_users_table.up.sql
└── 000001_create_users_table.down.sql
```

- `-ext sql` — file extension
- `-dir migrations` — directory to store migration files
- `-seq` — use sequential numbering (1, 2, 3...) instead of timestamps

### Write the SQL

**`000001_create_users_table.up.sql`:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**`000001_create_users_table.down.sql`:**
```sql
DROP TABLE IF EXISTS users;
```

### Add another migration

```bash
migrate create -ext sql -dir migrations -seq add_bio_to_users
```

**`000002_add_bio_to_users.up.sql`:**
```sql
ALTER TABLE users ADD COLUMN bio TEXT NOT NULL DEFAULT '';
```

**`000002_add_bio_to_users.down.sql`:**
```sql
ALTER TABLE users DROP COLUMN bio;
```

---

## 4. Running Migrations from the CLI

```bash
# Apply all pending migrations
migrate -database "sqlite://app.db" -path migrations up

# Apply the next N migrations
migrate -database "sqlite://app.db" -path migrations up 1

# Roll back the last migration
migrate -database "sqlite://app.db" -path migrations down 1

# Roll back all migrations
migrate -database "sqlite://app.db" -path migrations down

# Check current version
migrate -database "sqlite://app.db" -path migrations version

# Force a version (fixes dirty state — see below)
migrate -database "sqlite://app.db" -path migrations force 1
```

The tool creates a `schema_migrations` table in your database to track which version has been applied.

---

## 5. Running Migrations from Go Code

For production apps, you typically want migrations to run automatically at startup. Use the `migrate` library:

```go
package main

import (
    "database/sql"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create a driver instance for the migrate library
    driver, err := sqlite.WithInstance(db, &sqlite.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Point to migration files
    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "sqlite", driver,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Apply all pending migrations
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }

    log.Println("Migrations applied successfully")

    // Now use db normally...
}
```

**Note:** `m.Up()` returns `migrate.ErrNoChange` if all migrations are already applied. This isn't a real error — you need to check for it.

---

## 6. Embedding Migrations with `embed`

Having migration files on disk works, but it means you need to deploy the SQL files alongside your binary. Go's `embed` package lets you bake them into the compiled binary:

```go
package main

import (
    "database/sql"
    "embed"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    _ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create source from embedded files
    source, err := iofs.New(migrationFS, "migrations")
    if err != nil {
        log.Fatal(err)
    }

    driver, err := sqlite.WithInstance(db, &sqlite.Config{})
    if err != nil {
        log.Fatal(err)
    }

    m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
    if err != nil {
        log.Fatal(err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }

    log.Println("Migrations applied successfully")
}
```

The `//go:embed migrations/*.sql` directive tells the compiler to include all `.sql` files from the `migrations` directory inside the binary. At runtime, `migrationFS` acts like a filesystem.

---

## 7. Dirty State

If a migration partially fails (e.g., syntax error in the SQL), the database is marked as **dirty**. This means the migration tool doesn't know if the version was fully applied or not.

```bash
$ migrate -database "sqlite://app.db" -path migrations version
3 (dirty)
```

To fix it:
1. Manually check what state the database is in (did the migration partially apply?)
2. Fix the SQL file if needed
3. Force the version to the last known good state:

```bash
migrate -database "sqlite://app.db" -path migrations force 2
```

This sets the version to 2 (clean) without running any SQL, so the next `up` will re-attempt migration 3.

---

## 8. Migration Best Practices

### Each migration should be small and focused
```
GOOD: 000001_create_users_table
GOOD: 000002_add_email_index

BAD:  000001_create_all_tables_and_seed_data
```

### Migrations are immutable
Once a migration has been applied to any shared environment (staging, production, a teammate's machine), **never edit it**. Create a new migration instead. If migration 3 added the wrong column, don't edit migration 3 — create migration 4 to fix it.

### Down migrations should be the exact reverse of up
If up creates a table, down drops it. If up adds a column, down removes it. This makes rollbacks predictable.

### Test migrations on a fresh database
Run all migrations from scratch (up) and all the way back (down) to verify they work in both directions:

```bash
migrate -database "sqlite://test.db" -path migrations up
migrate -database "sqlite://test.db" -path migrations down
```

---

## Key Rules

1. **Never edit applied migrations** — create new ones to fix mistakes
2. **Always write both up and down** — you'll need rollbacks eventually
3. **One logical change per migration** — keep them small and focused
4. **Check for `migrate.ErrNoChange`** — it's not an error, it means you're up to date
5. **Use `embed` for production** — don't depend on files being on disk at deploy time

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Editing a migration after it's been applied | Other environments have the old version — causes drift |
| Writing up without down | You can't roll back when things go wrong |
| Putting all schema in one migration | Hard to reason about, can't partially roll back |
| Ignoring dirty state | Subsequent migrations will refuse to run |
| Not checking `migrate.ErrNoChange` | Your app crashes on startup when already up to date |

---

## Your Turn

### Exercise: Migration-Based Schema Setup

1. Install `golang-migrate` CLI (see section 2)
2. Create a `migrations` directory in your project
3. Use the CLI to create three migrations:
   - `create_users_table` — columns: `id` (integer PK autoincrement), `name` (text not null), `email` (text not null unique)
   - `create_posts_table` — columns: `id` (integer PK autoincrement), `user_id` (integer not null, foreign key to users), `title` (text not null), `body` (text not null)
   - `add_created_at_to_users` — adds a `created_at` column (datetime, not null, default `CURRENT_TIMESTAMP`)
4. Write the up and down SQL for each
5. Apply all migrations using the CLI, verify with `version`
6. Roll back one migration, verify with `version`
7. Apply again, then write a Go program that:
   - Opens the database
   - Runs migrations using the Go library (with embedded files)
   - Inserts a user and a post
   - Queries and prints them

---

<details>
<summary>Migration Files</summary>

**`000001_create_users_table.up.sql`:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);
```

**`000001_create_users_table.down.sql`:**
```sql
DROP TABLE IF EXISTS users;
```

**`000002_create_posts_table.up.sql`:**
```sql
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

**`000002_create_posts_table.down.sql`:**
```sql
DROP TABLE IF EXISTS posts;
```

**`000003_add_created_at_to_users.up.sql`:**
```sql
ALTER TABLE users ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
```

**`000003_add_created_at_to_users.down.sql`:**
```sql
ALTER TABLE users DROP COLUMN created_at;
```

</details>

<details>
<summary>Go Program</summary>

```go
package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migrations
	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	log.Println("Migrations applied")

	// Insert a user
	result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)",
		"Alice", "alice@example.com")
	if err != nil {
		log.Fatal(err)
	}
	userID, _ := result.LastInsertId()

	// Insert a post
	_, err = db.Exec("INSERT INTO posts (user_id, title, body) VALUES (?, ?, ?)",
		userID, "First Post", "Hello from Alice!")
	if err != nil {
		log.Fatal(err)
	}

	// Query and print
	var name, email, title, body string
	err = db.QueryRow(`
		SELECT u.name, u.email, p.title, p.body
		FROM users u
		JOIN posts p ON p.user_id = u.id
		WHERE u.id = ?
	`, userID).Scan(&name, &email, &title, &body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s (%s)\n", name, email)
	fmt.Printf("Post: %s — %s\n", title, body)
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| Migration | A versioned pair of up/down SQL scripts |
| `migrate create` | CLI command to generate migration files |
| `migrate up` | Apply pending migrations |
| `migrate down 1` | Roll back the last migration |
| `migrate version` | Check current schema version |
| `migrate force N` | Fix dirty state by setting version without running SQL |
| Go library | Run migrations programmatically at app startup |
| `embed.FS` | Bake migration files into the binary for production |
| Dirty state | Partial failure — must be resolved with `force` |
| Immutability | Never edit applied migrations — create new ones instead |
