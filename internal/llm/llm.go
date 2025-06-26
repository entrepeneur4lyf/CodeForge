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

// Provider interface that all LLM providers must implement
type Provider interface {
	Name() models.ModelProvider
	CreateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	SupportsModel(modelID models.ModelID) bool
	IsConfigured() bool
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

	if !provider.IsConfigured() {
		return nil, fmt.Errorf("provider %s not configured", model.Provider)
	}

	// Create the completion
	return provider.CreateCompletion(ctx, req)
}

// GetAvailableModels returns all models from configured providers
func GetAvailableModels() []models.Model {
	if manager == nil {
		return nil
	}

	var availableModels []models.Model
	for _, provider := range manager.providers {
		if !provider.IsConfigured() {
			continue
		}

		// Get all models for this provider
		providerModels := models.GetModelsByProvider(provider.Name())
		for _, model := range providerModels {
			if provider.SupportsModel(model.ID) {
				availableModels = append(availableModels, model)
			}
		}
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
		if !exists || !provider.IsConfigured() {
			continue
		}

		// Get the default model for this provider
		if defaultModel, exists := models.GetDefaultModelForProvider(providerName); exists {
			return defaultModel, nil
		}
	}

	return models.Model{}, fmt.Errorf("no configured providers available")
}

// initializeProviders sets up all available providers
func (m *Manager) initializeProviders() error {
	// For now, we'll add a basic OpenAI provider
	// We'll expand this to include all providers from OpenCode

	if providerCfg, exists := config.GetProvider(models.ProviderOpenAI); exists && !providerCfg.Disabled {
		openaiProvider := NewOpenAIProvider(providerCfg.APIKey)
		m.providers[models.ProviderOpenAI] = openaiProvider
	}

	// TODO: Add other providers (Anthropic, Gemini, etc.)
	// This will be expanded in the next iteration

	return nil
}
