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
	r.POST("/chatbots", CreateChatbot) 
	r.POST("/chatbots/:chatbotId/reindex", ReindexChatbot) 

	// Menu snapshot endpoints
	r.POST("/branches/:branchId/menu-snapshots", SaveMenuSnapshot)
	r.GET("/branches/:branchId/menu-snapshots/latest", GetLatestMenuSnapshot)
	
	
	
	r.POST("/branches/:branchId/query", QueryChatbot)
	r.POST("/branches/:branchId/query-with-history", QueryChatbotWithHistory)
}
