package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"

	//"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/types/known/structpb"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetExample(c *gin.Context) {
	_ = SupabaseClient
	_ = PineconeClient
	_ = GeminiClient
	c.JSON(http.StatusOK, gin.H{"message": "GET example endpoint with integrations"})
}

func PostExample(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "POST example endpoint"})
}

// CreateRestaurant creates a new restaurant in Supabase
func CreateRestaurant(c *gin.Context) {
	var restaurant Restaurant
	if err := c.ShouldBindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	insertData := map[string]interface{}{
		"id":          restaurant.ID,
		"name":        restaurant.Name,
		"description": restaurant.Description,
		"owner_id":    restaurant.OwnerID,
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

// Fix GetRestaurantBranches
func GetRestaurantBranches(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	var branches []Branch
	_, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("restaurant_id", restaurantID).
		ExecuteTo(&branches)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branches", "details": err.Error()})
		return
	}
	if len(branches) == 0 {
		c.JSON(http.StatusOK, []Branch{})
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
	req := &generativelanguagepb.EmbedContentRequest{
		Model: "models/text-embedding-004",
		Content: &generativelanguagepb.Content{
			Parts: []*generativelanguagepb.Part{
				{
					Data: &generativelanguagepb.Part_Text{
						Text: text,
					},
				},
			},
		},
	}

	resp, err := GeminiClient.EmbedContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding from Gemini: %w", err)
	}

	return resp.Embedding.Values, nil
}

// for testing func GetAllBranches
func GetAllBranches(c *gin.Context) {
	// Get all branches from Supabase
	log.Printf("Fetching all branches from Supabase")

	var branches []Branch
	count, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		ExecuteTo(&branches)

	log.Printf("Query result - Count: %d, Error: %v", count, err)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branches", "details": err.Error()})
		return
	}

	log.Printf("Found %d branches:", len(branches))
	for i, b := range branches {
		log.Printf("Branch %d: ID='%s', Name='%s', RestaurantID='%s', HasChatbot=%v",
			i, b.ID, b.Name, b.RestaurantID, b.HasChatbot)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    count,
		"branches": branches,
	})
}

// CreateChatbot handles the creation of a new chatbot for a branch
func CreateChatbot(c *gin.Context) {
	var req ChatbotContent
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Looking for branch ID: %s", req.BranchID)

	// Verify branch exists - Use results length, not count!
	var branches []Branch
	_, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", req.BranchID).
		ExecuteTo(&branches)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check branch", "details": err.Error()})
		return
	}

	// Use len(branches) instead of count!
	if len(branches) == 0 {
		log.Printf("Branch not found - ID: %s", req.BranchID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	branch := branches[0]
	log.Printf("Found branch: %s", branch.Name)

	// Get restaurant info
	var restaurants []Restaurant
	_, restErr := SupabaseClient.
		From("restaurants").
		Select("*", "", false).
		Eq("id", branch.RestaurantID).
		ExecuteTo(&restaurants)
	if restErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get restaurant", "details": restErr.Error()})
		return
	}
	if len(restaurants) == 0 { // Use len() here too!
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	restaurant := restaurants[0]

	// Create chatbot entry
	chatbotData := map[string]interface{}{
		"id":        uuid.New().String(),
		"branch_id": req.BranchID,
		"status":    "building",
	}

	var inserted []Chatbot
	_, chatbotErr := SupabaseClient.
		From("chatbots").
		Insert(chatbotData, false, "", "", "").
		ExecuteTo(&inserted)
	if chatbotErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chatbot", "details": chatbotErr.Error()})
		return
	}
	if len(inserted) == 0 { // Use len() here too!
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No chatbot rows inserted"})
		return
	}

	createdChatbot := inserted[0]

	// Process the content in a goroutine
	go func() {
		// Create namespace
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
	log.Printf("=== STORING VECTORS ===")
	log.Printf("Namespace: %s", namespace)
	log.Printf("Number of chunks: %d", len(chunks))

	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	// Create index connection using the host WITH namespace
	idxConnection, err := PineconeClient.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create IndexConnection for Host: %w", err)
	}

	// Convert chunks to Pinecone vectors
	vectors := make([]*pinecone.Vector, len(chunks))
	for i, chunk := range chunks {
		// Add debugging for each chunk
		log.Printf("Chunk %d: ID=%s", i, chunk.ID)
		log.Printf("  Text: %s", chunk.Text[:min(100, len(chunk.Text))])
		log.Printf("  Embedding length: %d", len(chunk.Embedding))
		log.Printf("  Restaurant ID: %s", chunk.Metadata.RestaurantID)
		log.Printf("  Branch ID: %s", chunk.Metadata.BranchID)

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

		_, err := idxConnection.UpsertVectors(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to upsert vectors to Pinecone: %w", err)
		}
		log.Printf("Successfully upserted batch of %d vectors to namespace '%s'!", len(batch), namespace)
	}

	log.Printf("=== STORAGE COMPLETE ===")
	return nil
}

func QueryChatbot(c *gin.Context) {
	branchID := c.Param("branchId")

	var query struct {
		Question string `json:"question" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get branch info
	var branches []Branch
	_, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", branchID).
		ExecuteTo(&branches)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch", "details": err.Error()})
		return
	}
	if len(branches) == 0 { // Use len() instead of count
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	branch := branches[0]

	// Get restaurant info
	var restaurants []Restaurant
	_, err = SupabaseClient.
		From("restaurants").
		Select("*", "", false).
		Eq("id", branch.RestaurantID).
		ExecuteTo(&restaurants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get restaurant", "details": err.Error()})
		return
	}
	if len(restaurants) == 0 { // Use len() instead of count
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

	// Query Pinecone - pass the user's question
	response, err := queryChatbotInPinecone(ctx, embedding, namespace, query.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query knowledge base"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update the return statement in queryChatbotInPinecone:
func queryChatbotInPinecone(ctx context.Context, embedding []float32, namespace string, userQuestion string) (gin.H, error) {
	log.Printf("=== QUERYING VECTORS ===")
	log.Printf("Query namespace: %s", namespace)
	log.Printf("User question: %s", userQuestion)
	log.Printf("Embedding length: %d", len(embedding))

	// Describe the index to get the host
	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	// Create index connection WITH namespace
	index, err := PineconeClient.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create IndexConnection for Host: %w", err)
	}

	// Query the index
	queryResp, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          embedding,
		TopK:            5,
		IncludeMetadata: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query Pinecone: %w", err)
	}

	log.Printf("Query returned %d matches", len(queryResp.Matches))

	var contextTexts []string
	var ids []string
	for _, match := range queryResp.Matches {
		ids = append(ids, match.Vector.Id)
		log.Printf("Match ID: %s, Score: %f", match.Vector.Id, match.Score)
	}

	// Fix the FetchVectors call
	if len(ids) > 0 {
		fetchResp, err := index.FetchVectors(ctx, ids)
		if err != nil {
			log.Printf("Fetch error: %v", err)
			return gin.H{
				"context": []string{},
				"message": "This is where you would use Gemini to generate a response",
				"error":   "Failed to fetch vector metadata",
			}, nil
		}

		log.Printf("Fetched %d vectors", len(fetchResp.Vectors))

		for _, vec := range fetchResp.Vectors {
			if vec.Metadata != nil {
				if textVal, ok := vec.Metadata.Fields["text"]; ok {
					text := textVal.GetStringValue()
					log.Printf("Found text: %s", text)
					contextTexts = append(contextTexts, text)
				}
			}
		}
	} else {
		log.Printf("No matching vectors found")
	}

	// Generate natural language response using the context
	var finalResponse string
	if len(contextTexts) > 0 {
		// Use improved prompt based on your Python reference
		prompt := createRestaurantPrompt(userQuestion, contextTexts)

		response, err := generateResponseWithGemini(ctx, prompt)
		if err != nil {
			log.Printf("Error generating response: %v", err)
			finalResponse = "I found some information but couldn't generate a proper response. Here's what I found: " + strings.Join(contextTexts, "; ")
		} else {
			finalResponse = response
		}
	} else {
		finalResponse = "I couldn't find any relevant information to answer your question."
	}

	log.Printf("=== QUERY COMPLETE ===")

	return gin.H{
		"response": finalResponse,
		"context":  contextTexts,
		"debug": gin.H{
			"namespace":     namespace,
			"matches":       len(queryResp.Matches),
			"context_count": len(contextTexts),
		},
	}, nil
}


func createRestaurantPrompt(userQuestion string, context []string) string {
	knowledgeContext := strings.Join(context, "\n")

	prompt := fmt.Sprintf(`You are a helpful assistant for a restaurant. You specialize in providing information about the restaurant's menu, services, hours, and general dining experience.

Respond in English.

Restaurant Knowledge (USE THIS INFORMATION TO ANSWER):
%s

Current User Question: %s

Instructions:
- Be friendly, helpful, and professional
- Focus on restaurant-related topics
- ALWAYS use the Restaurant Knowledge provided above to answer questions
- If the Restaurant Knowledge contains relevant information, use it directly in your response
- Only suggest contacting the restaurant if the specific information is not in the Restaurant Knowledge
- Keep responses concise but informative
- If asked about appetizers, focus on the appetizer information from the knowledge
- If asked about mains, focus on the main course information from the knowledge

Response:`, knowledgeContext, userQuestion)

	return prompt
}


func generateResponseWithGemini(ctx context.Context, prompt string) (string, error) {
	req := &generativelanguagepb.GenerateContentRequest{
		Model: "models/gemini-1.5-flash", // Change from gemini-2.5-flash to gemini-1.5-flash
		Contents: []*generativelanguagepb.Content{
			{
				Parts: []*generativelanguagepb.Part{
					{
						Data: &generativelanguagepb.Part_Text{
							Text: prompt,
						},
					},
				},
			},
		},
	}

	resp, err := GeminiClient.GenerateContent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return resp.Candidates[0].Content.Parts[0].GetText(), nil
}

func createPineconeIndex() error {
	ctx := context.Background()

	_, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err == nil {
		log.Printf("Index 'mindmenu-index' already exists")
		return nil
	}

	log.Printf("Creating Pinecone index 'mindmenu-index'...")

	dimension := int32(768)
	metric := pinecone.Cosine
	_, err = PineconeClient.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      "mindmenu-index",
		Dimension: &dimension,
		Metric:    &metric,
		Cloud:     pinecone.Aws,
		Region:    "us-east1",
	})

	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	log.Printf("Successfully created index 'mindmenu-index'")
	return nil
}
