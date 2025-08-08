package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// createPineconeIndex creates a new Pinecone index if it doesn't exist
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

// computeContentHash returns a stable hash representing the meaningful content
func computeContentHash(chunk TextChunk) string {
	// Hash fields that should cause an update when changed
	input := strings.Join([]string{
		strings.TrimSpace(chunk.Text),
		strings.TrimSpace(chunk.Metadata.Source),
		strings.TrimSpace(chunk.Metadata.Category),
	}, "|")
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

// computeDeterministicID creates a stable vector ID for a logical chunk
func computeDeterministicID(m Metadata) string {
	itemPart := ""
	if m.ItemKey != "" {
		itemPart = "key:" + m.ItemKey
	} else {
		itemPart = fmt.Sprintf("idx:%d", m.ItemIndex)
	}
	key := strings.Join([]string{
		strings.TrimSpace(m.RestaurantID),
		strings.TrimSpace(m.BranchID),
		strings.TrimSpace(m.Source),
		strings.TrimSpace(m.Category),
		itemPart,
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return "mm_" + hex.EncodeToString(sum[:])
}

// fetchExistingHashes gets existing content_hash for a set of IDs in this namespace
func fetchExistingHashes(ctx context.Context, index *pinecone.IndexConnection, ids []string) (map[string]string, error) {
	result := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	for i := 0; i < len(ids); i += 100 {
		end := i + 100
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]
		fetchResp, err := index.FetchVectors(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing vectors: %w", err)
		}
		for id, vec := range fetchResp.Vectors {
			if vec.Metadata != nil {
				if hv, ok := vec.Metadata.Fields["content_hash"]; ok {
					result[id] = hv.GetStringValue()
				}
			}
		}
	}
	return result, nil
}

// storeChunksInPinecone stores text chunks as vectors in Pinecone (selective upsert)
func storeChunksInPinecone(ctx context.Context, chunks []TextChunk, namespace string) error {
	log.Printf("=== STORING VECTORS (SELECTIVE) ===")
	log.Printf("Namespace: %s", namespace)
	log.Printf("Number of chunks: %d", len(chunks))

	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConnection, err := PineconeClient.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create IndexConnection: %w", err)
	}

	// Build vectors with deterministic IDs and content hashes
	vectors := make([]*pinecone.Vector, len(chunks))
	ids := make([]string, 0, len(chunks))

	for i, chunk := range chunks {
		// Ensure metadata has RestaurantID and BranchID set by caller
		newID := computeDeterministicID(chunk.Metadata)
		if chunk.ID != newID {
			log.Printf("Chunk %d: overriding ID %s -> %s", i, chunk.ID, newID)
			chunk.ID = newID
		}
		contentHash := computeContentHash(chunk)

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
			"item_key":      chunk.Metadata.ItemKey,
			"item_index":    chunk.Metadata.ItemIndex,
			"text":          chunk.Text,
			"content_hash":  contentHash,
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
		ids = append(ids, chunk.ID)
	}

	// Fetch existing hashes to diff
	existingHashes, err := fetchExistingHashes(ctx, idxConnection, ids)
	if err != nil {
		return err
	}

	// Determine which vectors to upsert
	var toUpsert []*pinecone.Vector
	var newCount, updatedCount, skipped int

	for _, v := range vectors {
		var incomingHash string
		if v.Metadata != nil {
			if hv, ok := v.Metadata.Fields["content_hash"]; ok {
				incomingHash = hv.GetStringValue()
			}
		}
		existingHash, exists := existingHashes[v.Id]
		switch {
		case !exists:
			newCount++
			toUpsert = append(toUpsert, v)
		case existingHash != incomingHash:
			updatedCount++
			toUpsert = append(toUpsert, v)
		default:
			skipped++
		}
	}

	log.Printf("Diff results: new=%d, updated=%d, unchanged=%d", newCount, updatedCount, skipped)

	// Upsert only changed/new vectors. Do NOT delete by default.
	for i := 0; i < len(toUpsert); i += 100 {
		end := i + 100
		if end > len(toUpsert) {
			end = len(toUpsert)
		}
		batch := toUpsert[i:end]
		if len(batch) == 0 {
			continue
		}
		if _, err := idxConnection.UpsertVectors(ctx, batch); err != nil {
			return fmt.Errorf("failed to upsert vectors: %w", err)
		}
		log.Printf("Upserted %d vectors to namespace '%s'", len(batch), namespace)
	}

	log.Printf("=== STORAGE COMPLETE ===")
	return nil
}

// Optional utility: delete specific vectors by ID
func deleteVectorsByID(ctx context.Context, namespace string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	idx, err := PineconeClient.DescribeIndex(ctx, "mindmenu-index")
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}
	index, err := PineconeClient.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create IndexConnection: %w", err)
	}
	for i := 0; i < len(ids); i += 100 {
		end := i + 100
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]
		if err := index.DeleteVectorsById(ctx, batch); err != nil {
			return fmt.Errorf("failed to delete vectors: %w", err)
		}
		log.Printf("Deleted %d vectors from namespace '%s'", len(batch), namespace)
	}
	return nil
}

// queryChatbotInPinecone queries the vector database and generates AI responses
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

func queryChatbotInPineconeWithHistory(ctx context.Context, embedding []float32, namespace string, userQuestion string, history []ChatHistory, language string) (gin.H, error) {
	log.Printf("=== QUERYING VECTORS WITH HISTORY ===")
	log.Printf("Query namespace: %s", namespace)
	log.Printf("User question: %s", userQuestion)
	log.Printf("Language: %s", language)
	log.Printf("History items: %d", len(history))
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

	// Fetch vector metadata
	if len(ids) > 0 {
		fetchResp, err := index.FetchVectors(ctx, ids)
		if err != nil {
			log.Printf("Fetch error: %v", err)
			return gin.H{
				"context": []string{},
				"message": "Failed to fetch vector metadata",
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

	// Generate natural language response using the context and history
	var finalResponse string
	if len(contextTexts) > 0 {
		// Use the enhanced prompt with history
		prompt := createRestaurantPromptWithHistory(userQuestion, contextTexts, history, language)

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

	log.Printf("=== QUERY WITH HISTORY COMPLETE ===")

	return gin.H{
		"response": finalResponse,
		"context":  contextTexts,
		"debug": gin.H{
			"namespace":     namespace,
			"matches":       len(queryResp.Matches),
			"context_count": len(contextTexts),
			"history_count": len(history),
			"language":      language,
		},
	}, nil
}
