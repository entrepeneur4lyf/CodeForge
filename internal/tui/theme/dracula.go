package theme

import "github.com/charmbracelet/lipgloss"

// DraculaTheme implements the popular Dracula color scheme
type DraculaTheme struct {
	BaseTheme
}

// NewDraculaTheme creates a new Dracula theme
func NewDraculaTheme() Theme {
	return &DraculaTheme{
		BaseTheme: BaseTheme{name: "Dracula"},
	}
}

func (t *DraculaTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#BD93F9") // Purple
}

func (t *DraculaTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#8BE9FD") // Cyan
}

func (t *DraculaTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36") // Background
}

func (t *DraculaTheme) BackgroundDarker() lipgloss.Color {
	return lipgloss.Color("#21222C") // Darker background
}

func (t *DraculaTheme) Text() lipgloss.Color {
	return lipgloss.Color("#F8F8F2") // Foreground
}

func (t *DraculaTheme) TextMuted() lipgloss.Color {
	return lipgloss.Color("#6272A4") // Comment
}

func (t *DraculaTheme) Border() lipgloss.Color {
	return lipgloss.Color("#44475A") // Current line
}

func (t *DraculaTheme) BorderFocused() lipgloss.Color {
	return lipgloss.Color("#BD93F9") // Purple
}

func (t *DraculaTheme) Success() lipgloss.Color {
	return lipgloss.Color("#50FA7B") // Green
}

func (t *DraculaTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#F1FA8C") // Yellow
}

func (t *DraculaTheme) Error() lipgloss.Color {
	return lipgloss.Color("#FF5555") // Red
}

func (t *DraculaTheme) Info() lipgloss.Color {
	return lipgloss.Color("#8BE9FD") // Cyan
}

func (t *DraculaTheme) GitAdded() lipgloss.Color {
	return lipgloss.Color("#50FA7B") // Green
}

func (t *DraculaTheme) GitModified() lipgloss.Color {
	return lipgloss.Color("#F1FA8C") // Yellow
}

func (t *DraculaTheme) GitDeleted() lipgloss.Color {
	return lipgloss.Color("#FF5555") // Red
}

func (t *DraculaTheme) GitUntracked() lipgloss.Color {
	return lipgloss.Color("#BD93F9") // Purple
}

func (t *DraculaTheme) LSPRunning() lipgloss.Color {
	return lipgloss.Color("#50FA7B") // Green
}

func (t *DraculaTheme) LSPWarning() lipgloss.Color {
	return lipgloss.Color("#F1FA8C") // Yellow
}

func (t *DraculaTheme) LSPError() lipgloss.Color {
	return lipgloss.Color("#FF5555") // Red
}
