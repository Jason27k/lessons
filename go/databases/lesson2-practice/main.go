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
		log.Fatalf("DB Opening Error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB Ping Error: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		author TEXT NOT NULL,
		year INTEGER NOT NULL
	)`)

	if err != nil {
		log.Fatalf("Create Table Error: %v", err)
	}

	result, err := db.Exec(`INSERT OR IGNORE INTO books (title, author, year) 
		VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)`,
		"Heated Rivalry", "Rachel Reid", 2024,
		"Theo of Golden: A Novel", "Allen Levi", 2025,
		"Dungeon Crawler Carl", "Matt Dinniman", 2025,
	)

	if err != nil {
		log.Fatalf("Insert Into Table Error: %v", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(id)

	affected, err := result.RowsAffected()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Rows affected %d\n", affected)

	var title string
	var author string
	var year int
	err = db.QueryRow("SELECT title, author, year FROM books WHERE id = ?", 1).Scan(&title, &author, &year)

	if err != nil {
		log.Fatalf("Query Table Error: %v", err)
	}

	fmt.Printf("Results: %s, %s, %d\n", title, author, year)

	rows, err := db.Query("SELECT title, author, year FROM books")

	if err != nil {
		log.Fatalf("Query Table Error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&title, &author, &year)
		if err != nil {
			log.Fatalf("Iterate Over Select Error: %v", err)
		}

		fmt.Printf("Results: %s, %s, %d\n", title, author, year)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Rows iteration error: %v", err)
	}
}
