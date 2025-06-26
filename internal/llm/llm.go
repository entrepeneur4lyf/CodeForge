package llm

import (
	"context"
	"fmt"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/models"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request for LLM completion
type CompletionRequest struct {
	Model       models.ModelID `json:"model"`
	Messages    []Message      `json:"messages"`
	MaxTokens   int64          `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
}

// CompletionResponse represents the response from an LLM
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	TokensUsed   int64  `json:"tokens_used,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// Manager handles multiple LLM providers
type Manager struct {
	providers map[models.ModelProvider]Provider
	config    *config.Config
}

// Global manager instance
var manager *Manager

// Initialize sets up the LLM manager with configuration
func Initialize(cfg *config.Config) error {
	manager = &Manager{
		providers: make(map[models.ModelProvider]Provider),
		config:    cfg,
	}

	// Initialize providers based on configuration
	if err := manager.initializeProviders(); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	return nil
}

// GetCompletion creates a completion using the specified model
func GetCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if manager == nil {
		return nil, fmt.Errorf("LLM manager not initialized")
	}

	return manager.CreateCompletion(ctx, req)
}

// CreateCompletion creates a completion using the appropriate provider
func (m *Manager) CreateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Get the model information
	model, exists := models.GetModel(req.Model)
	if !exists {
		return nil, fmt.Errorf("unknown model: %s", req.Model)
	}

	// Get the provider for this model
	provider, exists := m.providers[model.Provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not available", model.Provider)
	}

	// Convert messages to a single string for now (simplified interface)
	var messageContent string
	for _, msg := range req.Messages {
		messageContent += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// Send message using the simplified interface
	response, err := provider.SendMessage(ctx, messageContent)
	if err != nil {
		return nil, err
	}

	// Estimate token usage (rough approximation: 1 token ≈ 4 characters)
	inputTokens := int64(len(messageContent) / 4)
	outputTokens := int64(len(response) / 4)
	totalTokens := inputTokens + outputTokens

	return &CompletionResponse{
		Content:      response,
		Model:        string(req.Model),
		TokensUsed:   totalTokens,
		FinishReason: "stop",
	}, nil
}

// GetAvailableModels returns all models from configured providers
func GetAvailableModels() []models.Model {
	if manager == nil {
		return nil
	}

	var availableModels []models.Model
	for providerType, provider := range manager.providers {
		if provider == nil {
			continue
		}

		// Get all models for this provider type
		providerModels := models.GetModelsByProvider(providerType)
		availableModels = append(availableModels, providerModels...)
	}

	return availableModels
}

// GetDefaultModel returns the default model for the first available provider
func GetDefaultModel() (models.Model, error) {
	if manager == nil {
		return models.Model{}, fmt.Errorf("LLM manager not initialized")
	}

	// Provider priority order (same as OpenCode)
	providerOrder := []models.ModelProvider{
		models.ProviderCopilot,
		models.ProviderAnthropic,
		models.ProviderOpenAI,
		models.ProviderGemini,
		models.ProviderGROQ,
		models.ProviderOpenRouter,
		models.ProviderBedrock,
		models.ProviderAzure,
		models.ProviderVertexAI,
		models.ProviderXAI,
		models.ProviderLocal,
	}

	for _, providerName := range providerOrder {
		provider, exists := manager.providers[providerName]
		if !exists || provider == nil {
			continue
		}

		// Get the default model for this provider
		if defaultModel, exists := models.GetDefaultModelForProvider(providerName); exists {
			return defaultModel, nil
		}
	}

	return models.Model{}, fmt.Errorf("no configured providers available")
}

// initializeProviders sets up all available providers from Phase 1 specification
// Based on OpenCode's proven provider architecture (MIT licensed)
func (m *Manager) initializeProviders() error {
	// Initialize all Phase 1 providers using OpenCode's pattern
	providerTypes := []models.ModelProvider{
		models.ProviderCopilot,
		models.ProviderAnthropic,
		models.ProviderOpenAI,
		models.ProviderGemini,
		models.ProviderGROQ,
		models.ProviderOpenRouter,
		models.ProviderBedrock,
		models.ProviderAzure,
		models.ProviderVertexAI,
		models.ProviderXAI,
		models.ProviderLocal,
	}

	for _, providerType := range providerTypes {
		if providerCfg, exists := config.GetProvider(providerType); exists && !providerCfg.Disabled {
			// Get default model for this provider
			defaultModel, modelExists := models.GetDefaultModelForProvider(providerType)
			if !modelExists {
				continue // Skip if no default model available
			}

			// Create provider using OpenCode's factory pattern
			provider, err := NewProvider(providerType,
				WithAPIKey(providerCfg.APIKey),
				WithModel(defaultModel),
				WithMaxTokens(defaultModel.DefaultMaxTokens),
			)
			if err != nil {
				fmt.Printf("Failed to initialize provider %s: %v\n", providerType, err)
				continue
			}

			m.providers[providerType] = provider
		}
	}

	return nil
}
