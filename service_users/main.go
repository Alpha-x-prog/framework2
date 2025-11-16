package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	router := gin.Default()

	api := router.Group("/v1")
	{
		users := api.Group("/users")
		{
			users.POST("/register", handleRegister)
			users.POST("/login", handleLogin)
			users.GET("/me", AuthRequired(), handleMe)
		}
	}

	log.Println("service_users listening on", defaultPort)
	if err := router.Run(defaultPort); err != nil {
		log.Fatal(err)
	}
}
