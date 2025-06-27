package splash

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SplashCompleteMsg indicates the splash screen should close
type SplashCompleteMsg struct{}

// Model represents the splash screen
type Model struct {
	width  int
	height int
	dots   int
}

// New creates a new splash screen model
func New() Model {
	return Model{
		dots: 0,
	}
}

// Init initializes the splash screen
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return tickMsg{}
		}),
		// Auto-complete after 3 seconds to ensure smooth transition
		tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return SplashCompleteMsg{}
		}),
	)
}

type tickMsg struct{}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.dots = (m.dots + 1) % 4
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return tickMsg{}
		})

	case SplashCompleteMsg:
		return m, nil
	}

	return m, nil
}

// View renders the splash screen
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Create loading dots
	dots := ""
	for i := 0; i < m.dots; i++ {
		dots += "."
	}
	for i := m.dots; i < 3; i++ {
		dots += " "
	}

	logo := []string{
		" ██████╗ ██████╗ ██████╗ ███████╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗",
		"██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝",
		"██║     ██║   ██║██║  ██║█████╗  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗  ",
		"██║     ██║   ██║██║  ██║██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  ",
		"╚██████╗╚██████╔╝██████╔╝███████╗██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗",
		" ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝",
	}

	subtitle := "AI-Powered Code Assistant"
	loading := fmt.Sprintf("Loading%s", dots)

	// Styles
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Align(lipgloss.Center)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Align(lipgloss.Center)

	// Render logo lines
	logoLines := make([]string, len(logo))
	for i, line := range logo {
		logoLines[i] = logoStyle.Render(line)
	}

	// Create content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, logoLines...),
		"",
		subtitleStyle.Render(subtitle),
		"",
		"",
		loadingStyle.Render(loading),
	)

	// Center the content
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
