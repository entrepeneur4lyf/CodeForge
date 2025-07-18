package dialog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/search"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// SearchDialog is a simple file and text search dialog
type SearchDialog struct {
	theme       theme.Theme
	searchInput textinput.Model
	results     []SearchResult
	selected    int
	width       int
	height      int
	searchType  SearchType
	searcher    *search.Searcher
	searching   bool
	searchErr   error
}

// SearchType defines the type of search
type SearchType int

const (
	FileSearch SearchType = iota
	TextSearch
)

// SearchResult represents a search result
type SearchResult struct {
	Path    string
	Line    int
	Content string
	Match   string
}

// SearchSelectedMsg is sent when a search result is selected
type SearchSelectedMsg struct {
	Result SearchResult
}

// SearchResultsMsg is sent when search results are ready
type SearchResultsMsg struct {
	Results []SearchResult
	Err     error
}

// NewSearchDialog creates a new search dialog
func NewSearchDialog(theme theme.Theme, searchType SearchType) tea.Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	if searchType == FileSearch {
		ti.Placeholder = "Search for files..."
	} else {
		ti.Placeholder = "Search in files..."
	}

	return &SearchDialog{
		theme:       theme,
		searchInput: ti,
		searchType:  searchType,
		results:     []SearchResult{},
		searcher:    search.NewSearcher(),
	}
}

func (s *SearchDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (s *SearchDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.searchInput.Width = min(s.width-8, 60)

	case SearchResultsMsg:
		s.searching = false
		if msg.Err != nil {
			s.searchErr = msg.Err
			s.results = nil
		} else {
			s.results = msg.Results
			s.selected = 0
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, searchKeys.Cancel):
			return s, func() tea.Msg { return DialogCloseMsg{} }

		case key.Matches(msg, searchKeys.Select):
			if s.selected >= 0 && s.selected < len(s.results) {
				return s, func() tea.Msg {
					return SearchSelectedMsg{Result: s.results[s.selected]}
				}
			}

		case key.Matches(msg, searchKeys.Up):
			if s.selected > 0 {
				s.selected--
			}

		case key.Matches(msg, searchKeys.Down):
			if s.selected < len(s.results)-1 {
				s.selected++
			}

		default:
			// Update search input
			var cmd tea.Cmd
			s.searchInput, cmd = s.searchInput.Update(msg)
			cmds = append(cmds, cmd)

			// Perform search on input change
			if s.searchInput.Value() != "" {
				cmds = append(cmds, s.performSearch())
			} else {
				s.results = []SearchResult{}
				s.selected = 0
			}
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SearchDialog) View() string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	// Calculate dialog dimensions
	dialogWidth := min(s.width-4, 80)
	dialogHeight := min(s.height-4, 30)

	// Build content
	var content strings.Builder

	// Title
	title := "File Search"
	if s.searchType == TextSearch {
		title = "Text Search"
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(s.theme.TextEmphasized()).
		Width(dialogWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Search input
	inputStyle := lipgloss.NewStyle().Width(dialogWidth - 4).Align(lipgloss.Center)
	content.WriteString(inputStyle.Render(s.searchInput.View()))
	content.WriteString("\n\n")

	// Results or status
	if s.searching {
		searchingStyle := lipgloss.NewStyle().
			Foreground(s.theme.Info()).
			Width(dialogWidth - 4).
			Align(lipgloss.Center)
		content.WriteString(searchingStyle.Render("🔍 Searching..."))
	} else if s.searchErr != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(s.theme.Error()).
			Width(dialogWidth - 4).
			Align(lipgloss.Center)
		content.WriteString(errorStyle.Render(fmt.Sprintf("❌ Error: %v", s.searchErr)))
	} else if len(s.results) > 0 {
		resultsView := s.renderResults(dialogWidth-4, dialogHeight-10)
		content.WriteString(resultsView)
	} else if s.searchInput.Value() != "" {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(s.theme.TextMuted()).
			Width(dialogWidth - 4).
			Align(lipgloss.Center)
		content.WriteString(noResultsStyle.Render("No results found"))
	}

	content.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(s.theme.TextMuted()).
		Width(dialogWidth - 4).
		Align(lipgloss.Center)
	help := "↑/↓: Navigate • Enter: Select • Esc: Cancel"
	content.WriteString(helpStyle.Render(help))

	// Apply dialog style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.BorderNormal()).
		Width(dialogWidth).
		Height(dialogHeight).
		MaxWidth(dialogWidth).
		MaxHeight(dialogHeight)

	return dialogStyle.Render(content.String())
}

func (s *SearchDialog) renderResults(width, maxHeight int) string {
	var lines []string

	// Calculate visible range
	visibleItems := maxHeight - 2
	startIdx := 0
	if s.selected >= visibleItems {
		startIdx = s.selected - visibleItems + 1
	}
	endIdx := min(startIdx+visibleItems, len(s.results))

	// Scroll indicator
	if startIdx > 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.TextMuted()).
			Render("↑ more results above"))
	}

	// Render results
	for i := startIdx; i < endIdx; i++ {
		result := s.results[i]
		line := s.renderResult(result, i == s.selected, width)
		lines = append(lines, line)
	}

	// Bottom scroll indicator
	if endIdx < len(s.results) {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.TextMuted()).
			Render("↓ more results below"))
	}

	return strings.Join(lines, "\n")
}

func (s *SearchDialog) renderResult(result SearchResult, selected bool, width int) string {
	// Build result display
	var parts []string

	// File path
	pathStyle := lipgloss.NewStyle().
		Foreground(s.theme.TextMuted())
	if selected {
		pathStyle = pathStyle.Bold(true)
	}
	parts = append(parts, pathStyle.Render(result.Path))

	// Line number for text search
	if s.searchType == TextSearch && result.Line > 0 {
		lineStyle := lipgloss.NewStyle().
			Foreground(s.theme.TextMuted())
		parts = append(parts, lineStyle.Render(fmt.Sprintf(":%d", result.Line)))
	}

	// Content preview for text search
	if s.searchType == TextSearch && result.Content != "" {
		// Truncate content
		content := strings.TrimSpace(result.Content)
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		contentStyle := lipgloss.NewStyle()
		if selected {
			contentStyle = contentStyle.Foreground(s.theme.Primary())
		}

		parts = append(parts, "\n  "+contentStyle.Render(content))
	}

	// Join parts
	line := strings.Join(parts, "")

	// Apply selection style
	if selected {
		style := lipgloss.NewStyle().
			Background(s.theme.Primary()).
			Foreground(s.theme.Background()).
			Width(width)
		return style.Render(line)
	}

	style := lipgloss.NewStyle().
		Background(s.theme.Background()).
		Width(width)
	return style.Render(line)
}

func (s *SearchDialog) performSearch() tea.Cmd {
	query := s.searchInput.Value()
	if query == "" {
		return nil
	}

	s.searching = true
	s.searchErr = nil

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var results []SearchResult

		if s.searchType == FileSearch {
			// Search for files
			files, err := s.searcher.SearchFiles(ctx, query, ".")
			if err != nil {
				return SearchResultsMsg{Err: err}
			}
			
			// Convert to SearchResult
			for _, file := range files {
				results = append(results, SearchResult{
					Path: file,
				})
			}
		} else {
			// Search in files
			opts := search.Options{
				Query:         query,
				Path:          ".",
				CaseSensitive: false,
				MaxResults:    50,
				MaxLineLength: 200,
				UseFuzzy:      true,
				FuzzyThreshold: 60,
				Exclude:       []string{"*.bin", "*.exe", "*.so", "*.dylib", "*.dll", ".git/*", "vendor/*", "node_modules/*"},
			}
			
			searchResults, err := s.searcher.Search(ctx, opts)
			if err != nil {
				return SearchResultsMsg{Err: err}
			}
			
			// Convert to SearchResult
			for _, r := range searchResults {
				results = append(results, SearchResult{
					Path:    r.Path,
					Line:    r.Line,
					Content: r.Content,
					Match:   r.Match,
				})
			}
		}

		return SearchResultsMsg{
			Results: results,
		}
	}
}

// Key bindings
type searchKeyMap struct {
	Cancel key.Binding
	Select key.Binding
	Up     key.Binding
	Down   key.Binding
}

var searchKeys = searchKeyMap{
	Cancel: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "cancel"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
}
