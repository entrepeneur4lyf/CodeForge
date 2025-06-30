package api

import (
	"net/http"
	"os"
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
	providers := s.getAvailableProviders()

	s.writeJSON(w, map[string]interface{}{
		"providers": providers,
		"total":     len(providers),
	})
}

// getAvailableProviders returns the list of available LLM providers
func (s *Server) getAvailableProviders() []LLMProvider {
	providers := []LLMProvider{
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "Claude models for advanced reasoning",
			Status:      s.getProviderStatus("anthropic"),
			ModelCount:  s.getProviderModelCount("anthropic"),
			LastUpdated: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "openai",
			Name:        "OpenAI",
			Description: "GPT models for general AI tasks",
			Status:      s.getProviderStatus("openai"),
			ModelCount:  s.getProviderModelCount("openai"),
			LastUpdated: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          "openrouter",
			Name:        "OpenRouter",
			Description: "Access to 300+ models from multiple providers",
			Status:      s.getProviderStatus("openrouter"),
			ModelCount:  s.getProviderModelCount("openrouter"),
			LastUpdated: time.Now().Add(-15 * time.Minute),
		},
		{
			ID:          "ollama",
			Name:        "Ollama",
			Description: "Local models for privacy and speed",
			Status:      s.getProviderStatus("ollama"),
			ModelCount:  s.getProviderModelCount("ollama"),
			LastUpdated: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini",
			Description: "Google's multimodal AI models",
			Status:      s.getProviderStatus("gemini"),
			ModelCount:  s.getProviderModelCount("gemini"),
			LastUpdated: time.Now().Add(-45 * time.Minute),
		},
		{
			ID:          "groq",
			Name:        "Groq",
			Description: "Ultra-fast inference for open models",
			Status:      s.getProviderStatus("groq"),
			ModelCount:  s.getProviderModelCount("groq"),
			LastUpdated: time.Now().Add(-20 * time.Minute),
		},
	}

	return providers
}

// getProviderModelCount returns the number of models for a provider
func (s *Server) getProviderModelCount(providerID string) int {
	switch providerID {
	case "anthropic":
		return 5 // Claude models
	case "openai":
		return 8 // GPT models
	case "openrouter":
		return 300 // Multiple providers
	case "ollama":
		return s.getOllamaModelCount()
	case "gemini":
		return 4 // Gemini models
	case "groq":
		return 6 // Groq models
	default:
		return 0
	}
}

// getOllamaModelCount checks how many Ollama models are available
func (s *Server) getOllamaModelCount() int {
	// Check if Ollama endpoint is configured
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	// TODO: Actually check Ollama endpoint for available models
	// For now, return 0 if not available, or estimated count if available
	if s.getProviderStatus("ollama") == "available" {
		return 10 // Estimated local models
	}
	return 0
}

// getAllAvailableModels returns all available models from all providers
func (s *Server) getAllAvailableModels() []LLMModel {
	var models []LLMModel

	// Add models from each provider
	models = append(models, s.getAnthropicModels()...)
	models = append(models, s.getOpenAIModels()...)
	models = append(models, s.getOpenRouterModels()...)
	models = append(models, s.getGeminiModels()...)
	models = append(models, s.getGroqModels()...)

	// Add Ollama models if available
	if s.getProviderStatus("ollama") == "available" {
		models = append(models, s.getOllamaModels()...)
	}

	return models
}

// handleLLMModels returns all available models
func (s *Server) handleLLMModels(w http.ResponseWriter, r *http.Request) {
	models := s.getAllAvailableModels()

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
				ID:           "claude-3-5-sonnet-20241022",
				Name:         "Claude 3.5 Sonnet",
				Provider:     "anthropic",
				Description:  "Most intelligent model for complex reasoning",
				ContextSize:  200000,
				Capabilities: []string{"text", "code", "analysis", "reasoning"},
			},
			{
				ID:           "claude-3-haiku-20240307",
				Name:         "Claude 3 Haiku",
				Provider:     "anthropic",
				Description:  "Fastest model for simple tasks",
				ContextSize:  200000,
				Capabilities: []string{"text", "code", "speed"},
			},
		}
	case "openai":
		models = []LLMModel{
			{
				ID:           "gpt-4o",
				Name:         "GPT-4o",
				Provider:     "openai",
				Description:  "Multimodal flagship model",
				ContextSize:  128000,
				Capabilities: []string{"text", "code", "vision", "audio"},
			},
			{
				ID:           "gpt-4o-mini",
				Name:         "GPT-4o Mini",
				Provider:     "openai",
				Description:  "Affordable and intelligent small model",
				ContextSize:  128000,
				Capabilities: []string{"text", "code", "speed"},
			},
		}
	case "openrouter":
		models = []LLMModel{
			{
				ID:           "anthropic/claude-3.5-sonnet",
				Name:         "Claude 3.5 Sonnet",
				Provider:     "openrouter",
				Description:  "Claude 3.5 Sonnet via OpenRouter",
				ContextSize:  200000,
				Capabilities: []string{"text", "code", "analysis"},
			},
			{
				ID:           "openai/gpt-4o",
				Name:         "GPT-4o",
				Provider:     "openrouter",
				Description:  "GPT-4o via OpenRouter",
				ContextSize:  128000,
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

// getAnthropicModels returns available Anthropic models
func (s *Server) getAnthropicModels() []LLMModel {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return []LLMModel{}
	}

	return []LLMModel{
		{
			ID:           "claude-3-5-sonnet-20241022",
			Name:         "Claude 3.5 Sonnet",
			Provider:     "anthropic",
			Description:  "Most intelligent model for complex reasoning",
			ContextSize:  200000,
			InputCost:    3.0,
			OutputCost:   15.0,
			Capabilities: []string{"text", "code", "analysis", "reasoning"},
		},
		{
			ID:           "claude-3-5-haiku-20241022",
			Name:         "Claude 3.5 Haiku",
			Provider:     "anthropic",
			Description:  "Fastest model for simple tasks",
			ContextSize:  200000,
			InputCost:    0.25,
			OutputCost:   1.25,
			Capabilities: []string{"text", "code", "speed"},
		},
	}
}

// getOpenAIModels returns available OpenAI models
func (s *Server) getOpenAIModels() []LLMModel {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return []LLMModel{}
	}

	return []LLMModel{
		{
			ID:           "gpt-4o",
			Name:         "GPT-4o",
			Provider:     "openai",
			Description:  "Multimodal flagship model",
			ContextSize:  128000,
			InputCost:    5.0,
			OutputCost:   15.0,
			Capabilities: []string{"text", "code", "vision", "audio"},
		},
		{
			ID:           "gpt-4o-mini",
			Name:         "GPT-4o Mini",
			Provider:     "openai",
			Description:  "Affordable and intelligent small model",
			ContextSize:  128000,
			InputCost:    0.15,
			OutputCost:   0.6,
			Capabilities: []string{"text", "code", "speed"},
		},
	}
}

// getOpenRouterModels returns available OpenRouter models
func (s *Server) getOpenRouterModels() []LLMModel {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		return []LLMModel{}
	}

	return []LLMModel{
		{
			ID:           "anthropic/claude-3.5-sonnet",
			Name:         "Claude 3.5 Sonnet",
			Provider:     "openrouter",
			Description:  "Claude 3.5 Sonnet via OpenRouter",
			ContextSize:  200000,
			InputCost:    3.0,
			OutputCost:   15.0,
			Capabilities: []string{"text", "code", "analysis"},
		},
		{
			ID:           "openai/gpt-4o",
			Name:         "GPT-4o",
			Provider:     "openrouter",
			Description:  "GPT-4o via OpenRouter",
			ContextSize:  128000,
			InputCost:    5.0,
			OutputCost:   15.0,
			Capabilities: []string{"text", "code", "vision"},
		},
	}
}

// getGeminiModels returns available Gemini models
func (s *Server) getGeminiModels() []LLMModel {
	if os.Getenv("GEMINI_API_KEY") == "" {
		return []LLMModel{}
	}

	return []LLMModel{
		{
			ID:           "gemini-2.0-flash-exp",
			Name:         "Gemini 2.0 Flash",
			Provider:     "gemini",
			Description:  "Google's latest multimodal model",
			ContextSize:  1000000,
			InputCost:    0.075,
			OutputCost:   0.3,
			Capabilities: []string{"text", "code", "vision", "audio"},
		},
	}
}

// getGroqModels returns available Groq models
func (s *Server) getGroqModels() []LLMModel {
	if os.Getenv("GROQ_API_KEY") == "" {
		return []LLMModel{}
	}

	return []LLMModel{
		{
			ID:           "llama-3.1-70b-versatile",
			Name:         "Llama 3.1 70B",
			Provider:     "groq",
			Description:  "Meta's Llama 3.1 70B on Groq",
			ContextSize:  131072,
			InputCost:    0.59,
			OutputCost:   0.79,
			Capabilities: []string{"text", "code", "speed"},
		},
	}
}

// getOllamaModels returns available Ollama models
func (s *Server) getOllamaModels() []LLMModel {
	// TODO: Actually query Ollama API for available models
	return []LLMModel{
		{
			ID:           "llama3.1:8b",
			Name:         "Llama 3.1 8B",
			Provider:     "ollama",
			Description:  "Meta's Llama 3.1 8B (local)",
			ContextSize:  131072,
			Capabilities: []string{"text", "code", "local"},
		},
		{
			ID:           "codellama:13b",
			Name:         "Code Llama 13B",
			Provider:     "ollama",
			Description:  "Meta's Code Llama 13B (local)",
			ContextSize:  16384,
			Capabilities: []string{"code", "local"},
		},
	}
}
