package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var ErrInsufficientBalance = errors.New("insufficient Balance")
var ErrRowsAffected = errors.New("an unexpected number of rows were affected")

type Account struct {
	id      int
	name    string
	balance float64
}

func transfer(db *sql.DB, from string, to string, amount float64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var balance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE name = ?",
		from).Scan(&balance)
	if err != nil {
		return err
	}

	if balance < amount {
		return ErrInsufficientBalance
	}

	result, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE name = ?", amount, from)
	if err != nil {
		return err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if num != 1 {
		return ErrRowsAffected
	}

	result, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE name = ?", amount, to)
	if err != nil {
		return err
	}

	num, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if num != 1 {
		return ErrRowsAffected
	}

	return tx.Commit()
}

func main() {
	db, err := sql.Open("sqlite", "Accounts.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS accounts(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		balance REAL NOT NULL
	)`)

	if err != nil {
		log.Fatal(err)
	}

	preparedInsert, err := db.Prepare("INSERT OR IGNORE INTO accounts (name, balance) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer preparedInsert.Close()

	accounts := []Account{
		{name: "Alice", balance: 1000},
		{name: "Bob", balance: 500},
	}

	for _, account := range accounts {
		result, err := preparedInsert.Exec(account.name, account.balance)
		if err != nil {
			log.Fatal(err)
		}
		num, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Rows affected: %d\n", num)
	}

	preparedQuery, err := db.Prepare("SELECT id, name, balance FROM accounts WHERE name = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer preparedQuery.Close()

	names := []string{
		"Alice",
		"Bob",
	}
	err = transfer(db, "Alice", "Bob", 200)
	if errors.Is(err, ErrInsufficientBalance) {
		fmt.Printf("Transfer failed: %v\n", err)
	} else if err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		var account Account
		err := preparedQuery.QueryRow(name).Scan(&account.id, &account.name, &account.balance)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Account: %d, %s, %f\n", account.id, account.name, account.balance)
	}

	err = transfer(db, "Bob", "Alice", 10000)
	if errors.Is(err, ErrInsufficientBalance) {
		fmt.Printf("Transfer failed: %v\n", err)
	} else if err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		var account Account
		err := preparedQuery.QueryRow(name).Scan(&account.id, &account.name, &account.balance)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Account: %d, %s, %f\n", account.id, account.name, account.balance)
	}
}
