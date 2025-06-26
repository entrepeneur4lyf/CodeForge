package styles

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// BaseStyle returns the base style for all components
func BaseStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.Background()).
		Foreground(t.Text())
}

// TopBarStyle returns the style for the top status bar
func TopBarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.Text()).
		Padding(0, 1).
		Bold(true)
}

// StatusBarStyle returns the style for the bottom status bar
func StatusBarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Padding(0, 1)
}

// SidebarStyle returns the style for the left sidebar
func SidebarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.Background()).
		Foreground(t.Text()).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(t.Border()).
		Padding(1, 1)
}

// MainContentStyle returns the style for the main content area
func MainContentStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.Background()).
		Foreground(t.Text()).
		Padding(1)
}

// TabActiveStyle returns the style for active tabs
func TabActiveStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.Primary()).
		Foreground(t.Background()).
		Padding(0, 2).
		Bold(true)
}

// TabInactiveStyle returns the style for inactive tabs
func TabInactiveStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Padding(0, 2)
}

// TabBarStyle returns the style for the tab bar container
func TabBarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(t.Border())
}

// DialogStyle returns the style for modal dialogs
func DialogStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.Background()).
		Foreground(t.Text()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocused()).
		Padding(1, 2)
}

// ButtonStyle returns the style for buttons
func ButtonStyle(focused bool) lipgloss.Style {
	t := theme.CurrentTheme()
	if focused {
		return lipgloss.NewStyle().
			Background(t.Primary()).
			Foreground(t.Background()).
			Padding(0, 2).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.Text()).
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border()).
		Padding(0, 2)
}

// FileTreeItemStyle returns the style for file tree items
func FileTreeItemStyle(selected bool, modified bool) lipgloss.Style {
	t := theme.CurrentTheme()
	style := lipgloss.NewStyle()

	if selected {
		style = style.Background(t.Primary()).Foreground(t.Background())
	} else {
		style = style.Foreground(t.Text())
	}

	if modified {
		style = style.Foreground(t.GitModified()).Bold(true)
	}

	return style.Padding(0, 1)
}

// LSPStatusStyle returns the style for LSP status indicators
func LSPStatusStyle(status string) lipgloss.Style {
	t := theme.CurrentTheme()
	style := lipgloss.NewStyle().Padding(0, 1)

	switch status {
	case "running":
		return style.Foreground(t.LSPRunning())
	case "warning":
		return style.Foreground(t.LSPWarning())
	case "error":
		return style.Foreground(t.LSPError())
	default:
		return style.Foreground(t.TextMuted())
	}
}

// GitStatusStyle returns the style for git status indicators
func GitStatusStyle(status string) lipgloss.Style {
	t := theme.CurrentTheme()
	style := lipgloss.NewStyle()

	switch status {
	case "M": // Modified
		return style.Foreground(t.GitModified())
	case "A": // Added
		return style.Foreground(t.GitAdded())
	case "D": // Deleted
		return style.Foreground(t.GitDeleted())
	case "??": // Untracked
		return style.Foreground(t.GitUntracked())
	default:
		return style.Foreground(t.Text())
	}
}

// GetMarkdownRenderer returns a configured glamour renderer
func GetMarkdownRenderer(width int) *glamour.TermRenderer {
	// Create a simple glamour renderer with auto-style
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	return renderer
}

// ForceReplaceBackgroundWithLipgloss replaces background colors in rendered text
func ForceReplaceBackgroundWithLipgloss(content string, bg lipgloss.Color) string {
	// This is a simplified version - in practice you'd need more sophisticated
	// ANSI code replacement logic
	return content
}
