package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const defaultPort = ":8080"

var (
	usersServiceURL  string
	ordersServiceURL string
	jwtSecretString  string
)

func initConfig() {
	_ = godotenv.Load()

	usersServiceURL = getenv("USERS_SERVICE_URL", "http://localhost:8081")
	ordersServiceURL = getenv("ORDERS_SERVICE_URL", "http://localhost:8082")
	jwtSecretString = getenv("JWT_SECRET", "dev-secret-change-me")

	log.Printf("Gateway config: users=%s orders=%s", usersServiceURL, ordersServiceURL)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
