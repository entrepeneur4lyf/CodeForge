package layout

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Sizeable interface for components that can be resized
type Sizeable interface {
	SetSize(width, height int) tea.Cmd
}

// Bindings interface for components that provide key bindings
type Bindings interface {
	BindingKeys() []key.Binding
}

// Layout represents the main application layout
type Layout struct {
	width, height int

	// Layout areas
	topBarHeight    int
	bottomBarHeight int
	sidebarWidth    int

	// Content areas
	topBar      string
	bottomBar   string
	sidebar     string
	mainContent string

	// Responsive configuration
	minWidth     int
	minHeight    int
	compactWidth int // Below this width, use compact layout
	mediumWidth  int // Below this width, use medium layout

	// Layout state
	layoutMode LayoutMode
	tooSmall   bool
}

// LayoutMode defines different responsive layout modes
type LayoutMode int

const (
	LayoutModeMinimal LayoutMode = iota // Terminal too small
	LayoutModeCompact                   // Small terminal - minimal sidebar
	LayoutModeMedium                    // Medium terminal - normal layout
	LayoutModeFull                      // Large terminal - full features
)

// NewLayout creates a new layout with default dimensions
func NewLayout() *Layout {
	return &Layout{
		topBarHeight:    1,
		bottomBarHeight: 1,
		sidebarWidth:    30,

		// Responsive configuration
		minWidth:     80,  // Minimum usable width
		minHeight:    24,  // Minimum usable height
		compactWidth: 100, // Below this, use compact layout
		mediumWidth:  140, // Below this, use medium layout

		// Default to medium layout
		layoutMode: LayoutModeMedium,
	}
}

// SetSize updates the layout dimensions and calculates responsive layout
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.calculateResponsiveLayout()
}

// calculateResponsiveLayout determines layout mode and calculates responsive dimensions
func (l *Layout) calculateResponsiveLayout() {
	// Check if terminal is too small
	l.tooSmall = l.width < l.minWidth || l.height < l.minHeight

	if l.tooSmall {
		l.layoutMode = LayoutModeMinimal
		l.sidebarWidth = 0 // No sidebar in minimal mode
		return
	}

	// Calculate responsive sidebar width and layout mode
	if l.width < l.compactWidth {
		l.layoutMode = LayoutModeCompact
		l.sidebarWidth = max(15, min(25, l.width/5)) // Smaller sidebar for compact
	} else if l.width < l.mediumWidth {
		l.layoutMode = LayoutModeMedium
		l.sidebarWidth = max(25, min(40, l.width/4)) // Standard sidebar
	} else {
		l.layoutMode = LayoutModeFull
		l.sidebarWidth = max(30, min(50, l.width/5)) // Larger sidebar for full mode
	}
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max function is defined in container.go

// GetLayoutMode returns the current layout mode
func (l *Layout) GetLayoutMode() LayoutMode {
	return l.layoutMode
}

// IsTooSmall returns whether the terminal is too small for the UI
func (l *Layout) IsTooSmall() bool {
	return l.tooSmall
}

// GetMinimumSize returns the minimum required terminal size
func (l *Layout) GetMinimumSize() (width, height int) {
	return l.minWidth, l.minHeight
}

// RenderTooSmallWarning renders a warning when terminal is too small
func (l *Layout) RenderTooSmallWarning() string {
	if !l.tooSmall {
		return ""
	}

	warning := fmt.Sprintf("Terminal too small: %dx%d", l.width, l.height)
	requirement := fmt.Sprintf("Minimum required: %dx%d", l.minWidth, l.minHeight)
	instruction := "Please resize your terminal window"

	// Center the content
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		"⚠️  CodeForge",
		"",
		warning,
		requirement,
		"",
		instruction,
		"",
		"Press 'q' to quit",
		"",
	)

	// Style and center in available space
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")). // Red warning color
		Bold(true).
		Align(lipgloss.Center).
		Width(l.width).
		Height(l.height)

	return style.Render(content)
}

// GetContentDimensions returns the dimensions for the main content area
func (l *Layout) GetContentDimensions() (width, height int) {
	contentWidth := l.width - l.sidebarWidth
	contentHeight := l.height - l.topBarHeight - l.bottomBarHeight
	return contentWidth, contentHeight
}

// GetSidebarDimensions returns the dimensions for the sidebar
func (l *Layout) GetSidebarDimensions() (width, height int) {
	sidebarHeight := l.height - l.topBarHeight - l.bottomBarHeight
	return l.sidebarWidth, sidebarHeight
}

// SetTopBar sets the content for the top bar
func (l *Layout) SetTopBar(content string) {
	l.topBar = content
}

// SetBottomBar sets the content for the bottom bar
func (l *Layout) SetBottomBar(content string) {
	l.bottomBar = content
}

// SetSidebar sets the content for the sidebar
func (l *Layout) SetSidebar(content string) {
	l.sidebar = content
}

// SetMainContent sets the content for the main area
func (l *Layout) SetMainContent(content string) {
	l.mainContent = content
}

// Render renders the complete layout
func (l *Layout) Render() string {
	// Create the main content area (sidebar + content)
	contentWidth, contentHeight := l.GetContentDimensions()
	sidebarWidth, sidebarHeight := l.GetSidebarDimensions()

	// Ensure content fits within dimensions
	sidebar := l.fitContent(l.sidebar, sidebarWidth, sidebarHeight)
	mainContent := l.fitContent(l.mainContent, contentWidth, contentHeight)

	// Join sidebar and main content horizontally
	middleSection := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		mainContent,
	)

	// Create the complete layout
	var sections []string

	// Add top bar if present
	if l.topBar != "" {
		topBar := l.fitContent(l.topBar, l.width, l.topBarHeight)
		sections = append(sections, topBar)
	}

	// Add middle section
	sections = append(sections, middleSection)

	// Add bottom bar if present
	if l.bottomBar != "" {
		bottomBar := l.fitContent(l.bottomBar, l.width, l.bottomBarHeight)
		sections = append(sections, bottomBar)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// fitContent ensures content fits within the specified dimensions
func (l *Layout) fitContent(content string, width, height int) string {
	if content == "" {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Render("")
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(content)
}

// KeyMapToSlice converts a key map to a slice of key bindings
func KeyMapToSlice(keyMap interface{}) []key.Binding {
	// This would use reflection to extract key bindings from a struct
	// For now, return empty slice - implement based on specific key map types
	return []key.Binding{}
}

// Split creates a split layout (horizontal or vertical)
type Split struct {
	direction lipgloss.Position
	ratio     float64
	left      string
	right     string
	width     int
	height    int
}

// NewHorizontalSplit creates a horizontal split layout
func NewHorizontalSplit(ratio float64) *Split {
	return &Split{
		direction: lipgloss.Left,
		ratio:     ratio,
	}
}

// NewVerticalSplit creates a vertical split layout
func NewVerticalSplit(ratio float64) *Split {
	return &Split{
		direction: lipgloss.Top,
		ratio:     ratio,
	}
}

// SetSize sets the dimensions for the split
func (s *Split) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetLeft sets the content for the left/top pane
func (s *Split) SetLeft(content string) {
	s.left = content
}

// SetRight sets the content for the right/bottom pane
func (s *Split) SetRight(content string) {
	s.right = content
}

// Render renders the split layout
func (s *Split) Render() string {
	if s.direction == lipgloss.Left {
		// Horizontal split
		leftWidth := int(float64(s.width) * s.ratio)
		rightWidth := s.width - leftWidth

		leftContent := lipgloss.NewStyle().
			Width(leftWidth).
			Height(s.height).
			Render(s.left)

		rightContent := lipgloss.NewStyle().
			Width(rightWidth).
			Height(s.height).
			Render(s.right)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
	} else {
		// Vertical split
		topHeight := int(float64(s.height) * s.ratio)
		bottomHeight := s.height - topHeight

		topContent := lipgloss.NewStyle().
			Width(s.width).
			Height(topHeight).
			Render(s.left)

		bottomContent := lipgloss.NewStyle().
			Width(s.width).
			Height(bottomHeight).
			Render(s.right)

		return lipgloss.JoinVertical(lipgloss.Left, topContent, bottomContent)
	}
}
