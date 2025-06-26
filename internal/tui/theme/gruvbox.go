package theme

import "github.com/charmbracelet/lipgloss"

// GruvboxTheme implements the popular Gruvbox color scheme
type GruvboxTheme struct {
	BaseTheme
}

// NewGruvboxTheme creates a new Gruvbox theme
func NewGruvboxTheme() Theme {
	return &GruvboxTheme{
		BaseTheme: BaseTheme{name: "Gruvbox"},
	}
}

func (t *GruvboxTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#D3869B") // Purple
}

func (t *GruvboxTheme) Secondary() lipgloss.Color {
	return lipgloss.Color("#8EC07C") // Aqua
}

func (t *GruvboxTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282828") // Dark0
}

func (t *GruvboxTheme) BackgroundDarker() lipgloss.Color {
	return lipgloss.Color("#1D2021") // Dark0_hard
}

func (t *GruvboxTheme) Text() lipgloss.Color {
	return lipgloss.Color("#EBDBB2") // Light1
}

func (t *GruvboxTheme) TextMuted() lipgloss.Color {
	return lipgloss.Color("#928374") // Gray
}

func (t *GruvboxTheme) Border() lipgloss.Color {
	return lipgloss.Color("#3C3836") // Dark2
}

func (t *GruvboxTheme) BorderFocused() lipgloss.Color {
	return lipgloss.Color("#D3869B") // Purple
}

func (t *GruvboxTheme) Success() lipgloss.Color {
	return lipgloss.Color("#B8BB26") // Green
}

func (t *GruvboxTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#FABD2F") // Yellow
}

func (t *GruvboxTheme) Error() lipgloss.Color {
	return lipgloss.Color("#FB4934") // Red
}

func (t *GruvboxTheme) Info() lipgloss.Color {
	return lipgloss.Color("#83A598") // Blue
}

func (t *GruvboxTheme) GitAdded() lipgloss.Color {
	return lipgloss.Color("#B8BB26") // Green
}

func (t *GruvboxTheme) GitModified() lipgloss.Color {
	return lipgloss.Color("#FABD2F") // Yellow
}

func (t *GruvboxTheme) GitDeleted() lipgloss.Color {
	return lipgloss.Color("#FB4934") // Red
}

func (t *GruvboxTheme) GitUntracked() lipgloss.Color {
	return lipgloss.Color("#D3869B") // Purple
}

func (t *GruvboxTheme) LSPRunning() lipgloss.Color {
	return lipgloss.Color("#B8BB26") // Green
}

func (t *GruvboxTheme) LSPWarning() lipgloss.Color {
	return lipgloss.Color("#FABD2F") // Yellow
}

func (t *GruvboxTheme) LSPError() lipgloss.Color {
	return lipgloss.Color("#FB4934") // Red
}
