package util

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

// InfoType represents the type of info message
type InfoType int

const (
	InfoTypeInfo InfoType = iota
	InfoTypeWarn
	InfoTypeError
	InfoTypeSuccess
)

// InfoMsg represents an informational message
type InfoMsg struct {
	Type InfoType
	Msg  string
	TTL  time.Duration
}

// ClearStatusMsg clears the status message
type ClearStatusMsg struct{}

// CmdHandler wraps a message in a tea.Cmd
func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// ReportInfo creates a command to report an info message
func ReportInfo(msg string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeInfo,
		Msg:  msg,
		TTL:  3 * time.Second,
	})
}

// ReportWarn creates a command to report a warning message
func ReportWarn(msg string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeWarn,
		Msg:  msg,
		TTL:  5 * time.Second,
	})
}

// ReportError creates a command to report an error message
func ReportError(err error) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeError,
		Msg:  err.Error(),
		TTL:  10 * time.Second,
	})
}

// ReportSuccess creates a command to report a success message
func ReportSuccess(msg string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeSuccess,
		Msg:  msg,
		TTL:  3 * time.Second,
	})
}

// Clamp constrains a value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PadString pads a string to a specific length
func PadString(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	padding := length - len(s)
	padStr := string(padChar)
	for i := 0; i < padding; i++ {
		s += padStr
	}
	return s
}

// CenterString centers a string within a given width
func CenterString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	
	result := ""
	for i := 0; i < leftPad; i++ {
		result += " "
	}
	result += s
	for i := 0; i < rightPad; i++ {
		result += " "
	}
	return result
}

// FormatFileSize formats a file size in bytes to human readable format
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return string(rune(bytes/div)) + " " + "KMGTPE"[exp:exp+1] + "B"
}

// Debounce creates a debounced function that delays execution
func Debounce(fn func(), delay time.Duration) func() {
	var timer *time.Timer
	return func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, fn)
	}
}

// Throttle creates a throttled function that limits execution frequency
func Throttle(fn func(), interval time.Duration) func() {
	var lastCall time.Time
	return func() {
		if time.Since(lastCall) >= interval {
			fn()
			lastCall = time.Now()
		}
	}
}
