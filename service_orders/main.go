package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	initConfig()

	if err := initDB(); err != nil {
		log.Fatalf("failed to init orders database: %v", err)
	}

	// не Default, чтобы контролировать middleware сами
	router := gin.New()
	router.Use(
		gin.Recovery(),
		RequestIDMiddleware(),
		LoggingMiddleware(),
	)

	api := router.Group("/v1")
	{
		orders := api.Group("/orders")
		{
			orders.Use(AuthRequired())

			orders.POST("", handleCreateOrder)
			orders.GET("/:id", handleGetOrder)
			orders.GET("", handleListMyOrders)

			orders.PATCH("/:id/status", handleUpdateOrderStatus)
			orders.POST("/:id/cancel", handleCancelOrder)
			orders.DELETE("/:id", handleDeleteOrder)
		}
	}

	log.Println("service_orders listening on", defaultPort)
	if err := router.Run(defaultPort); err != nil {
		log.Fatal(err)
	}
}
