package mcp

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// InstalledTabModel represents the installed servers tab
type InstalledTabModel struct {
	width         int
	height        int
	servers       []MCPServer
	selectedIndex int
	expandedServers map[string]bool // Track which servers are expanded
	focused       bool
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]string
	AutoApprove bool
}

// Resource represents an MCP resource
type Resource struct {
	Name        string
	Description string
	URI         string
}

// NewInstalledTabModel creates a new installed tab model
func NewInstalledTabModel(servers []MCPServer) *InstalledTabModel {
	return &InstalledTabModel{
		servers:         servers,
		selectedIndex:   0,
		expandedServers: make(map[string]bool),
		focused:         false,
	}
}

// SetSize sets the size of the installed tab
func (itm *InstalledTabModel) SetSize(width, height int) {
	itm.width = width
	itm.height = height
}

// Init implements tea.Model
func (itm *InstalledTabModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (itm *InstalledTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if itm.selectedIndex > 0 {
				itm.selectedIndex--
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if itm.selectedIndex < len(itm.servers)-1 {
				itm.selectedIndex++
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			// Toggle server expansion
			if itm.selectedIndex < len(itm.servers) {
				serverID := itm.servers[itm.selectedIndex].ID
				itm.expandedServers[serverID] = !itm.expandedServers[serverID]
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
			// Toggle server status (on/off)
			if itm.selectedIndex < len(itm.servers) {
				return itm, itm.toggleServerStatus(itm.servers[itm.selectedIndex])
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Restart server
			if itm.selectedIndex < len(itm.servers) {
				return itm, itm.restartServer(itm.servers[itm.selectedIndex])
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			// Delete server
			if itm.selectedIndex < len(itm.servers) {
				return itm, itm.deleteServer(itm.servers[itm.selectedIndex])
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// Configure MCP servers
			return itm, itm.configureServers()
		}

	case tea.WindowSizeMsg:
		itm.SetSize(msg.Width, msg.Height)

	case ServerStatusToggleMsg:
		// Update server status
		for i, server := range itm.servers {
			if server.ID == msg.ServerID {
				itm.servers[i].Status = msg.NewStatus
				break
			}
		}
	}

	return itm, nil
}

// toggleServerStatus toggles the server on/off
func (itm *InstalledTabModel) toggleServerStatus(server MCPServer) tea.Cmd {
	return func() tea.Msg {
		var newStatus ServerStatus
		switch server.Status {
		case StatusOff:
			newStatus = StatusOn
		case StatusOn:
			newStatus = StatusOff
		case StatusError:
			newStatus = StatusOn // Try to restart from error
		}
		
		// TODO: Implement actual server toggle logic
		return ServerStatusToggleMsg{
			ServerID:  server.ID,
			NewStatus: newStatus,
		}
	}
}

// restartServer restarts a server
func (itm *InstalledTabModel) restartServer(server MCPServer) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual server restart logic
		return ServerRestartMsg{
			ServerID: server.ID,
			Success:  true,
		}
	}
}

// deleteServer deletes a server
func (itm *InstalledTabModel) deleteServer(server MCPServer) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual server deletion logic
		return ServerDeleteMsg{
			ServerID: server.ID,
			Success:  true,
		}
	}
}

// configureServers opens the server configuration
func (itm *InstalledTabModel) configureServers() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement server configuration dialog
		return ConfigureServersMsg{}
	}
}

// ServerStatusToggleMsg represents a server status toggle message
type ServerStatusToggleMsg struct {
	ServerID  string
	NewStatus ServerStatus
}

// ServerRestartMsg represents a server restart message
type ServerRestartMsg struct {
	ServerID string
	Success  bool
	Error    error
}

// ServerDeleteMsg represents a server deletion message
type ServerDeleteMsg struct {
	ServerID string
	Success  bool
	Error    error
}

// ConfigureServersMsg represents a configure servers message
type ConfigureServersMsg struct{}

// View implements tea.Model
func (itm *InstalledTabModel) View() string {
	if len(itm.servers) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().Secondary()).
			Italic(true).
			Align(lipgloss.Center)
		
		return emptyStyle.Render("No MCP servers installed.\nVisit the Marketplace tab to install servers.")
	}

	var sections []string

	// Render each server
	for i, server := range itm.servers {
		serverView := itm.renderServer(server, i == itm.selectedIndex, itm.expandedServers[server.ID])
		sections = append(sections, serverView)
	}

	// Configure button
	configStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Primary()).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 2).
		Margin(1, 0, 0, 0).
		Align(lipgloss.Center)
	
	configButton := configStyle.Render("⚙️ Configure MCP Servers")
	sections = append(sections, configButton)

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true).
		Margin(1, 0, 0, 0)
	
	helpText := "↑↓: navigate • Enter: expand/collapse • t: toggle on/off • r: restart • d: delete • c: configure"
	sections = append(sections, helpStyle.Render(helpText))

	return strings.Join(sections, "\n")
}

// renderServer renders a single server with expansion
func (itm *InstalledTabModel) renderServer(server MCPServer, selected bool, expanded bool) string {
	// Server header styling
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(itm.width - 4)
	
	if selected {
		headerStyle = headerStyle.BorderForeground(theme.CurrentTheme().Primary())
	} else {
		headerStyle = headerStyle.BorderForeground(theme.CurrentTheme().Secondary())
	}

	// Status indicator (clickable toggle)
	statusIndicator := server.Status.StatusIndicator()
	
	// Expansion indicator
	expandIcon := "▶"
	if expanded {
		expandIcon = "▼"
	}

	// Server name and status
	nameStyle := lipgloss.NewStyle().Bold(true)
	serverName := fmt.Sprintf("%s %s %s", expandIcon, server.Name, statusIndicator)
	
	// Tools and resources count
	countsStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary())
	counts := fmt.Sprintf("Tools (%d)    Resources (%d)", server.ToolsCount, server.ResourcesCount)

	// Header content
	headerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		nameStyle.Render(serverName),
		countsStyle.Render(counts),
	)

	// Expanded content
	var expandedContent string
	if expanded {
		expandedContent = itm.renderExpandedServer(server)
	}

	// Combine header and expanded content
	var content string
	if expanded {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			headerContent,
			"",
			expandedContent,
		)
	} else {
		content = headerContent
	}

	return headerStyle.Render(content)
}

// renderExpandedServer renders the expanded server details
func (itm *InstalledTabModel) renderExpandedServer(server MCPServer) string {
	var sections []string

	// Mock tools for demonstration
	tools := []Tool{
		{Name: "runtime-errors", Description: "Get application runtime errors", AutoApprove: true},
		{Name: "runtime-logs", Description: "Get application runtime logs", AutoApprove: true},
		{Name: "runtime-logs-by-location", Description: "Get logs for specific file and line", AutoApprove: false},
		{Name: "runtime-logs-and-errors", Description: "Get both logs and errors", AutoApprove: true},
	}

	// Tools section
	if len(tools) > 0 {
		toolsHeader := lipgloss.NewStyle().Bold(true).Render("🔧 Tools:")
		sections = append(sections, toolsHeader)
		
		for _, tool := range tools {
			checkBox := "☐"
			if tool.AutoApprove {
				checkBox = "☑"
			}
			
			toolStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text())
			toolLine := fmt.Sprintf("  %s %s - %s", checkBox, tool.Name, tool.Description)
			sections = append(sections, toolStyle.Render(toolLine))
		}
		sections = append(sections, "")
	}

	// Global settings
	settingsStyle := lipgloss.NewStyle().Bold(true)
	sections = append(sections, settingsStyle.Render("⚙️ Settings:"))
	
	autoApproveStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text())
	sections = append(sections, autoApproveStyle.Render("  ☑ Auto-approve all tools"))
	sections = append(sections, autoApproveStyle.Render("  Request Timeout: 1 minute"))
	sections = append(sections, "")

	// Action buttons
	buttonStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Secondary()).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Margin(0, 1, 0, 0)
	
	restartButton := buttonStyle.Background(lipgloss.Color("33")).Render("Restart Server")
	deleteButton := buttonStyle.Background(lipgloss.Color("196")).Render("Delete Server")
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, restartButton, deleteButton)
	sections = append(sections, buttons)

	return strings.Join(sections, "\n")
}
