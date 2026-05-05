package projects

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui/styles"
)

// Outgoing messages — handled by the root App model.
type StartSessionMsg struct{ ProjectID int64 }
type NewProjectMsg struct{}
type ArchiveMsg struct{ ProjectID int64 }
type ShowSessionLogMsg struct{ ProjectID int64 }

const nameColWidth = 28

// Model is the project list view.
type Model struct {
	items  []repository.ProjectTotal
	cursor int
}

// New returns an empty Model. Call SetItems before displaying.
func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

// SetItems replaces the project list and clamps the cursor if needed.
func (m Model) SetItems(items []repository.ProjectTotal) Model {
	m.items = items
	if len(items) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(items) {
		m.cursor = len(items) - 1
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "j", "down":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "s", "enter":
		if len(m.items) == 0 {
			return m, nil
		}
		id := m.items[m.cursor].ProjectID
		return m, func() tea.Msg { return StartSessionMsg{ProjectID: id} }
	case "n":
		return m, func() tea.Msg { return NewProjectMsg{} }
	case "a":
		if len(m.items) == 0 {
			return m, nil
		}
		id := m.items[m.cursor].ProjectID
		return m, func() tea.Msg { return ArchiveMsg{ProjectID: id} }
	case "l":
		if len(m.items) == 0 {
			return m, nil
		}
		id := m.items[m.cursor].ProjectID
		return m, func() tea.Msg { return ShowSessionLogMsg{ProjectID: id} }
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString(styles.Title.Render("hourglass") + "\n\n")

	if len(m.items) == 0 {
		sb.WriteString(styles.Muted.Render("  No projects yet.") + "\n")
	} else {
		for i, item := range m.items {
			total := styles.FormatDuration(item.Total)
			if i == m.cursor {
				padded := fmt.Sprintf("%-*s", nameColWidth, item.ProjectName)
				sb.WriteString("> " + styles.Selected.Render(padded) + "  " + styles.Accent.Render(total) + "\n")
			} else {
				fmt.Fprintf(&sb, "  %-*s  %s\n", nameColWidth, item.ProjectName, styles.Muted.Render(total))
			}
		}
	}

	sb.WriteString("\n" + styles.StatusBar.Render("  [s] start  [n] new  [a] archive  [l] log  [q] quit"))
	return sb.String()
}
