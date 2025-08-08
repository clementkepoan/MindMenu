package main

import (
	"encoding/json"
	"time"
)

// Restaurant represents a restaurant in the system
type Restaurant struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Branch represents a restaurant branch
type Branch struct {
	ID           string    `json:"id" db:"id"`
	RestaurantID string    `json:"restaurant_id" db:"restaurant_id"`
	Name         string    `json:"name" db:"name"`
	Address      string    `json:"address" db:"address"`
	HasChatbot   bool      `json:"has_chatbot" db:"has_chatbot"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Chatbot represents a chatbot instance
type Chatbot struct {
	ID          string    `json:"id" db:"id"`
	BranchID    string    `json:"branch_id" db:"branch_id"`
	Status      string    `json:"status" db:"status"`
	ContentHash string    `json:"content_hash" db:"content_hash"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ChatbotContent represents the request for creating a chatbot
type ChatbotContent struct {
	BranchID string          `json:"branch_id" binding:"required"`
	Content  json.RawMessage `json:"content" binding:"required"`
}

// TextChunk represents a chunk of text with embedding
type TextChunk struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
	Metadata  Metadata  `json:"metadata"`
}

// Metadata represents metadata for text chunks
type Metadata struct {
	RestaurantID string `json:"restaurant_id"`
	BranchID     string `json:"branch_id"`
	Source       string `json:"source"`
	Category     string `json:"category"`
	ItemKey   string `json:"item_key,omitempty"`
	ItemIndex int    `json:"item_index,omitempty"`
}
