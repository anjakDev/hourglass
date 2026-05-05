package newproject_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anjakDev/hourglass/internal/tui/views/newproject"
)

// typeRunes feeds individual character key messages into a model.
func typeRunes(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestNewProject_EnterWithName(t *testing.T) {
	var m tea.Model = newproject.New()
	m = typeRunes(m, "My Project")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m
	require.NotNil(t, cmd)
	assert.Equal(t, newproject.CreatedMsg{Name: "My Project"}, cmd())
}

func TestNewProject_EnterTrimsWhitespace(t *testing.T) {
	var m tea.Model = newproject.New()
	m = typeRunes(m, "  Alpha  ")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m
	require.NotNil(t, cmd)
	assert.Equal(t, newproject.CreatedMsg{Name: "Alpha"}, cmd())
}

func TestNewProject_EnterEmptyName_NoCmd(t *testing.T) {
	var m tea.Model = newproject.New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
}

func TestNewProject_Escape(t *testing.T) {
	var m tea.Model = newproject.New()
	m = typeRunes(m, "draft")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	_ = m
	require.NotNil(t, cmd)
	assert.Equal(t, newproject.CancelMsg{}, cmd())
}

func TestNewProject_View_ContainsHeading(t *testing.T) {
	m := newproject.New()
	assert.Contains(t, m.View(), "New project")
}
