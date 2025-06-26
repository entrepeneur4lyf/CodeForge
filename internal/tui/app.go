package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/animation"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/chat"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/dialog"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/filetree"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/tabs"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/layout"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// AppModel represents the main application model
type AppModel struct {
	// Core components
	fileTree      *filetree.FileTreeModel
	tabManager    *tabs.TabManager
	chatModel     *chat.ChatModel
	dialogManager *dialog.DialogManager
	animManager   *animation.Manager

	// Layout
	layout *layout.Layout
	width  int
	height int

	// State
	focused     string // "filetree", "tabs", "dialog"
	projectPath string
	initialized bool

	// Status
	lspStatus      map[string]string
	gitStatus      map[string]string
	vectorDBStatus string
	currentBranch  string
}

// NewApp creates a new application
func NewApp(projectPath string) *AppModel {
	// Initialize logging
	log.SetLevel(log.DebugLevel)
	log.Info("Starting CodeForge", "project", projectPath)

	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		log.Error("Failed to get absolute path", "error", err)
		absPath = projectPath
	}

	// Create components
	fileTree := filetree.NewFileTreeModel(absPath)
	chatModel := chat.NewChatModel()
	dialogManager := dialog.NewDialogManager()
	tabManager := tabs.NewTabManager()
	animManager := animation.NewManager()

	// Add tabs
	tabManager.AddTab("chat", "💬 Chat", chatModel)
	tabManager.AddTab("history", "📜 History", nil)    // TODO: Implement history
	tabManager.AddTab("viewer", "📄 File Viewer", nil) // TODO: Implement file viewer

	// Create layout
	appLayout := layout.NewLayout()

	app := &AppModel{
		fileTree:       fileTree,
		tabManager:     tabManager,
		chatModel:      chatModel,
		dialogManager:  dialogManager,
		animManager:    animManager,
		layout:         appLayout,
		projectPath:    absPath,
		focused:        "tabs",
		lspStatus:      make(map[string]string),
		gitStatus:      make(map[string]string),
		vectorDBStatus: "Ready",
		currentBranch:  "main",
	}

	return app
}

// Init implements tea.Model
func (app *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize components
	cmds = append(cmds, app.fileTree.Init())
	cmds = append(cmds, app.tabManager.Init())
	cmds = append(cmds, app.chatModel.Init())

	// Show initialization dialog if not initialized
	if !app.initialized {
		cmds = append(cmds, app.dialogManager.ShowInitDialog())
		app.focused = "dialog"
	}

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (app *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global shortcuts
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			log.Info("Quitting CodeForge")
			return app, tea.Quit
		case "ctrl+p":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.dialogManager.ShowCommandPalette())
				app.focused = "dialog"
			}
		case "ctrl+comma":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.dialogManager.ShowSettingsDialog())
				app.focused = "dialog"
			}
		case "ctrl+shift+p":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.dialogManager.ShowProviderSettingsDialog())
				app.focused = "dialog"
			}
		case "ctrl+e":
			if !app.dialogManager.IsVisible() {
				app.focused = "filetree"
				app.fileTree.SetFocused(true)
				app.tabManager.SetFocused(false)
			}
		case "ctrl+t":
			if !app.dialogManager.IsVisible() {
				app.focused = "tabs"
				app.fileTree.SetFocused(false)
				app.tabManager.SetFocused(true)
			}
		// Global tab switching shortcuts
		case "ctrl+1":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.tabManager.SwitchToIndex(0))
			}
		case "ctrl+2":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.tabManager.SwitchToIndex(1))
			}
		case "ctrl+3":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.tabManager.SwitchToIndex(2))
			}
		case "tab":
			if !app.dialogManager.IsVisible() && app.focused == "tabs" {
				// Switch to next tab
				nextTab := (app.tabManager.GetActiveTabIndex() + 1) % app.tabManager.GetTabCount()
				cmds = append(cmds, app.tabManager.SwitchToIndex(nextTab))
			}
		case "f1":
			if !app.dialogManager.IsVisible() {
				cmds = append(cmds, app.dialogManager.ShowConfirmDialog(
					"Help",
					"CodeForge Keyboard Shortcuts:\n\nCtrl+P - Command Palette\nCtrl+, - Settings\nCtrl+E - Focus File Tree\nCtrl+T - Focus Tabs\nCtrl+1-3 - Switch Tabs\nCtrl+Q - Quit",
				))
				app.focused = "dialog"
			}
		}

	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		app.updateSizes()

	case dialog.DialogClosedMsg:
		app.handleDialogResult(msg.Result)
		app.focused = "tabs"
		app.fileTree.SetFocused(false)
		app.tabManager.SetFocused(true)

		// Start fade-out animation for dialog
		app.animManager.StartAnimation("dialog-fade", animation.AnimationDialogFade, 0.0, 0.9, 0.2)

	case animation.AnimationTickMsg:
		cmd := app.animManager.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Update component animation values
		app.updateAnimationValues()

	case animation.AnimationCompleteMsg:
		log.Debug("Animation completed", "id", msg.ID, "type", msg.Type)

		// Reset animation values when complete
		switch msg.Type {
		case animation.AnimationTabSwitch:
			app.tabManager.SetAnimationValue(0.0)
		}

	case filetree.FileSelectedMsg:
		log.Info("File selected", "path", msg.Path)
		// TODO: Open file in viewer tab

	case chat.MessageSentMsg:
		log.Info("Message sent to AI", "content", msg.Content[:min(50, len(msg.Content))])
		// TODO: Send to AI service and get response

	case tabs.TabSwitchedMsg:
		log.Info("Tab switched", "tabID", msg.TabID, "index", msg.Index)
		// Start tab switch animation
		app.animManager.StartAnimation("tab-switch", animation.AnimationTabSwitch, 1.0, 0.8, 0.15)

	default:
		// Forward messages to appropriate components based on focus
		if app.dialogManager.IsVisible() {
			var cmd tea.Cmd
			app.dialogManager, cmd = app.dialogManager.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			switch app.focused {
			case "filetree":
				var cmd tea.Cmd
				app.fileTree, cmd = app.fileTree.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			case "tabs":
				var cmd tea.Cmd
				app.tabManager, cmd = app.tabManager.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return app, tea.Batch(cmds...)
}

// View implements tea.Model
func (app *AppModel) View() string {
	// Build the main layout
	app.layout.SetTopBar(app.renderTopBar())
	app.layout.SetBottomBar(app.renderBottomBar())
	app.layout.SetSidebar(app.fileTree.View())
	app.layout.SetMainContent(app.tabManager.View())

	mainView := app.layout.Render()

	// Overlay dialog if visible
	if app.dialogManager.IsVisible() {
		app.dialogManager.SetBackgroundContent(mainView)
		return app.dialogManager.View()
	}

	return mainView
}

// renderTopBar renders the top status bar
func (app *AppModel) renderTopBar() string {
	t := theme.CurrentTheme()

	var components []string

	// LSP status
	for name, status := range app.lspStatus {
		icon := "✓"
		color := t.LSPRunning()
		if status != "running" {
			icon = "❌"
			color = t.LSPError()
		}

		lspStyle := lipgloss.NewStyle().Foreground(color)
		components = append(components, lspStyle.Render(fmt.Sprintf("🔧 %s %s", name, icon)))
	}

	// Vector DB status
	dbStyle := lipgloss.NewStyle().Foreground(t.Info())
	components = append(components, dbStyle.Render(fmt.Sprintf("📊 Vector DB: %s", app.vectorDBStatus)))

	// Git branch
	gitStyle := lipgloss.NewStyle().Foreground(t.Success())
	components = append(components, gitStyle.Render(fmt.Sprintf("🌿 %s", app.currentBranch)))

	content := strings.Join(components, " | ")

	return styles.TopBarStyle().
		Width(app.width).
		Render(content)
}

// renderBottomBar renders the bottom status bar
func (app *AppModel) renderBottomBar() string {
	t := theme.CurrentTheme()

	// Left side - help text
	leftStyle := lipgloss.NewStyle().Foreground(t.TextMuted())
	left := leftStyle.Render("press enter to send, ctrl+q to quit")

	// Right side - model info
	rightStyle := lipgloss.NewStyle().Foreground(t.Text())
	right := rightStyle.Render("🤖 Claude 4 Sonnet | 💰 $0.05")

	// Calculate spacing
	totalWidth := app.width
	usedWidth := lipgloss.Width(left) + lipgloss.Width(right)
	spacing := totalWidth - usedWidth
	if spacing < 0 {
		spacing = 0
	}

	spacer := strings.Repeat(" ", spacing)
	content := left + spacer + right

	return styles.StatusBarStyle().
		Width(app.width).
		Render(content)
}

// updateSizes updates component sizes
func (app *AppModel) updateSizes() {
	app.layout.SetSize(app.width, app.height)

	// Get dimensions for components
	sidebarWidth, sidebarHeight := app.layout.GetSidebarDimensions()
	contentWidth, contentHeight := app.layout.GetContentDimensions()

	// Update component sizes
	app.fileTree.SetSize(sidebarWidth, sidebarHeight)
	app.tabManager.SetSize(contentWidth, contentHeight)
	app.dialogManager.SetSize(app.width, app.height)
}

// updateAnimationValues updates component animation values from the animation manager
func (app *AppModel) updateAnimationValues() {
	// Update tab switch animation
	if app.animManager.IsAnimating("tab-switch") {
		value := app.animManager.GetAnimationValue("tab-switch")
		app.tabManager.SetAnimationValue(value)
	}

	// Update dialog fade animation
	if app.animManager.IsAnimating("dialog-fade") {
		value := app.animManager.GetAnimationValue("dialog-fade")
		// TODO: Apply dialog fade animation
		_ = value
	}
}

// handleDialogResult handles the result of a dialog
func (app *AppModel) handleDialogResult(result dialog.DialogResult) {
	switch result.Type {
	case dialog.DialogTypeInit:
		if result.Confirmed {
			app.initialized = true
			log.Info("Project initialized")
			// TODO: Create OpenCode.md file
		}
	case dialog.DialogTypeSettings:
		if result.Confirmed {
			log.Info("Settings updated", "values", result.Values)
			// TODO: Apply settings
		}
	case dialog.DialogTypeCommand:
		if result.Confirmed {
			log.Info("Command executed", "values", result.Values)
			// TODO: Execute command
		}
	case dialog.DialogTypePermission:
		log.Info("Permission result", "confirmed", result.Confirmed)
		// TODO: Handle permission result
	case dialog.DialogTypeProviderSettings:
		if result.Confirmed {
			log.Info("Provider settings updated", "values", result.Values)
			// TODO: Apply provider settings
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
