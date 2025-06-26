package tui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// IsTTY returns true if the current environment is a TTY.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// Pane represents which pane is currently focused
type Pane int

const (
	FileBrowserPane Pane = iota
	CodeEditorPane
	AIChatPane
	TerminalPane
)

// Message types
type (
	llmResponseMsg   string
	errMsg           struct{ err error }
	fileSelectedMsg  string
	buildCompleteMsg string
)

func (e errMsg) Error() string { return e.err.Error() }

// File represents a file in the file browser
type File struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	Modified time.Time
}

// Tab represents an open file tab
type Tab struct {
	Name     string
	Path     string
	Content  string
	Modified bool
}

// Model represents the main TUI model with multi-pane layout
type Model struct {
	// Layout
	width  int
	height int

	// Current focus
	focusedPane Pane

	// File browser
	files        []File
	selectedFile int
	currentDir   string

	// Code editor
	tabs       []Tab
	activeTab  int
	codeEditor textarea.Model

	// AI chat
	chatMessages []string
	chatInput    textarea.Model
	chatViewport viewport.Model

	// Terminal/output
	terminalOutput   string
	terminalViewport viewport.Model

	// Styles
	styles Styles

	// State
	err error
}

// Styles holds all the styling for the TUI
type Styles struct {
	Border       lipgloss.Style
	BorderActive lipgloss.Style
	Title        lipgloss.Style
	File         lipgloss.Style
	FileSelected lipgloss.Style
	Tab          lipgloss.Style
	TabActive    lipgloss.Style
	Message      lipgloss.Style
	MessageUser  lipgloss.Style
	MessageAI    lipgloss.Style
}

// NewModel creates a new TUI model
func NewModel() Model {
	// Initialize styles
	styles := Styles{
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		BorderActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("69")),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")).
			Bold(true).
			Padding(0, 1),
		File: lipgloss.NewStyle().
			Padding(0, 1),
		FileSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("69")).
			Foreground(lipgloss.Color("15")),
		Tab: lipgloss.NewStyle().
			Padding(0, 1).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderForeground(lipgloss.Color("240")),
		TabActive: lipgloss.NewStyle().
			Padding(0, 1).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderForeground(lipgloss.Color("69")).
			Background(lipgloss.Color("235")),
		Message: lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 0, 1, 0),
		MessageUser: lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Background(lipgloss.Color("69")).
			Foreground(lipgloss.Color("15")),
		MessageAI: lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("15")),
	}

	// Initialize code editor
	codeEditor := textarea.New()
	codeEditor.Placeholder = "// Open a file to start editing..."
	codeEditor.ShowLineNumbers = true

	// Initialize chat input
	chatInput := textarea.New()
	chatInput.Placeholder = "Ask AI for help..."
	chatInput.SetHeight(3)

	// Initialize viewports
	chatViewport := viewport.New(0, 0)
	terminalViewport := viewport.New(0, 0)

	model := Model{
		focusedPane:      FileBrowserPane,
		currentDir:       ".",
		codeEditor:       codeEditor,
		chatInput:        chatInput,
		chatViewport:     chatViewport,
		terminalViewport: terminalViewport,
		styles:           styles,
		terminalOutput:   "CodeForge TUI initialized.\nPress Tab to switch panes, Ctrl+C to quit.\n",
	}

	// Load initial files
	model.loadFiles()

	// Add welcome tab
	model.tabs = append(model.tabs, Tab{
		Name:    "Welcome",
		Path:    "",
		Content: "// Welcome to CodeForge!\n// Press Tab to navigate between panes\n// Ctrl+1-4 for quick pane switching\n\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, CodeForge!\")\n}",
	})

	return model
}

// loadFiles loads files from the current directory
func (m *Model) loadFiles() {
	m.files = []File{}

	entries, err := os.ReadDir(m.currentDir)
	if err != nil {
		m.err = err
		return
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		file := File{
			Name:     entry.Name(),
			Path:     filepath.Join(m.currentDir, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
		}

		m.files = append(m.files, file)
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.loadFilesCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.nextPane()
		case "shift+tab":
			m.prevPane()
		case "ctrl+1":
			m.focusedPane = FileBrowserPane
		case "ctrl+2":
			m.focusedPane = CodeEditorPane
		case "ctrl+3":
			m.focusedPane = AIChatPane
		case "ctrl+4":
			m.focusedPane = TerminalPane
		}

		// Handle pane-specific input
		switch m.focusedPane {
		case FileBrowserPane:
			cmds = append(cmds, m.updateFileBrowser(msg))
		case CodeEditorPane:
			cmds = append(cmds, m.updateCodeEditor(msg))
		case AIChatPane:
			cmds = append(cmds, m.updateAIChat(msg))
		case TerminalPane:
			cmds = append(cmds, m.updateTerminal(msg))
		}

	case llmResponseMsg:
		m.chatMessages = append(m.chatMessages, "AI: "+string(msg))
		m.updateChatViewport()

	case errMsg:
		m.err = msg.err

	case fileSelectedMsg:
		cmds = append(cmds, m.openFile(string(msg)))
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.err)
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Calculate pane dimensions
	fileWidth := m.width / 4
	codeWidth := m.width - fileWidth - m.width/4
	chatWidth := m.width / 4
	terminalHeight := m.height / 4

	// Render panes
	fileBrowser := m.renderFileBrowser(fileWidth, m.height-terminalHeight)
	codeEditor := m.renderCodeEditor(codeWidth, m.height-terminalHeight)
	aiChat := m.renderAIChat(chatWidth, m.height-terminalHeight)
	terminal := m.renderTerminal(m.width, terminalHeight)

	// Combine panes
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, fileBrowser, codeEditor, aiChat)
	return lipgloss.JoinVertical(lipgloss.Left, topRow, terminal)
}

// Start starts the TUI
func Start() {
	// The LLM system should already be initialized by the root command
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
