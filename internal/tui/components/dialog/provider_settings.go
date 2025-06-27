package dialog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
	table "github.com/ionut-t/gotable"
)

// ProviderInfo represents a single AI provider
type ProviderInfo struct {
	Name       string
	Status     string // "Active", "Error", "Setup", "Limit"
	ModelCount int
	APIKey     string // Masked for security
	IsDefault  bool
	IsFavorite bool
	Models     []ModelInfo
}

// ModelInfo represents a single AI model
type ModelInfo struct {
	Name       string
	Context    string
	Cost       string
	Speed      string
	Quality    int // 1-5 stars
	IsFavorite bool
	IsEnabled  bool
}

// ProviderSettingsDialog manages the provider settings interface
type ProviderSettingsDialog struct {
	providerTable    table.Model
	modelTable       table.Model
	providers        []ProviderInfo
	selectedProvider int
	focused          string // "providers" or "models"
	width            int
	height           int
}

// ProviderSelectedMsg is sent when a provider is selected
type ProviderSelectedMsg struct {
	Provider ProviderInfo
}

// ModelToggledMsg is sent when a model is toggled
type ModelToggledMsg struct {
	ProviderName string
	ModelName    string
	Field        string // "favorite" or "enabled"
	Value        bool
}

// NewProviderSettingsDialog creates a new provider settings dialog
func NewProviderSettingsDialog() *ProviderSettingsDialog {
	providers := []ProviderInfo{
		{
			Name:       "Anthropic",
			Status:     "Active",
			ModelCount: 4,
			APIKey:     "sk-ant-***",
			IsDefault:  true,
			IsFavorite: true,
			Models: []ModelInfo{
				{Name: "Claude 4 Sonnet", Context: "200K", Cost: "$15.00", Speed: "Fast", Quality: 5, IsFavorite: true, IsEnabled: true},
				{Name: "Claude 3.5 Sonnet", Context: "200K", Cost: "$3.00", Speed: "Fast", Quality: 4, IsFavorite: false, IsEnabled: true},
				{Name: "Claude 3.5 Haiku", Context: "200K", Cost: "$0.25", Speed: "Ultra", Quality: 4, IsFavorite: true, IsEnabled: true},
				{Name: "Claude 3 Opus", Context: "200K", Cost: "$75.00", Speed: "Slow", Quality: 5, IsFavorite: false, IsEnabled: false},
			},
		},
		{
			Name:       "OpenAI",
			Status:     "Active",
			ModelCount: 8,
			APIKey:     "sk-***",
			IsDefault:  false,
			IsFavorite: false,
			Models: []ModelInfo{
				{Name: "GPT-4o", Context: "128K", Cost: "$5.00", Speed: "Fast", Quality: 4, IsFavorite: false, IsEnabled: true},
				{Name: "GPT-4o Mini", Context: "128K", Cost: "$0.15", Speed: "Ultra", Quality: 3, IsFavorite: false, IsEnabled: true},
				{Name: "GPT-4 Turbo", Context: "128K", Cost: "$10.00", Speed: "Medium", Quality: 4, IsFavorite: false, IsEnabled: true},
				{Name: "GPT-3.5 Turbo", Context: "16K", Cost: "$0.50", Speed: "Ultra", Quality: 3, IsFavorite: false, IsEnabled: true},
			},
		},
		{
			Name:       "Groq",
			Status:     "Limit",
			ModelCount: 3,
			APIKey:     "gsk_***",
			IsDefault:  false,
			IsFavorite: true,
			Models: []ModelInfo{
				{Name: "Llama 3.1 70B", Context: "8K", Cost: "Free", Speed: "Ultra", Quality: 3, IsFavorite: false, IsEnabled: true},
				{Name: "Llama 3.1 8B", Context: "8K", Cost: "Free", Speed: "Ultra", Quality: 2, IsFavorite: false, IsEnabled: true},
				{Name: "Mixtral 8x7B", Context: "32K", Cost: "Free", Speed: "Ultra", Quality: 3, IsFavorite: false, IsEnabled: true},
			},
		},
		{
			Name:       "Local",
			Status:     "Setup",
			ModelCount: 2,
			APIKey:     "N/A",
			IsDefault:  false,
			IsFavorite: false,
			Models: []ModelInfo{
				{Name: "Llama 3.1 8B", Context: "8K", Cost: "Free", Speed: "Medium", Quality: 2, IsFavorite: false, IsEnabled: false},
				{Name: "CodeLlama 7B", Context: "16K", Cost: "Free", Speed: "Medium", Quality: 2, IsFavorite: false, IsEnabled: false},
			},
		},
		{
			Name:       "Ollama",
			Status:     "Error",
			ModelCount: 0,
			APIKey:     "localhost:11434",
			IsDefault:  false,
			IsFavorite: false,
			Models:     []ModelInfo{},
		},
	}

	// Create provider table
	providerTable := table.New()
	providerTable.SetHeaders([]string{"Provider", "Status", "Models", "API Key", "Default", "⭐", "Actions"})

	// Create model table
	modelTable := table.New()
	modelTable.SetHeaders([]string{"Model", "Context", "Cost/1M", "Speed", "Quality", "⭐", "Use"})

	dialog := &ProviderSettingsDialog{
		providerTable: providerTable,
		modelTable:    modelTable,
		providers:     providers,
		focused:       "providers",
	}

	dialog.updateProviderTable()
	dialog.updateModelTable()

	log.Info("Created provider settings dialog", "providers", len(providers))

	return dialog
}

// Init implements tea.Model
func (psd *ProviderSettingsDialog) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (psd *ProviderSettingsDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			// Switch focus between tables
			if psd.focused == "providers" {
				psd.focused = "models"
			} else {
				psd.focused = "providers"
			}
			log.Debug("Focus switched", "focused", psd.focused)

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if psd.focused == "providers" && psd.selectedProvider > 0 {
				psd.selectedProvider--
				psd.updateModelTable()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if psd.focused == "providers" && psd.selectedProvider < len(psd.providers)-1 {
				psd.selectedProvider++
				psd.updateModelTable()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if psd.focused == "providers" {
				// Select provider
				provider := psd.providers[psd.selectedProvider]
				log.Info("Provider selected", "provider", provider.Name)
				return psd, func() tea.Msg {
					return ProviderSelectedMsg{Provider: provider}
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("f"))):
			// Toggle favorite
			if psd.focused == "providers" {
				psd.providers[psd.selectedProvider].IsFavorite = !psd.providers[psd.selectedProvider].IsFavorite
				psd.updateProviderTable()
				log.Debug("Provider favorite toggled", "provider", psd.providers[psd.selectedProvider].Name)
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			// Set as default
			if psd.focused == "providers" {
				// Clear all defaults first
				for i := range psd.providers {
					psd.providers[i].IsDefault = false
				}
				// Set current as default
				psd.providers[psd.selectedProvider].IsDefault = true
				psd.updateProviderTable()
				log.Info("Default provider set", "provider", psd.providers[psd.selectedProvider].Name)
			}
		}

	case tea.WindowSizeMsg:
		psd.width = msg.Width
		psd.height = msg.Height
	}

	return psd, nil
}

// View implements tea.Model
func (psd *ProviderSettingsDialog) View() string {
	t := theme.CurrentTheme()

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render("⚙️ Provider Settings")

	// Provider table section
	providerTitle := lipgloss.NewStyle().
		Foreground(t.Text()).
		Bold(true).
		Padding(0, 1).
		Render("Available Providers:")

	providerTableView := psd.renderProviderTable()

	// Model table section
	selectedProviderName := "None"
	if psd.selectedProvider < len(psd.providers) {
		selectedProviderName = psd.providers[psd.selectedProvider].Name
	}

	modelTitle := lipgloss.NewStyle().
		Foreground(t.Text()).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("Models for Selected Provider: %s", selectedProviderName))

	modelTableView := psd.renderModelTable()

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Italic(true).
		Padding(0, 1)

	help := helpStyle.Render("Tab: Switch tables • ↑/↓: Navigate • F: Toggle favorite • D: Set default • Enter: Select")

	// Action buttons
	buttonStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(0, 2).
		Margin(0, 1)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		buttonStyle.Render("[+ Add Provider]"),
		buttonStyle.Render("[Test Connection]"),
		buttonStyle.Render("[Import Config]"),
		buttonStyle.Render("[Export Config]"),
	)

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		providerTitle,
		providerTableView,
		"",
		modelTitle,
		modelTableView,
		"",
		buttons,
		"",
		help,
	)

	return styles.DialogStyle(psd.width, psd.height).
		Width(psd.width - 4).
		Height(psd.height - 4).
		Render(content)
}

// renderProviderTable renders the provider table with styling
func (psd *ProviderSettingsDialog) renderProviderTable() string {
	t := theme.CurrentTheme()

	// Style the table based on focus
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(1)

	if psd.focused == "providers" {
		tableStyle = tableStyle.BorderForeground(t.BorderFocused())
	}

	return tableStyle.Render(psd.providerTable.View())
}

// renderModelTable renders the model table with styling
func (psd *ProviderSettingsDialog) renderModelTable() string {
	t := theme.CurrentTheme()

	// Style the table based on focus
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(1)

	if psd.focused == "models" {
		tableStyle = tableStyle.BorderForeground(t.BorderFocused())
	}

	return tableStyle.Render(psd.modelTable.View())
}

// updateProviderTable updates the provider table with current data
func (psd *ProviderSettingsDialog) updateProviderTable() {
	var rows [][]string

	for i, provider := range psd.providers {
		status := psd.getStatusIcon(provider.Status)
		defaultIcon := ""
		if provider.IsDefault {
			defaultIcon = "✓"
		}
		favoriteIcon := ""
		if provider.IsFavorite {
			favoriteIcon = "❤️"
		}

		action := "[Edit]"
		switch provider.Status {
		case "Setup":
			action = "[Setup]"
		case "Error":
			action = "[Fix]"
		}

		// Highlight selected row
		rowStyle := ""
		if i == psd.selectedProvider && psd.focused == "providers" {
			rowStyle = "→ "
		}

		rows = append(rows, []string{
			rowStyle + provider.Name,
			status,
			fmt.Sprintf("%d", provider.ModelCount),
			provider.APIKey,
			defaultIcon,
			favoriteIcon,
			action,
		})
	}

	psd.providerTable.SetRows(rows)
}

// updateModelTable updates the model table with current provider's models
func (psd *ProviderSettingsDialog) updateModelTable() {
	var rows [][]string

	if psd.selectedProvider >= len(psd.providers) {
		psd.modelTable.SetRows(rows)
		return
	}

	provider := psd.providers[psd.selectedProvider]

	for _, model := range provider.Models {
		favoriteIcon := ""
		if model.IsFavorite {
			favoriteIcon = "❤️"
		}

		enabledIcon := ""
		if model.IsEnabled {
			enabledIcon = "✓"
		}

		quality := strings.Repeat("★", model.Quality) + strings.Repeat("☆", 5-model.Quality)

		rows = append(rows, []string{
			model.Name,
			model.Context,
			model.Cost,
			model.Speed,
			quality,
			favoriteIcon,
			enabledIcon,
		})
	}

	psd.modelTable.SetRows(rows)
}

// getStatusIcon returns a styled status indicator
func (psd *ProviderSettingsDialog) getStatusIcon(status string) string {
	switch status {
	case "Active":
		return "✅ Active"
	case "Error":
		return "❌ Error"
	case "Setup":
		return "🔄 Setup"
	case "Limit":
		return "⚠️ Limit"
	default:
		return "❓ Unknown"
	}
}

// SetSize sets the dialog dimensions
func (psd *ProviderSettingsDialog) SetSize(width, height int) {
	psd.width = width
	psd.height = height
}

// GetSelectedProvider returns the currently selected provider
func (psd *ProviderSettingsDialog) GetSelectedProvider() ProviderInfo {
	if psd.selectedProvider < len(psd.providers) {
		return psd.providers[psd.selectedProvider]
	}
	return ProviderInfo{}
}

// GetFavoriteModels returns all favorited models across all providers
func (psd *ProviderSettingsDialog) GetFavoriteModels() []ModelInfo {
	var favorites []ModelInfo
	for _, provider := range psd.providers {
		for _, model := range provider.Models {
			if model.IsFavorite {
				favorites = append(favorites, model)
			}
		}
	}
	return favorites
}
