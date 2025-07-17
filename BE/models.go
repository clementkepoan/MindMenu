package main

import "encoding/json"

type Example struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Restaurant represents a restaurant in the system
type Restaurant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
}

// Branch represents a branch of a restaurant
type Branch struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	HasChatbot   bool   `json:"has_chatbot"`
	CreatedAt    string `json:"created_at"`
}

// Chatbot represents a chatbot for a branch
type Chatbot struct {
	ID        string `json:"id"`
	BranchID  string `json:"branch_id"`
	Status    string `json:"status"` // "active", "building", "error"
	CreatedAt string `json:"created_at"`
}

// ChatbotContent is the input received from frontend for chatbot creation
type ChatbotContent struct {
	BranchID string          `json:"branch_id" binding:"required"`
	Content  json.RawMessage `json:"content" binding:"required"` // Raw JSON content
}

// TextChunk represents a chunk of text for vector storage
type TextChunk struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Metadata  Metadata  `json:"metadata"`
	Embedding []float32 `json:"-"` // Vector embedding, not stored in JSON
}

// Metadata for a text chunk
type Metadata struct {
	RestaurantID string `json:"restaurant_id"`
	BranchID     string `json:"branch_id"`
	Source       string `json:"source"`
	Category     string `json:"category"`
}
