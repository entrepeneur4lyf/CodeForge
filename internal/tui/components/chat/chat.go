package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
	"github.com/entrepeneur4lyf/codeforge/internal/tui/theme"
)

// Message represents a chat message
type Message struct {
	ID        string
	Content   string
	Role      string // "user" or "assistant"
	Timestamp time.Time
	Rendered  string // Cached rendered content
}

// ChatModel represents the chat component
type ChatModel struct {
	messages    []Message
	input       textarea.Model
	viewport    viewport.Model
	renderer    *glamour.TermRenderer
	width       int
	height      int
	focused     bool
	inputHeight int
	animation   harmonica.Spring
	scrollPos   float64

	// Model configuration
	currentModel *models.ModelSummary
	modelAPI     *models.ModelAPI
	modelChanged bool
}

// MessageSentMsg is sent when a message is submitted
type MessageSentMsg struct {
	Content string
}

// MessageReceivedMsg is sent when a response is received
type MessageReceivedMsg struct {
	Content string
	ID      string
}

// ModelChangedMsg is sent when the chat model is changed
type ModelChangedMsg struct {
	Model models.ModelSummary
}

// NewChatModel creates a new chat model
func NewChatModel(modelAPI *models.ModelAPI) *ChatModel {
	// Create input textarea
	input := textarea.New()
	input.Placeholder = "Ask me anything about your code..."
	input.Focus()
	input.CharLimit = 4000
	input.SetWidth(50)
	input.SetHeight(3)
	input.ShowLineNumbers = false

	// Create viewport for messages
	vp := viewport.New(50, 20)
	vp.SetContent("")

	// Create glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(50),
	)
	if err != nil {
		log.Error("Failed to create glamour renderer", "error", err)
		renderer = nil
	}

	log.Info("Created chat model")

	return &ChatModel{
		messages:     make([]Message, 0),
		input:        input,
		viewport:     vp,
		renderer:     renderer,
		inputHeight:  3,
		modelAPI:     modelAPI,
		currentModel: nil, // Will be set when a model is selected
		modelChanged: false,
	}
}

// Init implements tea.Model
func (cm *ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
func (cm *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if !cm.focused {
			return cm, nil
		}

		// Handle mouse events in chat area
		switch msg.Action {
		case tea.MouseActionPress:
			if msg.Button == tea.MouseButtonLeft {
				// Check if click is in input area (bottom of chat)
				inputY := cm.height - 3 // Input is typically at the bottom
				if msg.Y >= inputY {
					log.Debug("Mouse click in chat input area", "x", msg.X, "y", msg.Y)
					// Focus the input and forward mouse event
					var cmd tea.Cmd
					cm.input, cmd = cm.input.Update(msg)
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				} else {
					log.Debug("Mouse click in chat messages area", "x", msg.X, "y", msg.Y)
					// Click in messages area - could be used for message selection later
				}
			}
		case tea.MouseActionMotion:
			if msg.Button == tea.MouseButtonWheelUp {
				// Handle scroll up
				log.Debug("Mouse scroll up in chat")
				cm.viewport.LineUp(3)
			} else if msg.Button == tea.MouseButtonWheelDown {
				// Handle scroll down
				log.Debug("Mouse scroll down in chat")
				cm.viewport.LineDown(3)
			}
		}

	case tea.KeyMsg:
		if !cm.focused {
			return cm, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return cm, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+l"))):
			// Clear chat
			cm.messages = nil
			cm.updateViewport()
			log.Debug("Chat cleared")
			return cm, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Send message
			content := strings.TrimSpace(cm.input.Value())
			if content != "" {
				cm.addUserMessage(content)
				cm.input.Reset()
				log.Debug("Message sent", "content", content[:min(50, len(content))])
				return cm, func() tea.Msg {
					return MessageSentMsg{Content: content}
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+enter"))):
			// Add new line (let textarea handle it)
			fallthrough
		default:
			// Forward to input if it's focused
			var cmd tea.Cmd
			cm.input, cmd = cm.input.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case MessageReceivedMsg:
		cm.addAssistantMessage(msg.Content, msg.ID)
		log.Debug("Message received", "id", msg.ID)

	case ModelChangedMsg:
		cm.setCurrentModel(&msg.Model)
		log.Info("Chat model changed", "model", msg.Model.Name, "provider", msg.Model.Provider)

	case tea.WindowSizeMsg:
		cm.width = msg.Width
		cm.height = msg.Height
		cm.updateSizes()

	default:
		// Update viewport for scrolling
		var cmd tea.Cmd
		cm.viewport, cmd = cm.viewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cm, tea.Batch(cmds...)
}

// View implements tea.Model
func (cm *ChatModel) View() string {
	// Messages viewport
	messagesView := cm.viewport.View()

	// Input area
	inputView := cm.renderInput()

	// Combine with proper spacing
	return lipgloss.JoinVertical(
		lipgloss.Left,
		messagesView,
		inputView,
	)
}

// addUserMessage adds a user message to the chat
func (cm *ChatModel) addUserMessage(content string) {
	message := Message{
		ID:        fmt.Sprintf("user-%d", time.Now().UnixNano()),
		Content:   content,
		Role:      "user",
		Timestamp: time.Now(),
	}

	message.Rendered = cm.renderMessage(message)
	cm.messages = append(cm.messages, message)
	cm.updateViewport()
}

// addAssistantMessage adds an assistant message to the chat
func (cm *ChatModel) addAssistantMessage(content, id string) {
	message := Message{
		ID:        id,
		Content:   content,
		Role:      "assistant",
		Timestamp: time.Now(),
	}

	message.Rendered = cm.renderMessage(message)
	cm.messages = append(cm.messages, message)
	cm.updateViewport()
}

// renderMessage renders a single message with styling
func (cm *ChatModel) renderMessage(msg Message) string {
	t := theme.CurrentTheme()

	// Render content with glamour if available
	var content string
	if cm.renderer != nil && msg.Role == "assistant" {
		rendered, err := cm.renderer.Render(msg.Content)
		if err != nil {
			log.Warn("Failed to render markdown", "error", err)
			content = msg.Content
		} else {
			content = rendered
		}
	} else {
		content = msg.Content
	}

	// Create message header
	var headerStyle lipgloss.Style
	var roleIcon string

	switch msg.Role {
	case "user":
		headerStyle = lipgloss.NewStyle().
			Foreground(t.Info()).
			Bold(true)
		roleIcon = "👤"
	case "system":
		headerStyle = lipgloss.NewStyle().
			Foreground(t.Warning()).
			Bold(true)
		roleIcon = "⚙️"
	default: // assistant
		headerStyle = lipgloss.NewStyle().
			Foreground(t.Primary()).
			Bold(true)
		roleIcon = "🤖"
	}

	timestamp := msg.Timestamp.Format("15:04")
	header := headerStyle.Render(fmt.Sprintf("%s %s • %s", roleIcon, strings.Title(msg.Role), timestamp))

	// Message content style
	contentStyle := lipgloss.NewStyle().
		Foreground(t.Text()).
		Padding(0, 2)

	styledContent := contentStyle.Render(content)

	// Combine header and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledContent,
		"", // Empty line for spacing
	)
}

// renderInput renders the input area
func (cm *ChatModel) renderInput() string {
	t := theme.CurrentTheme()

	// Input prompt
	promptStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true)

	prompt := promptStyle.Render("> ")

	// Input area
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(0, 1)

	if cm.focused {
		inputStyle = inputStyle.BorderForeground(t.BorderFocused())
	}

	input := inputStyle.Render(cm.input.View())

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Italic(true)

	help := helpStyle.Render("Enter to send • Shift+Enter for new line • Ctrl+L to clear")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		prompt,
		input,
		help,
	)
}

// updateViewport updates the viewport content with all messages
func (cm *ChatModel) updateViewport() {
	var renderedMessages []string

	for _, msg := range cm.messages {
		renderedMessages = append(renderedMessages, msg.Rendered)
	}

	content := strings.Join(renderedMessages, "\n")
	cm.viewport.SetContent(content)

	// Scroll to bottom
	cm.viewport.GotoBottom()
}

// updateSizes updates component sizes based on available space (responsive)
func (cm *ChatModel) updateSizes() {
	if cm.width <= 0 || cm.height <= 0 {
		return
	}

	// Responsive padding based on terminal size
	var padding int
	if cm.width < 60 {
		padding = 2 // Minimal padding for small terminals
	} else if cm.width < 100 {
		padding = 4 // Standard padding
	} else {
		padding = 6 // More padding for large terminals
	}

	// Calculate viewport height (total - input area - padding)
	viewportHeight := cm.height - cm.inputHeight - padding - 2 // 2 for help text
	if viewportHeight < 3 {
		viewportHeight = 3 // Absolute minimum
	}

	// Update viewport with responsive dimensions
	cm.viewport.Width = max(20, cm.width-padding) // Minimum 20 chars
	cm.viewport.Height = viewportHeight

	// Update input with responsive width
	inputWidth := max(10, cm.width-padding-4) // Minimum 10 chars
	cm.input.SetWidth(inputWidth)

	// Update glamour renderer width (responsive word wrap)
	if cm.renderer != nil {
		wrapWidth := max(40, cm.width-padding-4) // Minimum 40 chars for readability
		newRenderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(wrapWidth),
		)
		if err == nil {
			cm.renderer = newRenderer
		}
	}

	// Re-render all messages with new width
	for i := range cm.messages {
		cm.messages[i].Rendered = cm.renderMessage(cm.messages[i])
	}

	cm.updateViewport()
}

// SetFocused sets the focus state
func (cm *ChatModel) SetFocused(focused bool) {
	cm.focused = focused
	if focused {
		cm.input.Focus()
	} else {
		cm.input.Blur()
	}
}

// SetSize sets the dimensions
func (cm *ChatModel) SetSize(width, height int) {
	cm.width = width
	cm.height = height
	cm.updateSizes()
}

// GetMessages returns all messages
func (cm *ChatModel) GetMessages() []Message {
	return cm.messages
}

// setCurrentModel sets the current model for the chat
func (cm *ChatModel) setCurrentModel(model *models.ModelSummary) {
	cm.currentModel = model
	cm.modelChanged = true

	// Add a system message to indicate model change
	if model != nil {
		systemMessage := Message{
			ID:        fmt.Sprintf("system-%d", time.Now().UnixNano()),
			Content:   fmt.Sprintf("🤖 Switched to %s (%s)", model.Name, model.Provider),
			Role:      "system",
			Timestamp: time.Now(),
		}
		systemMessage.Rendered = cm.renderMessage(systemMessage)
		cm.messages = append(cm.messages, systemMessage)
		cm.updateViewport()
	}
}

// GetCurrentModel returns the currently selected model
func (cm *ChatModel) GetCurrentModel() *models.ModelSummary {
	return cm.currentModel
}

// HasModelChanged returns whether the model has been changed
func (cm *ChatModel) HasModelChanged() bool {
	return cm.modelChanged
}

// ClearModelChanged clears the model changed flag
func (cm *ChatModel) ClearModelChanged() {
	cm.modelChanged = false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
