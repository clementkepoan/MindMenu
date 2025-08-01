package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Example endpoints
	r.GET("/example", GetExample)
	r.POST("/example", PostExample)

	// Restaurant endpoints
	r.POST("/restaurants", CreateRestaurant)
	r.GET("/restaurants/:restaurantId/branches", GetRestaurantBranches)

	// Branch endpoints
	r.POST("/branches", CreateBranch)

	// Chatbot endpoints
	r.POST("/chatbots", CreateChatbot)
	r.POST("/branches/:branchId/query", QueryChatbot)
}
