package theme

import "github.com/charmbracelet/lipgloss"

// TokyoNightTheme implements the popular Tokyo Night color scheme
type TokyoNightTheme struct {
	BaseTheme
}

// NewTokyoNightTheme creates a new Tokyo Night theme
func NewTokyoNightTheme() Theme {
	return &TokyoNightTheme{
		BaseTheme: BaseTheme{name: "Tokyo Night"},
	}
}

func (t *TokyoNightTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#BB9AF7") // Purple
}

func (t *TokyoNightTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#7DCFFF") // Cyan
}

func (t *TokyoNightTheme) Background() lipgloss.Color {
	return lipgloss.Color("#1A1B26") // Background
}

func (t *TokyoNightTheme) BackgroundDarker() lipgloss.Color {
	return lipgloss.Color("#16161E") // Darker background
}

func (t *TokyoNightTheme) Text() lipgloss.Color {
	return lipgloss.Color("#C0CAF5") // Foreground
}

func (t *TokyoNightTheme) TextMuted() lipgloss.Color {
	return lipgloss.Color("#565F89") // Comment
}

func (t *TokyoNightTheme) Border() lipgloss.Color {
	return lipgloss.Color("#292E42") // Border
}

func (t *TokyoNightTheme) BorderFocused() lipgloss.Color {
	return lipgloss.Color("#BB9AF7") // Purple
}

func (t *TokyoNightTheme) Success() lipgloss.Color {
	return lipgloss.Color("#9ECE6A") // Green
}

func (t *TokyoNightTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#E0AF68") // Yellow
}

func (t *TokyoNightTheme) Error() lipgloss.Color {
	return lipgloss.Color("#F7768E") // Red
}

func (t *TokyoNightTheme) Info() lipgloss.Color {
	return lipgloss.Color("#7AA2F7") // Blue
}

func (t *TokyoNightTheme) GitAdded() lipgloss.Color {
	return lipgloss.Color("#9ECE6A") // Green
}

func (t *TokyoNightTheme) GitModified() lipgloss.Color {
	return lipgloss.Color("#E0AF68") // Yellow
}

func (t *TokyoNightTheme) GitDeleted() lipgloss.Color {
	return lipgloss.Color("#F7768E") // Red
}

func (t *TokyoNightTheme) GitUntracked() lipgloss.Color {
	return lipgloss.Color("#BB9AF7") // Purple
}

func (t *TokyoNightTheme) LSPRunning() lipgloss.Color {
	return lipgloss.Color("#9ECE6A") // Green
}

func (t *TokyoNightTheme) LSPWarning() lipgloss.Color {
	return lipgloss.Color("#E0AF68") // Yellow
}

func (t *TokyoNightTheme) LSPError() lipgloss.Color {
	return lipgloss.Color("#F7768E") // Red
}
