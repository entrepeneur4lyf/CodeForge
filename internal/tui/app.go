package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/animation"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/chat"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/dialog"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/filetree"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/fileviewer"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/splash"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/tabs"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/toast"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/layout"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// AIResponseMsg is sent when AI responds to a message
type AIResponseMsg struct {
	Content string
	ID      string
	Error   error
}

// AppState represents the current state of the application
type AppState int

const (
	StateSplash AppState = iota
	StateMain
)

// Wrapper types to make components compatible with tea.Model
type FilePickerWrapper struct {
	filepicker.Model
}

func (w *FilePickerWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.Model.Update(msg)
	w.Model = model
	return w, cmd
}

type TabManagerWrapper struct {
	*tabs.TabManager
}

func (w *TabManagerWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.TabManager.Update(msg)
	w.TabManager = model
	return w, cmd
}

// AppModel represents the main application model
type AppModel struct {
	// Application state
	state AppState

	// Splash screen
	splashModel splash.Model

	// Core components
	filePicker   filepicker.Model
	fileViewer   *fileviewer.FileViewer
	tabManager   *tabs.TabManager
	chatModel    *chat.ChatModel
	animManager  *animation.Manager
	toastManager *toast.ToastManager

	// Dialogs (OpenCode pattern)
	showInitDialog         bool
	showProviderDialog     bool
	initDialog             dialog.InitDialogCmp
	providerSettingsDialog *dialog.ProviderSettingsDialog

	// Layout (modern OpenCode-style)
	splitLayout layout.SplitPaneLayout
	width       int
	height      int

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
	// Initialize logging (only show errors to reduce console clutter)
	log.SetLevel(log.ErrorLevel)

	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		log.Error("Failed to get absolute path", "error", err)
		absPath = projectPath
	}

	// Create splash screen
	splashModel := splash.New()

	// Create components (but don't initialize them yet)
	filePicker := filepicker.New()
	// Don't set CurrentDirectory yet - will be set after splash completes to avoid blocking UI
	chatModel := chat.NewChatModel()
	fileViewer := fileviewer.New()
	initDialog := dialog.NewInitDialogCmp()
	providerSettingsDialog := dialog.NewProviderSettingsDialog()
	tabManager := tabs.NewTabManager()
	animManager := animation.NewManager()
	toastManager := toast.NewToastManager()

	// Add tabs
	tabManager.AddTab("chat", "💬 Chat", chatModel)
	tabManager.AddTab("history", "📜 History", nil) // TODO: Implement history
	tabManager.AddTab("viewer", "📄 File Viewer", fileViewer)

	// Create modern split layout like OpenCode
	splitLayout := layout.NewSplitPaneLayout(
		layout.WithRatio(0.25),         // 25% sidebar, 75% main content
		layout.WithVerticalRatio(0.85), // 85% main area, 15% status
	)

	// Create wrapper models to make them compatible with tea.Model
	sidebarModel := &FilePickerWrapper{filePicker}
	mainModel := &TabManagerWrapper{tabManager}

	// Create containers for each panel
	sidebarContainer := layout.NewContainer(sidebarModel,
		layout.WithRoundedBorder(),
		layout.WithPaddingAll(1),
	)

	mainContainer := layout.NewContainer(mainModel,
		layout.WithRoundedBorder(),
		layout.WithPaddingAll(1),
	)

	// Set up the split layout with panels
	splitLayout.SetLeftPanel(sidebarContainer)
	splitLayout.SetRightPanel(mainContainer)

	// Initialize with a reasonable default size (will be updated on first WindowSizeMsg)
	splitLayout.SetSize(120, 30)

	app := &AppModel{
		state:                  StateSplash, // Start with splash screen
		splashModel:            splashModel,
		filePicker:             filePicker,
		fileViewer:             fileViewer,
		tabManager:             tabManager,
		chatModel:              chatModel,
		animManager:            animManager,
		toastManager:           toastManager,
		initDialog:             initDialog,
		providerSettingsDialog: providerSettingsDialog,
		showInitDialog:         false, // Will be set after splash completes
		showProviderDialog:     false,
		splitLayout:            splitLayout,
		projectPath:            absPath,
		focused:                "tabs",
		lspStatus:              make(map[string]string),
		gitStatus:              make(map[string]string),
		vectorDBStatus:         "Ready",
		currentBranch:          "main",
	}

	// Set dialog visibility based on initialization status
	app.showInitDialog = !app.initialized

	return app
}

// Init implements tea.Model
func (app *AppModel) Init() tea.Cmd {
	if app.state == StateSplash {
		// Only initialize splash screen initially
		return app.splashModel.Init()
	}

	var cmds []tea.Cmd

	// Initialize components
	cmds = append(cmds, app.tabManager.Init())
	cmds = append(cmds, app.chatModel.Init())

	// Don't load files immediately - wait until dialog is closed to avoid blocking UI

	// Initialize dialog (OpenCode pattern)
	cmds = append(cmds, app.initDialog.Init())

	// Show initialization dialog if not initialized
	if app.showInitDialog {
		app.focused = "dialog"
	}

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (app *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle splash screen state
	if app.state == StateSplash {
		var cmd tea.Cmd
		app.splashModel, cmd = app.splashModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check for splash completion
		if _, ok := msg.(splash.SplashCompleteMsg); ok {
			app.state = StateMain
			// Initialize main components now
			cmds = append(cmds, app.tabManager.Init())
			cmds = append(cmds, app.chatModel.Init())
			cmds = append(cmds, app.initDialog.Init())

			// Set dialog visibility and load files
			app.showInitDialog = !app.initialized
			if app.showInitDialog {
				app.focused = "dialog"
			}

			// Load files after splash completes
			cmds = append(cmds, func() tea.Msg {
				return "load_files"
			})
		}

		return app, tea.Batch(cmds...)
	}

	// Handle dialog first (OpenCode pattern) - this must come BEFORE main key handling
	if app.showInitDialog {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// Let global quit keys bypass the dialog
			if keyMsg.String() == "ctrl+c" || keyMsg.String() == "ctrl+q" {
				// Don't forward to dialog, let main handler process it
			} else {
				var cmd tea.Cmd
				d, cmd := app.initDialog.Update(msg)
				app.initDialog = d.(dialog.InitDialogCmp)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				// Don't process other keys when dialog is showing
				return app, tea.Batch(cmds...)
			}
		}
	}

	// Handle provider settings dialog
	if app.showProviderDialog {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// Let global quit keys bypass the dialog
			if keyMsg.String() == "ctrl+c" || keyMsg.String() == "ctrl+q" {
				// Don't forward to dialog, let main handler process it
			} else if keyMsg.String() == "esc" {
				// Close provider dialog on escape
				app.showProviderDialog = false
				app.focused = "tabs"
				return app, nil
			} else {
				var cmd tea.Cmd
				model, cmd := app.providerSettingsDialog.Update(msg)
				app.providerSettingsDialog = model.(*dialog.ProviderSettingsDialog)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				// Don't process other keys when dialog is showing
				return app, tea.Batch(cmds...)
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global shortcuts
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return app, tea.Quit
		case "ctrl+p":
			// Show command palette via toast
			cmds = append(cmds, toast.NewInfoToast(
				"Available commands:\n• Ctrl+Shift+P - Provider Settings\n• Ctrl+, - Settings\n• Ctrl+E - File Explorer\n• Tab - Switch Tabs\n• Ctrl+C/Q - Quit",
				toast.WithTitle("Command Palette"),
				toast.WithDuration(5*time.Second),
			))
		case "ctrl+comma":
			// Show settings info via toast
			cmds = append(cmds, toast.NewInfoToast(
				"Settings:\n• Use Ctrl+Shift+P for Provider Settings\n• Theme: CodeForge (default)\n• Config: ~/.config/codeforge/config.yaml",
				toast.WithTitle("Settings Info"),
				toast.WithDuration(4*time.Second),
			))
		case "ctrl+shift+p":
			// Show provider settings dialog
			app.showProviderDialog = true
			app.focused = "provider_dialog"
			cmds = append(cmds, toast.NewInfoToast(
				"Provider settings dialog opened",
				toast.WithTitle("Provider Settings"),
				toast.WithDuration(2*time.Second),
			))
		case "ctrl+e":
			if !app.showInitDialog {
				app.focused = "filetree"
				app.tabManager.SetFocused(false)
			}
		case "ctrl+t":
			if !app.showInitDialog {
				app.focused = "tabs"
				app.tabManager.SetFocused(true)
			}
		// Global tab switching shortcuts
		case "ctrl+1":
			if !app.showInitDialog {
				cmds = append(cmds, app.tabManager.SwitchToIndex(0))
			}
		case "ctrl+2":
			if !app.showInitDialog {
				cmds = append(cmds, app.tabManager.SwitchToIndex(1))
			}
		case "ctrl+3":
			if !app.showInitDialog {
				cmds = append(cmds, app.tabManager.SwitchToIndex(2))
			}
		case "tab":
			if !app.showInitDialog && app.focused == "tabs" {
				// Switch to next tab
				nextTab := (app.tabManager.GetActiveTabIndex() + 1) % app.tabManager.GetTabCount()
				cmds = append(cmds, app.tabManager.SwitchToIndex(nextTab))
			}
		case "f1":
			if !app.showInitDialog {
				// Show help information via toast
				cmds = append(cmds, toast.NewInfoToast(
					"CodeForge Help:\n• F1 - Help\n• Ctrl+P - Command Palette\n• Ctrl+Shift+P - Provider Settings\n• Ctrl+, - Settings\n• Ctrl+E - File Explorer\n• Ctrl+T - Focus Tabs\n• Tab - Switch Focus\n• Ctrl+1/2/3 - Switch Tabs\n• Ctrl+C/Q - Quit",
					toast.WithTitle("Help"),
					toast.WithDuration(6*time.Second),
				))
			}
		}

	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		app.toastManager.SetSize(msg.Width, msg.Height)
		app.updateSizes()

	case dialog.CloseInitDialogMsg:
		// Handle init dialog close (OpenCode pattern)
		app.showInitDialog = false
		if msg.Initialize {
			app.initialized = true
			// Create AGENT.md file
			cmd := app.createOpenCodeFile()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		app.focused = "tabs"
		app.tabManager.SetFocused(true)

		// Start fade-out animation for dialog
		app.animManager.StartAnimation("dialog-fade", animation.AnimationDialogFade, 0.0, 0.9, 0.2)

		// Now load files asynchronously after dialog closes
		cmds = append(cmds, func() tea.Msg {
			return "load_files"
		})

	case dialog.DialogClosedMsg:
		cmd := app.handleDialogResult(msg.Result)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		app.focused = "tabs"
		app.tabManager.SetFocused(true)

		// Start fade-out animation for dialog
		app.animManager.StartAnimation("dialog-fade", animation.AnimationDialogFade, 0.0, 0.9, 0.2)

	case dialog.ProviderSelectedMsg:
		// Apply provider selection
		app.showProviderDialog = false
		app.focused = "tabs"
		// Show success toast
		cmds = append(cmds, toast.NewSuccessToast(
			fmt.Sprintf("Switched to %s provider", msg.Provider.Name),
			toast.WithTitle("Provider Changed"),
			toast.WithDuration(2*time.Second),
		))

	case dialog.ModelToggledMsg:
		// Apply model toggle (silent operation)

	case toast.ShowToastMsg, toast.DismissToastMsg:
		var cmd tea.Cmd
		app.toastManager, cmd = app.toastManager.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case animation.AnimationTickMsg:
		cmd := app.animManager.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Update component animation values
		app.updateAnimationValues()

	case animation.AnimationCompleteMsg:
		// Reset animation values when complete (silent operation)
		switch msg.Type {
		case animation.AnimationTabSwitch:
			app.tabManager.SetAnimationValue(0.0)
		}

	case filetree.FileSelectedMsg:
		// Open file in viewer tab
		app.fileViewer.LoadFile(msg.Path)
		app.tabManager.SwitchToTab("viewer")
		app.focused = "tabs"
		// Show info toast
		fileName := filepath.Base(msg.Path)
		cmds = append(cmds, toast.NewInfoToast(
			fmt.Sprintf("Opened %s", fileName),
			toast.WithTitle("File Opened"),
			toast.WithDuration(2*time.Second),
		))

	case chat.MessageSentMsg:
		// Send to AI service and get response (silent operation)
		return app, app.sendToAI(msg.Content)

	case AIResponseMsg:
		if msg.Error != nil {
			// Show error toast and send error message to chat
			cmds = append(cmds, toast.NewErrorToast(
				fmt.Sprintf("AI Error: %s", msg.Error.Error()),
				toast.WithTitle("AI Response Error"),
				toast.WithDuration(4*time.Second),
			))
			return app, func() tea.Msg {
				return chat.MessageReceivedMsg{
					Content: fmt.Sprintf("❌ Error: %s", msg.Error.Error()),
					ID:      fmt.Sprintf("error-%d", time.Now().UnixNano()),
				}
			}
		} else {
			// Send AI response to chat (silent operation)
			return app, func() tea.Msg {
				return chat.MessageReceivedMsg{
					Content: msg.Content,
					ID:      msg.ID,
				}
			}
		}

	case chat.MessageReceivedMsg:
		// Forward to tab manager to display in chat (silent operation)
		var cmd tea.Cmd
		app.tabManager, cmd = app.tabManager.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		// Mouse support disabled - focus on keyboard navigation

	// Note: Standard filepicker handles file selection internally

	case fileviewer.FileLoadedMsg:
		// File loaded successfully (silent operation)

	case tabs.TabSwitchedMsg:
		// Start tab switch animation (silent operation)
		app.animManager.StartAnimation("tab-switch", animation.AnimationTabSwitch, 1.0, 0.8, 0.15)

	case string:
		if msg == "force_file_refresh" {
			// Force file picker to refresh its directory listing (silent operation)
			cmd := app.filePicker.Init()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else if msg == "load_files" {
			// Initialize file picker directory and trigger initialization (silent operation)
			app.filePicker.CurrentDirectory = app.projectPath
			// Call Init to trigger directory reading
			initCmd := app.filePicker.Init()
			if initCmd != nil {
				cmds = append(cmds, initCmd)
			}
		}

	default:
		// Always forward file picker messages regardless of focus
		// This ensures readDirMsg and other file picker messages are processed
		var cmd tea.Cmd
		app.filePicker, cmd = app.filePicker.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Forward messages to appropriate components based on focus
		if !app.showInitDialog {
			switch app.focused {
			case "filetree":
				var cmd tea.Cmd
				app.filePicker, cmd = app.filePicker.Update(msg)
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
	// Show splash screen if in splash state
	if app.state == StateSplash {
		return app.splashModel.View()
	}

	// Render the modern split layout
	mainView := app.splitLayout.View()

	// Add top and bottom bars
	topBar := app.renderTopBar()
	bottomBar := app.renderBottomBar()

	// Combine with vertical layout and ensure full width
	if topBar != "" && bottomBar != "" {
		mainView = lipgloss.JoinVertical(lipgloss.Left, topBar, mainView, bottomBar)
	} else if topBar != "" {
		mainView = lipgloss.JoinVertical(lipgloss.Left, topBar, mainView)
	} else if bottomBar != "" {
		mainView = lipgloss.JoinVertical(lipgloss.Left, mainView, bottomBar)
	}

	// Ensure the main view uses the full terminal width
	mainView = lipgloss.NewStyle().
		Width(app.width).
		Height(app.height).
		Render(mainView)

	// Overlay dialog if visible (OpenCode pattern)
	if app.showInitDialog {
		overlay := app.initDialog.View()
		// Center the dialog overlay
		return layout.PlaceOverlay(
			app.width/2-lipgloss.Width(overlay)/2,
			app.height/2-lipgloss.Height(overlay)/2,
			overlay,
			mainView,
			true,
		)
	}

	// Overlay provider settings dialog if visible
	if app.showProviderDialog {
		app.providerSettingsDialog.SetSize(app.width, app.height)
		overlay := app.providerSettingsDialog.View()
		// Center the dialog overlay
		return layout.PlaceOverlay(
			app.width/2-lipgloss.Width(overlay)/2,
			app.height/2-lipgloss.Height(overlay)/2,
			overlay,
			mainView,
			true,
		)
	}

	// Add toast overlay
	mainView = app.toastManager.RenderOverlay(mainView)

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
	// Update the split layout size
	app.splitLayout.SetSize(app.width, app.height)

	// Update dialog sizes
	app.initDialog.SetSize(app.width, app.height)
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
func (app *AppModel) handleDialogResult(result dialog.DialogResult) tea.Cmd {
	switch result.Type {
	case dialog.DialogTypeInit:
		if result.Confirmed {
			app.initialized = true
			// Create AGENT.md file
			return app.createOpenCodeFile()
		}
	case dialog.DialogTypeSettings:
		if result.Confirmed {
			// Apply settings
			if themeName, ok := result.Values["theme"].(string); ok {
				newTheme := theme.LoadTheme(themeName)
				theme.SetTheme(newTheme)
				// Show success toast
				return toast.NewSuccessToast(
					fmt.Sprintf("Theme changed to %s", themeName),
					toast.WithTitle("Theme Updated"),
					toast.WithDuration(2*time.Second),
				)
			}
		}
	case dialog.DialogTypeCommand:
		if result.Confirmed {
			// Execute command
			if command, ok := result.Values["command"].(string); ok {
				// Simple command execution - could be expanded to run actual commands
				switch command {
				case "refresh":
					// Refresh file picker by changing directory to itself
					app.filePicker.CurrentDirectory = app.projectPath
					return toast.NewSuccessToast(
						"File picker refreshed",
						toast.WithTitle("Refresh Complete"),
						toast.WithDuration(2*time.Second),
					)
				case "clear_chat":
					// Clear chat history
					return toast.NewSuccessToast(
						"Chat history cleared",
						toast.WithTitle("Chat Cleared"),
						toast.WithDuration(2*time.Second),
					)
				default:
					return toast.NewWarningToast(
						fmt.Sprintf("Unknown command: %s", command),
						toast.WithTitle("Command Error"),
						toast.WithDuration(3*time.Second),
					)
				}
			}
		}
	case dialog.DialogTypePermission:
		// Handle permission result (silent operation)
		if result.Confirmed {
			// Execute the action that required permission
		} else {
			// Cancel the action
		}
	case dialog.DialogTypeProviderSettings:
		if result.Confirmed {
			// Apply provider settings
			if provider, ok := result.Values["provider"].(string); ok {
				// TODO: Update the chat model with new provider when LLM integration is complete
				return toast.NewSuccessToast(
					fmt.Sprintf("Switched to %s provider", provider),
					toast.WithTitle("Provider Updated"),
					toast.WithDuration(2*time.Second),
				)
			}
			if themeName, ok := result.Values["theme"].(string); ok {
				// Apply theme change
				newTheme := theme.LoadTheme(themeName)
				theme.SetTheme(newTheme)
				return toast.NewSuccessToast(
					fmt.Sprintf("Theme changed to %s", themeName),
					toast.WithTitle("Theme Updated"),
					toast.WithDuration(2*time.Second),
				)
			}
		}
	}
	return nil
}

// sendToAI sends a message to the AI service and returns a command that will send the response
func (app *AppModel) sendToAI(message string) tea.Cmd {
	return func() tea.Msg {
		// Get the default agent configuration
		cfg := config.Get()
		if cfg == nil {
			return AIResponseMsg{
				Error: fmt.Errorf("configuration not loaded"),
			}
		}

		// Get the coder agent configuration
		agent, exists := cfg.Agents[config.AgentCoder]
		if !exists {
			return AIResponseMsg{
				Error: fmt.Errorf("coder agent not configured"),
			}
		}

		// Create completion request
		req := llm.CompletionRequest{
			Model: agent.Model,
			Messages: []llm.Message{
				{
					Role:    "user",
					Content: message,
				},
			},
			MaxTokens:   agent.MaxTokens,
			Temperature: 0.7,
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get completion from LLM
		resp, err := llm.GetCompletion(ctx, req)
		if err != nil {
			return AIResponseMsg{
				Error: fmt.Errorf("AI request failed: %w", err),
			}
		}

		// Return successful response
		return AIResponseMsg{
			Content: resp.Content,
			ID:      fmt.Sprintf("ai-%d", time.Now().UnixNano()),
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

// createOpenCodeFile creates the AGENT.md memory file for the project
func (app *AppModel) createOpenCodeFile() tea.Cmd {
	agentContent := `# CodeForge Agent Memory

This file serves as memory for AI agents working in this CodeForge repository.

## Project Overview
- **Type**: Go application with TUI interface
- **Purpose**: AI-powered code assistant and development tool
- **Architecture**: Bubble Tea TUI with modular component system
- **Framework**: Built on Charm ecosystem (Bubble Tea, Lipgloss, Bubbles)

## Build Commands
- Build: ` + "`go build ./cmd/codeforge`" + `
- Run: ` + "`./codeforge tui`" + `
- Test: ` + "`go test ./...`" + `
- Clean: ` + "`go clean`" + `
- Format: ` + "`gofmt -w .`" + `

## Code Style Guidelines
- Use Go standard formatting (` + "`gofmt`" + `)
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Use structured logging: ` + "`log.Info()`, `log.Error()`, `log.Debug()`" + `
- Import organization: standard library, third-party, local packages
- Keep functions focused and under 50 lines when possible
- Use interfaces for testability and modularity
- Handle errors explicitly, never ignore them
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

## Architecture Components
- **TUI Framework**: Bubble Tea v2 with Lipgloss v2 styling
- **Chat Interface**: AI conversation with Glamour markdown rendering
- **File Management**: Bubbles filepicker with filtering
- **Dialog System**: Modal dialogs for settings, initialization, provider config
- **Theme System**: Pluggable color themes (CodeForge, Catppuccin, Dracula, etc.)
- **Animation**: Harmonica physics-based smooth animations
- **Provider System**: Multi-LLM support (OpenAI, Anthropic, Groq, Local, etc.)

## Key Dependencies
- ` + "`github.com/charmbracelet/bubbletea/v2`" + ` - TUI framework
- ` + "`github.com/charmbracelet/lipgloss/v2`" + ` - Styling and layout
- ` + "`github.com/charmbracelet/bubbles`" + ` - UI components
- ` + "`github.com/charmbracelet/glamour`" + ` - Markdown rendering
- ` + "`github.com/charmbracelet/harmonica`" + ` - Animation physics
- ` + "`github.com/charmbracelet/log`" + ` - Structured logging

## Development Patterns
- Follow OpenCode architectural patterns for consistency
- Use the provider pattern for AI model integration
- Implement proper error handling with context
- Write comprehensive tests for UI components
- Maintain responsive design principles
- Use dependency injection for testability
- Keep state management centralized in app model
- Use message passing for component communication

## File Structure
- ` + "`cmd/codeforge/`" + ` - Main application entry point
- ` + "`internal/tui/`" + ` - TUI implementation and components
- ` + "`internal/llm/`" + ` - LLM provider implementations
- ` + "`internal/models/`" + ` - Data models and types
- ` + "`internal/config/`" + ` - Configuration management

## Testing Guidelines
- Write unit tests for all business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies (LLM APIs, file system)
- Test UI components with bubble tea testing utilities
- Maintain >80% code coverage
- Use integration tests for critical user flows

## Performance Considerations
- Lazy load file trees for large repositories
- Implement virtual scrolling for large chat histories
- Use efficient rendering with Lipgloss caching
- Minimize allocations in hot paths
- Profile memory usage regularly
- Optimize startup time with async initialization
`

	filePath := filepath.Join(app.projectPath, "AGENT.md")
	err := os.WriteFile(filePath, []byte(agentContent), 0644)
	if err != nil {
		// Show error toast
		return tea.Batch(toast.NewErrorToast(
			fmt.Sprintf("Failed to create AGENT.md: %v", err),
			toast.WithTitle("File Creation Error"),
		))
	} else {
		// Show success toast
		return tea.Batch(toast.NewSuccessToast(
			"AGENT.md file created successfully",
			toast.WithTitle("Project Initialized"),
			toast.WithDuration(3*time.Second),
		))
	}
}
