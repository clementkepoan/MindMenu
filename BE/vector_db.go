package main

import (
	"context"
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

// storeChunksInPinecone stores text chunks as vectors in Pinecone
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
