package chat

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/providers"
)

// ModelSelector handles interactive model selection
type ModelSelector struct {
	providers         []ProviderInfo
	models            []ModelInfo
	openRouterFilters []OpenRouterFilter
	selectedIndex     int
	mode              SelectionMode
	selectedProvider  string
	selectedFilter    string
	favorites         *Favorites
	result            chan SelectionResult
}

type SelectionMode int

const (
	SelectingProvider SelectionMode = iota
	SelectingOpenRouterFilter
	SelectingModel
)

type ProviderInfo struct {
	Name      string
	ID        string
	Available bool
	Favorite  bool
}

// OpenRouterFilter represents a provider filter for OpenRouter
type OpenRouterFilter struct {
	Name        string
	ProviderKey string
	Description string
}

type ModelInfo struct {
	Name     string
	ID       string
	Provider string
	Favorite bool
}

type SelectionResult struct {
	Provider string
	Model    string
	Canceled bool
}

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Bold(true)

	favoriteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	availableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	unavailableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Strikethrough(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

// NewModelSelector creates a new model selector
func NewModelSelector(favorites *Favorites) *ModelSelector {
	ms := &ModelSelector{
		favorites: favorites,
		mode:      SelectingProvider,
		result:    make(chan SelectionResult, 1),
	}
	ms.loadOpenRouterFilters()
	return ms
}

// SelectModel shows the interactive model selector and returns the selected provider/model
func (ms *ModelSelector) SelectModel() (string, string, error) {
	// Load providers
	ms.loadProviders()

	// Start the TUI
	p := tea.NewProgram(ms, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return "", "", fmt.Errorf("failed to run model selector: %w", err)
	}

	// Get result
	result := <-ms.result
	if result.Canceled {
		return "", "", fmt.Errorf("selection canceled")
	}

	return result.Provider, result.Model, nil
}

// loadProviders loads available providers and checks their availability
func (ms *ModelSelector) loadProviders() {
	providerNames := []string{"anthropic", "openai", "gemini", "groq", "github", "openrouter"}

	for _, name := range providerNames {
		available := ms.isProviderAvailable(name)
		favorite := ms.favorites.IsProviderFavorite(name)

		ms.providers = append(ms.providers, ProviderInfo{
			Name:      strings.Title(name),
			ID:        name,
			Available: available,
			Favorite:  favorite,
		})
	}

	// Sort providers: favorites first, then available, then unavailable
	sort.Slice(ms.providers, func(i, j int) bool {
		if ms.providers[i].Favorite != ms.providers[j].Favorite {
			return ms.providers[i].Favorite
		}
		if ms.providers[i].Available != ms.providers[j].Available {
			return ms.providers[i].Available
		}
		return ms.providers[i].Name < ms.providers[j].Name
	})
}

// loadOpenRouterFilters loads the available provider filters for OpenRouter
func (ms *ModelSelector) loadOpenRouterFilters() {
	ms.openRouterFilters = []OpenRouterFilter{
		{Name: "🌟 All Providers", ProviderKey: "", Description: "Show models from all providers"},
		{Name: "🤖 Anthropic", ProviderKey: "anthropic", Description: "Claude models via OpenRouter"},
		{Name: "🔥 OpenAI", ProviderKey: "openai", Description: "GPT models via OpenRouter"},
		{Name: "💎 Google", ProviderKey: "google", Description: "Gemini models via OpenRouter"},
		{Name: "🦙 Meta/Llama", ProviderKey: "meta-llama", Description: "Llama models via OpenRouter"},
		{Name: "🌊 Mistral", ProviderKey: "mistralai", Description: "Mistral models via OpenRouter"},
		{Name: "🧠 DeepSeek", ProviderKey: "deepseek", Description: "DeepSeek models via OpenRouter"},
		{Name: "⚡ xAI", ProviderKey: "x-ai", Description: "Grok models via OpenRouter"},
		{Name: "🔮 Cohere", ProviderKey: "cohere", Description: "Command models via OpenRouter"},
		{Name: "🚀 Others", ProviderKey: "others", Description: "Other providers via OpenRouter"},
	}
}

// loadOpenRouterModels loads OpenRouter models filtered by provider
func (ms *ModelSelector) loadOpenRouterModels(providerFilter string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get OpenRouter models
	models, err := providers.GetTopOpenRouterModelsByRanking(ctx, "", 20)
	if err != nil {
		// Fallback to empty list on error
		ms.models = []ModelInfo{}
		return
	}

	ms.models = []ModelInfo{}
	for _, model := range models {
		// Apply provider filter
		if providerFilter == "" || providerFilter == "others" {
			// Show all models or "others" category
			if providerFilter == "others" {
				// Only show models not from major providers
				majorProviders := []string{"anthropic", "openai", "google", "meta-llama", "mistralai", "deepseek", "x-ai", "cohere"}
				isMajor := false
				for _, major := range majorProviders {
					if strings.HasPrefix(model.ID, major+"/") {
						isMajor = true
						break
					}
				}
				if isMajor {
					continue
				}
			}
		} else {
			// Filter by specific provider
			if !strings.HasPrefix(model.ID, providerFilter+"/") {
				continue
			}
		}

		ms.models = append(ms.models, ModelInfo{
			Name:     model.Name,
			ID:       model.ID,
			Provider: "openrouter",
			Favorite: ms.favorites.IsModelFavorite(model.ID),
		})
	}

	// Sort models: favorites first, then alphabetically
	sort.Slice(ms.models, func(i, j int) bool {
		if ms.models[i].Favorite != ms.models[j].Favorite {
			return ms.models[i].Favorite
		}
		return ms.models[i].Name < ms.models[j].Name
	})
}

// isProviderAvailable checks if a provider has an API key
func (ms *ModelSelector) isProviderAvailable(provider string) bool {
	switch provider {
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY") != ""
	case "openai":
		return os.Getenv("OPENAI_API_KEY") != ""
	case "gemini":
		return os.Getenv("GEMINI_API_KEY") != ""
	case "groq":
		return os.Getenv("GROQ_API_KEY") != ""
	case "github":
		return os.Getenv("GITHUB_TOKEN") != ""
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY") != ""
	default:
		return false
	}
}

// loadModels loads models for the selected provider
func (ms *ModelSelector) loadModels(providerID string) {
	ms.models = []ModelInfo{}

	// Use default models for now (model registry integration can be added later)
	ms.addDefaultModels(providerID)

	// Sort models: favorites first, then alphabetically
	sort.Slice(ms.models, func(i, j int) bool {
		if ms.models[i].Favorite != ms.models[j].Favorite {
			return ms.models[i].Favorite
		}
		return ms.models[i].Name < ms.models[j].Name
	})
}

// addDefaultModels adds default models when registry is empty
func (ms *ModelSelector) addDefaultModels(providerID string) {
	// Special handling for OpenRouter - fetch dynamic models
	if providerID == "openrouter" {
		ms.addOpenRouterModels()
		return
	}
	defaults := map[string][]ModelInfo{
		"anthropic": {
			{Name: "Claude 3.5 Sonnet", ID: "claude-3-5-sonnet-20241022", Provider: providerID},
			{Name: "Claude 3.5 Haiku", ID: "claude-3-5-haiku-20241022", Provider: providerID},
			{Name: "Claude 3 Opus", ID: "claude-3-opus-20240229", Provider: providerID},
		},
		"openai": {
			{Name: "GPT-4o (Latest)", ID: "gpt-4o-2024-08-06", Provider: providerID},
			{Name: "GPT-4o Mini (Latest)", ID: "gpt-4o-mini-2024-07-18", Provider: providerID},
			{Name: "o1 Preview", ID: "o1-preview-2024-09-12", Provider: providerID},
			{Name: "o1 Mini", ID: "o1-mini-2024-09-12", Provider: providerID},
			{Name: "ChatGPT-4o Latest", ID: "chatgpt-4o-latest", Provider: providerID},
		},
		"gemini": {
			{Name: "Gemini 2.5 Pro (Latest)", ID: "gemini-2.5-pro", Provider: providerID},
			{Name: "Gemini 2.5 Flash (Latest)", ID: "gemini-2.5-flash", Provider: providerID},
			{Name: "Gemini 1.5 Pro", ID: "gemini-1.5-pro-latest", Provider: providerID},
		},
		"groq": {
			{Name: "Llama 3.3 70B (Latest)", ID: "llama-3.3-70b-versatile", Provider: providerID},
			{Name: "Llama 3.1 70B", ID: "llama-3.1-70b-versatile", Provider: providerID},
			{Name: "Llama 3.1 8B", ID: "llama-3.1-8b-instant", Provider: providerID},
		},
		"github": {
			{Name: "GPT-4o (Latest)", ID: "gpt-4o-2024-08-06", Provider: providerID},
			{Name: "GPT-4o Mini (Latest)", ID: "gpt-4o-mini-2024-07-18", Provider: providerID},
			{Name: "o1 Preview", ID: "o1-preview-2024-09-12", Provider: providerID},
		},
		"xai": {
			{Name: "Grok 3 (Latest)", ID: "grok-3", Provider: providerID},
			{Name: "Grok 3 Mini", ID: "grok-3-mini", Provider: providerID},
		},
		"mistral": {
			{Name: "Mistral Large 2407 (Latest)", ID: "mistral-large-2407", Provider: providerID},
			{Name: "Mistral Small 3.2 24B", ID: "mistral-small-3.2-24b-instruct", Provider: providerID},
			{Name: "Magistral Medium", ID: "magistral-medium-2506", Provider: providerID},
		},
		"deepseek": {
			{Name: "DeepSeek R1 (Latest)", ID: "deepseek-r1-0528", Provider: providerID},
			{Name: "DeepSeek R1 Distill", ID: "deepseek-r1-distill-qwen-7b", Provider: providerID},
		},
		"ollama": {
			{Name: "Llama 3.1 8B", ID: "llama3.1:8b", Provider: providerID},
			{Name: "Llama 3.1 70B", ID: "llama3.1:70b", Provider: providerID},
			{Name: "Code Llama", ID: "codellama:13b", Provider: providerID},
			{Name: "Mistral 7B", ID: "mistral:7b", Provider: providerID},
			{Name: "DeepSeek Coder", ID: "deepseek-coder:6.7b", Provider: providerID},
		},
	}

	if models, exists := defaults[providerID]; exists {
		for _, model := range models {
			model.Favorite = ms.favorites.IsModelFavorite(model.ID)
			ms.models = append(ms.models, model)
		}
	}
}

// Bubble Tea interface implementation
func (ms *ModelSelector) Init() tea.Cmd {
	return nil
}

func (ms *ModelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			ms.result <- SelectionResult{Canceled: true}
			return ms, tea.Quit

		case "up", "k":
			if ms.selectedIndex > 0 {
				ms.selectedIndex--
			}

		case "down", "j":
			maxIndex := 0
			if ms.mode == SelectingProvider {
				maxIndex = len(ms.providers) - 1
			} else if ms.mode == SelectingOpenRouterFilter {
				maxIndex = len(ms.openRouterFilters) - 1
			} else {
				maxIndex = len(ms.models) - 1
			}
			if ms.selectedIndex < maxIndex {
				ms.selectedIndex++
			}

		case "enter":
			if ms.mode == SelectingProvider {
				// Select provider and load models
				if ms.selectedIndex < len(ms.providers) {
					provider := ms.providers[ms.selectedIndex]
					if provider.Available {
						if provider.ID == "openrouter" {
							// Show OpenRouter filter menu
							ms.mode = SelectingOpenRouterFilter
							ms.selectedProvider = provider.ID
							ms.selectedIndex = 0
						} else {
							// Load models directly for other providers
							ms.loadModels(provider.ID)
							ms.mode = SelectingModel
							ms.selectedIndex = 0
						}
					}
				}
			} else if ms.mode == SelectingOpenRouterFilter {
				// Select OpenRouter filter and load filtered models
				if ms.selectedIndex < len(ms.openRouterFilters) {
					filter := ms.openRouterFilters[ms.selectedIndex]
					ms.selectedFilter = filter.ProviderKey
					ms.loadOpenRouterModels(filter.ProviderKey)
					ms.mode = SelectingModel
					ms.selectedIndex = 0
				}
			} else {
				// Select model and finish
				if ms.selectedIndex < len(ms.models) {
					model := ms.models[ms.selectedIndex]
					ms.result <- SelectionResult{
						Provider: model.Provider,
						Model:    model.ID,
						Canceled: false,
					}
					return ms, tea.Quit
				}
			}

		case " ":
			// Toggle favorite
			if ms.mode == SelectingProvider {
				if ms.selectedIndex < len(ms.providers) {
					provider := &ms.providers[ms.selectedIndex]
					provider.Favorite = !provider.Favorite
					if provider.Favorite {
						ms.favorites.AddProviderFavorite(provider.ID)
					} else {
						ms.favorites.RemoveProviderFavorite(provider.ID)
					}
				}
			} else {
				if ms.selectedIndex < len(ms.models) {
					model := &ms.models[ms.selectedIndex]
					model.Favorite = !model.Favorite
					if model.Favorite {
						ms.favorites.AddModelFavorite(model.ID)
					} else {
						ms.favorites.RemoveModelFavorite(model.ID)
					}
				}
			}

		case "backspace":
			// Go back to previous selection level
			if ms.mode == SelectingModel {
				if ms.selectedProvider == "openrouter" {
					ms.mode = SelectingOpenRouterFilter
				} else {
					ms.mode = SelectingProvider
				}
				ms.selectedIndex = 0
			} else if ms.mode == SelectingOpenRouterFilter {
				ms.mode = SelectingProvider
				ms.selectedIndex = 0
			}
		}
	}

	return ms, nil
}

func (ms *ModelSelector) View() string {
	var b strings.Builder

	if ms.mode == SelectingProvider {
		b.WriteString(titleStyle.Render("🤖 Select AI Provider"))
		b.WriteString("\n\n")

		for i, provider := range ms.providers {
			line := ""

			// Add favorite indicator
			if provider.Favorite {
				line += favoriteStyle.Render("★ ")
			} else {
				line += "  "
			}

			// Add provider name with availability styling
			if provider.Available {
				line += availableStyle.Render(provider.Name)
			} else {
				line += unavailableStyle.Render(provider.Name + " (no API key)")
			}

			// Highlight selected item
			if i == ms.selectedIndex {
				line = selectedStyle.Render(" " + line + " ")
			}

			b.WriteString(line + "\n")
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • space: favorite • q: quit"))

	} else if ms.mode == SelectingOpenRouterFilter {
		b.WriteString(titleStyle.Render("🌐 OpenRouter - Select Provider Filter"))
		b.WriteString("\n\n")

		for i, filter := range ms.openRouterFilters {
			line := "  " + filter.Name

			// Highlight selected item
			if i == ms.selectedIndex {
				line = selectedStyle.Render(" " + line + " ")
			}

			b.WriteString(line + "\n")
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • backspace: back • q: quit"))

	} else {
		b.WriteString(titleStyle.Render("🎯 Select Model"))
		b.WriteString("\n\n")

		for i, model := range ms.models {
			line := ""

			// Add favorite indicator
			if model.Favorite {
				line += favoriteStyle.Render("★ ")
			} else {
				line += "  "
			}

			line += model.Name

			// Highlight selected item
			if i == ms.selectedIndex {
				line = selectedStyle.Render(" " + line + " ")
			}

			b.WriteString(line + "\n")
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • space: favorite • backspace: back • q: quit"))
	}

	return b.String()
}

// addOpenRouterModels fetches and adds OpenRouter models dynamically
func (ms *ModelSelector) addOpenRouterModels() {
	// Try to get OpenRouter API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		// Fallback to hardcoded models if no API key
		ms.addOpenRouterFallbackModels()
		return
	}

	// Fetch top 20 models by ranking
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	models, err := providers.GetTopOpenRouterModelsByRanking(ctx, apiKey, 20)
	if err != nil {
		// Fallback to hardcoded models on error
		ms.addOpenRouterFallbackModels()
		return
	}

	// Convert OpenRouter models to ModelInfo
	for _, model := range models {
		modelInfo := ModelInfo{
			Name:     model.Name,
			ID:       model.ID,
			Provider: "openrouter",
			Favorite: ms.favorites.IsModelFavorite(model.ID),
		}
		ms.models = append(ms.models, modelInfo)
	}
}

// addOpenRouterFallbackModels adds hardcoded OpenRouter models as fallback (June 2025)
func (ms *ModelSelector) addOpenRouterFallbackModels() {
	fallbackModels := []ModelInfo{
		{Name: "Claude 3.5 Sonnet (Latest)", ID: "anthropic/claude-3.5-sonnet-20241022", Provider: "openrouter"},
		{Name: "GPT-4o (Latest)", ID: "openai/gpt-4o-2024-08-06", Provider: "openrouter"},
		{Name: "GPT-4o Mini (Latest)", ID: "openai/gpt-4o-mini-2024-07-18", Provider: "openrouter"},
		{Name: "o1 Preview", ID: "openai/o1-preview-2024-09-12", Provider: "openrouter"},
		{Name: "Gemini 2.5 Pro (Latest)", ID: "google/gemini-2.5-pro", Provider: "openrouter"},
		{Name: "Gemini 2.5 Flash (Latest)", ID: "google/gemini-2.5-flash", Provider: "openrouter"},
		{Name: "Llama 3.3 70B (Latest)", ID: "meta-llama/llama-3.3-70b-instruct", Provider: "openrouter"},
		{Name: "Mistral Large 2407", ID: "mistralai/mistral-large-2407", Provider: "openrouter"},
		{Name: "DeepSeek R1 (Latest)", ID: "deepseek/deepseek-r1-0528", Provider: "openrouter"},
		{Name: "Grok 3 (Latest)", ID: "x-ai/grok-3", Provider: "openrouter"},
		{Name: "Command R+ (Latest)", ID: "cohere/command-r-plus-08-2024", Provider: "openrouter"},
		{Name: "MiniMax M1", ID: "minimax/minimax-m1", Provider: "openrouter"},
		{Name: "Inception Mercury", ID: "inception/mercury", Provider: "openrouter"},
	}

	for _, model := range fallbackModels {
		model.Favorite = ms.favorites.IsModelFavorite(model.ID)
		ms.models = append(ms.models, model)
	}
}
