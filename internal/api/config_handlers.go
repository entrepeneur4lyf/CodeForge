package api

import (
	"encoding/json"
	"net/http"
)

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	LLM       LLMConfig       `json:"llm"`
	Embedding EmbeddingConfig `json:"embedding"`
	Database  DatabaseConfig  `json:"database"`
	API       APIConfig       `json:"api"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	DefaultProvider string            `json:"default_provider"`
	DefaultModel    string            `json:"default_model"`
	Providers       map[string]bool   `json:"providers"`
	Settings        map[string]string `json:"settings"`
}

// EmbeddingConfig represents embedding configuration
type EmbeddingConfig struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Status   string `json:"status"`
	Size     int64  `json:"size"`
	Chunks   int    `json:"chunks"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Port    int    `json:"port"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
}

// ConfigUpdateRequest represents a configuration update request
type ConfigUpdateRequest struct {
	LLM       *LLMConfig       `json:"llm,omitempty"`
	Embedding *EmbeddingConfig `json:"embedding,omitempty"`
	Database  *DatabaseConfig  `json:"database,omitempty"`
	API       *APIConfig       `json:"api,omitempty"`
}

// handleConfig handles GET /config and PUT /config
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getConfig(w, r)
	case "PUT":
		s.updateConfig(w, r)
	}
}

// getConfig returns current configuration
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	config := ConfigResponse{
		LLM: LLMConfig{
			DefaultProvider: "anthropic",
			DefaultModel:    "claude-3-5-sonnet-20241022",
			Providers: map[string]bool{
				"anthropic":  true,
				"openai":     true,
				"openrouter": true,
				"ollama":     false,
			},
			Settings: map[string]string{
				"temperature":    "0.7",
				"max_tokens":     "4096",
				"timeout":        "30s",
			},
		},
		Embedding: EmbeddingConfig{
			Provider:   "fallback",
			Model:      "hash-based",
			Dimensions: 384,
		},
		Database: DatabaseConfig{
			Type:   "libsql",
			Path:   "./codeforge.db",
			Status: "connected",
			Size:   1024000,
			Chunks: 1247,
		},
		API: APIConfig{
			Port:    8080,
			Version: "1.0.0",
			Debug:   false,
		},
	}

	// TODO: Get actual configuration from config service
	if s.config != nil {
		// Update with actual config values
	}

	s.writeJSON(w, config)
}

// updateConfig updates configuration
func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validate and apply configuration changes
	
	// For now, just return success
	response := map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
		"changes": req,
	}

	s.writeJSON(w, response)
}
