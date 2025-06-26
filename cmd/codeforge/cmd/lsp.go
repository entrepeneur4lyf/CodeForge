package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/spf13/cobra"
)

// lspCmd represents the lsp command
var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "LSP operations for code intelligence",
	Long: `Language Server Protocol operations for advanced code intelligence including:
- Symbol search and navigation
- Go-to-definition and references
- Code completion and hover information
- Refactoring and code transformations`,
}

// symbolsCmd searches for symbols in the workspace
var symbolsCmd = &cobra.Command{
	Use:   "symbols [query]",
	Short: "Search for symbols in the workspace",
	Long:  "Search for symbols (functions, classes, variables) across the entire workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		language, _ := cmd.Flags().GetString("language")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		manager := lsp.GetManager()
		if manager == nil {
			return fmt.Errorf("LSP manager not initialized")
		}

		client := manager.GetClientForLanguage(language)
		if client == nil {
			return fmt.Errorf("no LSP client available for language: %s", language)
		}

		symbols, err := client.GetWorkspaceSymbols(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to search symbols: %w", err)
		}

		if len(symbols) == 0 {
			fmt.Printf("No symbols found for query: %s\n", query)
			return nil
		}

		fmt.Printf("Found %d symbols for '%s':\n\n", len(symbols), query)
		for _, symbol := range symbols {
			location := symbol.Location
			// Convert URI to relative path
			filePath := strings.TrimPrefix(string(location.URI), "file://")
			if relPath, err := filepath.Rel(".", filePath); err == nil {
				filePath = relPath
			}

			fmt.Printf("📍 %s (%s)\n", symbol.Name, symbol.Kind)
			fmt.Printf("   %s:%d:%d\n", filePath, location.Range.Start.Line+1, location.Range.Start.Character+1)
			if symbol.ContainerName != "" {
				fmt.Printf("   in %s\n", symbol.ContainerName)
			}
			fmt.Println()
		}

		return nil
	},
}

// definitionCmd finds the definition of a symbol
var definitionCmd = &cobra.Command{
	Use:   "definition <file> <line> <character>",
	Short: "Go to definition of symbol at position",
	Long:  "Find the definition of the symbol at the specified position in the file",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		line, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid line number: %s", args[1])
		}
		character, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid character position: %s", args[2])
		}

		// Convert to 0-based indexing (LSP uses 0-based)
		line--
		character--

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		manager := lsp.GetManager()
		if manager == nil {
			return fmt.Errorf("LSP manager not initialized")
		}

		client := manager.GetClientForFile(filePath)
		if client == nil {
			return fmt.Errorf("no LSP client available for file: %s", filePath)
		}

		// Ensure file is opened in LSP
		if err := openFileInLSP(ctx, client, filePath); err != nil {
			return fmt.Errorf("failed to open file in LSP: %w", err)
		}

		locations, err := client.GetDefinition(ctx, filePath, line, character)
		if err != nil {
			return fmt.Errorf("failed to get definition: %w", err)
		}

		if len(locations) == 0 {
			fmt.Println("No definition found at the specified position")
			return nil
		}

		fmt.Printf("Found %d definition(s):\n\n", len(locations))
		for i, location := range locations {
			filePath := strings.TrimPrefix(string(location.URI), "file://")
			if relPath, err := filepath.Rel(".", filePath); err == nil {
				filePath = relPath
			}

			fmt.Printf("%d. %s:%d:%d\n", i+1, filePath,
				location.Range.Start.Line+1, location.Range.Start.Character+1)
		}

		return nil
	},
}

// referencesCmd finds all references to a symbol
var referencesCmd = &cobra.Command{
	Use:   "references <file> <line> <character>",
	Short: "Find all references to symbol at position",
	Long:  "Find all references to the symbol at the specified position in the file",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		line, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid line number: %s", args[1])
		}
		character, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid character position: %s", args[2])
		}

		// Convert to 0-based indexing
		line--
		character--

		includeDeclaration, _ := cmd.Flags().GetBool("include-declaration")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		manager := lsp.GetManager()
		if manager == nil {
			return fmt.Errorf("LSP manager not initialized")
		}

		client := manager.GetClientForFile(filePath)
		if client == nil {
			return fmt.Errorf("no LSP client available for file: %s", filePath)
		}

		// Ensure file is opened in LSP
		if err := openFileInLSP(ctx, client, filePath); err != nil {
			return fmt.Errorf("failed to open file in LSP: %w", err)
		}

		locations, err := client.GetReferences(ctx, filePath, line, character, includeDeclaration)
		if err != nil {
			return fmt.Errorf("failed to get references: %w", err)
		}

		if len(locations) == 0 {
			fmt.Println("No references found for the symbol at the specified position")
			return nil
		}

		fmt.Printf("Found %d reference(s):\n\n", len(locations))
		for i, location := range locations {
			filePath := strings.TrimPrefix(string(location.URI), "file://")
			if relPath, err := filepath.Rel(".", filePath); err == nil {
				filePath = relPath
			}

			fmt.Printf("%d. %s:%d:%d\n", i+1, filePath,
				location.Range.Start.Line+1, location.Range.Start.Character+1)
		}

		return nil
	},
}

// openFileInLSP ensures a file is opened in the LSP server
func openFileInLSP(ctx context.Context, client *lsp.Client, filePath string) error {
	// Check if file is already open
	if client.IsFileOpen(filePath) {
		return nil
	}

	// Open file in LSP
	return client.OpenFile(ctx, filePath)
}

func init() {
	rootCmd.AddCommand(lspCmd)

	// Add subcommands
	lspCmd.AddCommand(symbolsCmd)
	lspCmd.AddCommand(definitionCmd)
	lspCmd.AddCommand(referencesCmd)

	// Add flags
	symbolsCmd.Flags().StringP("language", "l", "", "Language to search in (go, rust, python, etc.)")
	referencesCmd.Flags().BoolP("include-declaration", "d", true, "Include declaration in results")
}
