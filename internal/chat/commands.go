package chat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/entrepeneur4lyf/codeforge/internal/builder"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/ml"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
)

// CommandRouter handles natural language commands and routes them to appropriate functionality
type CommandRouter struct {
	workingDir string
}

// NewCommandRouter creates a new command router
func NewCommandRouter(workingDir string) *CommandRouter {
	return &CommandRouter{
		workingDir: workingDir,
	}
}

// RouteDirectCommand handles commands that should be executed directly (build, file ops)
func (cr *CommandRouter) RouteDirectCommand(ctx context.Context, userInput string) (string, bool) {
	input := strings.ToLower(strings.TrimSpace(userInput))

	// Build commands - these are direct actions
	if cr.isBuildCommand(input) {
		return cr.handleBuildCommand(ctx, userInput)
	}

	// File operations - these are direct actions
	if cr.isFileCommand(input) {
		return cr.handleFileCommand(ctx, userInput)
	}

	// Not a direct command
	return "", false
}

// GatherContext collects relevant context using ML-powered code intelligence
func (cr *CommandRouter) GatherContext(ctx context.Context, userInput string) string {
	// Try to get ML service for intelligent context
	mlService := ml.GetService()
	if mlService != nil && mlService.IsEnabled() {
		// Use ML-powered intelligent context gathering
		context := mlService.GetIntelligentContext(ctx, userInput, 10)

		// If ML context is empty, try smart search as fallback
		if context == "" {
			context = mlService.SmartSearch(ctx, userInput)
		}

		if context != "" {
			return context
		}
	}

	// Graceful degradation - return empty context if ML is not available
	// This allows the existing chat system to work normally
	return ""
}

// Build command detection and handling
func (cr *CommandRouter) isBuildCommand(input string) bool {
	buildKeywords := []string{
		"build", "compile", "make", "cargo build", "go build", "npm run build",
		"mvn compile", "tsc", "cmake", "fix build", "build error", "compilation",
	}

	for _, keyword := range buildKeywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}
	return false
}

func (cr *CommandRouter) handleBuildCommand(ctx context.Context, userInput string) (string, bool) {
	// Execute build
	output, err := builder.Build(cr.workingDir)

	if err != nil {
		// Build failed - provide detailed error analysis
		errorOutput := string(output)
		result := fmt.Sprintf("🔨 Build failed in %s\n\n", cr.workingDir)
		result += "**Error Output:**\n```\n" + errorOutput + "\n```\n\n"

		// Try to parse and explain the error
		if errorOutput != "" {
			result += "**Analysis:**\n"
			result += cr.analyzeBuildError(errorOutput)
		}

		return result, true
	}

	// Build succeeded
	result := fmt.Sprintf("✅ Build successful in %s\n\n", cr.workingDir)
	if len(output) > 0 {
		result += "**Build Output:**\n```\n" + string(output) + "\n```"
	}

	return result, true
}

func (cr *CommandRouter) analyzeBuildError(errorOutput string) string {
	analysis := ""

	// Common error patterns
	if strings.Contains(errorOutput, "cannot find package") || strings.Contains(errorOutput, "no such file") {
		analysis += "- **Missing dependency**: The build is failing because a required package or file cannot be found.\n"
		analysis += "- **Solution**: Check your dependencies and ensure all required packages are installed.\n\n"
	}

	if strings.Contains(errorOutput, "syntax error") || strings.Contains(errorOutput, "expected") {
		analysis += "- **Syntax error**: There's a syntax error in your code.\n"
		analysis += "- **Solution**: Check the file and line number mentioned in the error.\n\n"
	}

	if strings.Contains(errorOutput, "undefined") || strings.Contains(errorOutput, "not declared") {
		analysis += "- **Undefined symbol**: A variable, function, or type is being used but not defined.\n"
		analysis += "- **Solution**: Check for typos or missing imports/declarations.\n\n"
	}

	if analysis == "" {
		analysis = "- Review the error output above for specific details about what went wrong.\n"
		analysis += "- Check the file paths and line numbers mentioned in the error.\n"
	}

	return analysis
}

// Search command detection and handling
func (cr *CommandRouter) isSearchCommand(input string) bool {
	searchKeywords := []string{
		"search", "find", "look for", "locate", "grep", "search for",
		"find code", "search code", "semantic search", "vector search",
	}

	for _, keyword := range searchKeywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}
	return false
}

func (cr *CommandRouter) handleSearchCommand(ctx context.Context, userInput string) (string, bool) {
	// Extract search query from user input
	query := cr.extractSearchQuery(userInput)
	if query == "" {
		return "❌ Could not extract search query. Please specify what you want to search for.", true
	}

	// Get embedding for the search query using the package-level function
	embedding, err := embeddings.GetEmbedding(ctx, query)
	if err != nil {
		return fmt.Sprintf("❌ Failed to generate embedding: %v", err), true
	}

	// Search vector database
	vdb := vectordb.Get()
	if vdb == nil {
		return "❌ Vector database not available", true
	}

	results, err := vdb.SearchSimilarChunks(ctx, embedding, 5, map[string]string{})
	if err != nil {
		return fmt.Sprintf("❌ Search failed: %v", err), true
	}

	if len(results) == 0 {
		return fmt.Sprintf("🔍 No results found for: %s", query), true
	}

	// Format results
	response := fmt.Sprintf("🔍 Search results for: **%s**\n\n", query)
	for i, result := range results {
		response += fmt.Sprintf("**%d. %s** (Score: %.3f)\n", i+1, result.Chunk.FilePath, result.Score)
		response += fmt.Sprintf("```%s\n%s\n```\n\n", result.Chunk.Language, result.Chunk.Content)
	}

	return response, true
}

func (cr *CommandRouter) extractSearchQuery(input string) string {
	// Remove common search prefixes
	prefixes := []string{
		"search for ", "find ", "look for ", "locate ", "search ",
		"find code ", "search code ", "semantic search ", "vector search ",
	}

	query := input
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(query), prefix) {
			query = query[len(prefix):]
			break
		}
	}

	return strings.TrimSpace(query)
}

// LSP command detection and handling
func (cr *CommandRouter) isLSPCommand(input string) bool {
	lspKeywords := []string{
		"definition", "find definition", "go to definition", "goto definition",
		"references", "find references", "find usages", "where is used",
		"hover", "type info", "symbol info", "documentation",
		"completion", "autocomplete", "code completion",
	}

	for _, keyword := range lspKeywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}
	return false
}

func (cr *CommandRouter) handleLSPCommand(ctx context.Context, userInput string) (string, bool) {
	lspManager := lsp.GetManager()
	if lspManager == nil {
		return "❌ LSP manager not available", true
	}

	// For now, provide general LSP information
	// In a full implementation, this would parse the command and execute specific LSP operations
	response := "🔧 **LSP Features Available:**\n\n"
	response += "- **Find Definition**: Locate where symbols are defined\n"
	response += "- **Find References**: Find all usages of a symbol\n"
	response += "- **Hover Information**: Get type and documentation info\n"
	response += "- **Code Completion**: Get autocomplete suggestions\n\n"
	response += "💡 **Note**: LSP features work best when you specify a file and position.\n"
	response += "Example: \"Find definition of MyFunction in main.go at line 25\""

	return response, true
}

// File command detection and handling
func (cr *CommandRouter) isFileCommand(input string) bool {
	fileKeywords := []string{
		"list files", "show files", "what files", "file tree", "directory",
		"ls", "dir", "files in", "show directory", "project structure",
	}

	for _, keyword := range fileKeywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}
	return false
}

func (cr *CommandRouter) handleFileCommand(ctx context.Context, userInput string) (string, bool) {
	// List files in the working directory
	files, err := cr.listProjectFiles(cr.workingDir)
	if err != nil {
		return fmt.Sprintf("❌ Failed to list files: %v", err), true
	}

	response := fmt.Sprintf("📁 **Project files in %s:**\n\n", cr.workingDir)
	for _, file := range files {
		response += fmt.Sprintf("- %s\n", file)
	}

	return response, true
}

func (cr *CommandRouter) listProjectFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common build/cache directories
		skipDirs := []string{"node_modules", "target", "build", "dist", ".git", "__pycache__"}
		for _, skipDir := range skipDirs {
			if info.IsDir() && info.Name() == skipDir {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() {
			// Get relative path
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				relPath = path
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}
