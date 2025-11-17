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

	router := gin.New()
	router.Use(
		gin.Recovery(),
		RequestIDMiddleware(),
		LoggingMiddleware(),
	)

	api := router.Group("/v1")
	{
		users := api.Group("/users")
		{
			// публичные
			users.POST("/register", handleRegister)
			users.POST("/login", handleLogin)

			// защищённые
			users.Use(AuthRequired())
			users.GET("/me", handleMe)
			users.PATCH("/me", handleUpdateProfile)
			users.GET("", AdminRequired(), handleGetUsers)
		}
	}

	log.Println("service_users listening on", defaultPort)
	if err := router.Run(defaultPort); err != nil {
		log.Fatal(err)
	}
}
