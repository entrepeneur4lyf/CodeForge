package fileviewer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// FileViewer represents a file viewer component
type FileViewer struct {
	viewport     viewport.Model
	content      string
	filePath     string
	fileName     string
	width        int
	height       int
	renderer     *glamour.TermRenderer
	isMarkdown   bool
	lineNumbers  bool
	focused      bool
}

// FileLoadedMsg is sent when a file is loaded
type FileLoadedMsg struct {
	Path    string
	Content string
}

// New creates a new file viewer
func New() *FileViewer {
	vp := viewport.New(80, 24)
	
	// Create glamour renderer for markdown files
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		log.Error("Failed to create glamour renderer", "error", err)
		renderer = nil
	}

	return &FileViewer{
		viewport:    vp,
		renderer:    renderer,
		lineNumbers: true,
		focused:     false,
	}
}

// SetSize sets the dimensions of the file viewer
func (fv *FileViewer) SetSize(width, height int) {
	fv.width = width
	fv.height = height
	fv.viewport.Width = width
	fv.viewport.Height = height - 2 // Leave space for header
}

// SetFocused sets the focus state
func (fv *FileViewer) SetFocused(focused bool) {
	fv.focused = focused
}

// LoadFile loads a file for viewing
func (fv *FileViewer) LoadFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return FileLoadedMsg{
				Path:    path,
				Content: fmt.Sprintf("Error loading file: %v", err),
			}
		}

		return FileLoadedMsg{
			Path:    path,
			Content: string(content),
		}
	}
}

// isMarkdownFile checks if the file is a markdown file
func (fv *FileViewer) isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

// renderContent renders the file content with optional line numbers
func (fv *FileViewer) renderContent(content string) string {
	if fv.isMarkdown && fv.renderer != nil {
		// Render markdown
		rendered, err := fv.renderer.Render(content)
		if err != nil {
			log.Error("Failed to render markdown", "error", err)
			return fv.renderWithLineNumbers(content)
		}
		return rendered
	}

	return fv.renderWithLineNumbers(content)
}

// renderWithLineNumbers adds line numbers to content
func (fv *FileViewer) renderWithLineNumbers(content string) string {
	if !fv.lineNumbers {
		return content
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder
	
	// Calculate the width needed for line numbers
	maxLineNum := len(lines)
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))
	
	lineNumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(lineNumWidth).
		Align(lipgloss.Right)

	for i, line := range lines {
		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		result.WriteString(fmt.Sprintf("%s │ %s\n", lineNum, line))
	}

	return result.String()
}

// Init implements tea.Model
func (fv *FileViewer) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (fv *FileViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case FileLoadedMsg:
		fv.filePath = msg.Path
		fv.fileName = filepath.Base(msg.Path)
		fv.content = msg.Content
		fv.isMarkdown = fv.isMarkdownFile(msg.Path)
		
		// Render and set content
		renderedContent := fv.renderContent(msg.Content)
		fv.viewport.SetContent(renderedContent)
		fv.viewport.GotoTop()

	case tea.KeyMsg:
		if fv.focused {
			// Handle key events for navigation
			fv.viewport, cmd = fv.viewport.Update(msg)
		}

	case tea.WindowSizeMsg:
		fv.SetSize(msg.Width, msg.Height)

	default:
		// Forward other messages to viewport
		if fv.focused {
			fv.viewport, cmd = fv.viewport.Update(msg)
		}
	}

	return fv, cmd
}

// View implements tea.Model
func (fv *FileViewer) View() string {
	if fv.filePath == "" {
		// Show empty state
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Align(lipgloss.Center).
			Width(fv.width).
			Height(fv.height)

		return emptyStyle.Render("📄 No file selected\n\nSelect a file from the file picker to view its contents")
	}

	// Header with file info
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	var header string
	if fv.isMarkdown {
		header = fmt.Sprintf("📝 %s (Markdown)", fv.fileName)
	} else {
		header = fmt.Sprintf("📄 %s", fv.fileName)
	}

	// Viewport content
	viewportContent := fv.viewport.View()

	// Combine header and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(header),
		viewportContent,
	)
}

// GetFilePath returns the currently loaded file path
func (fv *FileViewer) GetFilePath() string {
	return fv.filePath
}

// GetFileName returns the currently loaded file name
func (fv *FileViewer) GetFileName() string {
	return fv.fileName
}

// ToggleLineNumbers toggles line number display
func (fv *FileViewer) ToggleLineNumbers() {
	fv.lineNumbers = !fv.lineNumbers
	if fv.content != "" {
		// Re-render content with updated line number setting
		renderedContent := fv.renderContent(fv.content)
		fv.viewport.SetContent(renderedContent)
	}
}
