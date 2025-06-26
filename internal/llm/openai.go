package llm

import (
	"context"
	"fmt"

	"github.com/entrepeneur4lyf/codeforge/internal/models"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	apiKey string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	var client *openai.Client
	if apiKey != "" {
		client = openai.NewClient(apiKey)
	}

	return &OpenAIProvider{
		client: client,
		apiKey: apiKey,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() models.ModelProvider {
	return models.ProviderOpenAI
}

// IsConfigured returns true if the provider is properly configured
func (p *OpenAIProvider) IsConfigured() bool {
	return p.client != nil && p.apiKey != ""
}

// SupportsModel returns true if the provider supports the given model
func (p *OpenAIProvider) SupportsModel(modelID models.ModelID) bool {
	model, exists := models.GetModel(modelID)
	if !exists {
		return false
	}
	return model.Provider == models.ProviderOpenAI
}

// CreateCompletion creates a completion using OpenAI's API
func (p *OpenAIProvider) CreateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	// Get the model information
	model, exists := models.GetModel(req.Model)
	if !exists {
		return nil, fmt.Errorf("unknown model: %s", req.Model)
	}

	if model.Provider != models.ProviderOpenAI {
		return nil, fmt.Errorf("model %s is not an OpenAI model", req.Model)
	}

	// Convert our messages to OpenAI format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Set up the request
	openaiReq := openai.ChatCompletionRequest{
		Model:    model.APIModel,
		Messages: openaiMessages,
	}

	// Set max tokens if specified
	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = int(req.MaxTokens)
	} else if model.DefaultMaxTokens > 0 {
		openaiReq.MaxTokens = int(model.DefaultMaxTokens)
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		openaiReq.Temperature = float32(req.Temperature)
	}

	// Handle reasoning models (O1 series) differently
	if model.CanReason {
		// O1 models don't support temperature or max_tokens in the same way
		openaiReq.Temperature = 0
		if req.MaxTokens > 0 {
			// O1 models use max_completion_tokens instead
			openaiReq.MaxCompletionTokens = int(req.MaxTokens)
		}
	}

	// Create the completion
	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	// Calculate total tokens used
	var tokensUsed int64
	if resp.Usage.TotalTokens > 0 {
		tokensUsed = int64(resp.Usage.TotalTokens)
	}

	return &CompletionResponse{
		Content:      resp.Choices[0].Message.Content,
		Model:        resp.Model,
		TokensUsed:   tokensUsed,
		FinishReason: string(resp.Choices[0].FinishReason),
	}, nil
}
