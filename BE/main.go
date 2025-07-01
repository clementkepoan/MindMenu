package main

import (
	"context"
	"log"
	"os"

	generativelanguage "cloud.google.com/go/ai/generativelanguage/apiv1beta"
	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"github.com/supabase-community/supabase-go"
	"google.golang.org/api/option"
)

var (
	SupabaseClient *supabase.Client
	PineconeClient *pinecone.Client
	GeminiClient   *generativelanguage.Client
)

func initClients() {
	// Supabase
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	SupabaseClient = supabase.NewClient(supabaseUrl, supabaseKey)

	// Pinecone
	pineconeApiKey := os.Getenv("PINECONE_API_KEY")
	pineconeEnv := os.Getenv("PINECONE_ENVIRONMENT")
	var err error
	PineconeClient, err = pinecone.NewClient(pineconeApiKey, pineconeEnv)
	if err != nil {
		log.Fatalf("Failed to initialize Pinecone client: %v", err)
	}

	// Gemini (Google Generative Language)
	ctx := context.Background()
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	GeminiClient, err = generativelanguage.NewClient(ctx, option.WithAPIKey(geminiApiKey))
	if err != nil {
		log.Fatalf("Failed to initialize Gemini client: %v", err)
	}
}

func main() {
	initClients()
	r := gin.Default()
	RegisterRoutes(r)
	r.Run() // listen and serve on 0.0.0.0:8080
}
