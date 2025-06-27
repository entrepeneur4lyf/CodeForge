package layout

import (
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
}

// NewLayout creates a new layout with default dimensions
func NewLayout() *Layout {
	return &Layout{
		topBarHeight:    1,
		bottomBarHeight: 1,
		sidebarWidth:    30,
	}
}

// SetSize updates the layout dimensions and calculates responsive sidebar width
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height

	// Calculate responsive sidebar width
	// Small terminals: 25% of width, min 20, max 35
	// Large terminals: 20% of width, min 25, max 50
	if width < 80 {
		l.sidebarWidth = max(20, min(35, width/4))
	} else if width < 120 {
		l.sidebarWidth = max(25, min(40, width/4))
	} else {
		l.sidebarWidth = max(30, min(50, width/5))
	}
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
