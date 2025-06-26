package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/layout"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// DialogType represents the type of dialog
type DialogType int

const (
	DialogTypeInit DialogType = iota
	DialogTypeSettings
	DialogTypePermission
	DialogTypeCommand
	DialogTypeConfirm
	DialogTypeProviderSettings
)

// DialogResult represents the result of a dialog
type DialogResult struct {
	Type      DialogType
	Confirmed bool
	Values    map[string]interface{}
	Error     error
}

// DialogClosedMsg is sent when a dialog is closed
type DialogClosedMsg struct {
	Result DialogResult
}

// DialogManager manages modal dialogs
type DialogManager struct {
	currentDialog     tea.Model
	dialogType        DialogType
	visible           bool
	width             int
	height            int
	backgroundContent string
}

// NewDialogManager creates a new dialog manager
func NewDialogManager() *DialogManager {
	return &DialogManager{
		visible: false,
	}
}

// ShowInitDialog shows the project initialization dialog
func (dm *DialogManager) ShowInitDialog() tea.Cmd {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("initialize").
				Title("Initialize Project").
				Description("Initialization generates a new OpenCode.md file that contains information about your codebase. This file serves as memory for each project, you can freely add to it to help the agents be better at their job.\n\nWould you like to initialize this project?").
				Affirmative("Yes").
				Negative("No"),
		),
	).WithTheme(dm.getHuhTheme())

	dm.currentDialog = form
	dm.dialogType = DialogTypeInit
	dm.visible = true

	log.Info("Showing initialization dialog")

	return form.Init()
}

// ShowSettingsDialog shows the settings dialog
func (dm *DialogManager) ShowSettingsDialog() tea.Cmd {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("theme").
				Title("Theme").
				Description("Choose your preferred color theme").
				Options(
					huh.NewOption("CodeForge", "codeforge"),
					huh.NewOption("Catppuccin", "catppuccin"),
					huh.NewOption("Dracula", "dracula"),
					huh.NewOption("Gruvbox", "gruvbox"),
					huh.NewOption("Tokyo Night", "tokyonight"),
				),

			huh.NewSelect[string]().
				Key("provider").
				Title("LLM Provider").
				Description("Choose your AI provider").
				Options(
					huh.NewOption("OpenAI", "openai"),
					huh.NewOption("Anthropic", "anthropic"),
					huh.NewOption("Local", "local"),
				),

			huh.NewInput().
				Key("api_key").
				Title("API Key").
				Description("Enter your API key").
				Password(true).
				Placeholder("sk-..."),

			huh.NewConfirm().
				Key("auto_save").
				Title("Auto-save").
				Description("Automatically save files when modified").
				Affirmative("Yes").
				Negative("No"),
		),
	).WithTheme(dm.getHuhTheme())

	dm.currentDialog = form
	dm.dialogType = DialogTypeSettings
	dm.visible = true

	log.Info("Showing settings dialog")

	return form.Init()
}

// ShowPermissionDialog shows a permission request dialog
func (dm *DialogManager) ShowPermissionDialog(toolName, description string) tea.Cmd {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("permission").
				Title("Permission Request").
				Description("The AI wants to execute: "+toolName+"\n\n"+description+"\n\nHow would you like to proceed?").
				Options(
					huh.NewOption("Allow", "allow"),
					huh.NewOption("Allow for Session", "allow_session"),
					huh.NewOption("Deny", "deny"),
				),
		),
	).WithTheme(dm.getHuhTheme())

	dm.currentDialog = form
	dm.dialogType = DialogTypePermission
	dm.visible = true

	log.Info("Showing permission dialog", "tool", toolName)

	return form.Init()
}

// ShowCommandPalette shows the command palette
func (dm *DialogManager) ShowCommandPalette() tea.Cmd {
	commands := []huh.Option[string]{
		huh.NewOption("🔄 Refresh File Tree", "refresh"),
		huh.NewOption("🎨 Change Theme", "theme"),
		huh.NewOption("⚙️  Open Settings", "settings"),
		huh.NewOption("📊 Show Diagnostics", "diagnostics"),
		huh.NewOption("🌿 Switch Git Branch", "branch"),
		huh.NewOption("📁 Open Project", "open"),
		huh.NewOption("💾 Save All", "save-all"),
		huh.NewOption("🔍 Search Files", "search"),
		huh.NewOption("📝 New File", "new-file"),
		huh.NewOption("📂 New Folder", "new-folder"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("command").
				Title("Command Palette").
				Description("Choose a command to execute").
				Options(commands...),
		),
	).WithTheme(dm.getHuhTheme())

	dm.currentDialog = form
	dm.dialogType = DialogTypeCommand
	dm.visible = true

	log.Info("Showing command palette")

	return form.Init()
}

// ShowProviderSettingsDialog shows the provider settings dialog
func (dm *DialogManager) ShowProviderSettingsDialog() tea.Cmd {
	providerDialog := NewProviderSettingsDialog()

	dm.currentDialog = providerDialog
	dm.dialogType = DialogTypeProviderSettings
	dm.visible = true

	log.Info("Showing provider settings dialog")

	return providerDialog.Init()
}

// ShowConfirmDialog shows a simple confirmation dialog
func (dm *DialogManager) ShowConfirmDialog(title, description string) tea.Cmd {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("confirm").
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No"),
		),
	).WithTheme(dm.getHuhTheme())

	dm.currentDialog = form
	dm.dialogType = DialogTypeConfirm
	dm.visible = true

	log.Info("Showing confirm dialog", "title", title)

	return form.Init()
}

// Update implements tea.Model
func (dm *DialogManager) Update(msg tea.Msg) (*DialogManager, tea.Cmd) {
	if !dm.visible || dm.currentDialog == nil {
		return dm, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Close dialog without saving
			dm.closeDialog(false, nil)
			return dm, func() tea.Msg {
				return DialogClosedMsg{
					Result: DialogResult{
						Type:      dm.dialogType,
						Confirmed: false,
					},
				}
			}
		}

	case tea.WindowSizeMsg:
		dm.width = msg.Width
		dm.height = msg.Height
	}

	// Update the current dialog
	var cmd tea.Cmd
	dm.currentDialog, cmd = dm.currentDialog.Update(msg)

	// Check if form is complete
	if form, ok := dm.currentDialog.(*huh.Form); ok {
		if form.State == huh.StateCompleted {
			// Extract values and close dialog
			values := dm.extractFormValues(form)
			dm.closeDialog(true, values)

			return dm, func() tea.Msg {
				return DialogClosedMsg{
					Result: DialogResult{
						Type:      dm.dialogType,
						Confirmed: true,
						Values:    values,
					},
				}
			}
		}
	}

	return dm, cmd
}

// View implements tea.Model
func (dm *DialogManager) View() string {
	if !dm.visible || dm.currentDialog == nil {
		return ""
	}

	// Get the dialog content
	dialogContent := dm.currentDialog.View()

	// Style the dialog
	dialogStyle := styles.DialogStyle()
	styledDialog := dialogStyle.Render(dialogContent)

	// Calculate position for centering
	dialogWidth := lipgloss.Width(styledDialog)
	dialogHeight := lipgloss.Height(styledDialog)

	x := (dm.width - dialogWidth) / 2
	y := (dm.height - dialogHeight) / 2

	// Place overlay on background
	return layout.PlaceOverlay(x, y, styledDialog, dm.backgroundContent, true)
}

// SetBackgroundContent sets the background content for the overlay
func (dm *DialogManager) SetBackgroundContent(content string) {
	dm.backgroundContent = content
}

// SetSize sets the dimensions
func (dm *DialogManager) SetSize(width, height int) {
	dm.width = width
	dm.height = height
}

// IsVisible returns whether a dialog is currently visible
func (dm *DialogManager) IsVisible() bool {
	return dm.visible
}

// closeDialog closes the current dialog
func (dm *DialogManager) closeDialog(confirmed bool, values map[string]interface{}) {
	dm.visible = false
	dm.currentDialog = nil

	log.Debug("Dialog closed", "confirmed", confirmed, "type", dm.dialogType)
}

// extractFormValues extracts values from a completed form
func (dm *DialogManager) extractFormValues(form *huh.Form) map[string]interface{} {
	values := make(map[string]interface{})

	// This is a simplified extraction - in practice, you'd need to
	// iterate through form groups and fields to extract all values
	// For now, we'll return empty map and implement specific extraction
	// based on dialog type when needed

	return values
}

// getHuhTheme returns a Huh theme based on the current CodeForge theme
func (dm *DialogManager) getHuhTheme() *huh.Theme {
	t := theme.CurrentTheme()

	// Create a custom Huh theme based on current theme
	huhTheme := huh.ThemeCharm()

	// Customize colors to match CodeForge theme
	huhTheme.Focused.Base = lipgloss.NewStyle().BorderForeground(t.Primary())
	huhTheme.Focused.Title = lipgloss.NewStyle().Foreground(t.Primary()).Bold(true)
	huhTheme.Focused.Description = lipgloss.NewStyle().Foreground(t.Text())

	return huhTheme
}
