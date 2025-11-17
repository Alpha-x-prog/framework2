package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() error {
	d, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	d.SetMaxOpenConns(1)

	if err := d.Ping(); err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		items_json TEXT NOT NULL,
		status TEXT NOT NULL,
		total_amount REAL NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	if _, err := d.Exec(schema); err != nil {
		return err
	}

	db = d
	log.Println("SQLite for orders initialized at", dbPath)
	return nil
}
