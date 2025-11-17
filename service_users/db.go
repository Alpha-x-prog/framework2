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

	// SQLite не любит много одновременных коннектов
	d.SetMaxOpenConns(1)

	if err := d.Ping(); err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		roles TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	if _, err := d.Exec(schema); err != nil {
		return err
	}

	db = d
	log.Println("SQLite initialized at", dbPath)
	return nil
}
