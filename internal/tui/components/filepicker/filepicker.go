// Package filepicker provides a file picker component for CodeForge
// Based on the Bubbles filepicker component
package filepicker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New returns a new filepicker model with default styling and key bindings.
func New() Model {
	return Model{
		id:               nextID(),
		CurrentDirectory: ".",
		Cursor:           ">",
		AllowedTypes:     []string{},
		selected:         0,
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       false,
		DirAllowed:       true, // Allow directory selection for CodeForge
		FileAllowed:      true,
		AutoHeight:       true,
		Height:           0,
		max:              0,
		min:              0,
		selectedStack:    newStack(),
		minStack:         newStack(),
		maxStack:         newStack(),
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

// FileSelectedMsg is sent when a file is selected
type FileSelectedMsg struct {
	Path string
}

const (
	marginBottom  = 5
	fileSizeWidth = 7
	paddingLeft   = 2
)

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding
}

// DefaultKeyMap defines the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer defines the default styling for the file picker,
// with a given Lip Gloss renderer.
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		DisabledCursor:   r.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           r.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          r.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        r.NewStyle().Foreground(lipgloss.Color("99")),
		File:             r.NewStyle(),
		DisabledFile:     r.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: r.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       r.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         r.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         r.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory:   r.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."),
	}
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	files           []os.DirEntry
	ShowPermissions bool
	ShowSize        bool
	ShowHidden      bool
	DirAllowed      bool
	FileAllowed     bool

	FileSelected  string
	selected      int
	selectedStack stack

	min      int
	max      int
	maxStack stack
	minStack stack

	// Height of the picker.
	Height     int
	AutoHeight bool

	Cursor string
	Styles Styles

	// CodeForge specific fields
	focused bool
	width   int
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m *Model) pushView(selected, minimum, maximum int) {
	m.selectedStack.Push(selected)
	m.minStack.Push(minimum)
	m.maxStack.Push(maximum)
}

func (m *Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

// SetFocused sets the focus state
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// SetSize sets the dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.Height = height
}

// LoadDirectory directly loads files from a directory (synchronous)
func (m *Model) LoadDirectory(path string) error {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
			return dirEntries[i].Name() < dirEntries[j].Name()
		}
		return dirEntries[i].IsDir()
	})

	var filteredEntries []os.DirEntry
	for _, entry := range dirEntries {
		if !m.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	m.files = filteredEntries
	m.selected = 0
	m.min = 0
	if m.Height > 0 {
		m.max = m.Height - 1
	} else {
		m.max = len(m.files) - 1 // Fallback if height not set
	}

	// Files loaded successfully
	return nil
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			m.selected = len(m.files) - 1
			m.min = len(m.files) - m.Height
			m.max = len(m.files) - 1
		case key.Matches(msg, m.KeyMap.Down):
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			if m.selected > m.max {
				m.min++
				m.max++
			}
		case key.Matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.min {
				m.min--
				m.max--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			m.selected += m.Height
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			m.min = m.selected - m.Height + 1
			m.max = m.selected
		case key.Matches(msg, m.KeyMap.PageUp):
			m.selected -= m.Height
			if m.selected < 0 {
				m.selected = 0
			}
			m.min = m.selected
			m.max = m.selected + m.Height - 1
		case key.Matches(msg, m.KeyMap.Back):
			m.pushView(m.selected, m.min, m.max)
			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			return m, m.readDir(m.CurrentDirectory)
		case key.Matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}
			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath := filepath.Join(m.CurrentDirectory, f.Name())
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if isDir {
				m.pushView(m.selected, m.min, m.max)
				m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
				m.selected = 0
				m.min = 0
				m.max = m.Height - 1
				return m, m.readDir(m.CurrentDirectory)
			}

			// File selected
			if m.FileAllowed {
				m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				return m, func() tea.Msg {
					return FileSelectedMsg{Path: m.Path}
				}
			}
		case key.Matches(msg, m.KeyMap.Select):
			if len(m.files) == 0 {
				break
			}
			f := m.files[m.selected]
			if f.IsDir() && m.DirAllowed {
				m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				return m, func() tea.Msg {
					return FileSelectedMsg{Path: m.Path}
				}
			} else if !f.IsDir() && m.FileAllowed {
				m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				return m, func() tea.Msg {
					return FileSelectedMsg{Path: m.Path}
				}
			}
		}

	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		if m.selectedStack.Length() > 0 {
			m.selected, m.min, m.max = m.popView()
		} else {
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
		}

	case errorMsg:
		// Handle error
		break
	}

	return m, nil
}

func (m Model) readDir(path string) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		var filteredEntries []os.DirEntry
		for _, entry := range dirEntries {
			if !m.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			filteredEntries = append(filteredEntries, entry)
		}

		return readDirMsg{id: m.id, entries: filteredEntries}
	}
}

// View implements tea.Model
func (m Model) View() string {
	if len(m.files) == 0 {
		return m.Styles.EmptyDirectory.String()
	}

	var s strings.Builder

	start := m.min
	if start < 0 {
		start = 0
	}
	end := m.max + 1
	if end > len(m.files) {
		end = len(m.files)
	}

	for i := start; i < end; i++ {
		if i >= len(m.files) {
			break
		}

		var symlinkPath string
		f := m.files[i]
		info, err := f.Info()
		if err != nil {
			continue
		}

		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath = filepath.Join(m.CurrentDirectory, f.Name())
			info, err := os.Stat(symlinkPath)
			if err == nil && info.IsDir() {
				isDir = true
			}
		}

		var name string
		if isDir {
			name = m.Styles.Directory.Render(f.Name())
		} else if isSymlink {
			name = m.Styles.Symlink.Render(f.Name())
		} else {
			name = m.Styles.File.Render(f.Name())
		}

		// Add cursor
		cursor := " "
		if i == m.selected {
			if m.focused {
				cursor = m.Styles.Cursor.Render(m.Cursor)
				name = m.Styles.Selected.Render(name)
			} else {
				cursor = m.Styles.DisabledCursor.Render(m.Cursor)
				name = m.Styles.DisabledSelected.Render(name)
			}
		}

		line := cursor + " " + name

		// Add file size
		if m.ShowSize && !isDir {
			size := humanize.Bytes(uint64(info.Size()))
			line += m.Styles.FileSize.Render(size)
		}

		// Add permissions
		if m.ShowPermissions {
			perms := info.Mode().String()
			line += " " + m.Styles.Permission.Render(perms)
		}

		s.WriteString(line)
		if i < end-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}
