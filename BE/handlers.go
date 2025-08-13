package main

import (
	"context"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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
		// No branches found, return empty array with count
		c.JSON(http.StatusOK, gin.H{
			"count":    0,
			"branches": []Branch{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(branches),
		"branches": branches,
	})
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

	// Use branch_id as chatbot id; initialize version=1
	chatbotData := map[string]interface{}{
		"id":           req.BranchID,
		"branch_id":    req.BranchID,
		"status":       "building",
		"content_hash": hash,
		"version":      1,
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

// New: Lightweight creation without content. Returns chatbot_id.
func CreateChatbotLite(c *gin.Context) {
	var body struct {
		BranchID string `json:"branch_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure branch exists
	var branches []Branch
	_, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", body.BranchID).
		ExecuteTo(&branches)
	if err != nil || len(branches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	chatbot := map[string]interface{}{
		"id":                    uuid.New().String(),
		"branch_id":             body.BranchID,
		"status":                "idle",
		"content_hash":          "",
		"active_version_id":     nil,
		"last_indexed_version_id": nil,
	}

	var inserted []Chatbot
	_, err = SupabaseClient.
		From("chatbots").
		Insert(chatbot, false, "", "", "").
		ExecuteTo(&inserted)
	if err != nil || len(inserted) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chatbot"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"chatbot_id": inserted[0].ID})
}

// New: Add a version (store-only) for a chatbot. No vector DB work.
func AddChatbotVersion(c *gin.Context) {
	chatbotID := c.Param("chatbotId")
	var req struct {
		Content json.RawMessage `json:"content" binding:"required"`
		Notes   string          `json:"notes"`
		CreatedBy string        `json:"created_by"`
		MakeActive bool         `json:"make_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure chatbot exists
	var bots []Chatbot
	_, err := SupabaseClient.
		From("chatbots").
		Select("*", "", false).
		Eq("id", chatbotID).
		ExecuteTo(&bots)
	if err != nil || len(bots) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatbot not found"})
		return
	}

	h := generateHash(req.Content)
	version := map[string]interface{}{
		"id":           uuid.New().String(),
		"chatbot_id":   chatbotID,
		"content":      req.Content,
		"content_hash": h,
		"notes":        req.Notes,
		"created_by":   req.CreatedBy,
	}
	var vInserted []ChatbotVersion
	_, err = SupabaseClient.
		From("chatbot_versions").
		Insert(version, false, "", "", "").
		ExecuteTo(&vInserted)
	if err != nil || len(vInserted) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add version"})
		return
	}

	// Optionally mark as active
	if req.MakeActive {
		update := map[string]interface{}{
			"active_version_id": vInserted[0].ID,
		}
		var updated []Chatbot
		_, err = SupabaseClient.
			From("chatbots").
			Update(update, "", "").
			Eq("id", chatbotID).
			ExecuteTo(&updated)
		if err != nil {
			log.Printf("Warning: failed to set active version: %v", err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"version_id": vInserted[0].ID})
}

// New: Reindex a chatbot for a given version (or the active one) into Pinecone with selective upsert.
func ReindexChatbot(c *gin.Context) {
	chatbotID := c.Param("chatbotId") // equal to branch_id in simplified model
	var body struct {
		Content json.RawMessage `json:"content"` // optional; if omitted, use last stored content in chat_history or skip
		Prune   bool            `json:"prune"`
	}
	_ = c.ShouldBindJSON(&body) // accept empty

	// Load chatbot
	var bots []Chatbot
	_, err := SupabaseClient.
		From("chatbots").
		Select("*", "", false).
		Eq("id", chatbotID).
		ExecuteTo(&bots)
	if err != nil || len(bots) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chatbot not found"})
		return
	}
	bot := bots[0]

	// Fetch branch/restaurant for namespace
	var branches []Branch
	_, err = SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", bot.BranchID).
		ExecuteTo(&branches)
	if err != nil || len(branches) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load branch"})
		return
	}
	branch := branches[0]

	var restaurants []Restaurant
	_, err = SupabaseClient.
		From("restaurants").
		Select("*", "", false).
		Eq("id", branch.RestaurantID).
		ExecuteTo(&restaurants)
	if err != nil || len(restaurants) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load restaurant"})
		return
	}
	restaurant := restaurants[0]

	// Spawn background job: chunk -> embed -> upsert with selective diff
	go func() {
		// set status building
		var tmp []Chatbot
		_, _ = SupabaseClient.
			From("chatbots").
			Update(map[string]interface{}{"status": "building"}, "", "").
			Eq("id", chatbotID).
			ExecuteTo(&tmp)

		namespace := fmt.Sprintf("%s_%s", restaurant.ID, strings.ReplaceAll(branch.Name, " ", "_"))
		ctx := context.Background()
		var content json.RawMessage
		switch {
		case len(body.Content) > 0:
			content = body.Content
		default:
			// Try to fetch latest menu snapshot for this branch
			var snaps []MenuSnapshot
			_, err := SupabaseClient.
				From("menu_snapshots").
				Select("*", "", false).
				Eq("branch_id", branch.ID).
				ExecuteTo(&snaps)
			if err != nil || len(snaps) == 0 {
				log.Printf("Reindex: no content provided and no menu snapshot found; abort")
				updateChatbotStatus(chatbotID, "error")
				return
			}
			latest := snaps[0]
			for _, s := range snaps[1:] {
				if s.CreatedAt.After(latest.CreatedAt) {
					latest = s
				}
			}
			content = latest.Content
		}

		chunks, err := chunkContent(content)
		if err != nil {
			log.Printf("Reindex: chunking error: %v", err)
			updateChatbotStatus(chatbotID, "error")
			return
		}
		for i := range chunks {
			chunks[i].Metadata.RestaurantID = restaurant.ID
			chunks[i].Metadata.BranchID = branch.ID
		}

		chunks, err = generateEmbeddings(ctx, chunks)
		if err != nil {
			log.Printf("Reindex: embeddings error: %v", err)
			updateChatbotStatus(chatbotID, "error")
			return
		}

		if err := storeChunksInPinecone(ctx, chunks, namespace); err != nil {
			log.Printf("Reindex: store error: %v", err)
			updateChatbotStatus(chatbotID, "error")
			return
		}

		// Optionally prune: not implemented here; would require computing missing IDs.

		// success: if content changed (hash differs), increment version
		newHash := generateHash(content)
		newVersion := bot.Version
		if newHash != bot.ContentHash {
			newVersion = bot.Version + 1
		}
		update := map[string]interface{}{
			"status":       "active",
			"content_hash": newHash,
			"version":      newVersion,
		}
		var updated []Chatbot
		_, err = SupabaseClient.
			From("chatbots").
			Update(update, "", "").
			Eq("id", chatbotID).
			ExecuteTo(&updated)
		if err != nil {
			log.Printf("Reindex: failed to update chatbot: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Reindex started",
		"chatbot_id": chatbotID,
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

// SaveMenuSnapshot stores a raw JSON snapshot for a branch so users can review latest prices
func SaveMenuSnapshot(c *gin.Context) {
	branchID := c.Param("branchId")
	var body struct {
		Content   json.RawMessage `json:"content" binding:"required"`
		Notes     string          `json:"notes"`
		CreatedBy string          `json:"created_by"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure branch exists and user owns it via RLS
	var branches []Branch
	_, err := SupabaseClient.
		From("branches").
		Select("*", "", false).
		Eq("id", branchID).
		ExecuteTo(&branches)
	if err != nil || len(branches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	snapshot := map[string]interface{}{
		"branch_id":    branchID,
		"content":      body.Content,
		"content_hash": generateHash(body.Content),
		"notes":        body.Notes,
		"created_by":   body.CreatedBy,
	}

	var inserted []MenuSnapshot
	_, err = SupabaseClient.
		From("menu_snapshots").
		Insert(snapshot, false, "", "", "").
		ExecuteTo(&inserted)
	if err != nil || len(inserted) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save snapshot"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"snapshot_id": inserted[0].ID,
		"content_hash": inserted[0].ContentHash,
		"created_at":   inserted[0].CreatedAt,
	})
}

// GetLatestMenuSnapshot returns the latest stored JSON snapshot for a branch
func GetLatestMenuSnapshot(c *gin.Context) {
	branchID := c.Param("branchId")
	var rows []MenuSnapshot
	_, err := SupabaseClient.
		From("menu_snapshots").
		Select("*", "", false).
		Eq("branch_id", branchID).
		ExecuteTo(&rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch snapshot", "details": err.Error()})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No snapshots found"})
		return
	}
	// Pick the row with max CreatedAt
	latest := rows[0]
	for _, r := range rows[1:] {
		if r.CreatedAt.After(latest.CreatedAt) {
			latest = r
		}
	}
	c.JSON(http.StatusOK, gin.H{"snapshot": latest})
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
