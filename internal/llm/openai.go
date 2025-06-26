package llm

import (
	"context"
	"fmt"

	"github.com/entrepeneur4lyf/codeforge/internal/models"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client  *openai.Client
	options ProviderOptions
}

// NewOpenAIProvider creates a new OpenAI provider using the new pattern
func NewOpenAIProvider(options *ProviderOptions) Provider {
	config := openai.DefaultConfig(options.APIKey)

	// Set custom base URL if provided (for Groq, OpenRouter, etc.)
	if options.BaseURL != "" {
		config.BaseURL = options.BaseURL
	}

	client := openai.NewClientWithConfig(config)

	return &OpenAIProvider{
		client:  client,
		options: *options,
	}
}

// Model returns the configured model
func (p *OpenAIProvider) Model() models.Model {
	return p.options.Model
}

// SendMessage sends a message and returns the response
func (p *OpenAIProvider) SendMessage(ctx context.Context, message string) (string, error) {
	if p.client == nil || p.options.APIKey == "" {
		return "", fmt.Errorf("OpenAI provider not configured")
	}

	// Create messages array
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		},
	}

	// Add system message if provided
	if p.options.SystemMessage != "" {
		messages = append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: p.options.SystemMessage,
			},
		}, messages...)
	}

	// Set up the request
	req := openai.ChatCompletionRequest{
		Model:    p.options.Model.APIModel,
		Messages: messages,
	}

	// Set max tokens
	if p.options.MaxTokens > 0 {
		req.MaxTokens = int(p.options.MaxTokens)
	}

	// Handle reasoning models (O1 series) differently
	if p.options.Model.CanReason {
		req.Temperature = 0
		if p.options.MaxTokens > 0 {
			req.MaxCompletionTokens = int(p.options.MaxTokens)
		}
	}

	// Create the completion
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
