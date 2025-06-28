package mcp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// SortOption represents different sorting options
type SortOption int

const (
	SortByName SortOption = iota
	SortByStars
	SortByDownloads
	SortByNewest
	SortByCategory
)

// String returns the string representation of SortOption
func (s SortOption) String() string {
	switch s {
	case SortByName:
		return "Name"
	case SortByStars:
		return "GitHub Stars"
	case SortByDownloads:
		return "Most Installs"
	case SortByNewest:
		return "Newest"
	case SortByCategory:
		return "Category"
	default:
		return "Name"
	}
}

// MarketplaceTabModel represents the marketplace tab
type MarketplaceTabModel struct {
	width           int
	height          int
	servers         []MCPServer
	filteredServers []MCPServer
	searchInput     textinput.Model
	serverList      list.Model
	focused         bool
	showSearch      bool
	showFilters     bool
	fetcher         *mcp.RepositoryFetcher
	loading         bool
	error           error

	// Filter and sort state
	selectedCategory MCPCategory
	categoryIndex    int
	sortOption       SortOption
	sortIndex        int

	// Available options
	categories  []MCPCategory
	sortOptions []SortOption

	// Server detail modal
	showServerDetail bool
	selectedServer   *MCPServer
}

// ServerItem implements list.Item for the server list
type ServerItem struct {
	server MCPServer
}

// FilterValue implements list.Item
func (si ServerItem) FilterValue() string {
	return si.server.Name + " " + si.server.Description + " " + strings.Join(si.server.Tags, " ")
}

// Title implements list.Item
func (si ServerItem) Title() string {
	return si.server.Name
}

// Description implements list.Item
func (si ServerItem) Description() string {
	return si.server.Description
}

// NewMarketplaceTabModel creates a new marketplace tab model
func NewMarketplaceTabModel(servers []MCPServer) *MarketplaceTabModel {
	// Create search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search MCPs..."
	searchInput.CharLimit = 100
	searchInput.Width = 50

	// Convert servers to list items
	items := make([]list.Item, len(servers))
	for i, server := range servers {
		items[i] = ServerItem{server: server}
	}

	// Create server list
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(4) // Taller items for server cards

	serverList := list.New(items, delegate, 0, 0)
	serverList.Title = "Available MCP Servers"
	serverList.SetShowStatusBar(true)
	serverList.SetFilteringEnabled(true)
	serverList.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme().Primary())

	// Initialize categories and sort options
	categories := []MCPCategory{
		CategoryGeneral, // "All Categories"
		CategoryOfficial,
		CategoryReference,
		CategoryThirdParty,
		CategoryCommunity,
		CategoryFrameworks,
	}

	sortOptions := []SortOption{
		SortByName,
		SortByStars,
		SortByDownloads,
		SortByNewest,
		SortByCategory,
	}

	model := &MarketplaceTabModel{
		servers:          servers,
		filteredServers:  servers, // Initially show all servers
		searchInput:      searchInput,
		serverList:       serverList,
		focused:          false,
		showSearch:       false,
		showFilters:      false,
		selectedCategory: CategoryGeneral, // "All Categories"
		categoryIndex:    0,
		sortOption:       SortByName,
		sortIndex:        0,
		categories:       categories,
		sortOptions:      sortOptions,
	}

	// Apply initial sorting
	model.applySortAndFilter()

	return model
}

// applySortAndFilter applies the current sort and filter settings
func (mtm *MarketplaceTabModel) applySortAndFilter() {
	// Start with all servers
	filtered := make([]MCPServer, 0, len(mtm.servers))

	// Apply category filter
	for _, server := range mtm.servers {
		if mtm.selectedCategory == CategoryGeneral || server.Category == mtm.selectedCategory {
			filtered = append(filtered, server)
		}
	}

	// Apply sorting
	mtm.sortServers(filtered)

	// Update filtered servers
	mtm.filteredServers = filtered

	// Update list items
	items := make([]list.Item, len(filtered))
	for i, server := range filtered {
		items[i] = ServerItem{server: server}
	}
	mtm.serverList.SetItems(items)
}

// sortServers sorts the servers based on the current sort option
func (mtm *MarketplaceTabModel) sortServers(servers []MCPServer) {
	switch mtm.sortOption {
	case SortByName:
		sort.Slice(servers, func(i, j int) bool {
			return strings.ToLower(servers[i].Name) < strings.ToLower(servers[j].Name)
		})
	case SortByStars:
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].GitHubStars > servers[j].GitHubStars
		})
	case SortByDownloads:
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].DownloadCount > servers[j].DownloadCount
		})
	case SortByNewest:
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].LastUpdated.After(servers[j].LastUpdated)
		})
	case SortByCategory:
		sort.Slice(servers, func(i, j int) bool {
			if servers[i].Category == servers[j].Category {
				return strings.ToLower(servers[i].Name) < strings.ToLower(servers[j].Name)
			}
			return servers[i].Category < servers[j].Category
		})
	}
}

// SetSize sets the size of the marketplace tab
func (mtm *MarketplaceTabModel) SetSize(width, height int) {
	mtm.width = width
	mtm.height = height

	// Update search input width
	mtm.searchInput.Width = width - 4

	// Update server list size
	listHeight := height - 6 // Account for search bar and padding
	if mtm.showSearch {
		listHeight -= 3
	}
	mtm.serverList.SetSize(width-2, listHeight)
}

// Init implements tea.Model
func (mtm *MarketplaceTabModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (mtm *MarketplaceTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			// Toggle search
			mtm.showSearch = !mtm.showSearch
			if mtm.showSearch {
				mtm.searchInput.Focus()
				return mtm, textinput.Blink
			} else {
				mtm.searchInput.Blur()
				mtm.searchInput.SetValue("")
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if mtm.showServerDetail {
				mtm.showServerDetail = false
				mtm.selectedServer = nil
			} else if mtm.showSearch {
				mtm.showSearch = false
				mtm.searchInput.Blur()
				mtm.searchInput.SetValue("")
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if mtm.showSearch {
				// Apply search filter
				searchTerm := mtm.searchInput.Value()
				mtm.filterServers(searchTerm)
				mtm.showSearch = false
				mtm.searchInput.Blur()
			} else if mtm.showServerDetail {
				// Close server detail modal
				mtm.showServerDetail = false
				mtm.selectedServer = nil
			} else {
				// Show server details
				if selectedItem, ok := mtm.serverList.SelectedItem().(ServerItem); ok {
					mtm.selectedServer = &selectedItem.server
					mtm.showServerDetail = true
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("i"))):
			// Install selected server
			if mtm.showServerDetail && mtm.selectedServer != nil {
				return mtm, mtm.installServer(*mtm.selectedServer)
			} else if selectedItem, ok := mtm.serverList.SelectedItem().(ServerItem); ok {
				return mtm, mtm.installServer(selectedItem.server)
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			// Show server details
			if !mtm.showServerDetail {
				if selectedItem, ok := mtm.serverList.SelectedItem().(ServerItem); ok {
					mtm.selectedServer = &selectedItem.server
					mtm.showServerDetail = true
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("f"))):
			// Toggle filters
			mtm.showFilters = !mtm.showFilters

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// Cycle through categories
			mtm.categoryIndex = (mtm.categoryIndex + 1) % len(mtm.categories)
			mtm.selectedCategory = mtm.categories[mtm.categoryIndex]
			mtm.applySortAndFilter()

		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			// Cycle through sort options
			mtm.sortIndex = (mtm.sortIndex + 1) % len(mtm.sortOptions)
			mtm.sortOption = mtm.sortOptions[mtm.sortIndex]
			mtm.applySortAndFilter()

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Reset filters and sorting
			mtm.selectedCategory = CategoryGeneral
			mtm.categoryIndex = 0
			mtm.sortOption = SortByName
			mtm.sortIndex = 0
			mtm.applySortAndFilter()
		}

		// Handle input in search mode
		if mtm.showSearch {
			var cmd tea.Cmd
			mtm.searchInput, cmd = mtm.searchInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			// Handle list navigation
			var cmd tea.Cmd
			mtm.serverList, cmd = mtm.serverList.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		mtm.SetSize(msg.Width, msg.Height)
	}

	return mtm, tea.Batch(cmds...)
}

// filterServers filters the server list based on search term
func (mtm *MarketplaceTabModel) filterServers(searchTerm string) {
	if searchTerm == "" {
		// Reset to current filter/sort settings
		mtm.applySortAndFilter()
		return
	}

	// Apply search on top of current filters
	var filtered []MCPServer
	searchLower := strings.ToLower(searchTerm)

	for _, server := range mtm.filteredServers {
		if strings.Contains(strings.ToLower(server.Name), searchLower) ||
			strings.Contains(strings.ToLower(server.Description), searchLower) ||
			strings.Contains(strings.ToLower(server.Author), searchLower) ||
			mtm.containsTag(server.Tags, searchLower) {
			filtered = append(filtered, server)
		}
	}

	// Update list items with search results
	items := make([]list.Item, len(filtered))
	for i, server := range filtered {
		items[i] = ServerItem{server: server}
	}
	mtm.serverList.SetItems(items)
}

// containsTag checks if any tag contains the search term
func (mtm *MarketplaceTabModel) containsTag(tags []string, searchTerm string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), searchTerm) {
			return true
		}
	}
	return false
}

// installServer handles server installation
func (mtm *MarketplaceTabModel) installServer(server MCPServer) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual installation logic
		return ServerInstallMsg{ServerID: server.ID, Success: true}
	}
}

// ServerInstallMsg represents a server installation message
type ServerInstallMsg struct {
	ServerID string
	Success  bool
	Error    error
}

// View implements tea.Model
func (mtm *MarketplaceTabModel) View() string {
	// Show server detail modal if active
	if mtm.showServerDetail && mtm.selectedServer != nil {
		return mtm.renderServerDetailModal()
	}

	var sections []string

	// Filters and sorting bar
	filtersBar := mtm.renderFiltersBar()
	sections = append(sections, filtersBar)
	sections = append(sections, "") // Empty line

	// Search bar (if shown)
	if mtm.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.CurrentTheme().Primary()).
			Padding(0, 1)

		searchBar := searchStyle.Render(mtm.searchInput.View())
		sections = append(sections, searchBar)
		sections = append(sections, "") // Empty line
	}

	// Server list with custom rendering
	if len(mtm.serverList.Items()) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().Secondary()).
			Italic(true).
			Align(lipgloss.Center)

		empty := emptyStyle.Render("No servers found. Try adjusting your search.")
		sections = append(sections, empty)
	} else {
		// Custom server card rendering
		sections = append(sections, mtm.renderServerCards())
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true)

	var helpText string
	if mtm.showSearch {
		helpText = "Enter: apply filter • Esc: cancel search"
	} else {
		helpText = "/: search • c: category • s: sort • f: filters • r: reset • i/Enter: install • ↑↓: navigate"
	}

	sections = append(sections, "")
	sections = append(sections, helpStyle.Render(helpText))

	return strings.Join(sections, "\n")
}

// renderFiltersBar renders the filters and sorting bar
func (mtm *MarketplaceTabModel) renderFiltersBar() string {
	// Category filter
	categoryStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Primary()).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Margin(0, 1, 0, 0)

	categoryText := "All Categories"
	if mtm.selectedCategory != CategoryGeneral {
		categoryText = mtm.selectedCategory.String()
	}
	categoryFilter := categoryStyle.Render(fmt.Sprintf("📂 %s", categoryText))

	// Sort option
	sortStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Secondary()).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Margin(0, 1, 0, 0)

	sortFilter := sortStyle.Render(fmt.Sprintf("🔄 %s", mtm.sortOption.String()))

	// Server count
	countStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true)

	serverCount := countStyle.Render(fmt.Sprintf("(%d servers)", len(mtm.filteredServers)))

	// Combine filters
	filtersLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		categoryFilter,
		sortFilter,
		serverCount,
	)

	return filtersLine
}

// renderServerCards renders server cards in a custom format
func (mtm *MarketplaceTabModel) renderServerCards() string {
	var cards []string

	// Get visible items from the list
	items := mtm.serverList.Items()
	start, end := mtm.serverList.Paginator.GetSliceBounds(len(items))

	for i := start; i < end; i++ {
		if serverItem, ok := items[i].(ServerItem); ok {
			card := mtm.renderServerCard(serverItem.server, i == mtm.serverList.Index())
			cards = append(cards, card)
		}
	}

	return strings.Join(cards, "\n\n")
}

// renderServerCard renders a single server card
func (mtm *MarketplaceTabModel) renderServerCard(server MCPServer, selected bool) string {
	// Card styling
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(mtm.width - 4)

	if selected {
		cardStyle = cardStyle.BorderForeground(theme.CurrentTheme().Primary())
	} else {
		cardStyle = cardStyle.BorderForeground(theme.CurrentTheme().Secondary())
	}

	// Header with logo, name, and install button
	headerStyle := lipgloss.NewStyle().Bold(true)
	header := fmt.Sprintf("%s %s", server.Logo, server.Name)

	// Author and stats
	statsStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary())
	stats := fmt.Sprintf("by %s • ⭐ %d • ⬇ %d", server.Author, server.GitHubStars, server.DownloadCount)

	// Description
	descStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text())
	description := descStyle.Render(server.Description)

	// Tags
	var tagStrings []string
	tagStyle := lipgloss.NewStyle().
		Background(theme.CurrentTheme().Primary()).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Margin(0, 1, 0, 0)

	for _, tag := range server.Tags {
		tagStrings = append(tagStrings, tagStyle.Render(tag))
	}
	tags := strings.Join(tagStrings, "")

	// Install button
	installStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("46")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 2).
		Align(lipgloss.Center)

	var installButton string
	if server.IsInstalled {
		installButton = installStyle.Background(lipgloss.Color("240")).Render("INSTALLED")
	} else {
		installButton = installStyle.Render("INSTALL")
	}

	// Combine elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(header),
		statsStyle.Render(stats),
		"",
		description,
		"",
		tags,
		"",
		installButton,
	)

	return cardStyle.Render(content)
}
