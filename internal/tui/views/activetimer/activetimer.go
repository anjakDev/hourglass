package activetimer

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	tmr "github.com/anjakDev/hourglass/internal/timer"
	"github.com/anjakDev/hourglass/internal/tui/styles"
)

// TickMsg is sent every second to refresh the display. Exported so tests can
// inject a known time without relying on time.Now().
type TickMsg time.Time

// StopSessionMsg is returned when the user presses s. The App handles the DB
// write and view switch.
type StopSessionMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
}

// Model is the live timer view.
type Model struct {
	projectName string
	timer       *tmr.Timer
	now         time.Time
}

// New returns a Model for the given project. t must already be started; now is
// the start instant (used for the initial display before the first tick).
func New(projectName string, t *tmr.Timer, now time.Time) Model {
	return Model{projectName: projectName, timer: t, now: now}
}

// Init starts the per-second tick.
func (m Model) Init() tea.Cmd { return tickCmd() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.now = time.Time(msg)
		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			switch m.timer.State() {
			case tmr.StateRunning:
				_ = m.timer.Break(m.now)
			case tmr.StateOnBreak:
				_ = m.timer.Resume(m.now)
			}
		case "s":
			return m, func() tea.Msg { return StopSessionMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	onBreak := m.timer.State() == tmr.StateOnBreak

	work := m.timer.WorkDuration(m.now)
	clockStr := formatClock(work)

	var timerStyle = styles.TimerLarge
	if onBreak {
		timerStyle = styles.TimerBreak
	}

	var sb strings.Builder
	sb.WriteString(styles.Accent.Render("● "+m.projectName) + "\n\n")
	sb.WriteString(timerStyle.Render(clockStr) + "\n")

	if onBreak {
		breakDur := m.timer.BreakDuration(m.now)
		sb.WriteString(styles.Muted.Render(fmt.Sprintf("  on break  +%s", styles.FormatDuration(breakDur))) + "\n")
	} else {
		sb.WriteString(styles.Muted.Render("  work time") + "\n")
	}

	sb.WriteString("\n")
	if onBreak {
		sb.WriteString(styles.StatusBar.Render("  [b] resume   [s] stop"))
	} else {
		sb.WriteString(styles.StatusBar.Render("  [b] break   [s] stop"))
	}
	return sb.String()
}

// formatClock formats a duration as "HH:MM:SS".
func formatClock(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
