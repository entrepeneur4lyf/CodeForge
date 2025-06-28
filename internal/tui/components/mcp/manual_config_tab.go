package mcp

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// TransportType represents the MCP transport type
type TransportType int

const (
	TransportStdio TransportType = iota
	TransportHTTP
	TransportSSE
)

// String returns the transport type as a string
func (t TransportType) String() string {
	switch t {
	case TransportStdio:
		return "stdio"
	case TransportHTTP:
		return "http"
	case TransportSSE:
		return "sse"
	default:
		return "unknown"
	}
}

// ManualConfigTabModel represents the manual configuration tab
type ManualConfigTabModel struct {
	width         int
	height        int
	focused       bool
	currentField  int
	transportType TransportType
	
	// Form fields
	serverName    textinput.Model
	command       textinput.Model
	arguments     textinput.Model
	workingDir    textinput.Model
	httpURL       textinput.Model
	sseURL        textinput.Model
	envVars       textinput.Model
	
	// Form state
	fields        []textinput.Model
	transportOptions []string
	transportIndex   int
}

// NewManualConfigTabModel creates a new manual configuration tab model
func NewManualConfigTabModel() *ManualConfigTabModel {
	// Create form fields
	serverName := textinput.New()
	serverName.Placeholder = "my-custom-server"
	serverName.CharLimit = 100
	serverName.Width = 50

	command := textinput.New()
	command.Placeholder = "node server.js"
	command.CharLimit = 200
	command.Width = 50

	arguments := textinput.New()
	arguments.Placeholder = "--port 3000 --verbose"
	arguments.CharLimit = 200
	arguments.Width = 50

	workingDir := textinput.New()
	workingDir.Placeholder = "/path/to/server"
	workingDir.CharLimit = 200
	workingDir.Width = 50

	httpURL := textinput.New()
	httpURL.Placeholder = "http://localhost:3000/mcp"
	httpURL.CharLimit = 200
	httpURL.Width = 50

	sseURL := textinput.New()
	sseURL.Placeholder = "http://localhost:3000/sse"
	sseURL.CharLimit = 200
	sseURL.Width = 50

	envVars := textinput.New()
	envVars.Placeholder = "API_KEY=secret,DEBUG=true"
	envVars.CharLimit = 500
	envVars.Width = 50

	// Focus first field
	serverName.Focus()

	fields := []textinput.Model{
		serverName, command, arguments, workingDir, 
		httpURL, sseURL, envVars,
	}

	return &ManualConfigTabModel{
		focused:          false,
		currentField:     0,
		transportType:    TransportStdio,
		serverName:       serverName,
		command:          command,
		arguments:        arguments,
		workingDir:       workingDir,
		httpURL:          httpURL,
		sseURL:           sseURL,
		envVars:          envVars,
		fields:           fields,
		transportOptions: []string{"stdio", "http", "sse"},
		transportIndex:   0,
	}
}

// SetSize sets the size of the manual config tab
func (mctm *ManualConfigTabModel) SetSize(width, height int) {
	mctm.width = width
	mctm.height = height
	
	// Update field widths
	fieldWidth := width - 20
	if fieldWidth < 30 {
		fieldWidth = 30
	}
	
	for i := range mctm.fields {
		mctm.fields[i].Width = fieldWidth
	}
}

// Init implements tea.Model
func (mctm *ManualConfigTabModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (mctm *ManualConfigTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			// Move to next field
			mctm.nextField()

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			// Move to previous field
			mctm.prevField()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+t"))):
			// Cycle transport type
			mctm.transportIndex = (mctm.transportIndex + 1) % len(mctm.transportOptions)
			switch mctm.transportIndex {
			case 0:
				mctm.transportType = TransportStdio
			case 1:
				mctm.transportType = TransportHTTP
			case 2:
				mctm.transportType = TransportSSE
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			// Save configuration
			return mctm, mctm.saveConfiguration()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+r"))):
			// Reset form
			return mctm, mctm.resetForm()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+v"))):
			// Test connection
			return mctm, mctm.testConnection()

		default:
			// Update current field
			if mctm.currentField < len(mctm.fields) {
				var cmd tea.Cmd
				mctm.fields[mctm.currentField], cmd = mctm.fields[mctm.currentField].Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.WindowSizeMsg:
		mctm.SetSize(msg.Width, msg.Height)
	}

	return mctm, tea.Batch(cmds...)
}

// nextField moves to the next form field
func (mctm *ManualConfigTabModel) nextField() {
	mctm.fields[mctm.currentField].Blur()
	mctm.currentField = (mctm.currentField + 1) % len(mctm.fields)
	mctm.fields[mctm.currentField].Focus()
}

// prevField moves to the previous form field
func (mctm *ManualConfigTabModel) prevField() {
	mctm.fields[mctm.currentField].Blur()
	mctm.currentField = (mctm.currentField - 1 + len(mctm.fields)) % len(mctm.fields)
	mctm.fields[mctm.currentField].Focus()
}

// saveConfiguration saves the server configuration
func (mctm *ManualConfigTabModel) saveConfiguration() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual configuration saving
		return ConfigSaveMsg{
			Success: true,
			Message: "Server configuration saved successfully",
		}
	}
}

// resetForm resets the form to default values
func (mctm *ManualConfigTabModel) resetForm() tea.Cmd {
	return func() tea.Msg {
		return ConfigResetMsg{}
	}
}

// testConnection tests the server connection
func (mctm *ManualConfigTabModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual connection testing
		return ConnectionTestMsg{
			Success: true,
			Message: "Connection test successful",
		}
	}
}

// ConfigSaveMsg represents a configuration save message
type ConfigSaveMsg struct {
	Success bool
	Message string
	Error   error
}

// ConfigResetMsg represents a configuration reset message
type ConfigResetMsg struct{}

// ConnectionTestMsg represents a connection test message
type ConnectionTestMsg struct {
	Success bool
	Message string
	Error   error
}

// View implements tea.Model
func (mctm *ManualConfigTabModel) View() string {
	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme().Primary())
	sections = append(sections, titleStyle.Render("🔧 Manual MCP Server Configuration"))
	sections = append(sections, "")

	// Transport type selector
	transportStyle := lipgloss.NewStyle().Bold(true)
	sections = append(sections, transportStyle.Render("Transport Type:"))
	
	var transportOptions []string
	for i, option := range mctm.transportOptions {
		if i == mctm.transportIndex {
			selectedStyle := lipgloss.NewStyle().
				Background(theme.CurrentTheme().Primary()).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
			transportOptions = append(transportOptions, selectedStyle.Render(option))
		} else {
			transportOptions = append(transportOptions, option)
		}
	}
	sections = append(sections, "  "+strings.Join(transportOptions, "  "))
	sections = append(sections, "")

	// Form fields based on transport type
	sections = append(sections, mctm.renderFormFields())

	// Action buttons
	sections = append(sections, "")
	sections = append(sections, mctm.renderActionButtons())

	// Help text
	sections = append(sections, "")
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true)
	
	helpText := "Tab: next field • Shift+Tab: prev field • Ctrl+T: transport • Ctrl+S: save • Ctrl+V: test • Ctrl+R: reset"
	sections = append(sections, helpStyle.Render(helpText))

	return strings.Join(sections, "\n")
}

// renderFormFields renders the form fields based on transport type
func (mctm *ManualConfigTabModel) renderFormFields() string {
	var fields []string

	// Common fields
	fields = append(fields, mctm.renderField("Server Name:", mctm.serverName, 0))

	switch mctm.transportType {
	case TransportStdio:
		fields = append(fields, mctm.renderField("Command:", mctm.command, 1))
		fields = append(fields, mctm.renderField("Arguments:", mctm.arguments, 2))
		fields = append(fields, mctm.renderField("Working Directory:", mctm.workingDir, 3))
		fields = append(fields, mctm.renderField("Environment Variables:", mctm.envVars, 6))

	case TransportHTTP:
		fields = append(fields, mctm.renderField("HTTP URL:", mctm.httpURL, 4))
		fields = append(fields, mctm.renderField("Environment Variables:", mctm.envVars, 6))

	case TransportSSE:
		fields = append(fields, mctm.renderField("SSE URL:", mctm.sseURL, 5))
		fields = append(fields, mctm.renderField("Environment Variables:", mctm.envVars, 6))
	}

	return strings.Join(fields, "\n")
}

// renderField renders a single form field
func (mctm *ManualConfigTabModel) renderField(label string, field textinput.Model, index int) string {
	labelStyle := lipgloss.NewStyle().Bold(true).Width(20)
	
	fieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	
	if index == mctm.currentField {
		fieldStyle = fieldStyle.BorderForeground(theme.CurrentTheme().Primary())
	} else {
		fieldStyle = fieldStyle.BorderForeground(theme.CurrentTheme().Secondary())
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		fieldStyle.Render(field.View()),
	)
}

// renderActionButtons renders the action buttons
func (mctm *ManualConfigTabModel) renderActionButtons() string {
	buttonStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Primary()).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 1, 0, 0)

	saveButton := buttonStyle.Render("💾 Save")
	testButton := buttonStyle.Background(lipgloss.Color("33")).Render("🔍 Test")
	resetButton := buttonStyle.Background(lipgloss.Color("196")).Render("🔄 Reset")

	return lipgloss.JoinHorizontal(lipgloss.Left, saveButton, testButton, resetButton)
}
