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
	"github.com/entrepeneur4lyf/codeforge/internal/tui/styles"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
	"github.com/gabriel-vasile/mimetype"
	ignore "github.com/sabhiram/go-gitignore"
)

// TreeNode represents a file or directory in the tree
type TreeNode struct {
	Name      string
	Path      string
	IsDir     bool
	Children  []*TreeNode
	Parent    *TreeNode
	Expanded  bool
	Modified  bool
	GitStatus string // "M", "A", "D", "??"
}

// FileTreeModel represents the file tree component
type FileTreeModel struct {
	root         *TreeNode
	cursor       int
	selected     string
	width        int
	height       int
	focused      bool
	visibleNodes []*TreeNode
	rootPath     string
	showHidden   bool
	gitStatus    map[string]string
	gitignore    *ignore.GitIgnore
	scrollOffset int // For scrolling support
}

// FileSelectedMsg is sent when a file is selected
type FileSelectedMsg struct {
	Path string
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(rootPath string) *FileTreeModel {
	model := &FileTreeModel{
		rootPath:   rootPath,
		showHidden: false,
		gitStatus:  make(map[string]string),
	}

	// Load gitignore if it exists
	model.loadGitignore()
	model.loadTree()
	model.updateVisibleNodes()

	return model
}

// loadGitignore loads the .gitignore file if it exists
func (m *FileTreeModel) loadGitignore() {
	gitignorePath := filepath.Join(m.rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		if gitignore, err := ignore.CompileIgnoreFile(gitignorePath); err == nil {
			m.gitignore = gitignore
		}
	}
}

// isTextFile checks if a file is text-based and editable
func (m *FileTreeModel) isTextFile(path string) bool {
	// Use mimetype to detect if file is text-based
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		// Fallback to extension-based detection
		return m.isTextByExtension(path)
	}

	// Allow text files and common code formats
	switch {
	case strings.HasPrefix(mtype.String(), "text/"):
		return true
	case mtype.Is("application/json"):
		return true
	case mtype.Is("application/xml"):
		return true
	case mtype.Is("application/yaml"):
		return true
	case mtype.Is("application/javascript"):
		return true
	case mtype.Is("application/typescript"):
		return true
	default:
		return false
	}
}

// isTextByExtension fallback method for text detection
func (m *FileTreeModel) isTextByExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := []string{
		".txt", ".md", ".go", ".rs", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
		".json", ".yaml", ".yml", ".toml", ".xml", ".html", ".css", ".scss", ".sass",
		".sh", ".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd", ".dockerfile", ".sql",
		".log", ".conf", ".config", ".ini", ".env", ".gitignore", ".gitattributes",
	}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

// shouldShowInTree determines if a file/directory should be shown in the editor tree
func (m *FileTreeModel) shouldShowInTree(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check gitignore first
	if m.gitignore != nil && m.gitignore.MatchesPath(path) {
		return false
	}

	// Always show directories (but filter their contents)
	if info.IsDir() {
		// Skip common build/cache directories
		basename := filepath.Base(path)
		skipDirs := []string{
			"node_modules", ".git", "target", "build", "dist", ".next", ".nuxt",
			"__pycache__", ".pytest_cache", "vendor", ".cargo", ".vscode", ".idea",
			".DS_Store", "coverage", ".nyc_output", "tmp", "temp",
		}

		for _, skip := range skipDirs {
			if basename == skip {
				return false
			}
		}

		return true
	}

	// For files, only show if they're text-editable
	return m.isTextFile(path)
}

// getCodeFileIcon returns an appropriate icon for code files
func (m *FileTreeModel) getCodeFileIcon(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".go":
		return "🐹"
	case ".rs":
		return "🦀"
	case ".js", ".mjs":
		return "📜"
	case ".ts":
		return "📘"
	case ".py":
		return "🐍"
	case ".java":
		return "☕"
	case ".c", ".h":
		return "🔧"
	case ".cpp", ".hpp", ".cc":
		return "⚙️"
	case ".json":
		return "📋"
	case ".yaml", ".yml":
		return "⚙️"
	case ".toml":
		return "🔧"
	case ".xml":
		return "📄"
	case ".md":
		return "📝"
	case ".txt":
		return "📝"
	case ".css", ".scss", ".sass":
		return "🎨"
	case ".html", ".htm":
		return "🌐"
	case ".sql":
		return "🗃️"
	case ".sh", ".bash", ".zsh", ".fish":
		return "📜"
	case ".dockerfile":
		return "🐳"
	case ".env":
		return "🔐"
	case ".log":
		return "📋"
	default:
		return "📄"
	}
}

// adjustScrollOffset adjusts the scroll offset to keep the cursor visible
func (m *FileTreeModel) adjustScrollOffset() {
	// Calculate available space for file list (subtract header and padding)
	availableHeight := m.height - 4 // Header + padding + modified files section
	if availableHeight <= 0 {
		return
	}

	// Ensure cursor is visible within the scroll window
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+availableHeight {
		m.scrollOffset = m.cursor - availableHeight + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	// Ensure scroll offset doesn't exceed the maximum
	maxOffset := len(m.visibleNodes) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

// Init implements tea.Model
func (m *FileTreeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *FileTreeModel) Update(msg tea.Msg) (*FileTreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
				m.adjustScrollOffset()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.visibleNodes)-1 {
				m.cursor++
				m.adjustScrollOffset()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			pageSize := m.height - 4
			if pageSize <= 0 {
				pageSize = 1
			}
			m.cursor -= pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScrollOffset()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))):
			pageSize := m.height - 4
			if pageSize <= 0 {
				pageSize = 1
			}
			m.cursor += pageSize
			if m.cursor >= len(m.visibleNodes) {
				m.cursor = len(m.visibleNodes) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScrollOffset()
		case key.Matches(msg, key.NewBinding(key.WithKeys("home"))):
			m.cursor = 0
			m.adjustScrollOffset()
		case key.Matches(msg, key.NewBinding(key.WithKeys("end"))):
			m.cursor = len(m.visibleNodes) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScrollOffset()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.cursor < len(m.visibleNodes) {
				node := m.visibleNodes[m.cursor]
				if node.IsDir {
					node.Expanded = !node.Expanded
					m.updateVisibleNodes()
				} else {
					m.selected = node.Path
					return m, func() tea.Msg {
						return FileSelectedMsg{Path: node.Path}
					}
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("space"))):
			if m.cursor < len(m.visibleNodes) {
				node := m.visibleNodes[m.cursor]
				if node.IsDir {
					node.Expanded = !node.Expanded
					m.updateVisibleNodes()
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+r"))):
			m.loadTree()
			m.updateVisibleNodes()
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+h"))):
			m.showHidden = !m.showHidden
			m.loadTree()
			m.updateVisibleNodes()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View implements tea.Model
func (m *FileTreeModel) View() string {
	if m.root == nil {
		return "Loading..."
	}

	t := theme.CurrentTheme()
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1)

	header := headerStyle.Render("📁 Files")
	lines = append(lines, header)
	lines = append(lines, "")

	// File tree with scrolling
	availableHeight := m.height - 4 // Leave space for header and modified files
	if availableHeight <= 0 {
		availableHeight = 1
	}

	endIndex := m.scrollOffset + availableHeight
	if endIndex > len(m.visibleNodes) {
		endIndex = len(m.visibleNodes)
	}

	for i := m.scrollOffset; i < endIndex; i++ {
		if i >= len(m.visibleNodes) {
			break
		}

		node := m.visibleNodes[i]
		line := m.renderNode(node, i == m.cursor)
		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if len(m.visibleNodes) > availableHeight {
		if m.scrollOffset > 0 {
			lines[1] = "↑ " + lines[1] // Add up arrow to indicate more content above
		}
		if endIndex < len(m.visibleNodes) {
			lines = append(lines, "↓ More files...") // Add down arrow to indicate more content below
		}
	}

	// Modified files section
	if len(m.getModifiedFiles()) > 0 {
		lines = append(lines, "")
		modifiedStyle := lipgloss.NewStyle().
			Foreground(t.GitModified()).
			Bold(true).
			Padding(0, 1)

		lines = append(lines, modifiedStyle.Render("Modified:"))

		for _, file := range m.getModifiedFiles() {
			fileStyle := lipgloss.NewStyle().
				Foreground(t.GitModified()).
				Padding(0, 1)
			lines = append(lines, fileStyle.Render("• "+filepath.Base(file)))
		}
	}

	content := strings.Join(lines, "\n")

	// Ensure content fits within the allocated space
	style := styles.SidebarStyle().
		Width(m.width).
		Height(m.height)

	// Add border and focus styling
	if m.focused {
		style = style.BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary())
	} else {
		style = style.BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border())
	}

	return style.Render(content)
}

// renderNode renders a single tree node
func (m *FileTreeModel) renderNode(node *TreeNode, selected bool) string {

	// Calculate indentation based on depth
	depth := m.getNodeDepth(node)
	indent := strings.Repeat("  ", depth)

	// Icon and name
	var icon string
	if node.IsDir {
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
	} else {
		icon = m.getCodeFileIcon(node.Path)
	}

	// Git status indicator
	gitIndicator := ""
	if status, exists := m.gitStatus[node.Path]; exists {
		gitIndicator = fmt.Sprintf(" %s", status)
	}

	// Build the line
	line := fmt.Sprintf("%s%s %s%s", indent, icon, node.Name, gitIndicator)

	// Apply styling
	style := styles.FileTreeItemStyle(selected, node.Modified)
	if status, exists := m.gitStatus[node.Path]; exists {
		style = style.Foreground(m.getGitStatusColor(status))
	}

	return style.Render(line)
}

// getFileIcon returns an appropriate icon for the file type
func (m *FileTreeModel) getFileIcon(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".go":
		return "🐹"
	case ".rs":
		return "🦀"
	case ".js", ".ts":
		return "📜"
	case ".py":
		return "🐍"
	case ".md":
		return "📝"
	case ".json":
		return "📋"
	case ".yaml", ".yml":
		return "⚙️"
	case ".toml":
		return "🔧"
	default:
		return "📄"
	}
}

// getGitStatusColor returns the color for a git status
func (m *FileTreeModel) getGitStatusColor(status string) lipgloss.Color {
	t := theme.CurrentTheme()

	switch status {
	case "M":
		return t.GitModified()
	case "A":
		return t.GitAdded()
	case "D":
		return t.GitDeleted()
	case "??":
		return t.GitUntracked()
	default:
		return t.Text()
	}
}

// getNodeDepth calculates the depth of a node in the tree
func (m *FileTreeModel) getNodeDepth(node *TreeNode) int {
	depth := 0
	current := node.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// getModifiedFiles returns a list of modified files
func (m *FileTreeModel) getModifiedFiles() []string {
	var modified []string
	for path, status := range m.gitStatus {
		if status == "M" || status == "A" || status == "D" {
			modified = append(modified, path)
		}
	}
	sort.Strings(modified)
	return modified
}

// SetFocused sets the focus state of the file tree
func (m *FileTreeModel) SetFocused(focused bool) {
	m.focused = focused
}

// SetSize sets the dimensions of the file tree
func (m *FileTreeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetGitStatus updates the git status for files
func (m *FileTreeModel) SetGitStatus(status map[string]string) {
	m.gitStatus = status
	m.updateModifiedStatus()
}

// loadTree loads the file tree from the root path
func (m *FileTreeModel) loadTree() {
	m.root = m.buildTree(m.rootPath, nil)
}

// buildTree recursively builds the tree structure
func (m *FileTreeModel) buildTree(path string, parent *TreeNode) *TreeNode {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	name := filepath.Base(path)
	if path == m.rootPath {
		name = filepath.Base(m.rootPath)
	}

	node := &TreeNode{
		Name:   name,
		Path:   path,
		IsDir:  info.IsDir(),
		Parent: parent,
	}

	if node.IsDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return node
		}

		// Sort entries: directories first, then files
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir() != entries[j].IsDir() {
				return entries[i].IsDir()
			}
			return entries[i].Name() < entries[j].Name()
		})

		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())

			// Use smart filtering to determine if we should show this file/directory
			if !m.shouldShowInTree(childPath) {
				continue
			}

			// Skip hidden files unless showHidden is true
			if !m.showHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			child := m.buildTree(childPath, node)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}

		// Auto-expand root directory
		if parent == nil {
			node.Expanded = true
		}
	}

	return node
}

// updateVisibleNodes updates the list of visible nodes based on expansion state
func (m *FileTreeModel) updateVisibleNodes() {
	m.visibleNodes = nil
	if m.root != nil {
		m.collectVisibleNodes(m.root)
	}
}

// collectVisibleNodes recursively collects visible nodes
func (m *FileTreeModel) collectVisibleNodes(node *TreeNode) {
	// Don't show the root node itself
	if node.Parent != nil {
		m.visibleNodes = append(m.visibleNodes, node)
	}

	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			m.collectVisibleNodes(child)
		}
	}
}

// updateModifiedStatus updates the modified status of nodes based on git status
func (m *FileTreeModel) updateModifiedStatus() {
	if m.root != nil {
		m.updateNodeModifiedStatus(m.root)
	}
}

// updateNodeModifiedStatus recursively updates modified status
func (m *FileTreeModel) updateNodeModifiedStatus(node *TreeNode) {
	// Check if this node is modified
	if _, exists := m.gitStatus[node.Path]; exists {
		node.Modified = true
	} else {
		node.Modified = false
	}

	// Recursively update children
	for _, child := range node.Children {
		m.updateNodeModifiedStatus(child)
	}
}

// GetSelectedPath returns the currently selected file path
func (m *FileTreeModel) GetSelectedPath() string {
	return m.selected
}

// ExpandAll expands all directories in the tree
func (m *FileTreeModel) ExpandAll() {
	if m.root != nil {
		m.expandNode(m.root)
		m.updateVisibleNodes()
	}
}

// CollapseAll collapses all directories in the tree
func (m *FileTreeModel) CollapseAll() {
	if m.root != nil {
		m.collapseNode(m.root)
		m.updateVisibleNodes()
	}
}

// expandNode recursively expands a node and its children
func (m *FileTreeModel) expandNode(node *TreeNode) {
	if node.IsDir {
		node.Expanded = true
		for _, child := range node.Children {
			m.expandNode(child)
		}
	}
}

// collapseNode recursively collapses a node and its children
func (m *FileTreeModel) collapseNode(node *TreeNode) {
	if node.IsDir {
		node.Expanded = false
		for _, child := range node.Children {
			m.collapseNode(child)
		}
	}
}
