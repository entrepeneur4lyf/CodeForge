package status

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/git"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// EnhancedStatusBar provides a comprehensive status bar with LSP, Git, and Model information
type EnhancedStatusBar struct {
	// Core state
	width  int
	height int

	// Data sources
	lspManager *lsp.Manager
	gitRepo    *git.Repository
	modelAPI   *models.ModelAPI

	// Current status
	lspStatus    LSPStatus
	gitStatus    *git.GitStatus
	modelStatus  ModelStatus
	systemStatus SystemStatus

	// Update intervals
	lastLSPUpdate    time.Time
	lastGitUpdate    time.Time
	lastModelUpdate  time.Time
	lastSystemUpdate time.Time

	// Configuration
	showLSP    bool
	showGit    bool
	showModel  bool
	showSystem bool
}

// LSPStatus represents the status of LSP clients
type LSPStatus struct {
	ActiveClients   int
	TotalClients    int
	CurrentLanguage string
	Status          string // "ready", "starting", "error", "none"
	ErrorCount      int
}

// ModelStatus represents the current model status
type ModelStatus struct {
	CurrentModel  string
	Provider      string
	Status        string // "ready", "loading", "error"
	TokensUsed    int
	EstimatedCost float64
	LastResponse  time.Duration
}

// SystemStatus represents system-level status
type SystemStatus struct {
	CPUUsage      float64
	MemoryUsage   float64
	DiskUsage     float64
	NetworkStatus string // "online", "offline", "limited"
}

// NewEnhancedStatusBar creates a new enhanced status bar
func NewEnhancedStatusBar(lspManager *lsp.Manager, gitRepo *git.Repository, modelAPI *models.ModelAPI) *EnhancedStatusBar {
	return &EnhancedStatusBar{
		lspManager: lspManager,
		gitRepo:    gitRepo,
		modelAPI:   modelAPI,
		showLSP:    true,
		showGit:    true,
		showModel:  true,
		showSystem: false, // Disabled by default to reduce clutter
		height:     1,
	}
}

// Init implements tea.Model
func (esb *EnhancedStatusBar) Init() tea.Cmd {
	return tea.Batch(
		esb.updateLSPStatus(),
		esb.updateGitStatus(),
		esb.updateModelStatus(),
		esb.startPeriodicUpdates(),
	)
}

// Update implements tea.Model
func (esb *EnhancedStatusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		esb.width = msg.Width

	case LSPStatusMsg:
		esb.lspStatus = msg.Status
		esb.lastLSPUpdate = time.Now()

	case GitStatusMsg:
		esb.gitStatus = msg.Status
		esb.lastGitUpdate = time.Now()

	case ModelStatusMsg:
		esb.modelStatus = msg.Status
		esb.lastModelUpdate = time.Now()

	case SystemStatusMsg:
		esb.systemStatus = msg.Status
		esb.lastSystemUpdate = time.Now()

	case PeriodicUpdateMsg:
		// Trigger periodic updates
		cmds = append(cmds, esb.updateLSPStatus())
		cmds = append(cmds, esb.updateGitStatus())
		cmds = append(cmds, esb.updateModelStatus())
		cmds = append(cmds, esb.scheduleNextUpdate())
	}

	return esb, tea.Batch(cmds...)
}

// View implements tea.Model
func (esb *EnhancedStatusBar) View() string {
	if esb.width == 0 {
		return ""
	}

	var segments []string

	// Left side: LSP and Git status
	if esb.showLSP {
		segments = append(segments, esb.renderLSPStatus())
	}

	if esb.showGit {
		segments = append(segments, esb.renderGitStatus())
	}

	// Right side: Model and System status
	var rightSegments []string

	if esb.showModel {
		rightSegments = append(rightSegments, esb.renderModelStatus())
	}

	if esb.showSystem {
		rightSegments = append(rightSegments, esb.renderSystemStatus())
	}

	// Combine left and right segments
	leftSide := strings.Join(segments, " │ ")
	rightSide := strings.Join(rightSegments, " │ ")

	// Calculate spacing
	totalContentWidth := lipgloss.Width(leftSide) + lipgloss.Width(rightSide)
	spacingWidth := esb.width - totalContentWidth - 2 // Account for padding

	if spacingWidth < 0 {
		spacingWidth = 0
	}

	spacing := strings.Repeat(" ", spacingWidth)

	// Create status bar
	statusContent := leftSide + spacing + rightSide

	// Style the status bar
	statusBar := lipgloss.NewStyle().
		Background(theme.CurrentTheme().BackgroundDarker()).
		Foreground(theme.CurrentTheme().Text()).
		Width(esb.width).
		Padding(0, 1).
		Render(statusContent)

	return statusBar
}

// SetSize sets the status bar size
func (esb *EnhancedStatusBar) SetSize(width, height int) {
	esb.width = width
	esb.height = height
}

// Toggle methods
func (esb *EnhancedStatusBar) ToggleLSP()    { esb.showLSP = !esb.showLSP }
func (esb *EnhancedStatusBar) ToggleGit()    { esb.showGit = !esb.showGit }
func (esb *EnhancedStatusBar) ToggleModel()  { esb.showModel = !esb.showModel }
func (esb *EnhancedStatusBar) ToggleSystem() { esb.showSystem = !esb.showSystem }

// Custom messages
type LSPStatusMsg struct{ Status LSPStatus }
type GitStatusMsg struct{ Status *git.GitStatus }
type ModelStatusMsg struct{ Status ModelStatus }
type SystemStatusMsg struct{ Status SystemStatus }
type PeriodicUpdateMsg struct{}

// Commands

func (esb *EnhancedStatusBar) updateLSPStatus() tea.Cmd {
	return func() tea.Msg {
		if esb.lspManager == nil {
			return LSPStatusMsg{Status: LSPStatus{Status: "none"}}
		}

		// Get LSP client information
		// Note: This assumes methods exist on the LSP manager
		// You may need to implement these methods

		status := LSPStatus{
			Status:          "ready",
			ActiveClients:   0,
			TotalClients:    0,
			CurrentLanguage: "unknown",
			ErrorCount:      0,
		}

		// Try to get actual status from LSP manager
		// This is a simplified implementation

		return LSPStatusMsg{Status: status}
	}
}

func (esb *EnhancedStatusBar) updateGitStatus() tea.Cmd {
	return func() tea.Msg {
		if esb.gitRepo == nil {
			return GitStatusMsg{Status: nil}
		}

		// Get git status
		status, err := esb.gitRepo.GetStatus(nil) // Using nil context for now
		if err != nil {
			return GitStatusMsg{Status: nil}
		}

		return GitStatusMsg{Status: status}
	}
}

func (esb *EnhancedStatusBar) updateModelStatus() tea.Cmd {
	return func() tea.Msg {
		if esb.modelAPI == nil {
			return ModelStatusMsg{Status: ModelStatus{Status: "none"}}
		}

		// Get current model preferences
		prefs := esb.modelAPI.GetPreferences()

		status := ModelStatus{
			CurrentModel:  string(prefs.DefaultModel),
			Provider:      "unknown",
			Status:        "ready",
			TokensUsed:    0,
			EstimatedCost: 0.0,
			LastResponse:  0,
		}

		// Get more detailed status if available
		// This would require additional methods on the model API

		return ModelStatusMsg{Status: status}
	}
}

func (esb *EnhancedStatusBar) startPeriodicUpdates() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return PeriodicUpdateMsg{}
	})
}

func (esb *EnhancedStatusBar) scheduleNextUpdate() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return PeriodicUpdateMsg{}
	})
}

// Render methods

func (esb *EnhancedStatusBar) renderLSPStatus() string {
	var icon, text string
	var style lipgloss.Style

	switch esb.lspStatus.Status {
	case "ready":
		icon = "🟢"
		text = fmt.Sprintf("LSP %d/%d", esb.lspStatus.ActiveClients, esb.lspStatus.TotalClients)
		if esb.lspStatus.CurrentLanguage != "" && esb.lspStatus.CurrentLanguage != "unknown" {
			text += fmt.Sprintf(" (%s)", esb.lspStatus.CurrentLanguage)
		}
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Success())

	case "starting":
		icon = "🟡"
		text = "LSP Starting..."
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Warning())

	case "error":
		icon = "🔴"
		text = fmt.Sprintf("LSP Error (%d)", esb.lspStatus.ErrorCount)
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Error())

	default:
		icon = "⚫"
		text = "LSP Off"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().TextMuted())
	}

	return style.Render(icon + " " + text)
}

func (esb *EnhancedStatusBar) renderGitStatus() string {
	if esb.gitStatus == nil {
		return lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().TextMuted()).
			Render("⚫ No Git")
	}

	var icon, text string
	var style lipgloss.Style

	// Determine status icon and color
	switch esb.gitStatus.Status {
	case "clean":
		icon = "🟢"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Success())
	case "modified", "staged":
		icon = "🟡"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Warning())
	case "untracked":
		icon = "🔵"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Info())
	default:
		icon = "🟠"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Warning())
	}

	// Build status text
	text = esb.gitStatus.Branch

	// Add change indicators
	var changes []string
	if len(esb.gitStatus.Modified) > 0 {
		changes = append(changes, fmt.Sprintf("M%d", len(esb.gitStatus.Modified)))
	}
	if len(esb.gitStatus.Staged) > 0 {
		changes = append(changes, fmt.Sprintf("S%d", len(esb.gitStatus.Staged)))
	}
	if len(esb.gitStatus.Untracked) > 0 {
		changes = append(changes, fmt.Sprintf("U%d", len(esb.gitStatus.Untracked)))
	}

	if len(changes) > 0 {
		text += " (" + strings.Join(changes, ",") + ")"
	}

	// Add ahead/behind indicators
	if esb.gitStatus.Ahead > 0 || esb.gitStatus.Behind > 0 {
		text += fmt.Sprintf(" ↑%d↓%d", esb.gitStatus.Ahead, esb.gitStatus.Behind)
	}

	return style.Render(icon + " " + text)
}

func (esb *EnhancedStatusBar) renderModelStatus() string {
	var icon, text string
	var style lipgloss.Style

	switch esb.modelStatus.Status {
	case "ready":
		icon = "🤖"
		text = esb.modelStatus.CurrentModel
		if esb.modelStatus.Provider != "" && esb.modelStatus.Provider != "unknown" {
			text += fmt.Sprintf(" (%s)", esb.modelStatus.Provider)
		}
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Success())

	case "loading":
		icon = "⏳"
		text = "Loading Model..."
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Warning())

	case "error":
		icon = "❌"
		text = "Model Error"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Error())

	default:
		icon = "⚫"
		text = "No Model"
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().TextMuted())
	}

	// Add cost information if available
	if esb.modelStatus.EstimatedCost > 0 {
		text += fmt.Sprintf(" ($%.3f)", esb.modelStatus.EstimatedCost)
	}

	return style.Render(icon + " " + text)
}

func (esb *EnhancedStatusBar) renderSystemStatus() string {
	if esb.systemStatus.CPUUsage == 0 && esb.systemStatus.MemoryUsage == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().TextMuted()).
			Render("⚫ System")
	}

	text := fmt.Sprintf("CPU %.0f%% MEM %.0f%%",
		esb.systemStatus.CPUUsage,
		esb.systemStatus.MemoryUsage)

	// Color based on usage levels
	var style lipgloss.Style
	maxUsage := max(esb.systemStatus.CPUUsage, esb.systemStatus.MemoryUsage)

	if maxUsage > 80 {
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Error())
	} else if maxUsage > 60 {
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Warning())
	} else {
		style = lipgloss.NewStyle().Foreground(theme.CurrentTheme().Success())
	}

	return style.Render("📊 " + text)
}

// Helper function
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
