package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// Tab represents a single tab
type Tab struct {
	ID      string
	Title   string
	Content tea.Model
	Active  bool
}

// TabManager manages multiple tabs
type TabManager struct {
	tabs      []Tab
	activeTab int
	width     int
	height    int
	focused   bool
	tabHeight int
	animValue float64 // Animation value for smooth transitions
}

// TabSwitchedMsg is sent when a tab is switched
type TabSwitchedMsg struct {
	TabID string
	Index int
}

// NewTabManager creates a new tab manager
func NewTabManager() *TabManager {
	return &TabManager{
		tabs:      make([]Tab, 0),
		activeTab: 0,
		tabHeight: 3, // Height for tab bar
	}
}

// AddTab adds a new tab
func (tm *TabManager) AddTab(id, title string, content tea.Model) {
	tab := Tab{
		ID:      id,
		Title:   title,
		Content: content,
		Active:  len(tm.tabs) == 0, // First tab is active by default
	}

	tm.tabs = append(tm.tabs, tab)

	// If this is the first tab, make it active
	if len(tm.tabs) == 1 {
		tm.activeTab = 0
	}
}

// RemoveTab removes a tab by ID
func (tm *TabManager) RemoveTab(id string) {
	for i, tab := range tm.tabs {
		if tab.ID == id {
			tm.tabs = append(tm.tabs[:i], tm.tabs[i+1:]...)

			// Adjust active tab if necessary
			if tm.activeTab >= len(tm.tabs) {
				tm.activeTab = len(tm.tabs) - 1
			}
			if tm.activeTab < 0 {
				tm.activeTab = 0
			}

			tm.updateActiveStates()
			break
		}
	}
}

// SwitchToTab switches to a tab by ID
func (tm *TabManager) SwitchToTab(id string) tea.Cmd {
	for i, tab := range tm.tabs {
		if tab.ID == id {
			tm.activeTab = i
			tm.updateActiveStates()
			return func() tea.Msg {
				return TabSwitchedMsg{
					TabID: id,
					Index: i,
				}
			}
		}
	}
	return nil
}

// SwitchToIndex switches to a tab by index
func (tm *TabManager) SwitchToIndex(index int) tea.Cmd {
	if index >= 0 && index < len(tm.tabs) {
		tm.activeTab = index
		tm.updateActiveStates()
		return func() tea.Msg {
			return TabSwitchedMsg{
				TabID: tm.tabs[index].ID,
				Index: index,
			}
		}
	}
	return nil
}

// GetActiveTab returns the currently active tab
func (tm *TabManager) GetActiveTab() *Tab {
	if tm.activeTab >= 0 && tm.activeTab < len(tm.tabs) {
		return &tm.tabs[tm.activeTab]
	}
	return nil
}

// GetActiveContent returns the content of the active tab
func (tm *TabManager) GetActiveContent() tea.Model {
	if tab := tm.GetActiveTab(); tab != nil {
		return tab.Content
	}
	return nil
}

// Init implements tea.Model
func (tm *TabManager) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize all tab contents
	for i := range tm.tabs {
		if tm.tabs[i].Content != nil {
			if cmd := tm.tabs[i].Content.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (tm *TabManager) Update(msg tea.Msg) (*TabManager, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tm.focused {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+1"))):
				return tm, tm.SwitchToIndex(0)
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+2"))):
				return tm, tm.SwitchToIndex(1)
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+3"))):
				return tm, tm.SwitchToIndex(2)
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+4"))):
				return tm, tm.SwitchToIndex(3)
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+5"))):
				return tm, tm.SwitchToIndex(4)
			case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
				nextTab := (tm.activeTab + 1) % len(tm.tabs)
				return tm, tm.SwitchToIndex(nextTab)
			case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
				prevTab := tm.activeTab - 1
				if prevTab < 0 {
					prevTab = len(tm.tabs) - 1
				}
				return tm, tm.SwitchToIndex(prevTab)
			}
		}

		// Forward key events to active tab content
		if activeTab := tm.GetActiveTab(); activeTab != nil {
			var cmd tea.Cmd
			activeTab.Content, cmd = activeTab.Content.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		tm.width = msg.Width
		tm.height = msg.Height

		// Update size for all tab contents
		contentHeight := tm.height - tm.tabHeight
		for i := range tm.tabs {
			if sizeable, ok := tm.tabs[i].Content.(interface {
				SetSize(width, height int)
			}); ok {
				sizeable.SetSize(tm.width, contentHeight)
			}
		}

	default:
		// Forward other messages to active tab content
		if activeTab := tm.GetActiveTab(); activeTab != nil {
			var cmd tea.Cmd
			activeTab.Content, cmd = activeTab.Content.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return tm, tea.Batch(cmds...)
}

// View implements tea.Model
func (tm *TabManager) View() string {
	if len(tm.tabs) == 0 {
		return "No tabs available"
	}

	// Render tab bar
	tabBar := tm.renderTabBar()

	// Render active content
	var content string
	if activeTab := tm.GetActiveTab(); activeTab != nil {
		content = activeTab.Content.View()
	}

	// Combine tab bar and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
	)
}

// renderTabBar renders the tab bar
func (tm *TabManager) renderTabBar() string {
	t := theme.CurrentTheme()
	var renderedTabs []string

	for i, tab := range tm.tabs {
		var style lipgloss.Style
		if i == tm.activeTab {
			style = styles.TabActiveStyle()

			// Apply animation effects to active tab
			if tm.animValue > 0 {
				// Subtle glow effect during transition
				glowIntensity := int(tm.animValue * 255)
				if glowIntensity > 255 {
					glowIntensity = 255
				}

				// Enhanced active tab styling with animation
				style = style.
					BorderForeground(t.Primary()).
					Foreground(t.Primary())
			}
		} else {
			style = styles.TabInactiveStyle()

			// Fade effect for inactive tabs during transition
			if tm.animValue > 0 {
				opacity := 1.0 - (tm.animValue * 0.3) // Slight fade
				if opacity < 0.7 {
					opacity = 0.7
				}

				// Apply fade effect
				style = style.Foreground(t.TextMuted())
			}
		}

		renderedTabs = append(renderedTabs, style.Render(tab.Title))
	}

	// Add gap to fill remaining space
	tabContent := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	remainingWidth := tm.width - lipgloss.Width(tabContent)

	if remainingWidth > 0 {
		gap := lipgloss.NewStyle().
			Background(t.BackgroundDarker()).
			Width(remainingWidth).
			Render("")
		tabContent = lipgloss.JoinHorizontal(lipgloss.Top, tabContent, gap)
	}

	return styles.TabBarStyle().
		Width(tm.width).
		Render(tabContent)
}

// updateActiveStates updates the active state of all tabs
func (tm *TabManager) updateActiveStates() {
	for i := range tm.tabs {
		tm.tabs[i].Active = i == tm.activeTab
	}
}

// SetFocused sets the focus state
func (tm *TabManager) SetFocused(focused bool) {
	tm.focused = focused
}

// SetAnimationValue sets the current animation value
func (tm *TabManager) SetAnimationValue(value float64) {
	tm.animValue = value
}

// GetActiveTabIndex returns the index of the active tab
func (tm *TabManager) GetActiveTabIndex() int {
	return tm.activeTab
}

// GetTabCount returns the number of tabs
func (tm *TabManager) GetTabCount() int {
	return len(tm.tabs)
}

// SetSize sets the dimensions
func (tm *TabManager) SetSize(width, height int) {
	tm.width = width
	tm.height = height

	// Update size for all tab contents
	contentHeight := tm.height - tm.tabHeight
	for i := range tm.tabs {
		if sizeable, ok := tm.tabs[i].Content.(interface {
			SetSize(width, height int)
		}); ok {
			sizeable.SetSize(tm.width, contentHeight)
		}
	}
}

// GetTabByID returns a tab by its ID
func (tm *TabManager) GetTabByID(id string) *Tab {
	for i := range tm.tabs {
		if tm.tabs[i].ID == id {
			return &tm.tabs[i]
		}
	}
	return nil
}
