package main

import "time"

const (
	defaultPort     = ":8081"                  // порт сервиса пользователей
	jwtSecretString = "super-secret-change-me" // в реале брать из ENV
	dbPath          = "users.db"               // путь к SQLite файлу
)

var tokenTTL = 24 * time.Hour
