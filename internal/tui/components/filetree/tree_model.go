package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
	"github.com/gabriel-vasile/mimetype"
	ignore "github.com/sabhiram/go-gitignore"
)

// TreeModel represents a file tree using lipgloss tree component
type TreeModel struct {
	width     int
	height    int
	focused   bool
	rootPath  string
	tree      *tree.Tree
	flatItems []FileItem // Flattened list for navigation
	selected  int
	gitignore *ignore.GitIgnore
}

// FileItem represents a file or directory in the tree
type FileItem struct {
	Path     string
	Name     string
	IsDir    bool
	Level    int
	Expanded bool
	Hidden   bool
}

// NewTreeModel creates a new file tree model
func NewTreeModel(rootPath string) *TreeModel {
	model := &TreeModel{
		rootPath:  rootPath,
		selected:  0,
		flatItems: []FileItem{},
	}

	// Load .gitignore if it exists
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		if gi, err := ignore.CompileIgnoreFile(gitignorePath); err == nil {
			model.gitignore = gi
		}
	}

	model.buildTree()
	return model
}

// SetSize sets the size of the tree model
func (tm *TreeModel) SetSize(width, height int) {
	tm.width = width
	tm.height = height
}

// Focus sets focus on the tree model
func (tm *TreeModel) Focus() {
	tm.focused = true
}

// Blur removes focus from the tree model
func (tm *TreeModel) Blur() {
	tm.focused = false
}

// Init implements tea.Model
func (tm *TreeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (tm *TreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !tm.focused {
		return tm, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if tm.selected > 0 {
				tm.selected--
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if tm.selected < len(tm.flatItems)-1 {
				tm.selected++
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			if tm.selected < len(tm.flatItems) {
				item := tm.flatItems[tm.selected]
				if item.IsDir {
					// Toggle directory expansion
					tm.toggleDirectory(item.Path)
					tm.buildTree()
				} else {
					// Select file
					return tm, func() tea.Msg {
						return FileSelectedMsg{Path: item.Path}
					}
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Refresh tree
			tm.buildTree()
		}

	case tea.WindowSizeMsg:
		tm.SetSize(msg.Width, msg.Height)
	}

	return tm, nil
}

// toggleDirectory toggles the expansion state of a directory
func (tm *TreeModel) toggleDirectory(dirPath string) {
	for i, item := range tm.flatItems {
		if item.Path == dirPath && item.IsDir {
			tm.flatItems[i].Expanded = !tm.flatItems[i].Expanded
			break
		}
	}
}

// buildTree builds the tree structure and flattened list
func (tm *TreeModel) buildTree() {
	tm.flatItems = []FileItem{}
	tm.tree = tree.New()

	// Build tree recursively
	tm.buildTreeRecursive(tm.rootPath, tm.tree, 0)

	// Apply styling
	tm.tree.
		Enumerator(tree.RoundedEnumerator).
		ItemStyleFunc(tm.itemStyleFunc).
		EnumeratorStyleFunc(tm.enumeratorStyleFunc)
}

// buildTreeRecursive builds the tree structure recursively
func (tm *TreeModel) buildTreeRecursive(dirPath string, parentTree *tree.Tree, level int) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	// Sort entries: directories first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Skip hidden files (starting with .)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip gitignored files
		if tm.gitignore != nil {
			relPath, err := filepath.Rel(tm.rootPath, fullPath)
			if err == nil && tm.gitignore.MatchesPath(relPath) {
				continue
			}
		}

		// Skip binary files for non-directories
		if !entry.IsDir() && !tm.isTextFile(fullPath) {
			continue
		}

		// Create file item
		item := FileItem{
			Path:     fullPath,
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Level:    level,
			Expanded: false,
			Hidden:   false,
		}

		// Add to flat list
		tm.flatItems = append(tm.flatItems, item)

		// Add to tree
		if entry.IsDir() {
			// Create directory node
			dirNode := tree.Root(tm.formatDirName(entry.Name()))

			// Check if this directory should be expanded
			expanded := tm.isDirExpanded(fullPath)
			if expanded {
				item.Expanded = true
				tm.buildTreeRecursive(fullPath, dirNode, level+1)
			}

			parentTree.Child(dirNode)
		} else {
			// Create file node
			parentTree.Child(tm.formatFileName(entry.Name()))
		}
	}
}

// isDirExpanded checks if a directory is currently expanded
func (tm *TreeModel) isDirExpanded(dirPath string) bool {
	for _, item := range tm.flatItems {
		if item.Path == dirPath && item.IsDir {
			return item.Expanded
		}
	}
	return false
}

// isTextFile checks if a file is a text file that should be displayed
func (tm *TreeModel) isTextFile(filePath string) bool {
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		// If we can't detect, assume it's text based on extension
		ext := strings.ToLower(filepath.Ext(filePath))
		textExts := []string{
			".go", ".rs", ".py", ".js", ".ts", ".jsx", ".tsx", ".html", ".css", ".scss",
			".json", ".yaml", ".yml", ".toml", ".xml", ".md", ".txt", ".sh", ".bash",
			".zsh", ".fish", ".ps1", ".bat", ".cmd", ".dockerfile", ".makefile",
			".gitignore", ".gitattributes", ".editorconfig", ".env", ".ini", ".conf",
			".cfg", ".properties", ".sql", ".graphql", ".proto", ".thrift",
		}
		for _, textExt := range textExts {
			if ext == textExt {
				return true
			}
		}
		return false
	}

	// Check if it's a text MIME type
	return strings.HasPrefix(mtype.String(), "text/") ||
		strings.Contains(mtype.String(), "json") ||
		strings.Contains(mtype.String(), "xml") ||
		strings.Contains(mtype.String(), "yaml") ||
		strings.Contains(mtype.String(), "toml")
}

// formatDirName formats a directory name with an icon
func (tm *TreeModel) formatDirName(name string) string {
	return fmt.Sprintf("📁 %s", name)
}

// formatFileName formats a file name with an appropriate icon
func (tm *TreeModel) formatFileName(name string) string {
	ext := strings.ToLower(filepath.Ext(name))

	var icon string
	switch ext {
	case ".go":
		icon = "🐹"
	case ".rs":
		icon = "🦀"
	case ".py":
		icon = "🐍"
	case ".js", ".jsx":
		icon = "📜"
	case ".ts", ".tsx":
		icon = "📘"
	case ".html":
		icon = "🌐"
	case ".css", ".scss":
		icon = "🎨"
	case ".json":
		icon = "📋"
	case ".md":
		icon = "📝"
	case ".yaml", ".yml":
		icon = "⚙️"
	case ".toml":
		icon = "🔧"
	case ".xml":
		icon = "📄"
	case ".sh", ".bash", ".zsh", ".fish":
		icon = "🐚"
	case ".dockerfile":
		icon = "🐳"
	case ".makefile":
		icon = "🔨"
	default:
		icon = "📄"
	}

	return fmt.Sprintf("%s %s", icon, name)
}

// itemStyleFunc returns the style for tree items based on selection
func (tm *TreeModel) itemStyleFunc(children tree.Children, index int) lipgloss.Style {
	if index == tm.selected {
		return lipgloss.NewStyle().
			Foreground(theme.CurrentTheme().Primary()).
			Bold(true)
	}
	return lipgloss.NewStyle().Foreground(theme.CurrentTheme().Text())
}

// enumeratorStyleFunc returns the style for tree enumerators
func (tm *TreeModel) enumeratorStyleFunc(children tree.Children, index int) lipgloss.Style {
	if index == tm.selected {
		return lipgloss.NewStyle().Foreground(theme.CurrentTheme().Primary())
	}
	return lipgloss.NewStyle().Foreground(theme.CurrentTheme().Secondary())
}

// View implements tea.Model
func (tm *TreeModel) View() string {
	if tm.tree == nil {
		return "Loading files..."
	}

	// Create a container with border
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme().Secondary()).
		Padding(1).
		Width(tm.width - 2).
		Height(tm.height - 2)

	if tm.focused {
		style = style.BorderForeground(theme.CurrentTheme().Primary())
	}

	content := tm.tree.String()

	// Add help text at the bottom
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme().Secondary()).
		Italic(true)

	help := helpStyle.Render("↑↓: navigate • Enter: select/expand • r: refresh")

	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			help,
		),
	)
}
