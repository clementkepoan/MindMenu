package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Restaurant endpoints
	r.POST("/restaurants", CreateRestaurant)
	r.GET("/restaurants/:restaurantId/branches", GetRestaurantBranches)

	// Branch endpoints
	r.POST("/branches", CreateBranch)
	r.GET("/branches", GetAllBranches)

	// Chatbot endpoints
	r.POST("/chatbots", CreateChatbot)
	r.POST("/branches/:branchId/query", QueryChatbot)

	// Debug endpoints
	r.GET("/debug/pinecone-indexes", listPineconeIndexes)
}
