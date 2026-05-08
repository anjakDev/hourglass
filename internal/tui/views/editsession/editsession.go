package editsession

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/tui/styles"
)

const timeFormat = "2006-01-02 15:04"

// SaveMsg is returned when the user is done — either confirming edits or
// cancelling. StartedAt/EndedAt hold the times to persist.
type SaveMsg struct {
	StartedAt time.Time
	EndedAt   time.Time
}

// Model is the post-stop session edit view.
type Model struct {
	projectName string
	origStart   time.Time
	origEnd     time.Time
	startedAt   time.Time
	endedAt     time.Time
	cursor      int // 0 = start field, 1 = end field
	editing     bool
	inputValue  string
	inputErr    string
}

// New creates a Model showing the stopped session. Times are stored in UTC.
func New(projectName string, startedAt, endedAt time.Time) Model {
	return Model{
		projectName: projectName,
		origStart:   startedAt.UTC(),
		origEnd:     endedAt.UTC(),
		startedAt:   startedAt.UTC(),
		endedAt:     endedAt.UTC(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

// Exported accessors used by tests and by app.go.
func (m Model) StartedAt() time.Time { return m.startedAt }
func (m Model) EndedAt() time.Time   { return m.endedAt }
func (m Model) IsEditing() bool      { return m.editing }
func (m Model) InputValue() string   { return m.inputValue }
func (m Model) InputErr() string     { return m.inputErr }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.editing {
		return m.updateEditing(key)
	}
	return m.updateSummary(key)
}

func (m Model) updateSummary(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		return m, func() tea.Msg { return SaveMsg{StartedAt: m.startedAt, EndedAt: m.endedAt} }
	case tea.KeyEsc:
		return m, func() tea.Msg { return SaveMsg{StartedAt: m.origStart, EndedAt: m.origEnd} }
	case tea.KeyTab:
		m.cursor = 1 - m.cursor
		return m, nil
	}
	switch key.String() {
	case "j", "down":
		if m.cursor == 0 {
			m.cursor = 1
		}
	case "k", "up":
		if m.cursor == 1 {
			m.cursor = 0
		}
	case "e":
		m.editing = true
		m.inputErr = ""
		if m.cursor == 0 {
			m.inputValue = m.startedAt.In(time.Local).Format(timeFormat)
		} else {
			m.inputValue = m.endedAt.In(time.Local).Format(timeFormat)
		}
	}
	return m, nil
}

func (m Model) updateEditing(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		parsed, err := time.ParseInLocation(timeFormat, m.inputValue, time.Local)
		if err != nil {
			m.inputErr = "invalid format – use YYYY-MM-DD HH:MM"
			return m, nil
		}
		parsed = parsed.UTC()
		if m.cursor == 0 {
			if !parsed.Before(m.endedAt) {
				m.inputErr = "start must be before end"
				return m, nil
			}
			m.startedAt = parsed
		} else {
			if !parsed.After(m.startedAt) {
				m.inputErr = "end must be after start"
				return m, nil
			}
			m.endedAt = parsed
		}
		m.editing = false
		m.inputErr = ""
		return m, nil
	case tea.KeyEsc:
		m.editing = false
		m.inputErr = ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.inputValue) > 0 {
			// Safe: inputValue is always ASCII from Format output
			m.inputValue = m.inputValue[:len(m.inputValue)-1]
		}
		return m, nil
	case tea.KeyCtrlU:
		m.inputValue = ""
		return m, nil
	case tea.KeyRunes:
		m.inputValue += string(key.Runes)
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(styles.Title.Render(fmt.Sprintf("Edit session – %s", m.projectName)) + "\n\n")

	startStr := m.startedAt.In(time.Local).Format(timeFormat)
	endStr := m.endedAt.In(time.Local).Format(timeFormat)
	dur := m.endedAt.Sub(m.startedAt)

	if m.editing {
		inputDisplay := styles.Selected.Render(fmt.Sprintf("[%s]", m.inputValue))
		if m.cursor == 0 {
			sb.WriteString(fmt.Sprintf("  Started   %s\n", inputDisplay))
			sb.WriteString(fmt.Sprintf("  Ended     %s\n", styles.Muted.Render(endStr)))
		} else {
			sb.WriteString(fmt.Sprintf("  Started   %s\n", styles.Muted.Render(startStr)))
			sb.WriteString(fmt.Sprintf("  Ended     %s\n", inputDisplay))
		}
	} else {
		if m.cursor == 0 {
			sb.WriteString(fmt.Sprintf("  Started   %s\n", styles.Selected.Render(startStr)))
			sb.WriteString(fmt.Sprintf("  Ended     %s\n", styles.Muted.Render(endStr)))
		} else {
			sb.WriteString(fmt.Sprintf("  Started   %s\n", styles.Muted.Render(startStr)))
			sb.WriteString(fmt.Sprintf("  Ended     %s\n", styles.Selected.Render(endStr)))
		}
	}

	sb.WriteString(fmt.Sprintf("  Duration  %s\n", styles.Accent.Render(styles.FormatDuration(dur))))

	if m.inputErr != "" {
		sb.WriteString("\n" + styles.Warn.Render("  "+m.inputErr) + "\n")
	}

	if m.editing {
		sb.WriteString("\n" + styles.StatusBar.Render("  [enter] confirm   [esc] discard"))
	} else {
		sb.WriteString("\n" + styles.StatusBar.Render("  [tab/j/k] switch   [e] edit   [enter] save   [esc] keep original"))
	}

	return sb.String()
}
