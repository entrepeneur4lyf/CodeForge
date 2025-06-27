package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/entrepeneur4lyf/codeforge/internal/web"
	"github.com/spf13/cobra"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface",
	Long: `Start the CodeForge web interface for browser-based interaction.

The web interface provides:
- Semantic code search with real-time results
- LSP integration for code intelligence
- MCP tool management and execution
- Real-time status monitoring
- Mobile-responsive design

Example:
  codeforge web --port 8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")

		// Initialize configuration (same as root command)
		cfg, err := config.Load(workingDir, debug)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Initialize all services
		if err := llm.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize LLM providers: %w", err)
		}

		if err := embeddings.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize embedding service: %w", err)
		}

		if err := lsp.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize LSP clients: %w", err)
		}

		// MCP server is now standalone - no initialization needed here

		if err := vectordb.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize vector database: %w", err)
		}

		// Create and start web server
		server := web.NewServer(cfg)

		// Handle graceful shutdown
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			fmt.Println("\n🛑 Shutting down web server...")
			os.Exit(0)
		}()

		fmt.Printf("🚀 CodeForge web interface ready!\n")
		fmt.Printf("📱 Open your browser to: http://localhost:%d\n", port)
		fmt.Printf("🔍 Features available:\n")
		fmt.Printf("   • Semantic code search\n")
		fmt.Printf("   • LSP code intelligence\n")
		fmt.Printf("   • MCP tool integration\n")
		fmt.Printf("   • Real-time WebSocket updates\n")
		fmt.Printf("\n💡 Press Ctrl+C to stop\n\n")

		// Start server (blocking)
		if err := server.Start(port); err != nil {
			return fmt.Errorf("web server failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	// Add flags
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
}
