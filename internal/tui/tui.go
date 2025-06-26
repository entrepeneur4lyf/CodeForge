package tui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/mattn/go-isatty"
)

// IsTTY returns true if the current environment is a TTY.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

type (
	llmResponseMsg string
	errMsg         struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

type model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []string
	senderStyle lipgloss.Style
	err         error
}

func InitialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	return model{
		textarea:    ta,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tacmd tea.Cmd
		vpcmd tea.Cmd
	)

	m.textarea, tacmd = m.textarea.Update(msg)
	m.viewport, vpcmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			userInput := m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+userInput)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			return m, func() tea.Msg {
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
	case llmResponseMsg:
		m.messages = append(m.messages, "AI: "+string(msg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tacmd, vpcmd)
}

func (m model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

func Start() {
	// The LLM system should already be initialized by the root command
	p := tea.NewProgram(InitialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
