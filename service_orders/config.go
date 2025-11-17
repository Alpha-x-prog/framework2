package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort = ":8082"     // пусть заказы слушают 8082
	dbPath      = "orders.db" // файл SQLite
)

var (
	jwtSecretString string
	tokenTTL        = 24 * time.Hour
)

func initConfig() {
	_ = godotenv.Load()

	jwtSecretString = getenv("JWT_SECRET", "dev-secret-change-me")

	log.Println("Config initialized for service_orders, JWT_SECRET length:", len(jwtSecretString))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
