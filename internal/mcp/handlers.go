package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool Handlers

// handleSemanticSearch handles semantic code search requests
func (cfs *CodeForgeServer) handleSemanticSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	// Get optional parameters
	maxResults := int(request.GetFloat("max_results", 10))
	language := request.GetString("language", "")
	chunkType := request.GetString("chunk_type", "")

	// Build filters
	filters := make(map[string]string)
	if language != "" {
		filters["language"] = language
	}
	if chunkType != "" {
		filters["chunk_type"] = chunkType
	}

	// TODO: Generate embedding for query using embedding service
	// For now, create a dummy embedding
	queryEmbedding := make([]float32, 256)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.1 // Placeholder
	}

	// Search using vector database
	results, err := cfs.vectorDB.SearchSimilarChunks(ctx, queryEmbedding, maxResults, filters)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Format results
	var resultTexts []string
	for _, result := range results {
		resultText := fmt.Sprintf("Query: %s\n\nFile: %s\nType: %s\nLanguage: %s\nScore: %.3f\n\nContent:\n%s\n---",
			query,
			result.Chunk.FilePath,
			result.Chunk.ChunkType,
			result.Chunk.Language,
			result.Score,
			result.Chunk.Content,
		)
		resultTexts = append(resultTexts, resultText)
	}

	if len(resultTexts) == 0 {
		return mcp.NewToolResultText("No similar code found for the given query."), nil
	}

	return mcp.NewToolResultText(strings.Join(resultTexts, "\n\n")), nil
}

// handleReadFile handles file reading requests
func (cfs *CodeForgeServer) handleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	// Validate and resolve path
	fullPath, err := cfs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid path: %v", err)), nil
	}

	// Check if file exists
	if !cfs.fileExists(fullPath) {
		return mcp.NewToolResultError(fmt.Sprintf("file not found: %s", path)), nil
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// handleWriteFile handles file writing requests
func (cfs *CodeForgeServer) handleWriteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content parameter is required"), nil
	}

	// Validate and resolve path
	fullPath, err := cfs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid path: %v", err)), nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directory: %v", err)), nil
	}

	// Write file content
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path)), nil
}

// handleCodeAnalysis handles code analysis requests
func (cfs *CodeForgeServer) handleCodeAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path parameter is required"), nil
	}

	// Validate and resolve path
	fullPath, err := cfs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid path: %v", err)), nil
	}

	// Check if file exists
	if !cfs.fileExists(fullPath) {
		return mcp.NewToolResultError(fmt.Sprintf("file not found: %s", path)), nil
	}

	// TODO: Implement actual code analysis using LSP or tree-sitter
	// For now, return basic file information
	info, err := os.Stat(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get file info: %v", err)), nil
	}

	analysis := map[string]interface{}{
		"file_path": path,
		"size":      info.Size(),
		"modified":  info.ModTime(),
		"language":  detectLanguage(path),
		"symbols":   []string{}, // TODO: Extract actual symbols
		"imports":   []string{}, // TODO: Extract actual imports
	}

	analysisJSON, _ := json.MarshalIndent(analysis, "", "  ")
	return mcp.NewToolResultText(string(analysisJSON)), nil
}

// handleProjectStructure handles project structure requests
func (cfs *CodeForgeServer) handleProjectStructure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := request.GetString("path", ".")
	maxDepth := int(request.GetFloat("max_depth", 3))

	// Validate and resolve path
	fullPath, err := cfs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid path: %v", err)), nil
	}

	// Build directory tree
	tree, err := cfs.buildDirectoryTree(fullPath, maxDepth)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to build directory tree: %v", err)), nil
	}

	return mcp.NewToolResultText(tree), nil
}

// Resource Handlers

// handleProjectMetadata handles project metadata resource requests
func (cfs *CodeForgeServer) handleProjectMetadata(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	metadata := map[string]interface{}{
		"name":           "CodeForge Project",
		"workspace_root": cfs.workspaceRoot,
		"version":        "0.1.0",
		"description":    "AI-powered code intelligence platform",
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(metadataJSON),
		},
	}, nil
}

// handleFileResource handles file resource requests
func (cfs *CodeForgeServer) handleFileResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract file path from URI (codeforge://files/{path})
	uri := request.Params.URI
	if !strings.HasPrefix(uri, "codeforge://files/") {
		return nil, fmt.Errorf("invalid file resource URI: %s", uri)
	}

	path := strings.TrimPrefix(uri, "codeforge://files/")

	// Validate and resolve path
	fullPath, err := cfs.validatePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Check if file exists
	if !cfs.fileExists(fullPath) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Determine MIME type based on file extension
	mimeType := detectMIMEType(path)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mimeType,
			Text:     string(content),
		},
	}, nil
}

// handleGitStatus handles git status resource requests
func (cfs *CodeForgeServer) handleGitStatus(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// TODO: Implement actual git status checking
	// For now, return placeholder data
	gitStatus := map[string]interface{}{
		"branch":    "main",
		"status":    "clean",
		"modified":  []string{},
		"untracked": []string{},
		"staged":    []string{},
	}

	statusJSON, _ := json.MarshalIndent(gitStatus, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(statusJSON),
		},
	}, nil
}

// Helper functions

// detectLanguage detects programming language from file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".mjs":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".c":
		return "c"
	case ".h", ".hpp":
		return "c_header"
	default:
		return "unknown"
	}
}

// detectMIMEType detects MIME type from file extension
func detectMIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".md":
		return "text/markdown"
	default:
		return "text/plain"
	}
}

// buildDirectoryTree builds a text representation of directory structure
func (cfs *CodeForgeServer) buildDirectoryTree(rootPath string, maxDepth int) (string, error) {
	var result strings.Builder

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth relative to root
		relPath, _ := filepath.Rel(rootPath, path)
		depth := strings.Count(relPath, string(filepath.Separator))

		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create indentation
		indent := strings.Repeat("  ", depth)

		if d.IsDir() {
			result.WriteString(fmt.Sprintf("%s%s/\n", indent, d.Name()))
		} else {
			result.WriteString(fmt.Sprintf("%s%s\n", indent, d.Name()))
		}

		return nil
	})

	return result.String(), err
}
