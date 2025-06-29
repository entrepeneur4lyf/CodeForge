package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
)

// EmbeddingProvider represents different embedding providers
type EmbeddingProvider int

const (
	ProviderOllama EmbeddingProvider = iota
	ProviderOpenAI
	ProviderFallback
)

// EmbeddingService handles text embedding generation with multiple providers
type EmbeddingService struct {
	provider    EmbeddingProvider
	initialized bool
	mu          sync.RWMutex
}

// OllamaEmbeddingRequest represents a request to Ollama's embedding API
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents a response from Ollama's embedding API
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// OpenAIEmbeddingRequest represents a request to OpenAI's embedding API
type OpenAIEmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"`
}

// OpenAIEmbeddingResponse represents a response from OpenAI's embedding API
type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// Global embedding service instance
var embeddingService *EmbeddingService

// Initialize sets up the embedding service with the best available provider
func Initialize(cfg *config.Config) error {
	embeddingService = &EmbeddingService{}

	// Check config preference first
	if cfg != nil && cfg.Embedding.Provider != "" {
		switch cfg.Embedding.Provider {
		case "ollama":
			if isOllamaAvailable() {
				embeddingService.provider = ProviderOllama
				fmt.Println("Using Ollama embedding service (configured)")
				embeddingService.initialized = true
				return nil
			}
			fmt.Println("Ollama configured but not available, falling back...")
		case "openai":
			if isOpenAIAvailable() {
				embeddingService.provider = ProviderOpenAI
				fmt.Println("Using OpenAI embedding service (configured)")
				embeddingService.initialized = true
				return nil
			}
			fmt.Println("OpenAI configured but not available, falling back...")
		}
	}

	// Default to fallback (conservative approach)
	embeddingService.provider = ProviderFallback
	fmt.Println("Using fallback embedding service (use /embedding to configure)")

	// Show available options
	if isOllamaAvailable() {
		fmt.Println("💡 Ollama detected - use '/embedding ollama' for better quality")
	} else if isOpenAIAvailable() {
		fmt.Println("💡 OpenAI API detected - use '/embedding openai' for better quality")
	}

	embeddingService.initialized = true
	return nil
}

// Get returns the global embedding service instance
func Get() *EmbeddingService {
	return embeddingService
}

// GetEmbedding generates an embedding using the best available service
func GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	if embeddingService == nil || !embeddingService.initialized {
		return nil, fmt.Errorf("embedding service not initialized")
	}

	switch embeddingService.provider {
	case ProviderOllama:
		return getOllamaEmbedding(ctx, text)
	case ProviderOpenAI:
		return getOpenAIEmbedding(ctx, text)
	default:
		return getFallbackEmbedding(text), nil
	}
}

// GetCodeEmbedding generates a code embedding using the best available service
func GetCodeEmbedding(ctx context.Context, code, language string) ([]float32, error) {
	// Preprocess code for better embeddings
	processedCode := preprocessCode(code, language)
	return GetEmbedding(ctx, processedCode)
}

// isOllamaAvailable checks if Ollama is running and has an embedding model
func isOllamaAvailable() bool {
	// Check if Ollama is running
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Check if we have an embedding model (nomic-embed-text is common)
	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return false
	}

	// Look for common embedding models
	embeddingModels := []string{"nomic-embed-text", "all-minilm", "mxbai-embed-large"}
	for _, model := range tags.Models {
		for _, embModel := range embeddingModels {
			if strings.Contains(model.Name, embModel) {
				return true
			}
		}
	}

	return false
}

// isOpenAIAvailable checks if OpenAI API key is available
func isOpenAIAvailable() bool {
	return os.Getenv("OPENAI_API_KEY") != ""
}

// getOllamaEmbedding gets an embedding from Ollama
func getOllamaEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Use nomic-embed-text as default, or first available embedding model
	model := "nomic-embed-text"

	req := OllamaEmbeddingRequest{
		Model:  model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert float64 to float32
	embedding := make([]float32, len(ollamaResp.Embedding))
	for i, v := range ollamaResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// getOpenAIEmbedding gets an embedding from OpenAI
func getOpenAIEmbedding(ctx context.Context, text string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	req := OpenAIEmbeddingRequest{
		Input: text,
		Model: "text-embedding-3-small", // Cheaper and faster than large
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(body))
	}

	var openaiResp OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// Convert float64 to float32
	embedding := make([]float32, len(openaiResp.Data[0].Embedding))
	for i, v := range openaiResp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// getFallbackEmbedding creates a simple hash-based embedding as fallback
func getFallbackEmbedding(text string) []float32 {
	// Simple hash-based pseudo-embedding for fallback
	// This is basic but ensures the system always works
	const embeddingDim = 384

	embedding := make([]float32, embeddingDim)
	text = strings.ToLower(text)

	// Use a simple hash function to generate pseudo-embeddings
	for i := 0; i < embeddingDim; i++ {
		hash := 0
		for j, char := range text {
			hash = hash*31 + int(char) + i + j
		}
		// Normalize to [-1, 1] range
		embedding[i] = float32((hash%2000)-1000) / 1000.0
	}

	return embedding
}

// preprocessCode preprocesses code for better embeddings
func preprocessCode(code, language string) string {
	// Remove excessive whitespace
	lines := strings.Split(code, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	processed := strings.Join(cleanLines, "\n")

	// Add language context
	return fmt.Sprintf("Language: %s\nCode:\n%s", language, processed)
}
