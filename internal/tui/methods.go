package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
)

// loadFilesCmd returns a command to load files
func (m Model) loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		return nil // Files are loaded synchronously in loadFiles()
	}
}

// updateSizes updates the sizes of all components
func (m *Model) updateSizes() {
	fileWidth := m.width / 4
	codeWidth := m.width - fileWidth - m.width/4
	chatWidth := m.width / 4
	terminalHeight := m.height / 4

	// Update code editor size
	m.codeEditor.SetWidth(codeWidth - 4)
	m.codeEditor.SetHeight(m.height - terminalHeight - 6)

	// Update chat input size
	m.chatInput.SetWidth(chatWidth - 4)

	// Update viewports
	m.chatViewport.Width = chatWidth - 4
	m.chatViewport.Height = m.height - terminalHeight - 10

	m.terminalViewport.Width = m.width - 4
	m.terminalViewport.Height = terminalHeight - 4
}

// nextPane moves focus to the next pane
func (m *Model) nextPane() {
	switch m.focusedPane {
	case FileBrowserPane:
		m.focusedPane = CodeEditorPane
		m.codeEditor.Focus()
		m.chatInput.Blur()
	case CodeEditorPane:
		m.focusedPane = AIChatPane
		m.codeEditor.Blur()
		m.chatInput.Focus()
	case AIChatPane:
		m.focusedPane = TerminalPane
		m.chatInput.Blur()
	case TerminalPane:
		m.focusedPane = FileBrowserPane
	}
}

// prevPane moves focus to the previous pane
func (m *Model) prevPane() {
	switch m.focusedPane {
	case FileBrowserPane:
		m.focusedPane = TerminalPane
	case CodeEditorPane:
		m.focusedPane = FileBrowserPane
		m.codeEditor.Blur()
	case AIChatPane:
		m.focusedPane = CodeEditorPane
		m.chatInput.Blur()
		m.codeEditor.Focus()
	case TerminalPane:
		m.focusedPane = AIChatPane
		m.chatInput.Focus()
	}
}

// updateFileBrowser handles file browser input
func (m *Model) updateFileBrowser(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if m.selectedFile > 0 {
			m.selectedFile--
		}
	case "down", "j":
		if m.selectedFile < len(m.files)-1 {
			m.selectedFile++
		}
	case "enter":
		if len(m.files) > 0 && m.selectedFile < len(m.files) {
			file := m.files[m.selectedFile]
			if file.IsDir {
				m.currentDir = file.Path
				m.selectedFile = 0
				m.loadFiles()
			} else {
				return func() tea.Msg {
					return fileSelectedMsg(file.Path)
				}
			}
		}
	case "backspace":
		if m.currentDir != "." {
			m.currentDir = filepath.Dir(m.currentDir)
			m.selectedFile = 0
			m.loadFiles()
		}
	}
	return nil
}

// updateCodeEditor handles code editor input
func (m *Model) updateCodeEditor(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.codeEditor, cmd = m.codeEditor.Update(msg)

	// Mark current tab as modified if content changed
	if len(m.tabs) > 0 && m.activeTab < len(m.tabs) {
		m.tabs[m.activeTab].Content = m.codeEditor.Value()
		m.tabs[m.activeTab].Modified = true
	}

	switch msg.String() {
	case "ctrl+s":
		return m.saveCurrentFile()
	case "ctrl+w":
		return m.closeCurrentTab()
	}

	return cmd
}

// updateAIChat handles AI chat input
func (m *Model) updateAIChat(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		if !msg.Alt {
			userInput := strings.TrimSpace(m.chatInput.Value())
			if userInput != "" {
				m.chatMessages = append(m.chatMessages, "You: "+userInput)
				m.updateChatViewport()
				m.chatInput.Reset()

				return func() tea.Msg {
					// Get the default model
					defaultModel, err := llm.GetDefaultModel()
					if err != nil {
						return errMsg{err}
					}

					// Create completion request
					req := llm.CompletionRequest{
						Model: defaultModel.ID,
						Messages: []llm.Message{
							{
								Role:    "user",
								Content: userInput,
							},
						},
						MaxTokens:   defaultModel.DefaultMaxTokens,
						Temperature: 0.7,
					}

					resp, err := llm.GetCompletion(context.Background(), req)
					if err != nil {
						return errMsg{err}
					}
					return llmResponseMsg(resp.Content)
				}
			}
			return nil
		}
	}

	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	return cmd
}

// updateTerminal handles terminal input
func (m *Model) updateTerminal(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+l":
		m.terminalOutput = ""
		m.terminalViewport.SetContent("")
	}

	var cmd tea.Cmd
	m.terminalViewport, cmd = m.terminalViewport.Update(msg)
	return cmd
}

// updateChatViewport updates the chat viewport content
func (m *Model) updateChatViewport() {
	content := strings.Join(m.chatMessages, "\n")
	m.chatViewport.SetContent(content)
	m.chatViewport.GotoBottom()
}

// openFile opens a file in a new tab
func (m *Model) openFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return errMsg{err}
		}

		// Check if file is already open
		for i, tab := range m.tabs {
			if tab.Path == path {
				m.activeTab = i
				m.codeEditor.SetValue(tab.Content)
				return nil
			}
		}

		// Add new tab
		tab := Tab{
			Name:    filepath.Base(path),
			Path:    path,
			Content: string(content),
		}

		m.tabs = append(m.tabs, tab)
		m.activeTab = len(m.tabs) - 1
		m.codeEditor.SetValue(tab.Content)

		return nil
	}
}

// saveCurrentFile saves the current file
func (m *Model) saveCurrentFile() tea.Cmd {
	if len(m.tabs) == 0 || m.activeTab >= len(m.tabs) {
		return nil
	}

	tab := &m.tabs[m.activeTab]
	if tab.Path == "" {
		return nil // Can't save files without a path
	}

	return func() tea.Msg {
		err := os.WriteFile(tab.Path, []byte(tab.Content), 0644)
		if err != nil {
			return errMsg{err}
		}

		tab.Modified = false
		m.terminalOutput += fmt.Sprintf("Saved: %s\n", tab.Path)
		m.terminalViewport.SetContent(m.terminalOutput)
		m.terminalViewport.GotoBottom()

		return nil
	}
}

// closeCurrentTab closes the current tab
func (m *Model) closeCurrentTab() tea.Cmd {
	if len(m.tabs) <= 1 {
		return nil // Keep at least one tab
	}

	// Remove current tab
	m.tabs = append(m.tabs[:m.activeTab], m.tabs[m.activeTab+1:]...)

	// Adjust active tab
	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}

	// Update editor content
	if len(m.tabs) > 0 {
		m.codeEditor.SetValue(m.tabs[m.activeTab].Content)
	}

	return nil
}
