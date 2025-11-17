package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort = ":8081"
	dbPath      = "users.db"
)

var (
	jwtSecretString string
	tokenTTL        = 24 * time.Hour
)

// загружаем .env и инициализируем глобальные конфиги
func initConfig() {
	_ = godotenv.Load()

	jwtSecretString = getenv("JWT_SECRET", "dev-secret-change-me")

	log.Println("Config initialized, JWT_SECRET length:", len(jwtSecretString))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
