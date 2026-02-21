# Databases in Go - Lesson Plan

## Prerequisites
- Go basics
- Goroutines and concurrency (completed)
- HTTP servers (completed)

## Topics

### 1. database/sql Fundamentals
- The `database/sql` package and driver model
- Opening a connection (`sql.Open` vs actual connection)
- Ping, Close, and connection lifecycle
- Why `database/sql` is an abstraction, not an ORM

### 2. CRUD Operations
- `db.Exec` for INSERT, UPDATE, DELETE
- `db.QueryRow` for single-row SELECT
- `db.Query` for multi-row SELECT
- Scanning rows into structs
- `LastInsertId` and `RowsAffected`

### 3. Prepared Statements & SQL Injection
- Why string formatting SQL is dangerous
- `db.Prepare` and placeholder syntax
- When to use prepared statements vs inline params
- Connection pinning tradeoff

### 4. Transactions
- `db.Begin`, `tx.Commit`, `tx.Rollback`
- The defer-rollback pattern
- When transactions matter (consistency, atomicity)
- Transaction isolation levels

### 5. Connection Pool Management
- How `database/sql` pools connections behind the scenes
- `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`
- Pool exhaustion and debugging
- `db.Stats()` for monitoring

### 6. Migrations
- Why schema changes need to be versioned
- `golang-migrate` tool and library
- Up/down migrations
- Embedding migrations with `embed`

### 7. Repository Pattern & Structuring DB Code
- Separating SQL from handlers
- The repository interface pattern
- Dependency injection with interfaces
- Testing with fake repositories

### 8. Building a CRUD API
- Combining HTTP + database into a REST API
- JSON request → DB insert → JSON response
- Error handling at the boundary (SQL errors → HTTP status codes)
- Graceful shutdown with DB cleanup

### 9. sqlx & Beyond
- `sqlx` — less boilerplate, struct scanning, named params
- When to use an ORM (GORM) vs query builder vs raw SQL
- `pgx` for Postgres-specific features
- Choosing the right tool for the job
