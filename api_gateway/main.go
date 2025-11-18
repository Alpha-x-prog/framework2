package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	initConfig()

	router := gin.New()
	router.Use(
		gin.Recovery(),
		RequestIDMiddleware(),
		LoggingMiddleware(),
		CORSMiddleware(),
		RateLimitMiddleware(),
	)

	// Публичные маршруты (без JWT)
	public := router.Group("/v1")
	{
		public.POST("/users/register", proxyToUsers)
		public.POST("/users/login", proxyToUsers)
	}

	// Защищённые маршруты (с JWT)
	protected := router.Group("/v1")
	protected.Use(JWTMiddleware())
	{
		// users
		protected.GET("/users/me", proxyToUsers)
		protected.PATCH("/users/me", proxyToUsers)
		protected.GET("/users", proxyToUsers)

		// orders
		protected.POST("/orders", proxyToOrders)
		protected.GET("/orders", proxyToOrders)
		protected.GET("/orders/:id", proxyToOrders)
		protected.PATCH("/orders/:id/status", proxyToOrders)
		protected.POST("/orders/:id/cancel", proxyToOrders)
		protected.DELETE("/orders/:id", proxyToOrders)
	}

	log.Println("api_gateway listening on", defaultPort)
	if err := router.Run(defaultPort); err != nil {
		log.Fatal(err)
	}
}
