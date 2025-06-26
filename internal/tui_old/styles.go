package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors inspired by OpenCode's clean design
var (
	// Base colors
	primaryColor   = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#656d76", Dark: "#8b949e"}
	accentColor    = lipgloss.AdaptiveColor{Light: "#8250df", Dark: "#a5a5ff"}
	
	// Status colors
	errorColor   = lipgloss.AdaptiveColor{Light: "#d1242f", Dark: "#f85149"}
	warningColor = lipgloss.AdaptiveColor{Light: "#bf8700", Dark: "#f0883e"}
	successColor = lipgloss.AdaptiveColor{Light: "#1a7f37", Dark: "#56d364"}
	infoColor    = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}
	
	// Text colors
	textColor         = lipgloss.AdaptiveColor{Light: "#24292f", Dark: "#f0f6fc"}
	textMutedColor    = lipgloss.AdaptiveColor{Light: "#656d76", Dark: "#8b949e"}
	textEmphasized    = lipgloss.AdaptiveColor{Light: "#1f2328", Dark: "#ffffff"}
	
	// Background colors
	backgroundColor          = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"}
	backgroundSecondaryColor = lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#161b22"}
	backgroundDarkerColor    = lipgloss.AdaptiveColor{Light: "#eaeef2", Dark: "#21262d"}
	
	// Border colors
	borderNormalColor  = lipgloss.AdaptiveColor{Light: "#d0d7de", Dark: "#30363d"}
	borderFocusedColor = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}
	borderDimColor     = lipgloss.AdaptiveColor{Light: "#eaeef2", Dark: "#21262d"}
)

// Style functions inspired by OpenCode
func baseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(backgroundColor).
		Foreground(textColor)
}

func paneBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderNormalColor).
		Padding(0, 1)
}

func focusedPaneBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderFocusedColor).
		Padding(0, 1)
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textEmphasized).
		Bold(true).
		Padding(0, 1)
}

func selectedItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(backgroundColor).
		Bold(true).
		Padding(0, 1)
}

func normalItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textColor).
		Padding(0, 1)
}

func mutedTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textMutedColor)
}

func helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textMutedColor).
		Italic(true).
		Padding(0, 1)
}

func statusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(backgroundSecondaryColor).
		Foreground(textColor).
		Padding(0, 1)
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)
}

func successStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)
}

func warningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true)
}

func infoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(infoColor).
		Bold(true)
}

// File type icons (simple text-based like OpenCode)
func getFileIcon(filename string) string {
	switch {
	case filename == "..":
		return "📁"
	case isDirectory(filename):
		return "📁"
	case isGoFile(filename):
		return "🐹"
	case isRustFile(filename):
		return "🦀"
	case isPythonFile(filename):
		return "🐍"
	case isJavaScriptFile(filename):
		return "📜"
	case isTypeScriptFile(filename):
		return "📘"
	case isJavaFile(filename):
		return "☕"
	case isCppFile(filename):
		return "⚡"
	case isConfigFile(filename):
		return "⚙️"
	case isMarkdownFile(filename):
		return "📝"
	case isImageFile(filename):
		return "🖼️"
	default:
		return "📄"
	}
}

// Helper functions for file type detection
func isDirectory(filename string) bool {
	// This should be checked from file info, but for now assume directories don't have extensions
	return filename == ".." || (len(filename) > 0 && filename[0] != '.' && !hasExtension(filename))
}

func hasExtension(filename string) bool {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return true
		}
		if filename[i] == '/' {
			break
		}
	}
	return false
}

func isGoFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".go"
}

func isRustFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".rs"
}

func isPythonFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".py"
}

func isJavaScriptFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".js"
}

func isTypeScriptFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".ts"
}

func isJavaFile(filename string) bool {
	return hasExtension(filename) && len(filename) > 5 && filename[len(filename)-5:] == ".java"
}

func isCppFile(filename string) bool {
	return hasExtension(filename) && (filename[len(filename)-4:] == ".cpp" || filename[len(filename)-2:] == ".h")
}

func isConfigFile(filename string) bool {
	return filename == "config.yaml" || filename == "config.yml" || filename == "Cargo.toml" || 
		   filename == "package.json" || filename == "go.mod" || filename == "requirements.txt"
}

func isMarkdownFile(filename string) bool {
	return hasExtension(filename) && filename[len(filename)-3:] == ".md"
}

func isImageFile(filename string) bool {
	if !hasExtension(filename) {
		return false
	}
	ext := filename[len(filename)-4:]
	return ext == ".png" || ext == ".jpg" || ext == ".gif" || ext == ".svg"
}
