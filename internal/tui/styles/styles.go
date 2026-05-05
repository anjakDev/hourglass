package styles

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Palette
var (
	colorAccent  = lipgloss.Color("86")  // green-cyan
	colorMuted   = lipgloss.Color("241") // dim gray
	colorDanger  = lipgloss.Color("196") // red
	colorRunning = lipgloss.Color("226") // yellow
)

// Shared styles
var (
	Bold   = lipgloss.NewStyle().Bold(true)
	Muted  = lipgloss.NewStyle().Foreground(colorMuted)
	Accent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	Danger = lipgloss.NewStyle().Foreground(colorDanger)

	// TimerLarge is used for the big elapsed-time display.
	TimerLarge = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			PaddingLeft(2)

	// TimerBreak dims the timer display while on a break.
	TimerBreak = lipgloss.NewStyle().
			Foreground(colorRunning).
			Bold(true).
			PaddingLeft(2)

	// StatusBar is the bottom key-hint line.
	StatusBar = lipgloss.NewStyle().Foreground(colorMuted)

	// Title is the top app name/screen title.
	Title = lipgloss.NewStyle().Bold(true).PaddingBottom(1)

	// Selected highlights the cursor row in a list.
	Selected = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	// TotalCol right-aligns duration totals.
	TotalCol = lipgloss.NewStyle().Foreground(colorMuted)
)

// FormatDuration formats a duration as "1h 23m" or "45m" or "0m".
// Sub-minute precision is discarded.
func FormatDuration(d time.Duration) string {
	d = d.Truncate(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
