package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	initConfig()

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

			// защищённые маршруты
			users.GET("/me", AuthRequired(), handleMe)

			// только админ: список всех пользователей
			users.GET("", AuthRequired(), AdminRequired(), handleGetUsers)
			// маршрут /v1/users (без суффиксов)
		}
	}

	log.Println("service_users listening on", defaultPort)
	if err := router.Run(defaultPort); err != nil {
		log.Fatal(err)
	}
}
