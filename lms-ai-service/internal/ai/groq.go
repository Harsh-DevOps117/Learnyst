package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

var GroqClient *openai.Client

func InitGroq() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("WARNING: GROQ_API_KEY is not set. AI features will fail until provided.")
		// We still initialize with dummy so it doesn't panic on nil client, 
		// but actual requests will fail auth.
	}

	config := openai.DefaultConfig(apiKey)
	// Override the Base URL to point to Groq's OpenAI-compatible endpoint
	config.BaseURL = "https://api.groq.com/openai/v1"

	GroqClient = openai.NewClientWithConfig(config)
}

func AskDoubt(question string) (string, error) {
	if GroqClient == nil {
		return "", fmt.Errorf("Groq client not initialized")
	}

	systemPrompt := `You are an expert Teaching Assistant for an online learning platform. 
Provide clear, concise, and highly educational answers to student questions. 
Keep your answers under 3 paragraphs and use simple formatting.`

	resp, err := GroqClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "llama3-8b-8192", // Fast LPU model
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("Groq API error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}
