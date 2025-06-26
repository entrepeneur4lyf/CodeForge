package cmd

import (
	"fmt"
	"os"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	workingDir string
)

var rootCmd = &cobra.Command{
	Use:   "codeforge",
	Short: "AI-powered coding assistant with multi-language support",
	Long: `CodeForge is a next-generation AI coding assistant that combines the best features
from leading tools to provide comprehensive development support.

Features:
- Multi-provider LLM support (OpenAI, Claude, Gemini, Groq, and more)
- Universal build system for 7+ programming languages
- Intelligent error detection and AI-powered fixing
- Advanced terminal UI with code intelligence
- Vector-based semantic code search
- Real-time collaboration and team features`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		cfg, err := config.Load(workingDir, debug)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Initialize LLM manager
		if err := llm.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize LLM providers: %w", err)
		}

		// Initialize embedding service
		if err := embeddings.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize embedding service: %w", err)
		}

		// Initialize LSP manager
		if err := lsp.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize LSP clients: %w", err)
		}

		// Initialize MCP manager
		if err := mcp.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize MCP clients: %w", err)
		}

		// Initialize vector database
		if err := vectordb.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize vector database: %w", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVar(&workingDir, "wd", wd, "Working directory")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
