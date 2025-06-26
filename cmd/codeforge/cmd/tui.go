package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
	"github.com/entrepeneur4lyf/codeforge/internal/tui"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the Terminal User Interface",
	Long: `Start the CodeForge Terminal User Interface (TUI) for interactive development.

The TUI provides a multi-pane interface with:
- File browser for project navigation
- Code editor with syntax highlighting and tabs
- AI chat assistant for real-time help
- Terminal output for build results and logs

Navigation:
- Tab: Switch between panes
- Ctrl+1: Focus file browser
- Ctrl+2: Focus code editor  
- Ctrl+3: Focus AI chat
- Ctrl+4: Focus terminal
- Ctrl+C: Quit

File Browser:
- ↑/↓: Navigate files
- Enter: Open file/directory
- Backspace: Go up directory

Code Editor:
- Ctrl+S: Save file
- Ctrl+W: Close tab

AI Chat:
- Enter: Send message
- Alt+Enter: New line

Example:
  codeforge tui`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration (same as other commands)
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

		if err := mcp.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize MCP clients: %w", err)
		}

		if err := vectordb.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize vector database: %w", err)
		}

		fmt.Println("🚀 Starting CodeForge TUI...")
		fmt.Println("📱 Multi-pane interface ready!")
		fmt.Println("🔍 Features available:")
		fmt.Println("   • File browser and code editor")
		fmt.Println("   • Real-time AI chat assistant")
		fmt.Println("   • Terminal output and build logs")
		fmt.Println("   • LSP code intelligence")
		fmt.Println("   • MCP tool integration")
		fmt.Println()
		fmt.Println("💡 Press Tab to switch panes, Ctrl+C to quit")
		fmt.Println()

		// Create the TUI application
		app := tui.NewApp(workingDir)

		// Start the Bubble Tea program
		p := tea.NewProgram(
			app,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		// Run the program
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running CodeForge TUI: %w", err)
		}

		fmt.Println("Thanks for using CodeForge! 🚀")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
