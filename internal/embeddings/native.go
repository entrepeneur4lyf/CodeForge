package embeddings

import (
	"context"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
)

//go:embed models/minilm-extracted/*
var embeddedModelData embed.FS

// NativeEmbeddingService handles text embedding generation using extracted model weights
type NativeEmbeddingService struct {
	embeddings   [][]float32 // [vocab_size][embedding_dim]
	vocab        map[string]int
	vocabSize    int
	embeddingDim int
	initialized  bool
	mu           sync.RWMutex
}

// EmbeddingMetadata represents the metadata for the embedding model
type EmbeddingMetadata struct {
	VocabSize    int    `json:"vocab_size"`
	EmbeddingDim int    `json:"embedding_dim"`
	DataType     string `json:"data_type"`
	ByteOrder    string `json:"byte_order"`
}

// Global native embedding service instance
var nativeEmbeddingService *NativeEmbeddingService

// InitializeNative sets up the native embedding service using extracted model weights
func InitializeNative(cfg *config.Config) error {
	nativeEmbeddingService = &NativeEmbeddingService{}

	// Load embedded model data
	if err := nativeEmbeddingService.loadEmbeddedModel(); err != nil {
		return fmt.Errorf("failed to load embedded model: %w", err)
	}

	// Mark as initialized before testing
	nativeEmbeddingService.initialized = true

	// Test the embedding service
	if err := nativeEmbeddingService.testEmbedding(); err != nil {
		nativeEmbeddingService.initialized = false
		return fmt.Errorf("failed to test native embedding service: %w", err)
	}
	return nil
}

// GetNative returns the global native embedding service instance
func GetNative() *NativeEmbeddingService {
	return nativeEmbeddingService
}

// loadEmbeddedModel loads the model data from embedded files
func (nes *NativeEmbeddingService) loadEmbeddedModel() error {
	// Load metadata
	metadataData, err := embeddedModelData.ReadFile("models/minilm-extracted/embeddings_metadata.json")
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata EmbeddingMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	nes.vocabSize = metadata.VocabSize
	nes.embeddingDim = metadata.EmbeddingDim

	// Load vocabulary
	vocabData, err := embeddedModelData.ReadFile("models/minilm-extracted/vocab.json")
	if err != nil {
		return fmt.Errorf("failed to read vocabulary: %w", err)
	}

	if err := json.Unmarshal(vocabData, &nes.vocab); err != nil {
		return fmt.Errorf("failed to parse vocabulary: %w", err)
	}

	// Load embedding weights
	embeddingData, err := embeddedModelData.ReadFile("models/minilm-extracted/embeddings.bin")
	if err != nil {
		return fmt.Errorf("failed to read embeddings: %w", err)
	}

	// Parse binary embedding data
	if err := nes.parseEmbeddingData(embeddingData); err != nil {
		return fmt.Errorf("failed to parse embedding data: %w", err)
	}

	fmt.Printf("Native embedding service loaded: vocab_size=%d, embedding_dim=%d\n",
		nes.vocabSize, nes.embeddingDim)

	return nil
}

// parseEmbeddingData parses the binary embedding data
func (nes *NativeEmbeddingService) parseEmbeddingData(data []byte) error {
	expectedSize := nes.vocabSize * nes.embeddingDim * 4 // 4 bytes per float32
	if len(data) != expectedSize {
		return fmt.Errorf("embedding data size mismatch: got %d, expected %d", len(data), expectedSize)
	}

	// Initialize embeddings slice
	nes.embeddings = make([][]float32, nes.vocabSize)
	for i := range nes.embeddings {
		nes.embeddings[i] = make([]float32, nes.embeddingDim)
	}

	// Parse binary data (little endian float32)
	for i := 0; i < nes.vocabSize; i++ {
		for j := 0; j < nes.embeddingDim; j++ {
			offset := (i*nes.embeddingDim + j) * 4
			bits := binary.LittleEndian.Uint32(data[offset : offset+4])
			nes.embeddings[i][j] = math.Float32frombits(bits)
		}
	}

	return nil
}

// GenerateEmbedding generates an embedding for the given text using the native model
func (nes *NativeEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Handle edge cases
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	if len(text) > 10000 { // Prevent excessive memory usage
		text = text[:10000]
	}

	nes.mu.RLock()
	if !nes.initialized {
		nes.mu.RUnlock()
		return nil, fmt.Errorf("native embedding service not initialized")
	}
	nes.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Optimized tokenization
	tokens := nes.tokenize(text)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no valid tokens found in text")
	}

	// Get embeddings for tokens
	var tokenEmbeddings [][]float32
	for _, token := range tokens {
		if tokenID, exists := nes.vocab[token]; exists && tokenID < len(nes.embeddings) {
			tokenEmbeddings = append(tokenEmbeddings, nes.embeddings[tokenID])
		}
	}

	if len(tokenEmbeddings) == 0 {
		return nil, fmt.Errorf("no valid token embeddings found")
	}

	// Average the token embeddings (simple pooling strategy)
	avgEmbedding := make([]float32, nes.embeddingDim)
	for _, tokenEmb := range tokenEmbeddings {
		for i, val := range tokenEmb {
			avgEmbedding[i] += val
		}
	}

	// Normalize by number of tokens
	numTokens := float32(len(tokenEmbeddings))
	for i := range avgEmbedding {
		avgEmbedding[i] /= numTokens
	}

	// L2 normalize the final embedding
	var norm float32 = 0.0
	for _, val := range avgEmbedding {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range avgEmbedding {
			avgEmbedding[i] /= norm
		}
	}

	// Pad to 1536 dimensions for Go libsql driver compatibility (OpenAI embedding size)
	paddedEmbedding := make([]float32, 1536)
	copy(paddedEmbedding, avgEmbedding)
	// The rest of the array is already zero-initialized

	return paddedEmbedding, nil
}

// tokenize performs improved tokenization with subword handling
func (nes *NativeEmbeddingService) tokenize(text string) []string {
	// Convert to lowercase for better vocabulary matching
	text = strings.ToLower(text)

	// Split into words first
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_')
	})

	var tokens []string
	for _, word := range words {
		if word == "" {
			continue
		}

		// Try exact word match first
		if _, exists := nes.vocab[word]; exists {
			tokens = append(tokens, word)
			continue
		}

		// Try common prefixes/suffixes for better coverage
		subwordTokens := nes.subwordTokenize(word)
		tokens = append(tokens, subwordTokens...)
	}

	// Fallback: add common tokens if we found nothing
	if len(tokens) == 0 {
		commonTokens := []string{"the", "a", "an", "and", "or", "in", "on", "at", "to", "for"}
		for _, token := range commonTokens {
			if _, exists := nes.vocab[token]; exists {
				tokens = append(tokens, token)
				break
			}
		}
	}

	return tokens
}

// subwordTokenize attempts to break words into subwords that exist in vocabulary
func (nes *NativeEmbeddingService) subwordTokenize(word string) []string {
	var tokens []string

	// Try progressively shorter prefixes
	for i := len(word); i >= 2; i-- {
		prefix := word[:i]
		if _, exists := nes.vocab[prefix]; exists {
			tokens = append(tokens, prefix)
			// Recursively tokenize the rest
			if i < len(word) {
				remaining := nes.subwordTokenize(word[i:])
				tokens = append(tokens, remaining...)
			}
			return tokens
		}
	}

	// Try single characters as last resort
	for _, char := range word {
		charStr := string(char)
		if _, exists := nes.vocab[charStr]; exists {
			tokens = append(tokens, charStr)
		}
	}

	return tokens
}

// GenerateCodeEmbedding generates an embedding specifically for code
func (nes *NativeEmbeddingService) GenerateCodeEmbedding(ctx context.Context, code, language string) ([]float32, error) {
	// Preprocess code for better embeddings
	processedCode := preprocessCodeForNative(code, language)
	return nes.GenerateEmbedding(ctx, processedCode)
}

// preprocessCodeForNative preprocesses code for better embedding generation
func preprocessCodeForNative(code, language string) string {
	// Remove excessive whitespace
	lines := strings.Split(code, "\n")
	var processedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
			processedLines = append(processedLines, trimmed)
		}
	}

	processed := strings.Join(processedLines, " ")

	// Add language context
	return fmt.Sprintf("%s code: %s", language, processed)
}

// testEmbedding tests the native embedding service
func (nes *NativeEmbeddingService) testEmbedding() error {
	ctx := context.Background()
	testText := "Hello, world!"

	embedding, err := nes.GenerateEmbedding(ctx, testText)
	if err != nil {
		return fmt.Errorf("native embedding test failed: %w", err)
	}

	// Check that we got a reasonable embedding
	if len(embedding) == 0 {
		return fmt.Errorf("received empty embedding")
	}

	// Check for padded dimension (1536) since we pad to match Go libsql driver expectations
	expectedDim := 1536
	if len(embedding) != expectedDim {
		return fmt.Errorf("unexpected embedding dimension: got %d, expected %d", len(embedding), expectedDim)
	}

	fmt.Printf("Native embedding service initialized successfully (dimension: %d)\n", len(embedding))
	return nil
}

// IsInitialized returns whether the native embedding service is initialized
func (nes *NativeEmbeddingService) IsInitialized() bool {
	if nes == nil {
		return false
	}
	nes.mu.RLock()
	defer nes.mu.RUnlock()
	return nes.initialized
}

// Close cleans up the native embedding service
func (nes *NativeEmbeddingService) Close() error {
	// Nothing to clean up for native service
	return nil
}
