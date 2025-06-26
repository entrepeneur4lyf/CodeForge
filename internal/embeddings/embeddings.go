package embeddings

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
)

//go:embed models/minilm-distilled/*
var embeddedModel embed.FS

// EmbeddingService handles text embedding generation
type EmbeddingService struct {
	modelPath   string
	pythonPath  string
	scriptPath  string
	tempDir     string
	initialized bool
	mu          sync.RWMutex
}

// EmbeddingResponse represents the response from the embedding service
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// Global embedding service instance
var embeddingService *EmbeddingService

// Initialize sets up the embedding service
func Initialize(cfg *config.Config) error {
	// Try to initialize native embedding service first
	if err := InitializeNative(cfg); err != nil {
		fmt.Printf("Warning: Failed to initialize native embedding service: %v\n", err)
		fmt.Println("Falling back to Python-based embedding service...")

		// Fallback to Python-based service
		return initializePythonService(cfg)
	}

	fmt.Println("Using native Go embedding service")
	return nil
}

// initializePythonService initializes the Python-based embedding service as fallback
func initializePythonService(cfg *config.Config) error {
	// Find Python executable
	pythonPath, err := findPython()
	if err != nil {
		return fmt.Errorf("failed to find Python executable: %w", err)
	}

	// Create temporary directory for extracted model
	tempDir, err := os.MkdirTemp("", "codeforge-model-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded model to temp directory
	modelPath := filepath.Join(tempDir, "minilm-distilled")
	if err := extractEmbeddedModel(modelPath); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to extract embedded model: %w", err)
	}

	// Create embedding service
	embeddingService = &EmbeddingService{
		modelPath:  modelPath,
		pythonPath: pythonPath,
		tempDir:    tempDir,
	}

	// Create the Python script for embedding generation
	if err := embeddingService.createEmbeddingScript(); err != nil {
		return fmt.Errorf("failed to create embedding script: %w", err)
	}

	// Mark as initialized before testing
	embeddingService.initialized = true

	// Test the embedding service
	if err := embeddingService.testEmbedding(); err != nil {
		embeddingService.initialized = false
		return fmt.Errorf("failed to test embedding service: %w", err)
	}
	return nil
}

// Get returns the global embedding service instance
func Get() *EmbeddingService {
	return embeddingService
}

// GetEmbedding generates an embedding using the best available service
func GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Try native service first
	if nativeService := GetNative(); nativeService != nil && nativeService.IsInitialized() {
		return nativeService.GenerateEmbedding(ctx, text)
	}

	// Fallback to Python service
	if embeddingService != nil && embeddingService.IsInitialized() {
		embedding64, err := embeddingService.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		return convertFloat64ToFloat32(embedding64), nil
	}

	return nil, fmt.Errorf("no embedding service available")
}

// convertFloat64ToFloat32 converts a float64 slice to float32
func convertFloat64ToFloat32(f64 []float64) []float32 {
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}
	return f32
}

// GetCodeEmbedding generates a code embedding using the best available service
func GetCodeEmbedding(ctx context.Context, code, language string) ([]float32, error) {
	// Try native service first
	if nativeService := GetNative(); nativeService != nil && nativeService.IsInitialized() {
		return nativeService.GenerateCodeEmbedding(ctx, code, language)
	}

	// Fallback to Python service
	if embeddingService != nil && embeddingService.IsInitialized() {
		embedding64, err := embeddingService.GenerateCodeEmbedding(ctx, code, language)
		if err != nil {
			return nil, err
		}
		return convertFloat64ToFloat32(embedding64), nil
	}

	return nil, fmt.Errorf("no embedding service available")
}

// GenerateEmbedding generates an embedding for the given text
func (es *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	es.mu.RLock()
	if !es.initialized {
		es.mu.RUnlock()
		return nil, fmt.Errorf("embedding service not initialized")
	}
	es.mu.RUnlock()

	// Prepare the command
	cmd := exec.CommandContext(ctx, es.pythonPath, es.scriptPath, text)

	// Run the command and capture output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Parse the JSON response
	var response EmbeddingResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("embedding generation error: %s", response.Error)
	}

	return response.Embedding, nil
}

// GenerateCodeEmbedding generates an embedding specifically for code
func (es *EmbeddingService) GenerateCodeEmbedding(ctx context.Context, code, language string) ([]float64, error) {
	// Preprocess code for better embeddings
	processedCode := preprocessCode(code, language)
	return es.GenerateEmbedding(ctx, processedCode)
}

// createEmbeddingScript creates the Python script for embedding generation
func (es *EmbeddingService) createEmbeddingScript() error {
	scriptContent := fmt.Sprintf(`#!/usr/bin/env python3
import sys
import json
import traceback
from sentence_transformers import SentenceTransformer

# Load the model2vec distilled model
MODEL_PATH = "%s"

try:
    model = SentenceTransformer(MODEL_PATH)
except Exception as e:
    print(json.dumps({"error": f"Failed to load model: {str(e)}"}))
    sys.exit(1)

def generate_embedding(text):
    try:
        # Generate embedding
        embedding = model.encode(text, convert_to_tensor=False)
        
        # Convert to list for JSON serialization
        embedding_list = embedding.tolist()
        
        return {"embedding": embedding_list}
    except Exception as e:
        return {"error": f"Failed to generate embedding: {str(e)}"}

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(json.dumps({"error": "Usage: script.py <text>"}))
        sys.exit(1)
    
    text = sys.argv[1]
    result = generate_embedding(text)
    print(json.dumps(result))
`, es.modelPath)

	// Create scripts directory
	scriptsDir := filepath.Join("scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Write the script
	scriptPath := filepath.Join(scriptsDir, "embed.py")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write embedding script: %w", err)
	}

	es.scriptPath = scriptPath
	return nil
}

// testEmbedding tests the embedding service with a simple text
func (es *EmbeddingService) testEmbedding() error {
	ctx := context.Background()
	testText := "Hello, world!"

	embedding, err := es.GenerateEmbedding(ctx, testText)
	if err != nil {
		return fmt.Errorf("embedding test failed: %w", err)
	}

	// Check that we got a reasonable embedding
	if len(embedding) == 0 {
		return fmt.Errorf("received empty embedding")
	}

	// Based on the config, we expect 256-dimensional embeddings
	expectedDim := 256
	if len(embedding) != expectedDim {
		return fmt.Errorf("unexpected embedding dimension: got %d, expected %d", len(embedding), expectedDim)
	}

	fmt.Printf("Embedding service initialized successfully (dimension: %d)\n", len(embedding))
	return nil
}

// findPython finds a suitable Python executable
func findPython() (string, error) {
	// Try different Python executables
	candidates := []string{"python3", "python", "python3.8", "python3.9", "python3.10", "python3.11", "python3.12"}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			// Test if sentence-transformers is available
			cmd := exec.Command(path, "-c", "import sentence_transformers")
			if err := cmd.Run(); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no suitable Python executable found with sentence-transformers installed")
}

// preprocessCode preprocesses code for better embedding generation
func preprocessCode(code, language string) string {
	// Remove excessive whitespace
	lines := strings.Split(code, "\n")
	var processedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}

	processed := strings.Join(processedLines, "\n")

	// Add language context for better embeddings
	return fmt.Sprintf("Language: %s\nCode:\n%s", language, processed)
}

// IsInitialized returns whether the embedding service is initialized
func (es *EmbeddingService) IsInitialized() bool {
	if es == nil {
		return false
	}
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.initialized
}

// extractEmbeddedModel extracts the embedded model files to the specified directory
func extractEmbeddedModel(targetDir string) error {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Walk through embedded files
	return fs.WalkDir(embeddedModel, "models/minilm-distilled", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "models/minilm-distilled" {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel("models/minilm-distilled", path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read embedded file
		data, err := embeddedModel.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Write to target location
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		return nil
	})
}

// Close cleans up the embedding service
func (es *EmbeddingService) Close() error {
	// Clean up the script file
	if es.scriptPath != "" {
		os.Remove(es.scriptPath)
	}

	// Clean up temporary directory
	if es.tempDir != "" {
		os.RemoveAll(es.tempDir)
	}

	return nil
}
