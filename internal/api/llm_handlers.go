package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// LLMProvider represents an LLM provider
type LLMProvider struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "available", "unavailable", "error"
	ModelCount  int       `json:"model_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// LLMModel represents an LLM model
type LLMModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Description  string                 `json:"description"`
	ContextSize  int                    `json:"context_size"`
	InputCost    float64                `json:"input_cost,omitempty"`
	OutputCost   float64                `json:"output_cost,omitempty"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// handleLLMProviders returns available LLM providers
func (s *Server) handleLLMProviders(w http.ResponseWriter, r *http.Request) {
	// TODO: Get actual providers from LLM handler
	providers := []LLMProvider{
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "Claude models for advanced reasoning",
			Status:      "available",
			ModelCount:  5,
			LastUpdated: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "openai",
			Name:        "OpenAI",
			Description: "GPT models for general AI tasks",
			Status:      "available",
			ModelCount:  8,
			LastUpdated: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          "openrouter",
			Name:        "OpenRouter",
			Description: "Access to 300+ models from multiple providers",
			Status:      "available",
			ModelCount:  300,
			LastUpdated: time.Now().Add(-15 * time.Minute),
		},
		{
			ID:          "ollama",
			Name:        "Ollama",
			Description: "Local models for privacy and speed",
			Status:      "unavailable",
			ModelCount:  0,
			LastUpdated: time.Now().Add(-5 * time.Minute),
		},
	}

	s.writeJSON(w, map[string]interface{}{
		"providers": providers,
		"total":     len(providers),
	})
}

// handleLLMModels returns all available models
func (s *Server) handleLLMModels(w http.ResponseWriter, r *http.Request) {
	// TODO: Get actual models from LLM handler
	models := []LLMModel{
		{
			ID:          "claude-3-5-sonnet-20241022",
			Name:        "Claude 3.5 Sonnet",
			Provider:    "anthropic",
			Description: "Most intelligent model for complex reasoning",
			ContextSize: 200000,
			InputCost:   3.0,
			OutputCost:  15.0,
			Capabilities: []string{"text", "code", "analysis", "reasoning"},
		},
		{
			ID:          "gpt-4o",
			Name:        "GPT-4o",
			Provider:    "openai",
			Description: "Multimodal flagship model",
			ContextSize: 128000,
			InputCost:   5.0,
			OutputCost:  15.0,
			Capabilities: []string{"text", "code", "vision", "audio"},
		},
		{
			ID:          "anthropic/claude-3.5-sonnet",
			Name:        "Claude 3.5 Sonnet (OpenRouter)",
			Provider:    "openrouter",
			Description: "Claude 3.5 Sonnet via OpenRouter",
			ContextSize: 200000,
			InputCost:   3.0,
			OutputCost:  15.0,
			Capabilities: []string{"text", "code", "analysis"},
		},
	}

	// Filter by provider if specified
	provider := r.URL.Query().Get("provider")
	if provider != "" {
		var filtered []LLMModel
		for _, model := range models {
			if model.Provider == provider {
				filtered = append(filtered, model)
			}
		}
		models = filtered
	}

	s.writeJSON(w, map[string]interface{}{
		"models": models,
		"total":  len(models),
	})
}

// handleProviderModels returns models for a specific provider
func (s *Server) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["provider"]

	// TODO: Get actual models for the provider
	var models []LLMModel

	switch providerID {
	case "anthropic":
		models = []LLMModel{
			{
				ID:          "claude-3-5-sonnet-20241022",
				Name:        "Claude 3.5 Sonnet",
				Provider:    "anthropic",
				Description: "Most intelligent model for complex reasoning",
				ContextSize: 200000,
				Capabilities: []string{"text", "code", "analysis", "reasoning"},
			},
			{
				ID:          "claude-3-haiku-20240307",
				Name:        "Claude 3 Haiku",
				Provider:    "anthropic",
				Description: "Fastest model for simple tasks",
				ContextSize: 200000,
				Capabilities: []string{"text", "code", "speed"},
			},
		}
	case "openai":
		models = []LLMModel{
			{
				ID:          "gpt-4o",
				Name:        "GPT-4o",
				Provider:    "openai",
				Description: "Multimodal flagship model",
				ContextSize: 128000,
				Capabilities: []string{"text", "code", "vision", "audio"},
			},
			{
				ID:          "gpt-4o-mini",
				Name:        "GPT-4o Mini",
				Provider:    "openai",
				Description: "Affordable and intelligent small model",
				ContextSize: 128000,
				Capabilities: []string{"text", "code", "speed"},
			},
		}
	case "openrouter":
		models = []LLMModel{
			{
				ID:          "anthropic/claude-3.5-sonnet",
				Name:        "Claude 3.5 Sonnet",
				Provider:    "openrouter",
				Description: "Claude 3.5 Sonnet via OpenRouter",
				ContextSize: 200000,
				Capabilities: []string{"text", "code", "analysis"},
			},
			{
				ID:          "openai/gpt-4o",
				Name:        "GPT-4o",
				Provider:    "openrouter",
				Description: "GPT-4o via OpenRouter",
				ContextSize: 128000,
				Capabilities: []string{"text", "code", "vision"},
			},
		}
	}

	s.writeJSON(w, map[string]interface{}{
		"provider": providerID,
		"models":   models,
		"total":    len(models),
	})
}
