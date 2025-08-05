package main

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
)

// getEmbeddingFromGemini generates embeddings using Gemini API
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

// generateResponseWithGemini generates text responses using Gemini API
func generateResponseWithGemini(ctx context.Context, prompt string) (string, error) {
	req := &generativelanguagepb.GenerateContentRequest{
		Model: "models/gemini-1.5-flash",
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

// createRestaurantPrompt creates a structured prompt for the restaurant AI assistant
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
