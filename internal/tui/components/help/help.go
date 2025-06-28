package help

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// HelpScreen provides a comprehensive help screen with all keyboard shortcuts
type HelpScreen struct {
	width   int
	height  int
	visible bool
}

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
	Category    string
}

// NewHelpScreen creates a new help screen
func NewHelpScreen() *HelpScreen {
	return &HelpScreen{
		visible: false,
	}
}

// Init implements tea.Model
func (h *HelpScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (h *HelpScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height

	case tea.KeyMsg:
		if h.visible {
			switch msg.String() {
			case "esc", "f1", "?", "q":
				h.Hide()
			}
		}
	}

	return h, nil
}

// View implements tea.Model
func (h *HelpScreen) View() string {
	if !h.visible {
		return ""
	}

	return h.renderHelp()
}

// Show displays the help screen
func (h *HelpScreen) Show() {
	h.visible = true
}

// Hide hides the help screen
func (h *HelpScreen) Hide() {
	h.visible = false
}

// IsVisible returns whether the help screen is visible
func (h *HelpScreen) IsVisible() bool {
	return h.visible
}

// SetSize sets the help screen size
func (h *HelpScreen) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// getKeyBindings returns all keyboard shortcuts organized by category
func (h *HelpScreen) getKeyBindings() map[string][]KeyBinding {
	return map[string][]KeyBinding{
		"General": {
			{"F1 / ?", "Show/Hide Help", ""},
			{"Ctrl+P", "Command Palette", ""},
			{"Ctrl+C / Ctrl+Q", "Quit Application", ""},
			{"Esc", "Close Dialog/Return", ""},
		},
		"Navigation": {
			{"Ctrl+E", "Focus File Explorer", ""},
			{"Ctrl+T", "Focus Tabs", ""},
			{"Tab", "Switch Focus/Next Tab", ""},
			{"Ctrl+1/2/3", "Switch to Tab 1/2/3", ""},
		},
		"Model Management": {
			{"Ctrl+M", "Open Model Selector", ""},
			{"Ctrl+Shift+M", "Open Model Settings", ""},
			{"F", "Toggle Favorites (in selector)", ""},
			{"Q", "Quick Select Mode", ""},
			{"D", "Toggle Details View", ""},
			{"C", "Comparison Mode", ""},
		},
		"Chat": {
			{"Enter", "Send Message", ""},
			{"Shift+Enter", "New Line", ""},
			{"Ctrl+L", "Clear Chat", ""},
		},
		"Settings": {
			{"Ctrl+Shift+P", "Provider Settings", ""},
			{"Ctrl+,", "Settings Info", ""},
		},
	}
}

// renderHelp renders the help screen content
func (h *HelpScreen) renderHelp() string {
	t := theme.CurrentTheme()

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary()).
		Align(lipgloss.Center).
		Margin(1, 0)

	title := titleStyle.Render("🔧 CodeForge - Keyboard Shortcuts")

	// Category header style
	categoryStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Info()).
		Margin(1, 0, 0, 0)

	// Key style
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Success()).
		Width(20).
		Align(lipgloss.Left)

	// Description style
	descStyle := lipgloss.NewStyle().
		Foreground(t.Text())

	// Build content sections
	var sections []string
	keyBindings := h.getKeyBindings()

	// Define category order for better organization
	categoryOrder := []string{"General", "Navigation", "Model Management", "Chat", "Settings"}

	for _, category := range categoryOrder {
		bindings := keyBindings[category]
		if len(bindings) == 0 {
			continue
		}

		// Category header
		sections = append(sections, categoryStyle.Render("▶ "+category))

		// Key bindings for this category
		for _, binding := range bindings {
			keyPart := keyStyle.Render(binding.Key)
			descPart := descStyle.Render(binding.Description)
			line := lipgloss.JoinHorizontal(lipgloss.Left, keyPart, descPart)
			sections = append(sections, "  "+line)
		}

		// Add spacing between categories
		sections = append(sections, "")
	}

	// Footer with instructions
	footerStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0)

	footer := footerStyle.Render("Press F1, ?, Esc, or Q to close this help screen")

	// Combine all content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		strings.Join(sections, "\n"),
		footer,
	)

	// Create dialog container
	dialogStyle := styles.DialogStyle(h.width, h.height).
		Width(min(80, h.width-4)).
		Height(min(35, h.height-4))

	return dialogStyle.Render(content)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
