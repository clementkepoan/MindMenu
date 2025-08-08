package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// updateChatbotStatus updates the status of a chatbot in the database
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

// chunkContent splits JSON content into text chunks for processing
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
					Source:    section,
					Category:  "general",
					ItemKey:   section, // logical key for this field
					ItemIndex: -1,
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
						Source:    section,
						Category:  "list",
						ItemKey:   "", // use index for stable ID
						ItemIndex: i,
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
						Source:    section,
						Category:  "object",
						ItemKey:   key, // use key for stable ID
						ItemIndex: -1,
					},
				}
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks, nil
}

// generateEmbeddings creates vector embeddings for text chunks using Gemini
func generateEmbeddings(ctx context.Context, chunks []TextChunk) ([]TextChunk, error) {

	for i := range chunks {
		// Using Gemini to generate embeddings
		embedding, err := getEmbeddingFromGemini(ctx, chunks[i].Text)
		if err != nil {
			return nil, err
		}
		chunks[i].Embedding = embedding
	}
	return chunks, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
