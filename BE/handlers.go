package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
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

// CreateRestaurant creates a new restaurant in Supabase
// CreateRestaurant creates a new restaurant in Supabase
func CreateRestaurant(c *gin.Context) {
	var restaurant Restaurant
	if err := c.ShouldBindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if restaurant.ID == "" {
		restaurant.ID = uuid.New().String()
	}

	if restaurant.OwnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner ID is required"})
		return
	}

	if restaurant.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Restaurant name is required"})
		return
	}

	// DON'T send timestamps - let the database handle them
	insertData := map[string]interface{}{
		"id":          restaurant.ID,
		"name":        restaurant.Name,
		"description": restaurant.Description,
		"owner_id":    restaurant.OwnerID,
		// Remove created_at and updated_at - let DB set them
	}

	log.Printf("Attempting to insert restaurant: %+v", insertData)

	// Insert into Supabase
	var inserted []Restaurant
	count, err := SupabaseClient.
		From("restaurants").
		Insert(insertData, false, "", "", "").
		ExecuteTo(&inserted)

	log.Printf("Insert operation - Count: %d, Error: %v", count, err)
	log.Printf("Inserted length: %d", len(inserted))

	if err != nil {
		log.Printf("Supabase insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create restaurant", "details": err.Error()})
		return
	}

	if count == 0 || len(inserted) == 0 {
		log.Printf("No rows inserted - Count: %d, Inserted length: %d", count, len(inserted))

		// Try to fetch the restaurant that should have been created
		var fetchedRestaurants []Restaurant
		fetchCount, fetchErr := SupabaseClient.
			From("restaurants").
			Select("*", "", false).
			Eq("id", restaurant.ID).
			ExecuteTo(&fetchedRestaurants)

		log.Printf("Fetch attempt - Count: %d, Error: %v", fetchCount, fetchErr)

		if fetchErr == nil && len(fetchedRestaurants) > 0 {
			// Restaurant was created but not returned due to RLS
			c.JSON(http.StatusCreated, fetchedRestaurants[0])
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No rows inserted - this may indicate a database constraint violation or RLS policy blocking return"})
		return
	}

	c.JSON(http.StatusCreated, inserted[0])
}

// CreateBranch creates a new branch for a restaurant
func CreateBranch(c *gin.Context) {
	var branch Branch
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if branch.ID == "" {
		branch.ID = uuid.New().String()
	}

	if branch.RestaurantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Restaurant ID is required"})
		return
	}

	if branch.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Branch name is required"})
		return
	}

	// DON'T send timestamps - let the database handle them
	insertData := map[string]interface{}{
		"id":            branch.ID,
		"restaurant_id": branch.RestaurantID,
		"name":          branch.Name,
		"address":       branch.Address,
		"has_chatbot":   false,
		// Remove created_at and updated_at - let DB set them
	}

	log.Printf("Attempting to insert branch: %+v", insertData)

	// Insert into Supabase
	var inserted []Branch
	count, err := SupabaseClient.
		From("branches").
		Insert(insertData, false, "", "", "").
		ExecuteTo(&inserted)

	log.Printf("Branch insert operation - Count: %d, Error: %v", count, err)
	log.Printf("Inserted length: %d", len(inserted))

	if err != nil {
		log.Printf("Supabase branch insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch", "details": err.Error()})
		return
	}

	if count == 0 || len(inserted) == 0 {
		log.Printf("No branch rows inserted - Count: %d, Inserted length: %d", count, len(inserted))

		// Try to fetch the branch that should have been created
		var fetchedBranches []Branch
		fetchCount, fetchErr := SupabaseClient.
			From("branches").
			Select("*", "", false).
			Eq("id", branch.ID).
			ExecuteTo(&fetchedBranches)

		log.Printf("Branch fetch attempt - Count: %d, Error: %v", fetchCount, fetchErr)

		if fetchErr == nil && len(fetchedBranches) > 0 {
			// Branch was created but not returned due to RLS
			c.JSON(http.StatusCreated, fetchedBranches[0])
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No branch rows inserted"})
		return
	}

	c.JSON(http.StatusCreated, inserted[0])
}

// GetRestaurantBranches gets all branches for a restaurant
func GetRestaurantBranches(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	var branches []Branch
	count, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("restaurant_id", restaurantID).
		ExecuteTo(&branches)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branches", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusOK, []Branch{}) // Return empty array instead of error for no branches
		return
	}

	c.JSON(http.StatusOK, branches)
}

// ChunkContent splits JSON content into text chunks
func chunkContent(content json.RawMessage) ([]TextChunk, error) {
	// Parse the raw JSON
	var contentMap map[string]interface{}
	if err := json.Unmarshal(content, &contentMap); err != nil {
		return nil, err
	}

	chunks := []TextChunk{}

	// Process different sections of the JSON content
	for section, data := range contentMap {
		switch v := data.(type) {
		case string:
			// For simple string values, create a single chunk
			chunk := TextChunk{
				ID:   uuid.New().String(),
				Text: fmt.Sprintf("%s: %s", section, v),
				Metadata: Metadata{
					Source:   section,
					Category: "general",
				},
			}
			chunks = append(chunks, chunk)
		case []interface{}:
			// For arrays, process each item
			for i, item := range v {
				itemStr, err := json.Marshal(item)
				if err != nil {
					continue
				}
				chunk := TextChunk{
					ID:   uuid.New().String(),
					Text: fmt.Sprintf("%s item %d: %s", section, i, string(itemStr)),
					Metadata: Metadata{
						Source:   section,
						Category: "list",
					},
				}
				chunks = append(chunks, chunk)
			}
		case map[string]interface{}:
			// For nested objects, process each field
			for key, val := range v {
				valStr, err := json.Marshal(val)
				if err != nil {
					continue
				}
				chunk := TextChunk{
					ID:   uuid.New().String(),
					Text: fmt.Sprintf("%s - %s: %s", section, key, string(valStr)),
					Metadata: Metadata{
						Source:   section,
						Category: "object",
					},
				}
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks, nil
}

// GenerateEmbeddings creates vector embeddings for text chunks using Gemini
func generateEmbeddings(ctx context.Context, chunks []TextChunk) ([]TextChunk, error) {
	// This is a simplified example - in a real implementation, you'd batch these requests
	for i := range chunks {
		// Using Gemini to generate embeddings
		// Note: This is simplified - you need to implement the actual API call
		// to get embeddings from your chosen provider
		embedding, err := getEmbeddingFromGemini(ctx, chunks[i].Text)
		if err != nil {
			return nil, err
		}
		chunks[i].Embedding = embedding
	}
	return chunks, nil
}

// Helper function to get embeddings from Gemini (simplified example)
func getEmbeddingFromGemini(ctx context.Context, text string) ([]float32, error) {
	// Implement your embedding generation here using GeminiClient
	// This is a placeholder
	embedding := make([]float32, 768) // Example dimension
	// Call your embedding model...
	return embedding, nil
}

// CreateChatbot handles the creation of a new chatbot for a branch
func CreateChatbot(c *gin.Context) {
	var req ChatbotContent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Looking for branch ID: %s", req.BranchID)

	// Verify branch exists
	var branches []Branch
	count, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", req.BranchID).
		ExecuteTo(&branches)
	log.Printf("Branch query result - Count: %d, Error: %v", count, err)
	if count > 0 {
		log.Printf("Found branch: %+v", branches[0])
	} else {
		log.Printf("No branches found for ID: %s", req.BranchID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check branch", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	branch := branches[0]

	// Get restaurant info for namespace creation
	var restaurants []Restaurant
	count, err = SupabaseClient.
		From("restaurants").
		Select("*", "", false).
		Eq("id", branch.RestaurantID).
		ExecuteTo(&restaurants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get restaurant", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	restaurant := restaurants[0]

	// Create a new chatbot entry - use map instead of struct
	chatbotData := map[string]interface{}{
		"id":        uuid.New().String(),
		"branch_id": req.BranchID,
		"status":    "building",
		// Remove timestamps - let DB handle them
	}

	var inserted []Chatbot
	count, err = SupabaseClient.
		From("chatbots").
		Insert(chatbotData, false, "", "", "").
		ExecuteTo(&inserted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chatbot", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No rows inserted"})
		return
	}

	createdChatbot := inserted[0]

	// Process the content in a goroutine
	go func() {
		// Create namespace in format RestaurantID + branch name
		namespace := fmt.Sprintf("%s_%s", restaurant.ID, strings.ReplaceAll(branch.Name, " ", "_"))

		// Chunk the content
		chunks, err := chunkContent(req.Content)
		if err != nil {
			log.Printf("Error chunking content: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		// Add metadata to chunks
		for i := range chunks {
			chunks[i].Metadata.RestaurantID = restaurant.ID
			chunks[i].Metadata.BranchID = branch.ID
		}

		// Generate embeddings
		ctx := context.Background()
		chunks, err = generateEmbeddings(ctx, chunks)
		if err != nil {
			log.Printf("Error generating embeddings: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		// Store in Pinecone
		err = storeChunksInPinecone(ctx, chunks, namespace)
		if err != nil {
			log.Printf("Error storing in Pinecone: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		// Update chatbot status
		updateChatbotStatus(createdChatbot.ID, "active")

		// Update branch to mark it as having a chatbot
		updateData := map[string]interface{}{
			"has_chatbot": true,
		}

		var updatedBranches []Branch
		_, err = SupabaseClient.
			From("branches").
			Update(updateData, "", "").
			Eq("id", req.BranchID).
			ExecuteTo(&updatedBranches)
		if err != nil {
			log.Printf("Error updating branch: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Chatbot creation started",
		"chatbot_id": createdChatbot.ID,
	})
}

// Helper function to update chatbot status
func updateChatbotStatus(chatbotID, status string) {
	updateData := map[string]interface{}{
		"status": status,
	}

	var updated []Chatbot
	_, err := SupabaseClient.
		From("chatbots").
		Update(updateData, "", "").
		Eq("id", chatbotID).
		ExecuteTo(&updated)
	if err != nil {
		log.Printf("Error updating chatbot status: %v", err)
	}
}

// Helper function to store chunks in Pinecone
func storeChunksInPinecone(ctx context.Context, chunks []TextChunk, namespace string) error {
	// Describe the index to get the host
	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	// Create index connection using the host
	idxConnection, err := PineconeClient.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to create IndexConnection for Host: %w", err)
	}

	// Convert chunks to Pinecone vectors
	vectors := make([]*pinecone.Vector, len(chunks))
	for i, chunk := range chunks {
		metadataMap := map[string]interface{}{
			"restaurant_id": chunk.Metadata.RestaurantID,
			"branch_id":     chunk.Metadata.BranchID,
			"source":        chunk.Metadata.Source,
			"category":      chunk.Metadata.Category,
			"text":          chunk.Text,
		}

		metadata, err := structpb.NewStruct(metadataMap)

		if err != nil {
			return fmt.Errorf("failed to create metadata struct: %w", err)
		}

		vectors[i] = &pinecone.Vector{
			Id:       chunk.ID,
			Values:   &chunk.Embedding,
			Metadata: metadata,
		}
	}

	// Upsert vectors in batches of 100
	batchSize := 100
	for i := 0; i < len(vectors); i += batchSize {
		end := i + batchSize
		if end > len(vectors) {
			end = len(vectors)
		}
		batch := vectors[i:end]

		count, err := idxConnection.UpsertVectors(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to upsert vectors to Pinecone: %w", err)
		}
		log.Printf("Successfully upserted %d vector(s)!", count)
	}

	return nil
}

// QueryChatbot handles queries to the chatbot
func QueryChatbot(c *gin.Context) {
	branchID := c.Param("branchId")

	// Get the query
	var query struct {
		Question string `json:"question" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get branch info
	var branches []Branch
	count, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", branchID).
		ExecuteTo(&branches)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	branch := branches[0]

	// Get restaurant info
	var restaurants []Restaurant
	count, err = SupabaseClient.
		From("restaurants").
		Select("*", "", false).
		Eq("id", branch.RestaurantID).
		ExecuteTo(&restaurants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get restaurant", "details": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	restaurant := restaurants[0]

	// Create namespace
	namespace := fmt.Sprintf("%s_%s", restaurant.ID, strings.ReplaceAll(branch.Name, " ", "_"))

	// Generate embedding for the query
	ctx := context.Background()
	embedding, err := getEmbeddingFromGemini(ctx, query.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query"})
		return
	}

	// Query Pinecone
	response, err := queryChatbotInPinecone(ctx, embedding, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query knowledge base"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to query the chatbot in Pinecone
func queryChatbotInPinecone(ctx context.Context, embedding []float32, namespace string) (gin.H, error) {
	// Describe the index to get the host (same as in storeChunksInPinecone)
	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	// Create index connection using the host
	index, err := PineconeClient.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create IndexConnection for Host: %w", err)
	}

	// Query the index - use the correct method signature
	queryResp, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          embedding,
		TopK:            5,
		IncludeMetadata: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query Pinecone: %w", err)
	}

	// Extract relevant text from matches
	// Step 1: Extract IDs from query result
	var contextTexts []string
	var ids []string
	for _, match := range queryResp.Matches {
		ids = append(ids, match.Vector.Id)
	}

	// Step 2: Fetch full metadata
	fetchResp, err := index.FetchVectors(ctx, ids)
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}

	// Step 3: Access metadata
	for _, vec := range fetchResp.Vectors {
		if vec.Metadata != nil {
			if textVal, ok := vec.Metadata.Fields["text"]; ok {
				fmt.Println("Found text:", textVal.GetStringValue())
				contextTexts = append(contextTexts, textVal.GetStringValue())
			}
		}

	}
	// Here you would typically use Gemini to generate a response based on the retrieved context
	// For now, just return the relevant texts
	return gin.H{
		"context": contextTexts,
		"message": "This is where you would use Gemini to generate a response",
	}, nil
}
