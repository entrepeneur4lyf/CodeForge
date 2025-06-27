package layout

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplitPaneLayout interface {
	tea.Model
	Sizeable
	Bindings
	SetLeftPanel(panel Container) tea.Cmd
	SetRightPanel(panel Container) tea.Cmd
	SetBottomPanel(panel Container) tea.Cmd

	ClearLeftPanel() tea.Cmd
	ClearRightPanel() tea.Cmd
	ClearBottomPanel() tea.Cmd
}

type splitPaneLayout struct {
	width         int
	height        int
	ratio         float64
	verticalRatio float64

	rightPanel  Container
	leftPanel   Container
	bottomPanel Container
}

type SplitPaneOption func(*splitPaneLayout)

func (s *splitPaneLayout) Init() tea.Cmd {
	var cmds []tea.Cmd

	if s.leftPanel != nil {
		cmds = append(cmds, s.leftPanel.Init())
	}

	if s.rightPanel != nil {
		cmds = append(cmds, s.rightPanel.Init())
	}

	if s.bottomPanel != nil {
		cmds = append(cmds, s.bottomPanel.Init())
	}

	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, s.SetSize(msg.Width, msg.Height)
	}

	if s.rightPanel != nil {
		u, cmd := s.rightPanel.Update(msg)
		s.rightPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.leftPanel != nil {
		u, cmd := s.leftPanel.Update(msg)
		s.leftPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.bottomPanel != nil {
		u, cmd := s.bottomPanel.Update(msg)
		s.bottomPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *splitPaneLayout) View() string {
	var topSection string

	if s.leftPanel != nil && s.rightPanel != nil {
		leftView := s.leftPanel.View()
		rightView := s.rightPanel.View()
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	} else if s.leftPanel != nil {
		topSection = s.leftPanel.View()
	} else if s.rightPanel != nil {
		topSection = s.rightPanel.View()
	} else {
		topSection = ""
	}

	if s.bottomPanel != nil {
		bottomView := s.bottomPanel.View()
		if topSection != "" {
			return lipgloss.JoinVertical(lipgloss.Left, topSection, bottomView)
		}
		return bottomView
	}

	return topSection
}

func (s *splitPaneLayout) SetSize(width, height int) tea.Cmd {
	s.width = width
	s.height = height

	var cmds []tea.Cmd

	// Calculate dimensions for panels
	var topHeight int
	if s.bottomPanel != nil {
		topHeight = int(float64(height) * s.verticalRatio)
		bottomHeight := height - topHeight
		cmd := s.bottomPanel.SetSize(width, bottomHeight)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		topHeight = height
	}

	// Set left and right panel sizes
	if s.leftPanel != nil && s.rightPanel != nil {
		leftWidth := int(float64(width) * s.ratio)
		rightWidth := width - leftWidth

		cmd := s.leftPanel.SetSize(leftWidth, topHeight)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		cmd = s.rightPanel.SetSize(rightWidth, topHeight)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else if s.leftPanel != nil {
		cmd := s.leftPanel.SetSize(width, topHeight)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else if s.rightPanel != nil {
		cmd := s.rightPanel.SetSize(width, topHeight)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) GetSize() (int, int) {
	return s.width, s.height
}

func (s *splitPaneLayout) BindingKeys() []key.Binding {
	var bindings []key.Binding

	if s.leftPanel != nil {
		bindings = append(bindings, s.leftPanel.BindingKeys()...)
	}

	if s.rightPanel != nil {
		bindings = append(bindings, s.rightPanel.BindingKeys()...)
	}

	if s.bottomPanel != nil {
		bindings = append(bindings, s.bottomPanel.BindingKeys()...)
	}

	return bindings
}

func (s *splitPaneLayout) SetLeftPanel(panel Container) tea.Cmd {
	s.leftPanel = panel
	if panel != nil {
		return panel.Init()
	}
	return nil
}

func (s *splitPaneLayout) SetRightPanel(panel Container) tea.Cmd {
	s.rightPanel = panel
	if panel != nil {
		return panel.Init()
	}
	return nil
}

func (s *splitPaneLayout) SetBottomPanel(panel Container) tea.Cmd {
	s.bottomPanel = panel
	if panel != nil {
		return panel.Init()
	}
	return nil
}

func (s *splitPaneLayout) ClearLeftPanel() tea.Cmd {
	s.leftPanel = nil
	return nil
}

func (s *splitPaneLayout) ClearRightPanel() tea.Cmd {
	s.rightPanel = nil
	return nil
}

func (s *splitPaneLayout) ClearBottomPanel() tea.Cmd {
	s.bottomPanel = nil
	return nil
}

func NewSplitPaneLayout(options ...SplitPaneOption) SplitPaneLayout {
	s := &splitPaneLayout{
		ratio:         0.3, // Default 30% left, 70% right
		verticalRatio: 0.8, // Default 80% top, 20% bottom
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func WithRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.ratio = ratio
	}
}

func WithVerticalRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.verticalRatio = ratio
	}
}
