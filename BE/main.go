package main

import (
	"context"
	"fmt"
	"log"
	"os"

	generativelanguage "cloud.google.com/go/ai/generativelanguage/apiv1"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/supabase-community/supabase-go"
	"google.golang.org/api/option"
)

var (
	SupabaseClient *supabase.Client
	PineconeClient *pinecone.Client
	GeminiClient   *generativelanguage.GenerativeClient
)

func InitializeClients() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Supabase
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY") // Use service role key for RLS bypass

	// Fixed: supabase.NewClient now returns 2 values
	var err error
	SupabaseClient, err = supabase.NewClient(supabaseUrl, supabaseKey, nil)
	if err != nil {
		log.Printf("Warning: Supabase client initialization issue: %v", err)
	}

	// Pinecone
	pineconeApiKey := os.Getenv("PINECONE_API_KEY")

	// Fixed: pinecone.NewClient field names (ApiKey, not APIKey)
	config := pinecone.NewClientParams{
		ApiKey: pineconeApiKey, // Fixed: ApiKey not APIKey
		// Environment field doesn't exist in newer versions
	}
	PineconeClient, err = pinecone.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to initialize Pinecone client: %v", err)
	}

	// Create Pinecone index if it doesn't exist
	err = createPineconeIndex()
	if err != nil {
		log.Printf("Failed to create Pinecone index: %v", err)
		return fmt.Errorf("failed to initialize Pinecone index: %w", err)
	}

	// Gemini (Google Generative Language)
	ctx := context.Background()
	GeminiClient, err = generativelanguage.NewGenerativeClient(
		ctx,
		option.WithAPIKey(os.Getenv("GEMINI_API_KEY")),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini client: %v", err)
	}

	return nil
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize clients
	if err := InitializeClients(); err != nil {
		log.Fatalf("Failed to initialize clients: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Register all routes
	RegisterRoutes(r)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
