package models

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// ModelSettingsDialog provides a UI for managing model preferences
type ModelSettingsDialog struct {
	// Core state
	modelAPI *models.ModelAPI
	form     *huh.Form
	width    int
	height   int
	visible  bool

	// Current preferences
	preferences models.UserPreferences

	// Form state
	formComplete bool

	// Available options
	availableModels    []models.ModelSummary
	availableProviders []string
}

// NewModelSettingsDialog creates a new model settings dialog
func NewModelSettingsDialog(modelAPI *models.ModelAPI) *ModelSettingsDialog {
	dialog := &ModelSettingsDialog{
		modelAPI: modelAPI,
		visible:  false,
	}

	// Load current preferences
	dialog.preferences = modelAPI.GetPreferences()

	return dialog
}

// Show displays the dialog
func (msd *ModelSettingsDialog) Show() tea.Cmd {
	msd.visible = true
	return tea.Batch(
		msd.loadAvailableOptions(),
		msd.buildForm(),
	)
}

// Hide hides the dialog
func (msd *ModelSettingsDialog) Hide() {
	msd.visible = false
	msd.formComplete = false
}

// IsVisible returns whether the dialog is visible
func (msd *ModelSettingsDialog) IsVisible() bool {
	return msd.visible
}

// Init implements tea.Model
func (msd *ModelSettingsDialog) Init() tea.Cmd {
	if msd.form != nil {
		return msd.form.Init()
	}
	return nil
}

// Update implements tea.Model
func (msd *ModelSettingsDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !msd.visible {
		return msd, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		msd.width = msg.Width
		msd.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if msd.form != nil && msd.form.State == huh.StateCompleted {
				// Form is completed, save and close
				return msd, msd.savePreferences()
			} else {
				// Cancel without saving
				msd.Hide()
				return msd, nil
			}
		}

	case OptionsLoadedMsg:
		msd.availableModels = msg.Models
		msd.availableProviders = msg.Providers
		return msd, msd.buildForm()

	case PreferencesSavedMsg:
		msd.Hide()
		return msd, nil
	}

	// Update form
	if msd.form != nil {
		var cmd tea.Cmd
		formModel, cmd := msd.form.Update(msg)
		msd.form = formModel.(*huh.Form)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check if form is completed
		if msd.form.State == huh.StateCompleted && !msd.formComplete {
			msd.formComplete = true
			cmds = append(cmds, msd.savePreferences())
		}
	}

	return msd, tea.Batch(cmds...)
}

// View implements tea.Model
func (msd *ModelSettingsDialog) View() string {
	if !msd.visible {
		return ""
	}

	if msd.form == nil {
		return msd.renderLoading()
	}

	// Render form in a dialog box
	formView := msd.form.View()

	// Create dialog container
	dialogStyle := styles.DialogStyle(msd.width, msd.height).
		Width(min(80, msd.width-4)).
		Height(min(30, msd.height-4))

	// Add title and instructions
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme().Primary()).
		Align(lipgloss.Center).
		Render("⚙️  Model Preferences")

	instructions := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Align(lipgloss.Center).
		Render("Configure your model selection preferences")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		instructions,
		"",
		formView,
		"",
		lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().Secondary()).
			Italic(true).
			Render("Press Tab/Shift+Tab to navigate, Enter to select, Esc when done"),
	)

	return dialogStyle.Render(content)
}

// Custom messages
type OptionsLoadedMsg struct {
	Models    []models.ModelSummary
	Providers []string
}

type PreferencesSavedMsg struct{}

// Commands

func (msd *ModelSettingsDialog) loadAvailableOptions() tea.Cmd {
	return func() tea.Msg {
		// Load available models
		modelOptions := models.ModelListOptions{
			SortBy:    "name",
			SortOrder: "asc",
		}

		modelList, err := msd.modelAPI.ListModels(modelOptions)
		if err != nil {
			modelList = []models.ModelSummary{}
		}

		// Extract unique providers
		providerSet := make(map[string]bool)
		for _, model := range modelList {
			providerSet[model.Provider] = true
		}

		var providers []string
		for provider := range providerSet {
			providers = append(providers, provider)
		}

		return OptionsLoadedMsg{
			Models:    modelList,
			Providers: providers,
		}
	}
}

func (msd *ModelSettingsDialog) buildForm() tea.Cmd {
	return func() tea.Msg {
		// Build model options for select
		var modelOptions []huh.Option[string]
		for _, model := range msd.availableModels {
			label := fmt.Sprintf("%s (%s)", model.Name, model.Provider)
			if model.IsFavorite {
				label = "⭐ " + label
			}
			modelOptions = append(modelOptions, huh.NewOption(label, string(model.ID)))
		}

		// Build provider options
		var providerOptions []huh.Option[string]
		for _, provider := range msd.availableProviders {
			providerOptions = append(providerOptions, huh.NewOption(provider, provider))
		}

		// Create temporary variables for form values
		defaultModel := string(msd.preferences.DefaultModel)
		maxCost := fmt.Sprintf("%.2f", msd.preferences.MaxCostPerMToken)
		preferredTier := msd.preferences.PreferredTier

		// Convert provider IDs to strings
		var preferredProviders []string
		for _, p := range msd.preferences.PreferredProviders {
			preferredProviders = append(preferredProviders, string(p))
		}

		// Create form groups
		msd.form = huh.NewForm(
			// Basic Preferences Group
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Default Model").
					Description("Your preferred model for new conversations").
					Options(modelOptions...).
					Value(&defaultModel),

				huh.NewInput().
					Title("Max Cost per Million Tokens").
					Description("Maximum cost you're willing to pay (USD)").
					Value(&maxCost).
					Validate(func(s string) error {
						if _, err := strconv.ParseFloat(s, 64); err != nil {
							return fmt.Errorf("must be a valid number")
						}
						return nil
					}),

				huh.NewSelect[string]().
					Title("Preferred Tier").
					Description("Your preferred cost tier").
					Options(
						huh.NewOption("Free", "free"),
						huh.NewOption("Low Cost", "low"),
						huh.NewOption("Medium Cost", "medium"),
						huh.NewOption("High Cost", "high"),
					).
					Value(&preferredTier),
			).Title("💰 Cost Preferences"),

			// Provider Preferences Group
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Preferred Providers").
					Description("Select your preferred providers in order of preference").
					Options(providerOptions...).
					Value(&preferredProviders),
			).Title("🏢 Provider Preferences"),

			// Feature Preferences Group
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Required Features").
					Description("Features that models must support").
					Options(
						huh.NewOption("Vision/Image Analysis", "vision"),
						huh.NewOption("Function Calling/Tools", "tools"),
						huh.NewOption("Advanced Reasoning", "reasoning"),
						huh.NewOption("Streaming Responses", "streaming"),
						huh.NewOption("Code Generation", "code"),
					).
					Value(&msd.preferences.RequiredFeatures),

				huh.NewConfirm().
					Title("Auto-Select Best Model").
					Description("Automatically select the best model for each task").
					Value(&msd.preferences.AutoSelectBest),

				huh.NewConfirm().
					Title("Enable Fallback Models").
					Description("Use alternative models if preferred model is unavailable").
					Value(&msd.preferences.FallbackEnabled),

				huh.NewConfirm().
					Title("Cache Preferences").
					Description("Remember your model selection decisions").
					Value(&msd.preferences.CachePreferences),
			).Title("🎯 Selection Preferences"),
		).WithTheme(huh.ThemeCharm())

		return msd.form.Init()
	}
}

func (msd *ModelSettingsDialog) savePreferences() tea.Cmd {
	return func() tea.Msg {
		// Extract values from form
		// Note: In a real implementation, you'd need to properly extract
		// the form values. This is simplified for the example.

		// Update preferences in the model API
		err := msd.modelAPI.UpdatePreferences(msd.preferences)
		if err != nil {
			// Handle error - could show a toast notification
			return PreferencesSavedMsg{}
		}

		return PreferencesSavedMsg{}
	}
}

// Helper methods

func (msd *ModelSettingsDialog) renderLoading() string {
	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("Loading model preferences...")

	dialogStyle := styles.DialogStyle(msd.width, msd.height).
		Width(40).
		Height(10)

	return dialogStyle.Render(content)
}

// SetSize sets the dialog size
func (msd *ModelSettingsDialog) SetSize(width, height int) {
	msd.width = width
	msd.height = height
}

// Helper function - using built-in min from Go 1.21+
