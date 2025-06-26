package theme

import "github.com/charmbracelet/lipgloss"

// CodeForgeTheme is the default theme for CodeForge
type CodeForgeTheme struct {
	BaseTheme
}

// NewCodeForgeTheme creates a new CodeForge theme
func NewCodeForgeTheme() Theme {
	return &CodeForgeTheme{
		BaseTheme: BaseTheme{name: "CodeForge"},
	}
}

func (t *CodeForgeTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#7D56F4") // Purple
}

func (t *CodeForgeTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#56B6C2") // Cyan
}

func (t *CodeForgeTheme) Background() lipgloss.Color {
	return lipgloss.Color("#1E1E2E") // Dark background
}

func (t *CodeForgeTheme) BackgroundDarker() lipgloss.Color {
	return lipgloss.Color("#181825") // Darker background
}

func (t *CodeForgeTheme) Text() lipgloss.Color {
	return lipgloss.Color("#CDD6F4") // Light text
}

func (t *CodeForgeTheme) TextMuted() lipgloss.Color {
	return lipgloss.Color("#6C7086") // Muted text
}

func (t *CodeForgeTheme) Border() lipgloss.Color {
	return lipgloss.Color("#313244") // Border
}

func (t *CodeForgeTheme) BorderFocused() lipgloss.Color {
	return lipgloss.Color("#7D56F4") // Focused border (same as primary)
}

func (t *CodeForgeTheme) Success() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CodeForgeTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CodeForgeTheme) Error() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}

func (t *CodeForgeTheme) Info() lipgloss.Color {
	return lipgloss.Color("#89B4FA") // Blue
}

func (t *CodeForgeTheme) GitAdded() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CodeForgeTheme) GitModified() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CodeForgeTheme) GitDeleted() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}

func (t *CodeForgeTheme) GitUntracked() lipgloss.Color {
	return lipgloss.Color("#CBA6F7") // Purple
}

func (t *CodeForgeTheme) LSPRunning() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CodeForgeTheme) LSPWarning() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CodeForgeTheme) LSPError() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}
