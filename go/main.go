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
