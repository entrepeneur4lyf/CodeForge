package theme

import "github.com/charmbracelet/lipgloss"

// CatppuccinTheme implements the popular Catppuccin color scheme
type CatppuccinTheme struct {
	BaseTheme
}

// NewCatppuccinTheme creates a new Catppuccin theme
func NewCatppuccinTheme() Theme {
	return &CatppuccinTheme{
		BaseTheme: BaseTheme{name: "Catppuccin"},
	}
}

func (t *CatppuccinTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#CBA6F7") // Mauve
}

func (t *CatppuccinTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#89B4FA") // Blue
}

func (t *CatppuccinTheme) Background() lipgloss.Color {
	return lipgloss.Color("#1E1E2E") // Base
}

func (t *CatppuccinTheme) BackgroundDarker() lipgloss.Color {
	return lipgloss.Color("#181825") // Mantle
}

func (t *CatppuccinTheme) Text() lipgloss.Color {
	return lipgloss.Color("#CDD6F4") // Text
}

func (t *CatppuccinTheme) TextMuted() lipgloss.Color {
	return lipgloss.Color("#6C7086") // Overlay1
}

func (t *CatppuccinTheme) Border() lipgloss.Color {
	return lipgloss.Color("#313244") // Surface0
}

func (t *CatppuccinTheme) BorderFocused() lipgloss.Color {
	return lipgloss.Color("#CBA6F7") // Mauve
}

func (t *CatppuccinTheme) Success() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CatppuccinTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CatppuccinTheme) Error() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}

func (t *CatppuccinTheme) Info() lipgloss.Color {
	return lipgloss.Color("#89B4FA") // Blue
}

func (t *CatppuccinTheme) GitAdded() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CatppuccinTheme) GitModified() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CatppuccinTheme) GitDeleted() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}

func (t *CatppuccinTheme) GitUntracked() lipgloss.Color {
	return lipgloss.Color("#CBA6F7") // Mauve
}

func (t *CatppuccinTheme) LSPRunning() lipgloss.Color {
	return lipgloss.Color("#A6E3A1") // Green
}

func (t *CatppuccinTheme) LSPWarning() lipgloss.Color {
	return lipgloss.Color("#F9E2AF") // Yellow
}

func (t *CatppuccinTheme) LSPError() lipgloss.Color {
	return lipgloss.Color("#F38BA8") // Red
}
