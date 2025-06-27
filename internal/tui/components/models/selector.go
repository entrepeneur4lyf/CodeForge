package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// ModelSelectorComponent provides a UI for selecting and managing models
type ModelSelectorComponent struct {
	// Core state
	modelAPI *models.ModelAPI
	list     list.Model
	width    int
	height   int
	focused  bool

	// View state
	viewMode    ViewMode
	filterMode  FilterMode
	showDetails bool

	// Data
	allModels      []models.ModelSummary
	filteredModels []models.ModelSummary
	selectedModel  *models.ModelSummary
	quickOptions   *models.QuickSelectOptions

	// Context for selection
	taskType         string
	requiredFeatures []string
	maxCost          float64
}

// ViewMode defines different ways to view models
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeGrid
	ViewModeQuickSelect
	ViewModeComparison
)

// FilterMode defines different filtering options
type FilterMode int

const (
	FilterModeAll FilterMode = iota
	FilterModeFavorites
	FilterModeProvider
	FilterModeFamily
	FilterModeFeatures
	FilterModeCost
)

// ModelItem implements list.Item for the model list
type ModelItem struct {
	models.ModelSummary
}

func (i ModelItem) FilterValue() string {
	return i.Name + " " + i.Family + " " + i.Provider + " " + strings.Join(i.Capabilities, " ")
}

func (i ModelItem) Title() string {
	title := i.Name
	if i.IsFavorite {
		title = "⭐ " + title
	}
	if i.IsRecommended {
		title = "🎯 " + title
	}
	return title
}

func (i ModelItem) Description() string {
	var parts []string

	// Provider and family
	parts = append(parts, fmt.Sprintf("%s • %s", i.Provider, i.Family))

	// Cost tier
	costIcon := map[string]string{
		"low":    "💰",
		"medium": "💳",
		"high":   "💎",
	}
	if icon, ok := costIcon[i.CostTier]; ok {
		parts = append(parts, icon+" "+i.CostTier)
	}

	// Quality tier
	qualityIcon := map[string]string{
		"basic":     "⚡",
		"good":      "🚀",
		"excellent": "✨",
	}
	if icon, ok := qualityIcon[i.QualityTier]; ok {
		parts = append(parts, icon+" "+i.QualityTier)
	}

	// Key capabilities
	if len(i.Capabilities) > 0 {
		caps := strings.Join(i.Capabilities[:min(3, len(i.Capabilities))], ", ")
		parts = append(parts, "• "+caps)
	}

	return strings.Join(parts, " ")
}

// NewModelSelectorComponent creates a new model selector component
func NewModelSelectorComponent(modelAPI *models.ModelAPI) *ModelSelectorComponent {
	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(theme.CurrentTheme().Primary()).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(theme.CurrentTheme().Secondary())

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "🤖 Model Selection"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme().Primary())
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary())
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text())

	// Add custom key bindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "favorites")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quick select")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "details")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compare")),
		}
	}

	return &ModelSelectorComponent{
		modelAPI:   modelAPI,
		list:       l,
		viewMode:   ViewModeList,
		filterMode: FilterModeAll,
		taskType:   "chat", // Default task type
	}
}

// Init implements tea.Model
func (ms *ModelSelectorComponent) Init() tea.Cmd {
	return ms.loadModels()
}

// Update implements tea.Model
func (ms *ModelSelectorComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ms.width = msg.Width
		ms.height = msg.Height
		ms.list.SetSize(msg.Width-4, msg.Height-6) // Account for borders and title

	case tea.KeyMsg:
		if !ms.focused {
			return ms, nil
		}

		switch msg.String() {
		case "f":
			// Toggle favorites filter
			if ms.filterMode == FilterModeFavorites {
				ms.filterMode = FilterModeAll
			} else {
				ms.filterMode = FilterModeFavorites
			}
			return ms, ms.applyFilter()

		case "q":
			// Switch to quick select mode
			ms.viewMode = ViewModeQuickSelect
			return ms, ms.loadQuickOptions()

		case "d":
			// Toggle details view
			ms.showDetails = !ms.showDetails

		case "c":
			// Switch to comparison mode
			ms.viewMode = ViewModeComparison

		case "enter":
			// Select current model
			if item, ok := ms.list.SelectedItem().(ModelItem); ok {
				ms.selectedModel = &item.ModelSummary
				return ms, ms.selectModel(item.ModelSummary)
			}

		case "esc":
			// Return to list mode
			ms.viewMode = ViewModeList
			ms.showDetails = false
		}

	case ModelsLoadedMsg:
		ms.allModels = msg.Models
		return ms, ms.applyFilter()

	case QuickOptionsLoadedMsg:
		ms.quickOptions = msg.Options

	case ModelSelectedMsg:
		// Model was selected, could trigger other actions
		return ms, nil
	}

	// Update list
	var cmd tea.Cmd
	ms.list, cmd = ms.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return ms, tea.Batch(cmds...)
}

// View implements tea.Model
func (ms *ModelSelectorComponent) View() string {
	if ms.width == 0 || ms.height == 0 {
		return "Loading model selector..."
	}

	switch ms.viewMode {
	case ViewModeQuickSelect:
		return ms.renderQuickSelect()
	case ViewModeComparison:
		return ms.renderComparison()
	default:
		return ms.renderList()
	}
}

// SetFocus sets the focus state
func (ms *ModelSelectorComponent) SetFocus(focused bool) {
	ms.focused = focused
}

// SetTaskContext sets the context for model selection
func (ms *ModelSelectorComponent) SetTaskContext(taskType string, features []string, maxCost float64) {
	ms.taskType = taskType
	ms.requiredFeatures = features
	ms.maxCost = maxCost
}

// GetSelectedModel returns the currently selected model
func (ms *ModelSelectorComponent) GetSelectedModel() *models.ModelSummary {
	return ms.selectedModel
}

// Custom messages
type ModelsLoadedMsg struct {
	Models []models.ModelSummary
}

type QuickOptionsLoadedMsg struct {
	Options *models.QuickSelectOptions
}

type ModelSelectedMsg struct {
	Model models.ModelSummary
}

// Commands

func (ms *ModelSelectorComponent) loadModels() tea.Cmd {
	return func() tea.Msg {
		options := models.ModelListOptions{
			SortBy:    "name",
			SortOrder: "asc",
		}

		modelList, err := ms.modelAPI.ListModels(options)
		if err != nil {
			// Handle error - for now just return empty list
			return ModelsLoadedMsg{Models: []models.ModelSummary{}}
		}

		return ModelsLoadedMsg{Models: modelList}
	}
}

func (ms *ModelSelectorComponent) loadQuickOptions() tea.Cmd {
	return func() tea.Msg {
		req := models.SelectionRequest{
			TaskType:         ms.taskType,
			RequiredFeatures: ms.requiredFeatures,
			MaxCost:          ms.maxCost,
			AllowFallback:    true,
		}

		options, err := ms.modelAPI.GetQuickSelectOptions(context.Background(), req)
		if err != nil {
			// Handle error
			return QuickOptionsLoadedMsg{Options: nil}
		}

		return QuickOptionsLoadedMsg{Options: options}
	}
}

func (ms *ModelSelectorComponent) selectModel(model models.ModelSummary) tea.Cmd {
	return func() tea.Msg {
		return ModelSelectedMsg{Model: model}
	}
}

func (ms *ModelSelectorComponent) applyFilter() tea.Cmd {
	return func() tea.Msg {
		var filtered []models.ModelSummary

		switch ms.filterMode {
		case FilterModeFavorites:
			for _, model := range ms.allModels {
				if model.IsFavorite {
					filtered = append(filtered, model)
				}
			}
		default:
			filtered = ms.allModels
		}

		// Convert to list items
		items := make([]list.Item, len(filtered))
		for i, model := range filtered {
			items[i] = ModelItem{ModelSummary: model}
		}

		ms.filteredModels = filtered
		ms.list.SetItems(items)

		return nil
	}
}

// Render methods

func (ms *ModelSelectorComponent) renderList() string {
	content := ms.list.View()

	if ms.showDetails && len(ms.filteredModels) > 0 {
		// Show details panel for selected model
		selectedIdx := ms.list.Index()
		if selectedIdx < len(ms.filteredModels) {
			details := ms.renderModelDetails(ms.filteredModels[selectedIdx])

			// Split view: list on left, details on right
			listWidth := ms.width / 2
			detailsWidth := ms.width - listWidth - 2

			ms.list.SetSize(listWidth-2, ms.height-4)
			content = ms.list.View()

			detailsPanel := lipgloss.NewStyle().
				Width(detailsWidth).
				Height(ms.height - 4).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.CurrentTheme().Secondary()).
				Padding(1).
				Render(details)

			content = lipgloss.JoinHorizontal(lipgloss.Top, content, detailsPanel)
		}
	}

	// Add filter indicator
	filterText := ms.getFilterText()
	if filterText != "" {
		header := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().Secondary()).
			Italic(true).
			Render("Filter: " + filterText)
		content = lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	return content
}

func (ms *ModelSelectorComponent) renderQuickSelect() string {
	if ms.quickOptions == nil {
		return "Loading quick options..."
	}

	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme().Primary()).
		Render("🚀 Quick Model Selection")
	sections = append(sections, title)

	// Task context
	context := fmt.Sprintf("Task: %s", ms.taskType)
	if len(ms.requiredFeatures) > 0 {
		context += fmt.Sprintf(" • Features: %s", strings.Join(ms.requiredFeatures, ", "))
	}
	if ms.maxCost > 0 {
		context += fmt.Sprintf(" • Max Cost: $%.2f/M tokens", ms.maxCost)
	}

	contextStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true)
	sections = append(sections, contextStyle.Render(context))
	sections = append(sections, "")

	// Quick options
	options := []struct {
		name  string
		icon  string
		model *models.ModelRecommendation
		desc  string
	}{
		{"Recommended", "🎯", ms.quickOptions.Recommended, "Best overall choice"},
		{"Fastest", "⚡", ms.quickOptions.Fastest, "Optimized for speed"},
		{"Cheapest", "💰", ms.quickOptions.Cheapest, "Most cost-effective"},
		{"Best Quality", "✨", ms.quickOptions.BestQuality, "Highest quality output"},
		{"Balanced", "⚖️", ms.quickOptions.Balanced, "Good balance of all factors"},
	}

	for i, opt := range options {
		if opt.model == nil {
			continue
		}

		// Option header
		header := fmt.Sprintf("%s %s", opt.icon, opt.name)
		if i == 0 { // Highlight recommended
			header = lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.CurrentTheme().Primary()).
				Render(header)
		} else {
			header = lipgloss.NewStyle().
				Bold(true).
				Render(header)
		}

		// Model info
		modelInfo := fmt.Sprintf("%s (%s) - %s",
			opt.model.Model.Name,
			opt.model.Provider,
			opt.desc)

		// Score and reasoning
		score := fmt.Sprintf("Score: %.1f/100", opt.model.Score)
		reasoning := ""
		if len(opt.model.Reasoning) > 0 {
			reasoning = "• " + strings.Join(opt.model.Reasoning, " • ")
		}

		optionBlock := lipgloss.JoinVertical(lipgloss.Left,
			header,
			lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text()).Render(modelInfo),
			lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary()).Render(score),
			lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary()).Render(reasoning),
		)

		sections = append(sections, optionBlock)
		sections = append(sections, "")
	}

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Render("Press 1-5 to select, Esc to return to list")
	sections = append(sections, instructions)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (ms *ModelSelectorComponent) renderComparison() string {
	// TODO: Implement comparison view
	return "Comparison view - Coming soon!"
}

func (ms *ModelSelectorComponent) renderModelDetails(model models.ModelSummary) string {
	var sections []string

	// Model name and basic info
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme().Primary()).
		Render(model.Name)
	sections = append(sections, title)

	// Provider and family
	info := fmt.Sprintf("Provider: %s\nFamily: %s", model.Provider, model.Family)
	sections = append(sections, info)

	// Tiers
	tiers := fmt.Sprintf("Cost: %s • Quality: %s • Speed: %s\nContext: %s",
		model.CostTier, model.QualityTier, model.SpeedTier, model.ContextSize)
	sections = append(sections, tiers)

	// Capabilities
	if len(model.Capabilities) > 0 {
		caps := "Capabilities:\n• " + strings.Join(model.Capabilities, "\n• ")
		sections = append(sections, caps)
	}

	// Tags
	if len(model.Tags) > 0 {
		tags := "Tags: " + strings.Join(model.Tags, ", ")
		sections = append(sections, tags)
	}

	// Description
	if model.Description != "" {
		sections = append(sections, model.Description)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (ms *ModelSelectorComponent) getFilterText() string {
	switch ms.filterMode {
	case FilterModeFavorites:
		return "Favorites only"
	default:
		return ""
	}
}

// Helper function - using built-in min from Go 1.21+
// func min(a, b int) int is now built-in
