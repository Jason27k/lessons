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
	rows, err := db.Query("SELECT id, title, author, year FROM books WHERE title LIKE ? OR AUTHOR LIKE ?",
		"%"+term+"%", "%"+term+"%")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		if err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func main() {
	books := []Book{
		{Title: "To Kill a Mockingbird", Author: "Harper Lee", Year: 1960},
		{Title: "1984", Author: "George Orwell", Year: 1949},
		{Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Year: 1925},
		{Title: "One Hundred Years of Solitude", Author: "Gabriel García Márquez", Year: 1967},
		{Title: "Brave New World", Author: "Aldous Huxley", Year: 1932},
		{Title: "The Great Alone", Author: "Kristin Hannah", Year: 2018},
		{Title: "New World Order", Author: "H.G. Wells", Year: 1940},
		{Title: "Go Set a Watchman", Author: "Harper Lee", Year: 2015},
		{Title: "Animal Farm", Author: "George Orwell", Year: 1945},
	}

	db, err := sql.Open("sqlite", "Books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		author TEXT NOT NULL,
		YEAR INTEGER NOT NULL
	)`)

	if err != nil {
		log.Fatal(err)
	}

	query, err := db.Prepare("INSERT OR IGNORE INTO books (title, author, year) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer query.Close()

	for _, book := range books {
		result, err := query.Exec(book.Title, book.Author, book.Year)
		if err != nil {
			log.Fatal(err)
		}

		numRows, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Number of rows affected: %v\n", numRows)
	}

	fmt.Println("\nQuery all books")
	rows, err := db.Query("SELECT id, title, author, year FROM books")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var book Book

		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%d, %s, %s, %d\n", book.ID, book.Title, book.Author, book.Year)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nBooks with term 'Lee' in them, searched for using searchBooks function")
	searchResults, err := searchBooks(db, "lee")

	if err != nil {
		log.Fatal(err)
	}

	for _, book := range searchResults {
		fmt.Printf("%d, %s, %s, %d\n", book.ID, book.Title, book.Author, book.Year)
	}
}
