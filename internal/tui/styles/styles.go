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

// PanelStyle returns a modern panel style
func PanelStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#161b22"}).
		Foreground(t.Text()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#d0d7de", Dark: "#30363d"}).
		Padding(2, 3).
		Margin(1)
}

// FocusedPanelStyle returns a focused panel style with emphasis
func FocusedPanelStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#0d1117"}).
		Foreground(t.Text()).
		Border(lipgloss.ThickBorder()).
		BorderForeground(t.Primary()).
		Padding(2, 3).
		Margin(1)
}

// TopBarStyle returns the style for the top status bar (modern design)
func TopBarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#21262d"}).
		Foreground(t.Text()).
		Padding(0, 2).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(t.Border()).
		Align(lipgloss.Left)
}

// StatusBarStyle returns the style for the bottom status bar
func StatusBarStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(t.Border())
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

// DialogStyle returns the style for modal dialogs (modern design)
func DialogStyle(terminalWidth, terminalHeight int) lipgloss.Style {
	t := theme.CurrentTheme()

	// Calculate responsive width (60-80% of terminal width, min 40, max 100)
	dialogWidth := max(40, min(terminalWidth-4, int(float64(terminalWidth)*0.7)))

	return lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"}).
		Foreground(t.Text()).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(t.Primary()).
		Padding(3, 4).
		Margin(2).
		Width(dialogWidth).
		Align(lipgloss.Center)
}

// Helper functions for responsive design
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Modern header styles for better visual hierarchy
func HeaderStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1).
		Margin(0, 0, 1, 0)
}

func SubHeaderStyle() lipgloss.Style {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Bold(true).
		Padding(0, 1)
}

// Modern file browser styles
func FileItemStyle(focused bool) lipgloss.Style {
	t := theme.CurrentTheme()
	if focused {
		return lipgloss.NewStyle().
			Background(t.Primary()).
			Foreground(t.Background()).
			Padding(0, 1).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Foreground(t.Text()).
		Padding(0, 1)
}

// Modern tab styles
func TabStyle(active bool) lipgloss.Style {
	t := theme.CurrentTheme()
	if active {
		return lipgloss.NewStyle().
			Background(t.Primary()).
			Foreground(t.Background()).
			Padding(0, 2).
			Bold(true).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(t.Primary())
	}
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		BorderForeground(t.Border())
}

// ButtonStyle returns the style for buttons
func ButtonStyle(focused bool) lipgloss.Style {
	t := theme.CurrentTheme()
	if focused {
		return lipgloss.NewStyle().
			Background(t.Primary()).
			Foreground(t.Background()).
			Padding(1, 3).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary()).
			Bold(true).
			Align(lipgloss.Center)
	}
	return lipgloss.NewStyle().
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(1, 3).
		Margin(0, 1).
		Align(lipgloss.Center)
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
