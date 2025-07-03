package main

import (
	"context"
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
	GeminiClient *generativelanguage.GenerativeClient
)

func initClients() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Supabase
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

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

	// Gemini (Google Generative Language)
	ctx := context.Background()
    GeminiClient, err = generativelanguage.NewGenerativeClient(
        ctx,
        option.WithAPIKey(os.Getenv("GEMINI_API_KEY")),
    )
	if err != nil {
		log.Fatalf("Failed to initialize Gemini client: %v", err)
	}
}

func main() {
	initClients()
	r := gin.Default()

	// Use the RegisterRoutes function from routes.go
	RegisterRoutes(r)

	r.Run() // listen and serve on 0.0.0.0:8080
}
