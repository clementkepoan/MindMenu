package main

import (
	"context"

	"fmt"
	"log"
	"net/http"
	"strings"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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



func CreateChatbot(c *gin.Context) {
	var req ChatbotContent
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate hash from the content to detect changes
	hash := generateHash(req.Content)

	// Check if a chatbot with the same content hash already exists
	var existing []Chatbot
	_, err := SupabaseClient.
		From("chatbots").
		Select("*", "", false).
		Eq("branch_id", req.BranchID).
		Eq("content_hash", hash).
		ExecuteTo(&existing)

	if err == nil && len(existing) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Content unchanged. Skipping chatbot regeneration.",
			"chatbot_id": existing[0].ID,
		})
		return
	}

	log.Printf("Looking for branch ID: %s", req.BranchID)

	var branches []Branch
	_, err = SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", req.BranchID).
		ExecuteTo(&branches)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check branch", "details": err.Error()})
		return
	}
	if len(branches) == 0 {
		log.Printf("Branch not found - ID: %s", req.BranchID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	branch := branches[0]
	log.Printf("Found branch: %s", branch.Name)

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
	if len(restaurants) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	restaurant := restaurants[0]

	chatbotData := map[string]interface{}{
		"id":           uuid.New().String(),
		"branch_id":    req.BranchID,
		"status":       "building",
		"content_hash": hash,
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
	if len(inserted) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No chatbot rows inserted"})
		return
	}

	createdChatbot := inserted[0]

	go func() {
		namespace := fmt.Sprintf("%s_%s", restaurant.ID, strings.ReplaceAll(branch.Name, " ", "_"))

		chunks, err := chunkContent(req.Content)
		if err != nil {
			log.Printf("Error chunking content: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		for i := range chunks {
			chunks[i].Metadata.RestaurantID = restaurant.ID
			chunks[i].Metadata.BranchID = branch.ID
		}

		ctx := context.Background()
		chunks, err = generateEmbeddings(ctx, chunks)
		if err != nil {
			log.Printf("Error generating embeddings: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		err = storeChunksInPinecone(ctx, chunks, namespace)
		if err != nil {
			log.Printf("Error storing in Pinecone: %v", err)
			updateChatbotStatus(createdChatbot.ID, "error")
			return
		}

		updateChatbotStatus(createdChatbot.ID, "active")

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

type QueryWithHistoryRequest struct {
	Question  string `json:"question" binding:"required"`
	SessionID string `json:"session_id"`
	Language  string `json:"language"`
}

func QueryChatbotWithHistory(c *gin.Context) {
    branchID := c.Param("branchId") // Change from "branch_id" to "branchId" to match the route
    var query QueryWithHistoryRequest
    if err := c.ShouldBindJSON(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Set defaults
    if query.SessionID == "" {
        query.SessionID = uuid.New().String()
    }
    if query.Language == "" {
        query.Language = "en"
    }

    ctx := context.Background()

    // Get branch information (same as QueryChatbot)
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
    if len(branches) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
        return
    }

    branch := branches[0]

    
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
    if len(restaurants) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
        return
    }

    restaurant := restaurants[0]

    
    namespace := fmt.Sprintf("%s_%s", restaurant.ID, strings.ReplaceAll(branch.Name, " ", "_"))

   
    history, err := getChatHistory(query.SessionID, 10)
    if err != nil {
        log.Printf("Error getting chat history: %v", err)
        history = []ChatHistory{} 
    }

    
    embedding, err := getEmbeddingFromGemini(ctx, query.Question)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embedding"})
        return
    }

    // Query vector database with correct namespace
    response, err := queryChatbotInPineconeWithHistory(ctx, embedding, namespace, query.Question, history, query.Language)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query knowledge base"})
        return
    }

    // Store the interaction
    if responseStr, ok := response["response"].(string); ok {
        err = storeInteraction(query.SessionID, query.Question, responseStr, query.Language)
        if err != nil {
            log.Printf("Warning: Failed to store interaction: %v", err)
        }
    }

    // Add session_id to response
    response["session_id"] = query.SessionID

    c.JSON(http.StatusOK, response)
}
func generateHash(content json.RawMessage) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}