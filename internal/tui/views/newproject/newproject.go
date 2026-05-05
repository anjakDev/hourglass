package newproject

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/tui/styles"
)

// CreatedMsg is returned when the user confirms a valid project name.
type CreatedMsg struct{ Name string }

// CancelMsg is returned when the user presses Escape.
type CancelMsg struct{}

// Model is the new-project form view.
type Model struct {
	input textinput.Model
}

// New returns a fresh, focused Model ready for input.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "project name"
	ti.CharLimit = 80
	ti.Width = 36
	_ = ti.Focus() // mutates the local value; Init() issues the blink cmd
	return Model{input: ti}
}

// Init starts the cursor blink animation (input is already focused by New).
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles key input. Enter confirms, Escape cancels; all other keys
// are forwarded to the text input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyEnter:
			name := strings.TrimSpace(m.input.Value())
			if name == "" {
				return m, nil
			}
			return m, func() tea.Msg { return CreatedMsg{Name: name} }
		case tea.KeyEscape:
			return m, func() tea.Msg { return CancelMsg{} }
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the form.
func (m Model) View() string {
	return styles.Title.Render("New project") + "\n\n" +
		"  " + m.input.View() + "\n\n" +
		styles.StatusBar.Render("  enter  create   esc  cancel")
}
