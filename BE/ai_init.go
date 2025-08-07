package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
)


type ChatHistory struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Query     string    `json:"query"`
	Response  string    `json:"response"`
	Language  string    `json:"language"`
	Timestamp time.Time `json:"timestamp"`
}

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
		Model: "models/gemini-2.5-flash",
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


func createChatHistoryTable() error {
	
	var result []ChatHistory
	_, err := SupabaseClient.
		From("chat_history").
		Select("id", "", false).
		Limit(1, "").
		ExecuteTo(&result)

	if err != nil {
		// Table might not exist, log warning
		log.Printf("Note: chat_history table may need to be created manually: %v", err)
		log.Printf("SQL to create table:")
		log.Printf(`
CREATE TABLE IF NOT EXISTS chat_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id TEXT NOT NULL,
    query TEXT NOT NULL,
    response TEXT NOT NULL,
    language TEXT DEFAULT 'en',
    timestamp TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_history_session_id ON chat_history(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_history_timestamp ON chat_history(timestamp);
        `)
		return fmt.Errorf("chat_history table may not exist: %w", err)
	}

	log.Printf("chat_history table exists or was created successfully")
	return nil
}


func getChatHistory(sessionID string, limit int) ([]ChatHistory, error) {
	if limit <= 0 {
		limit =3 
	}

	var history []ChatHistory
	_, err := SupabaseClient.
		From("chat_history").
		Select("*", "", false).
		Eq("session_id", sessionID).
		Order("timestamp",nil).
		Limit(limit, "").
		ExecuteTo(&history)

	if err != nil {
		log.Printf("Error retrieving chat history: %v", err)
		return []ChatHistory{}, fmt.Errorf("failed to retrieve chat history: %w", err)
	}

	
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	log.Printf("Retrieved %d messages for session %s", len(history), sessionID)
	return history, nil
}


func storeInteraction(sessionID, query, response, language string) error {
	if language == "" {
		language = "en" 
	}

	data := map[string]interface{}{
		"session_id": sessionID,
		"query":      query,
		"response":   response,
		"language":   language,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	var result []ChatHistory
	_, err := SupabaseClient.
		From("chat_history").
		Insert(data, false, "", "", "").
		ExecuteTo(&result)

	if err != nil {
		log.Printf("Error storing interaction: %v", err)
		return fmt.Errorf("failed to store interaction: %w", err)
	}

	log.Printf("Stored interaction for session %s", sessionID)
	return nil
}

// buildConversationContext builds conversation context from chat history
func buildConversationContext(history []ChatHistory) string {
	if len(history) == 0 {
		return "This is the start of a new conversation."
	}

	var contextParts []string
	// Use last 5 interactions to avoid token limits
	start := 0
	if len(history) > 5 {
		start = len(history) - 5
	}

	for _, interaction := range history[start:] {
		contextParts = append(contextParts, fmt.Sprintf("User: %s", interaction.Query))
		contextParts = append(contextParts, fmt.Sprintf("Assistant: %s", interaction.Response))
	}

	return strings.Join(contextParts, "\n")
}

// createRestaurantPromptWithHistory creates a prompt that includes conversation history
func createRestaurantPromptWithHistory(userQuestion string, context []string, history []ChatHistory, language string) string {
	knowledgeContext := strings.Join(context, "\n")
	conversationContext := buildConversationContext(history)

	// Language instructions
	languageInstructions := map[string]string{
		"en": "Respond in English",
		"zh": "请用中文回答",
		"ja": "日本語で回答してください",
		"ko": "한국어로 답변해주세요",
	}

	langInstruction := languageInstructions[language]
	if langInstruction == "" {
		langInstruction = "Respond in English"
	}

	prompt := fmt.Sprintf(`You are a helpful assistant for a restaurant. You specialize in providing information about the restaurant's menu, services, hours, and general dining experience.

%s.

Restaurant Knowledge (USE THIS INFORMATION TO ANSWER):
%s

Conversation History:
%s

Current User Question: %s

Instructions:
- Be friendly, helpful, and professional
- Focus on restaurant-related topics
- ALWAYS use the Restaurant Knowledge provided above to answer questions
- If the Restaurant Knowledge contains relevant information, use it directly in your response
- Only suggest contacting the restaurant if the specific information is not in the Restaurant Knowledge
- Keep responses concise but informative
- STRICTLY follow the language instructions provided
- Maintain the conversational context from previous messages

Response:`, langInstruction, knowledgeContext, conversationContext, userQuestion)

	return prompt
}
