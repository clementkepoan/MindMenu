package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetExample(c *gin.Context) {
	// Example: Use Supabase, Pinecone, Gemini clients
	_ = SupabaseClient
	_ = PineconeClient
	_ = GeminiClient
	c.JSON(http.StatusOK, gin.H{"message": "GET example endpoint with integrations"})
}

func PostExample(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "POST example endpoint"})
}
