package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/layout"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/util"
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

// CloseInitDialogMsg is sent when the init dialog is closed (OpenCode pattern)
type CloseInitDialogMsg struct {
	Initialize bool
}

// InitDialogCmp is a component that asks the user if they want to initialize the project (EXACT OpenCode copy)
type InitDialogCmp struct {
	width, height int
	selected      int
}

// NewInitDialogCmp creates a new InitDialogCmp.
func NewInitDialogCmp() InitDialogCmp {
	return InitDialogCmp{
		selected: 0,
	}
}

// Init implements tea.Model
func (m InitDialogCmp) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (EXACT OpenCode copy)
func (m InitDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, util.CmdHandler(CloseInitDialogMsg{Initialize: false})
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "left", "right", "h", "l"))):
			m.selected = (m.selected + 1) % 2
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, util.CmdHandler(CloseInitDialogMsg{Initialize: m.selected == 0})
		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			return m, util.CmdHandler(CloseInitDialogMsg{Initialize: true})
		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			return m, util.CmdHandler(CloseInitDialogMsg{Initialize: false})
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model
func (m InitDialogCmp) View() string {
	title := "Initialize Project"
	description := "Initialization generates a new OpenCode.md file that contains information about your codebase. This file serves as memory for each project, you can freely add to it to help the agents be better at their job."
	question := "Would you like to initialize this project?"

	// Create styles
	baseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(80)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(76)

	questionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255"))

	// Button styles
	var yesStyle, noStyle lipgloss.Style
	if m.selected == 0 { // 0 = Yes selected
		yesStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true)
		noStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)
	} else { // 1 = No selected
		yesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)
		noStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true)
	}

	// Create buttons
	yesButton := yesStyle.Render("Yes")
	noButton := noStyle.Render("No")
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButton, "  ", noButton)

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	help := helpStyle.Render("←/→ toggle • enter confirm • y Yes • n No • esc cancel")

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		descStyle.Render(description),
		"",
		questionStyle.Render(question),
		"",
		buttons,
		"",
		help,
	)

	return baseStyle.Render(content)
}

// SetSize sets the size of the component (OpenCode pattern)
func (m *InitDialogCmp) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SimpleDialog represents a simple yes/no dialog
type SimpleDialog struct {
	title       string
	description string
	options     []string
	selected    int
}

// Init implements tea.Model
func (sd *SimpleDialog) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (sd *SimpleDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyLeft:
			if sd.selected > 0 {
				sd.selected--
			}
		case tea.KeyRight:
			if sd.selected < len(sd.options)-1 {
				sd.selected++
			}
		case tea.KeyEnter:
			return sd, func() tea.Msg { return "confirm" }
		}

		// Handle character keys
		switch msg.String() {
		case "h":
			if sd.selected > 0 {
				sd.selected--
			}
		case "l":
			if sd.selected < len(sd.options)-1 {
				sd.selected++
			}
		case "y", "Y":
			sd.selected = 0 // Yes
			return sd, func() tea.Msg { return "confirm" }
		case "n", "N":
			sd.selected = 1 // No
			return sd, func() tea.Msg { return "confirm" }
		case " ": // Space bar to toggle
			if sd.selected == 0 {
				sd.selected = 1
			} else {
				sd.selected = 0
			}
		}
	}
	return sd, nil
}

// View implements tea.Model
func (sd *SimpleDialog) View() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(sd.title))
	s.WriteString("\n\n")

	// Description
	s.WriteString(sd.description)
	s.WriteString("\n\n")

	// Options with better styling
	for i, option := range sd.options {
		if i == sd.selected {
			// Selected option with background and bold
			s.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("212")).
				Bold(true).
				Padding(0, 1).
				Render("● " + option))
		} else {
			// Unselected option
			s.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("○ " + option))
		}
		if i < len(sd.options)-1 {
			s.WriteString("   ")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("←/→ h/l toggle • space toggle • enter submit • y Yes • n No"))

	return s.String()
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
	// Create a simple custom dialog instead of Huh (following OpenCode pattern)
	dialog := NewInitDialogCmp()

	dm.currentDialog = dialog
	dm.dialogType = DialogTypeInit
	dm.visible = true

	log.Info("Showing initialization dialog")

	return dialog.Init()
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

	case DialogClosedMsg:
		// Handle dialog closed messages from child dialogs
		dm.closeDialog(msg.Result.Confirmed, msg.Result.Values)
		return dm, func() tea.Msg {
			return msg // Forward the message to the parent
		}

	case string:
		// Handle messages from SimpleDialog
		if msg == "confirm" {
			if simpleDialog, ok := dm.currentDialog.(*SimpleDialog); ok {
				confirmed := simpleDialog.selected == 0 // 0 = Yes, 1 = No
				values := map[string]interface{}{
					"confirmed": confirmed,
				}
				dm.closeDialog(confirmed, values)
				return dm, func() tea.Msg {
					return DialogClosedMsg{
						Result: DialogResult{
							Type:      dm.dialogType,
							Confirmed: confirmed,
							Values:    values,
						},
					}
				}
			}
		}
	}

	// Update the current dialog
	var cmd tea.Cmd

	// Handle Huh forms properly
	if form, ok := dm.currentDialog.(*huh.Form); ok {
		updatedForm, formCmd := form.Update(msg)
		if f, ok := updatedForm.(*huh.Form); ok {
			dm.currentDialog = f
			cmd = formCmd

			// Check if form is complete
			if f.State == huh.StateCompleted {
				// Extract values and close dialog
				values := dm.extractFormValues(f)
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
	} else {
		// Handle other dialog types (like InitDialog)
		updatedDialog, dialogCmd := dm.currentDialog.Update(msg)
		dm.currentDialog = updatedDialog
		cmd = dialogCmd
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

	// Style the dialog (responsive)
	dialogStyle := styles.DialogStyle(dm.width, dm.height)
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

// CloseDialog closes the current dialog (public method)
func (dm *DialogManager) CloseDialog() {
	dm.closeDialog(false, nil)
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
