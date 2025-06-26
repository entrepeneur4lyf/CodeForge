package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the TUI
type Theme interface {
	Name() string
	
	// Core colors
	Primary() lipgloss.Color
	Secondary() lipgloss.Color
	Background() lipgloss.Color
	BackgroundDarker() lipgloss.Color
	Text() lipgloss.Color
	TextMuted() lipgloss.Color
	
	// UI elements
	Border() lipgloss.Color
	BorderFocused() lipgloss.Color
	
	// Status colors
	Success() lipgloss.Color
	Warning() lipgloss.Color
	Error() lipgloss.Color
	Info() lipgloss.Color
	
	// Git colors
	GitAdded() lipgloss.Color
	GitModified() lipgloss.Color
	GitDeleted() lipgloss.Color
	GitUntracked() lipgloss.Color
	
	// LSP colors
	LSPRunning() lipgloss.Color
	LSPWarning() lipgloss.Color
	LSPError() lipgloss.Color
}

// BaseTheme provides a base implementation
type BaseTheme struct {
	name string
}

func (t BaseTheme) Name() string {
	return t.name
}

// Current theme instance
var currentTheme Theme = NewCodeForgeTheme()

// CurrentTheme returns the currently active theme
func CurrentTheme() Theme {
	return currentTheme
}

// SetTheme changes the active theme
func SetTheme(theme Theme) {
	currentTheme = theme
}

// Available themes
var availableThemes = map[string]func() Theme{
	"codeforge":  NewCodeForgeTheme,
	"catppuccin": NewCatppuccinTheme,
	"dracula":    NewDraculaTheme,
	"gruvbox":    NewGruvboxTheme,
	"tokyonight": NewTokyoNightTheme,
}

// GetAvailableThemes returns a list of available theme names
func GetAvailableThemes() []string {
	themes := make([]string, 0, len(availableThemes))
	for name := range availableThemes {
		themes = append(themes, name)
	}
	return themes
}

// LoadTheme loads a theme by name
func LoadTheme(name string) Theme {
	if factory, exists := availableThemes[name]; exists {
		return factory()
	}
	return NewCodeForgeTheme() // fallback
}
