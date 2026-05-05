package sessionlog

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui/styles"
)

// BackMsg is returned when the user navigates back.
type BackMsg struct{}

// Model is the read-only session log view for one project.
type Model struct {
	projectName string
	sessions    []repository.Session
}

// New constructs the model with pre-loaded session data.
func New(projectName string, sessions []repository.Session) Model {
	return Model{projectName: projectName, sessions: sessions}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder

	header := fmt.Sprintf("Sessions — %s — today", m.projectName)
	sb.WriteString(styles.Title.Render(header) + "\n\n")

	if len(m.sessions) == 0 {
		sb.WriteString(styles.Muted.Render("  No sessions today.") + "\n")
	} else {
		var total time.Duration
		for _, s := range m.sessions {
			sb.WriteString(formatRow(s) + "\n")
			total += s.WorkDuration()
		}
		sb.WriteString("\n")
		sb.WriteString(styles.Bold.Render("  Total: "+styles.FormatDuration(total)) + "\n")
	}

	sb.WriteString("\n" + styles.StatusBar.Render("  [esc] back"))
	return sb.String()
}

func formatRow(s repository.Session) string {
	start := formatTime(s.StartedAt)
	if s.EndedAt == nil {
		return styles.Muted.Render(fmt.Sprintf("  %s → running", start))
	}
	end := formatTime(*s.EndedAt)
	dur := styles.FormatDuration(s.WorkDuration())
	row := fmt.Sprintf("  %s → %s    %s", start, end, dur)
	if s.BreakDurationSeconds > 0 {
		breakDur := styles.FormatDuration(time.Duration(s.BreakDurationSeconds) * time.Second)
		row += fmt.Sprintf("  (%s break)", breakDur)
	}
	return row
}

func formatTime(t time.Time) string {
	return t.Local().Format("15:04")
}
