package llm

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func Init(apiKey string) {
	client = openai.NewClient(apiKey)
}

func GetCompletion(prompt string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
