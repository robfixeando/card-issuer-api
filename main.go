package main

import (
	"card-issuer-api/database"
	"card-issuer-api/handlers"
	"card-issuer-api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {

	database.Connect()

	r := gin.Default()

	api := r.Group("/api/v1")
	api.Use(middleware.TenantMiddleware())
	{
		api.POST("/cards", handlers.CreateCard)
		api.POST("/cards/batch", handlers.BatchCreateCards)
		api.GET("/cards", handlers.ListCards)
		api.PATCH("/cards/:id/status", handlers.UpdateCardStatus)
		api.GET("/cards/:id", handlers.GetCardByID)
	}

	r.Run(":8085")
}
