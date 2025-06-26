package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFileBrowser renders the file browser pane
func (m Model) renderFileBrowser(width, height int) string {
	var content strings.Builder
	
	// Header
	content.WriteString(m.styles.Title.Render("📁 FILES"))
	content.WriteString("\n")
	
	// Current directory
	content.WriteString(fmt.Sprintf("📂 %s\n", m.currentDir))
	content.WriteString("\n")
	
	// Files
	for i, file := range m.files {
		var icon string
		if file.IsDir {
			icon = "📁"
		} else {
			switch {
			case strings.HasSuffix(file.Name, ".go"):
				icon = "🔧"
			case strings.HasSuffix(file.Name, ".rs"):
				icon = "🦀"
			case strings.HasSuffix(file.Name, ".py"):
				icon = "🐍"
			case strings.HasSuffix(file.Name, ".js"), strings.HasSuffix(file.Name, ".ts"):
				icon = "📜"
			case strings.HasSuffix(file.Name, ".md"):
				icon = "📄"
			case strings.HasSuffix(file.Name, ".yaml"), strings.HasSuffix(file.Name, ".yml"):
				icon = "⚙️"
			default:
				icon = "📄"
			}
		}
		
		line := fmt.Sprintf("%s %s", icon, file.Name)
		
		if i == m.selectedFile {
			line = m.styles.FileSelected.Render(line)
		} else {
			line = m.styles.File.Render(line)
		}
		
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Instructions
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("↑/↓: Navigate"))
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("Enter: Open"))
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("Backspace: Up dir"))
	
	// Apply border
	borderStyle := m.styles.Border
	if m.focusedPane == FileBrowserPane {
		borderStyle = m.styles.BorderActive
	}
	
	return borderStyle.
		Width(width - 2).
		Height(height - 2).
		Render(content.String())
}

// renderCodeEditor renders the code editor pane
func (m Model) renderCodeEditor(width, height int) string {
	var content strings.Builder
	
	// Tabs
	if len(m.tabs) > 0 {
		var tabs []string
		for i, tab := range m.tabs {
			tabName := tab.Name
			if tab.Modified {
				tabName += "*"
			}
			
			if i == m.activeTab {
				tabs = append(tabs, m.styles.TabActive.Render(tabName))
			} else {
				tabs = append(tabs, m.styles.Tab.Render(tabName))
			}
		}
		content.WriteString(strings.Join(tabs, ""))
		content.WriteString("\n")
	}
	
	// Editor content
	editorContent := m.codeEditor.View()
	content.WriteString(editorContent)
	
	// Instructions at bottom
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("Ctrl+S: Save | Ctrl+W: Close Tab"))
	
	// Apply border
	borderStyle := m.styles.Border
	if m.focusedPane == CodeEditorPane {
		borderStyle = m.styles.BorderActive
	}
	
	return borderStyle.
		Width(width - 2).
		Height(height - 2).
		Render(content.String())
}

// renderAIChat renders the AI chat pane
func (m Model) renderAIChat(width, height int) string {
	var content strings.Builder
	
	// Header
	content.WriteString(m.styles.Title.Render("🤖 AI ASSISTANT"))
	content.WriteString("\n\n")
	
	// Chat messages
	chatContent := m.chatViewport.View()
	content.WriteString(chatContent)
	
	// Input area
	content.WriteString("\n")
	content.WriteString("─────────────────────")
	content.WriteString("\n")
	content.WriteString(m.chatInput.View())
	
	// Instructions
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("Enter: Send | Alt+Enter: New line"))
	
	// Apply border
	borderStyle := m.styles.Border
	if m.focusedPane == AIChatPane {
		borderStyle = m.styles.BorderActive
	}
	
	return borderStyle.
		Width(width - 2).
		Height(height - 2).
		Render(content.String())
}

// renderTerminal renders the terminal/output pane
func (m Model) renderTerminal(width, height int) string {
	var content strings.Builder
	
	// Header with tabs
	content.WriteString(m.styles.Title.Render("💻 OUTPUT"))
	content.WriteString(" | ")
	content.WriteString(m.styles.File.Render("Terminal"))
	content.WriteString(" | ")
	content.WriteString(m.styles.File.Render("Problems"))
	content.WriteString("\n")
	
	// Terminal content
	terminalContent := m.terminalViewport.View()
	if terminalContent == "" {
		terminalContent = m.terminalOutput
	}
	content.WriteString(terminalContent)
	
	// Instructions
	content.WriteString("\n")
	content.WriteString(m.styles.File.Render("Ctrl+L: Clear"))
	
	// Apply border
	borderStyle := m.styles.Border
	if m.focusedPane == TerminalPane {
		borderStyle = m.styles.BorderActive
	}
	
	return borderStyle.
		Width(width - 2).
		Height(height - 2).
		Render(content.String())
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	var status strings.Builder
	
	// Current file info
	if len(m.tabs) > 0 && m.activeTab < len(m.tabs) {
		tab := m.tabs[m.activeTab]
		status.WriteString(fmt.Sprintf("📄 %s", tab.Name))
		if tab.Modified {
			status.WriteString(" (modified)")
		}
		status.WriteString(" | ")
	}
	
	// Current pane
	paneNames := map[Pane]string{
		FileBrowserPane: "Files",
		CodeEditorPane:  "Editor", 
		AIChatPane:     "AI Chat",
		TerminalPane:   "Terminal",
	}
	status.WriteString(fmt.Sprintf("Focus: %s", paneNames[m.focusedPane]))
	
	// Keyboard shortcuts
	status.WriteString(" | Tab: Switch | Ctrl+1-4: Quick switch | Ctrl+C: Quit")
	
	return m.styles.File.
		Width(m.width).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15")).
		Render(status.String())
}
