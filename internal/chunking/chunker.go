package chunking

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
)

// ChunkingStrategy defines different approaches to code chunking
type ChunkingStrategy int

const (
	StrategyTreeSitter ChunkingStrategy = iota // Semantic chunking using tree-sitter
	StrategyFunction                           // Function-level chunking
	StrategyClass                              // Class-level chunking
	StrategyFile                               // File-level chunking
	StrategyText                               // Simple text chunking
)

// ChunkingConfig holds configuration for the chunker
type ChunkingConfig struct {
	MaxChunkSize    int              // Maximum characters per chunk
	OverlapSize     int              // Overlap between chunks
	Strategy        ChunkingStrategy // Chunking strategy to use
	IncludeContext  bool             // Include surrounding context
	ExtractSymbols  bool             // Extract symbol information
	ExtractImports  bool             // Extract import statements
	ExtractComments bool             // Extract documentation comments
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() ChunkingConfig {
	return ChunkingConfig{
		MaxChunkSize:    2000,
		OverlapSize:     200,
		Strategy:        StrategyTreeSitter,
		IncludeContext:  true,
		ExtractSymbols:  true,
		ExtractImports:  true,
		ExtractComments: true,
	}
}

// CodeChunker handles intelligent code chunking using tree-sitter
type CodeChunker struct {
	config ChunkingConfig
}

// NewCodeChunker creates a new code chunker with the given configuration
func NewCodeChunker(config ChunkingConfig) *CodeChunker {
	return &CodeChunker{
		config: config,
	}
}

// ChunkFile chunks a file into semantic code chunks
func (c *CodeChunker) ChunkFile(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	switch c.config.Strategy {
	case StrategyTreeSitter:
		return c.chunkWithTreeSitter(ctx, filePath, content, language)
	case StrategyFunction:
		return c.chunkByFunction(ctx, filePath, content, language)
	case StrategyClass:
		return c.chunkByClass(ctx, filePath, content, language)
	case StrategyFile:
		return c.chunkByFile(ctx, filePath, content, language)
	case StrategyText:
		return c.chunkByText(ctx, filePath, content, language)
	default:
		return c.chunkWithTreeSitter(ctx, filePath, content, language)
	}
}

// chunkWithTreeSitter uses tree-sitter for semantic chunking
func (c *CodeChunker) chunkWithTreeSitter(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	// For now, fallback to text chunking since tree-sitter is not implemented yet
	// TODO: Implement tree-sitter parsing when dependencies are available
	return c.chunkByText(ctx, filePath, content, language)
}

// Fallback chunking strategies

// Tree-sitter functions will be implemented when dependencies are available

// Fallback chunking strategies

// chunkByFunction chunks code by function boundaries
func (c *CodeChunker) chunkByFunction(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	// This would use regex or simple parsing to find function boundaries
	// For now, fallback to text chunking
	return c.chunkByText(ctx, filePath, content, language)
}

// chunkByClass chunks code by class boundaries
func (c *CodeChunker) chunkByClass(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	// This would use regex or simple parsing to find class boundaries
	// For now, fallback to text chunking
	return c.chunkByText(ctx, filePath, content, language)
}

// chunkByFile creates a single chunk for the entire file
func (c *CodeChunker) chunkByFile(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	chunk := &vectordb.CodeChunk{
		ID:       fmt.Sprintf("%s_file", filepath.Base(filePath)),
		FilePath: filePath,
		Content:  content,
		ChunkType: vectordb.ChunkType{
			Type: "file",
			Data: map[string]interface{}{
				"full_file": true,
			},
		},
		Language: language,
		Location: vectordb.SourceLocation{
			StartLine:   1,
			EndLine:     strings.Count(content, "\n") + 1,
			StartColumn: 1,
			EndColumn:   1,
		},
		Metadata: map[string]string{
			"chunk_strategy": "file",
		},
	}

	return []*vectordb.CodeChunk{chunk}, nil
}

// chunkByText performs simple text-based chunking
func (c *CodeChunker) chunkByText(ctx context.Context, filePath, content, language string) ([]*vectordb.CodeChunk, error) {
	chunks := []*vectordb.CodeChunk{}
	lines := strings.Split(content, "\n")

	chunkSize := c.config.MaxChunkSize
	overlap := c.config.OverlapSize

	currentChunk := ""
	startLine := 1
	chunkIndex := 0

	for i, line := range lines {
		if len(currentChunk)+len(line)+1 > chunkSize && len(currentChunk) > 0 {
			// Create chunk
			chunk := &vectordb.CodeChunk{
				ID:       fmt.Sprintf("%s_text_%d", filepath.Base(filePath), chunkIndex),
				FilePath: filePath,
				Content:  strings.TrimSpace(currentChunk),
				ChunkType: vectordb.ChunkType{
					Type: "text",
					Data: map[string]interface{}{
						"chunk_index": chunkIndex,
					},
				},
				Language: language,
				Location: vectordb.SourceLocation{
					StartLine:   startLine,
					EndLine:     i,
					StartColumn: 1,
					EndColumn:   len(line),
				},
				Metadata: map[string]string{
					"chunk_strategy": "text",
				},
			}
			chunks = append(chunks, chunk)

			// Start new chunk with overlap
			overlapLines := []string{}
			if overlap > 0 && len(lines) > i-overlap {
				overlapStart := i - overlap
				if overlapStart < 0 {
					overlapStart = 0
				}
				overlapLines = lines[overlapStart:i]
			}

			currentChunk = strings.Join(overlapLines, "\n")
			if len(currentChunk) > 0 {
				currentChunk += "\n"
			}
			startLine = i - len(overlapLines) + 1
			chunkIndex++
		}

		currentChunk += line + "\n"
	}

	// Add final chunk if there's remaining content
	if len(strings.TrimSpace(currentChunk)) > 0 {
		chunk := &vectordb.CodeChunk{
			ID:       fmt.Sprintf("%s_text_%d", filepath.Base(filePath), chunkIndex),
			FilePath: filePath,
			Content:  strings.TrimSpace(currentChunk),
			ChunkType: vectordb.ChunkType{
				Type: "text",
				Data: map[string]interface{}{
					"chunk_index": chunkIndex,
				},
			},
			Language: language,
			Location: vectordb.SourceLocation{
				StartLine:   startLine,
				EndLine:     len(lines),
				StartColumn: 1,
				EndColumn:   len(lines[len(lines)-1]),
			},
			Metadata: map[string]string{
				"chunk_strategy": "text",
			},
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
