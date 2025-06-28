package mcp

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/components/tabs"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// MarketplaceModel represents the MCP marketplace interface
type MarketplaceModel struct {
	width      int
	height     int
	tabManager *tabs.TabManager
	focused    bool
}

// ServerStatus represents the three-state server status
type ServerStatus int

const (
	StatusOff   ServerStatus = iota // Red - disabled
	StatusOn                        // Green - connected
	StatusError                     // Yellow - error/warning
)

// String returns the status as a string
func (s ServerStatus) String() string {
	switch s {
	case StatusOff:
		return "OFF"
	case StatusOn:
		return "ON"
	case StatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Color returns the color for the status indicator
func (s ServerStatus) Color() lipgloss.Color {
	switch s {
	case StatusOff:
		return lipgloss.Color("196") // Red
	case StatusOn:
		return lipgloss.Color("46") // Green
	case StatusError:
		return lipgloss.Color("226") // Yellow
	default:
		return lipgloss.Color("240") // Gray
	}
}

// StatusIndicator returns a styled status indicator
func (s ServerStatus) StatusIndicator() string {
	style := lipgloss.NewStyle().
		Foreground(s.Color()).
		Bold(true)
	return style.Render("●")
}

// MCPCategory represents the category of an MCP server
type MCPCategory int

const (
	CategoryOfficial MCPCategory = iota
	CategoryReference
	CategoryThirdParty
	CategoryCommunity
	CategoryFrameworks
	CategoryGeneral
)

// String returns the string representation of MCPCategory
func (c MCPCategory) String() string {
	switch c {
	case CategoryOfficial:
		return "Official"
	case CategoryReference:
		return "Reference"
	case CategoryThirdParty:
		return "Third-party"
	case CategoryCommunity:
		return "Community"
	case CategoryFrameworks:
		return "Frameworks"
	case CategoryGeneral:
		return "General"
	default:
		return "Unknown"
	}
}

// ParseMCPCategory parses a string into MCPCategory
func ParseMCPCategory(s string) MCPCategory {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "official":
		return CategoryOfficial
	case "reference":
		return CategoryReference
	case "third-party", "thirdparty":
		return CategoryThirdParty
	case "community":
		return CategoryCommunity
	case "frameworks":
		return CategoryFrameworks
	case "general":
		return CategoryGeneral
	default:
		return CategoryGeneral
	}
}

// InstallationType represents the installation method for an MCP server
type InstallationType int

const (
	InstallationNPM InstallationType = iota
	InstallationPython
	InstallationDocker
	InstallationBinary
	InstallationGit
	InstallationUnknown
)

// String returns the string representation of InstallationType
func (i InstallationType) String() string {
	switch i {
	case InstallationNPM:
		return "npm"
	case InstallationPython:
		return "python"
	case InstallationDocker:
		return "docker"
	case InstallationBinary:
		return "binary"
	case InstallationGit:
		return "git"
	case InstallationUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ParseInstallationType parses a string into InstallationType
func ParseInstallationType(s string) InstallationType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "npm", "node", "javascript":
		return InstallationNPM
	case "python", "pip", "pypi":
		return InstallationPython
	case "docker", "container":
		return InstallationDocker
	case "binary", "executable":
		return InstallationBinary
	case "git", "github":
		return InstallationGit
	default:
		return InstallationUnknown
	}
}

// MCPServer represents an MCP server in the marketplace
type MCPServer struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Author         string           `json:"author"`
	Description    string           `json:"description"`
	Version        string           `json:"version"`
	Tags           []string         `json:"tags"`
	GitHubStars    int              `json:"github_stars"`
	DownloadCount  int              `json:"download_count"`
	Status         ServerStatus     `json:"status"`
	IsInstalled    bool             `json:"is_installed"`
	Logo           string           `json:"logo"` // URL or emoji
	ToolsCount     int              `json:"tools_count"`
	ResourcesCount int              `json:"resources_count"`
	Category       MCPCategory      `json:"category"`
	InstallType    InstallationType `json:"install_type"`
	InstallCommand string           `json:"install_command"`
	GitHubURL      string           `json:"github_url"`
	Language       string           `json:"language"`
	LastUpdated    time.Time        `json:"last_updated"`
}

// Validate validates the MCPServer configuration
func (s *MCPServer) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	if strings.TrimSpace(s.Author) == "" {
		return fmt.Errorf("server author cannot be empty")
	}

	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("server description cannot be empty")
	}

	if s.InstallType == InstallationUnknown && strings.TrimSpace(s.InstallCommand) == "" {
		return fmt.Errorf("install command required for unknown installation type")
	}

	// Validate GitHub URL format if provided
	if s.GitHubURL != "" && !strings.Contains(s.GitHubURL, "github.com") {
		return fmt.Errorf("invalid GitHub URL format")
	}

	return nil
}

// IsValid returns true if the server configuration is valid
func (s *MCPServer) IsValid() bool {
	return s.Validate() == nil
}

// GetCategoryColor returns the color for the server category
func (s *MCPServer) GetCategoryColor() lipgloss.Color {
	switch s.Category {
	case CategoryOfficial:
		return lipgloss.Color("46") // Green
	case CategoryReference:
		return lipgloss.Color("33") // Blue
	case CategoryThirdParty:
		return lipgloss.Color("226") // Yellow
	case CategoryCommunity:
		return lipgloss.Color("201") // Pink
	case CategoryFrameworks:
		return lipgloss.Color("196") // Red
	default:
		return lipgloss.Color("240") // Gray
	}
}

// GetInstallIcon returns an icon for the installation type
func (s *MCPServer) GetInstallIcon() string {
	switch s.InstallType {
	case InstallationNPM:
		return "📦"
	case InstallationPython:
		return "🐍"
	case InstallationDocker:
		return "🐳"
	case InstallationBinary:
		return "⚙️"
	case InstallationGit:
		return "🔗"
	default:
		return "❓"
	}
}

// NewMarketplaceModel creates a new marketplace model
func NewMarketplaceModel() *MarketplaceModel {
	mm := &MarketplaceModel{
		focused: false,
	}

	// Create tab manager with three tabs
	mm.tabManager = tabs.NewTabManager()

	// Add Marketplace tab
	mm.tabManager.AddTab("marketplace", "Marketplace", mm.createMarketplaceTab())

	// Add Installed tab
	mm.tabManager.AddTab("installed", "Installed", mm.createInstalledTab())

	// Add Manual Configuration tab
	mm.tabManager.AddTab("manual", "Manual Config", mm.createManualConfigTab())

	return mm
}

// createMarketplaceTab creates the marketplace tab content
func (mm *MarketplaceModel) createMarketplaceTab() tea.Model {
	return NewMarketplaceTabModel(mm.getMockServers())
}

// createInstalledTab creates the installed servers tab content
func (mm *MarketplaceModel) createInstalledTab() tea.Model {
	return NewInstalledTabModel(mm.getMockInstalledServers())
}

// createManualConfigTab creates the manual configuration tab content
func (mm *MarketplaceModel) createManualConfigTab() tea.Model {
	return NewManualConfigTabModel()
}

// getMockServers returns mock server data for development
func (mm *MarketplaceModel) getMockServers() []MCPServer {
	return []MCPServer{
		{
			ID:             "azure-services",
			Name:           "Azure Services",
			Author:         "Azure",
			Description:    "Comprehensive Azure cloud services integration with resource management capabilities",
			Version:        "1.2.0",
			Tags:           []string{"cloud", "azure", "infrastructure"},
			GitHubStars:    245,
			DownloadCount:  1200,
			Status:         StatusOff,
			IsInstalled:    false,
			Logo:           "☁️",
			ToolsCount:     12,
			ResourcesCount: 8,
			Category:       CategoryOfficial,
			InstallType:    InstallationNPM,
			InstallCommand: "npm install @azure/mcp-server",
			GitHubURL:      "https://github.com/Azure/mcp-server",
			Language:       "TypeScript",
			LastUpdated:    time.Now().AddDate(0, 0, -7),
		},
		{
			ID:             "github-integration",
			Name:           "GitHub Integration",
			Author:         "GitHub",
			Description:    "Complete GitHub API integration for repository management and automation",
			Version:        "2.1.0",
			Tags:           []string{"git", "github", "automation"},
			GitHubStars:    567,
			DownloadCount:  3400,
			Status:         StatusOff,
			IsInstalled:    false,
			Logo:           "🐙",
			ToolsCount:     18,
			ResourcesCount: 12,
			Category:       CategoryOfficial,
			InstallType:    InstallationNPM,
			InstallCommand: "npm install @github/mcp-server",
			GitHubURL:      "https://github.com/github/mcp-server",
			Language:       "JavaScript",
			LastUpdated:    time.Now().AddDate(0, 0, -3),
		},
		{
			ID:             "filesystem-tools",
			Name:           "Filesystem Tools",
			Author:         "Community",
			Description:    "Advanced filesystem operations and file management utilities",
			Version:        "1.0.5",
			Tags:           []string{"filesystem", "tools", "utilities"},
			GitHubStars:    123,
			DownloadCount:  890,
			Status:         StatusOff,
			IsInstalled:    false,
			Logo:           "📁",
			ToolsCount:     8,
			ResourcesCount: 4,
			Category:       CategoryCommunity,
			InstallType:    InstallationPython,
			InstallCommand: "pip install filesystem-mcp",
			GitHubURL:      "https://github.com/community/filesystem-mcp",
			Language:       "Python",
			LastUpdated:    time.Now().AddDate(0, 0, -14),
		},
	}
}

// getMockInstalledServers returns mock installed server data
func (mm *MarketplaceModel) getMockInstalledServers() []MCPServer {
	return []MCPServer{
		{
			ID:             "console-ninja",
			Name:           "console-ninja",
			Author:         "WallabyJS",
			Description:    "Runtime error and log monitoring for JavaScript applications",
			Version:        "1.0.0",
			Tags:           []string{"debugging", "javascript", "monitoring"},
			GitHubStars:    89,
			DownloadCount:  456,
			Status:         StatusOn,
			IsInstalled:    true,
			Logo:           "🥷",
			ToolsCount:     4,
			ResourcesCount: 0,
			Category:       CategoryThirdParty,
			InstallType:    InstallationNPM,
			InstallCommand: "npm install console-ninja",
			GitHubURL:      "https://github.com/wallabyjs/console-ninja",
			Language:       "JavaScript",
			LastUpdated:    time.Now().AddDate(0, 0, -2),
		},
		{
			ID:             "context7",
			Name:           "context7",
			Author:         "Community",
			Description:    "Context management and analysis tools",
			Version:        "0.8.2",
			Tags:           []string{"context", "analysis"},
			GitHubStars:    34,
			DownloadCount:  123,
			Status:         StatusError,
			IsInstalled:    true,
			Logo:           "🔍",
			ToolsCount:     6,
			ResourcesCount: 3,
		},
		{
			ID:             "server-memory",
			Name:           "server-memory",
			Author:         "DevTools",
			Description:    "Memory management and monitoring utilities",
			Version:        "1.1.0",
			Tags:           []string{"memory", "monitoring", "performance"},
			GitHubStars:    67,
			DownloadCount:  234,
			Status:         StatusOn,
			IsInstalled:    true,
			Logo:           "🧠",
			ToolsCount:     3,
			ResourcesCount: 2,
		},
	}
}

// SetSize sets the size of the marketplace model
func (mm *MarketplaceModel) SetSize(width, height int) {
	mm.width = width
	mm.height = height
	if mm.tabManager != nil {
		mm.tabManager.SetSize(width, height-2) // Account for title
	}
}

// Focus sets focus on the marketplace model
func (mm *MarketplaceModel) Focus() {
	mm.focused = true
}

// Blur removes focus from the marketplace model
func (mm *MarketplaceModel) Blur() {
	mm.focused = false
}

// Init implements tea.Model
func (mm *MarketplaceModel) Init() tea.Cmd {
	if mm.tabManager != nil {
		return mm.tabManager.Init()
	}
	return nil
}

// Update implements tea.Model
func (mm *MarketplaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !mm.focused {
		return mm, nil
	}

	if mm.tabManager != nil {
		var cmd tea.Cmd
		mm.tabManager, cmd = mm.tabManager.Update(msg)
		return mm, cmd
	}

	return mm, nil
}

// View implements tea.Model
func (mm *MarketplaceModel) View() string {
	if mm.tabManager == nil {
		return "Loading MCP Marketplace..."
	}

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme().Primary()).
		Render("🔌 MCP Servers")

	// Tab content
	content := mm.tabManager.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
	)
}
