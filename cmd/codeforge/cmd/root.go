package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/entrepeneur4lyf/codeforge/internal/chat"
	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/ml"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	workingDir string
)

var (
	quiet    bool
	model    string
	provider string
	format   string
)

var rootCmd = &cobra.Command{
	Use:   "codeforge [prompt]",
	Short: "AI-powered coding assistant",
	Long: `CodeForge is an AI coding assistant that helps with development tasks.

Usage:
  codeforge                    # Start interactive chat
  codeforge "your question"    # Get direct answer
  echo "question" | codeforge  # Pipe input

Features:
- 25+ LLM providers (OpenRouter, Anthropic, OpenAI, Google, Groq, and more)
- 300+ models with smart database caching
- Build and fix projects automatically
- Semantic code search and analysis
- LSP integration for code intelligence
- MCP tool integration`,
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	SilenceErrors:     false,
	Args:              cobra.ArbitraryArgs, // Accept any arguments
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

		// MCP server is now standalone - no initialization needed here

		// Initialize vector database
		if err := vectordb.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize vector database: %w", err)
		}

		// Initialize ML service (graceful degradation if it fails)
		if err := ml.Initialize(cfg); err != nil {
			// Don't fail the entire application if ML initialization fails
			fmt.Printf("⚠️  ML features disabled: %v\n", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Handle different input modes like Gemini CLI
		if len(args) > 0 {
			// Direct prompt mode: codeforge "question"
			prompt := strings.Join(args, " ")
			handleDirectPrompt(prompt)
		} else {
			// Check for piped input
			if hasStdinInput() {
				handlePipedInput()
			} else {
				// Interactive mode (default)
				startInteractiveMode()
			}
		}
	},
}

func init() {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	// Add flags for the new CLI pattern
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVar(&workingDir, "wd", wd, "Working directory")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode - output only the answer")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Specify the model to use")
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "", "Specify the provider (anthropic, openai, openrouter, etc.)")
	rootCmd.Flags().StringVar(&format, "format", "text", "Output format (text, json, markdown)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// handleDirectPrompt processes a direct prompt with real LLM integration
func handleDirectPrompt(prompt string) {
	// Determine model to use
	selectedModel := model
	if selectedModel == "" {
		selectedModel = chat.GetDefaultModel()
	}

	// Get API key for the model
	apiKey := chat.GetAPIKeyForModel(selectedModel)
	if apiKey == "" {
		if quiet {
			fmt.Println("Error: No API key found. Set one of the supported provider API keys.")
		} else {
			fmt.Println("❌ Error: No API key found")
			fmt.Println("Please set one of these environment variables:")
			fmt.Println("")
			fmt.Println("🌐 Multi-Provider Platforms:")
			fmt.Println("  - OPENROUTER_API_KEY (300+ models from 50+ providers)")
			fmt.Println("")
			fmt.Println("🏢 Direct Provider Keys:")
			fmt.Println("  - ANTHROPIC_API_KEY (Claude models)")
			fmt.Println("  - OPENAI_API_KEY (GPT models)")
			fmt.Println("  - GEMINI_API_KEY (Gemini models)")
			fmt.Println("  - GROQ_API_KEY (ultra-fast inference)")
			fmt.Println("")
			fmt.Println("⚡ Additional Providers:")
			fmt.Println("  - TOGETHER_API_KEY (Together AI)")
			fmt.Println("  - FIREWORKS_API_KEY (Fireworks AI)")
			fmt.Println("  - DEEPSEEK_API_KEY (DeepSeek)")
			fmt.Println("  - COHERE_API_KEY (Cohere)")
			fmt.Println("  - MISTRAL_API_KEY (Mistral AI)")
			fmt.Println("  - PERPLEXITY_API_KEY (Perplexity)")
			fmt.Println("  - CEREBRAS_API_KEY (Cerebras)")
			fmt.Println("  - SAMBANOVA_API_KEY (SambaNova)")
			fmt.Println("")
			fmt.Println("💡 Tip: OPENROUTER_API_KEY gives you access to the most models!")
		}
		os.Exit(1)
	}

	// Create chat session
	session, err := chat.NewChatSession(selectedModel, apiKey, provider, quiet, format)
	if err != nil {
		if quiet {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("❌ Error creating chat session: %v\n", err)
		}
		os.Exit(1)
	}

	// Process the message
	response, err := session.ProcessMessage(prompt)
	if err != nil {
		if quiet {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("❌ Error: %v\n", err)
		}
		os.Exit(1)
	}

	// In quiet mode, response is already printed during streaming
	// In non-quiet mode, we need to print it since streaming was shown
	if quiet {
		// Response was not streamed, so print it now
		fmt.Println(response)
	}
}

func hasStdinInput() bool {
	// Check if stdin is not a terminal (pipe or redirect)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// If stdin is not a character device, it's piped or redirected
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func handlePipedInput() {
	fmt.Println("Reading from stdin...")

	// Read all input from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading stdin: %v\n", err)
		return
	}

	if len(lines) == 0 {
		fmt.Println("No input received from stdin")
		return
	}

	// Join all lines into a single prompt
	prompt := strings.Join(lines, "\n")

	// Handle as direct prompt
	handleDirectPrompt(prompt)
}

func startInteractiveMode() {
	// Determine model to use
	selectedModel := model
	if selectedModel == "" {
		selectedModel = chat.GetDefaultModel()
	}

	// Get API key for the model
	apiKey := chat.GetAPIKeyForModel(selectedModel)
	if apiKey == "" {
		fmt.Println("❌ Error: No API key found")
		fmt.Println("Please set one of these environment variables:")
		fmt.Println("")
		fmt.Println("🌐 Multi-Provider Platforms:")
		fmt.Println("  - OPENROUTER_API_KEY (300+ models from 50+ providers)")
		fmt.Println("")
		fmt.Println("🏢 Direct Provider Keys:")
		fmt.Println("  - ANTHROPIC_API_KEY (Claude models)")
		fmt.Println("  - OPENAI_API_KEY (GPT models)")
		fmt.Println("  - GEMINI_API_KEY (Gemini models)")
		fmt.Println("  - GROQ_API_KEY (ultra-fast inference)")
		fmt.Println("")
		fmt.Println("⚡ Additional Providers:")
		fmt.Println("  - TOGETHER_API_KEY, FIREWORKS_API_KEY, DEEPSEEK_API_KEY")
		fmt.Println("  - COHERE_API_KEY, MISTRAL_API_KEY, PERPLEXITY_API_KEY")
		fmt.Println("  - CEREBRAS_API_KEY, SAMBANOVA_API_KEY")
		fmt.Println("")
		fmt.Println("💡 Tip: OPENROUTER_API_KEY gives you access to the most models!")
		os.Exit(1)
	}

	// Create chat session
	session, err := chat.NewChatSession(selectedModel, apiKey, provider, quiet, format)
	if err != nil {
		fmt.Printf("❌ Error creating chat session: %v\n", err)
		os.Exit(1)
	}

	// Start interactive chat
	if err := session.StartInteractive(); err != nil {
		fmt.Printf("❌ Error in interactive mode: %v\n", err)
		os.Exit(1)
	}
}

// init sets up signal handling for graceful shutdown
func init() {
	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🔄 Shutting down gracefully...")

		// Shutdown ML service
		ml.Shutdown()

		os.Exit(0)
	}()
}
