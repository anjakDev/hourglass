package projects_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui/views/projects"
)

func threeItems() []repository.ProjectTotal {
	return []repository.ProjectTotal{
		{ProjectID: 1, ProjectName: "Alpha", Total: 30 * time.Minute},
		{ProjectID: 2, ProjectName: "Beta", Total: 45 * time.Minute},
		{ProjectID: 3, ProjectName: "Gamma", Total: 0},
	}
}

func key(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// TestProjects_CursorMoves verifies j/k navigate the list by checking which
// project is selected when s is pressed.
func TestProjects_CursorMoves(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())

	// j moves down — cursor on Beta
	m, _ = m.Update(key('j'))
	_, cmd := m.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 2}, cmd())

	// k moves back up — cursor on Alpha
	m, _ = m.Update(key('k'))
	_, cmd = m.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 1}, cmd())
}

func TestProjects_CursorClampsAtBottom(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	// three j's on a 3-item list — last one should clamp
	m, _ = m.Update(key('j'))
	m, _ = m.Update(key('j'))
	m, _ = m.Update(key('j')) // already at bottom, no-op

	_, cmd := m.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 3}, cmd())
}

func TestProjects_CursorClampsAtTop(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	m, _ = m.Update(key('k')) // already at top, no-op

	_, cmd := m.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 1}, cmd())
}

func TestProjects_StartSession(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	_, cmd := m.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 1}, cmd())
}

func TestProjects_StartSession_Enter(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 1}, cmd())
}

func TestProjects_StartSession_EmptyList(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('s'))
	assert.Nil(t, cmd)
}

func TestProjects_NewProject(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('n'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.NewProjectMsg{}, cmd())
}

func TestProjects_Archive(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	_, cmd := m.Update(key('a'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.ArchiveMsg{ProjectID: 1}, cmd())
}

func TestProjects_Archive_EmptyList(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('a'))
	assert.Nil(t, cmd)
}

func TestProjects_ShowSessionLog(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	_, cmd := m.Update(key('l'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.ShowSessionLogMsg{ProjectID: 1}, cmd())
}

func TestProjects_ShowSessionLog_EmptyList(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('l'))
	assert.Nil(t, cmd)
}

func TestProjects_Quit(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('q'))
	require.NotNil(t, cmd)
	assert.Equal(t, tea.QuitMsg{}, cmd())
}

func TestProjects_SetItems_ClampsWhenShrinks(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	// move to last item
	m, _ = m.Update(key('j'))
	m, _ = m.Update(key('j'))

	// shrink to 1 item — cursor must clamp
	updated := m.(projects.Model).SetItems(threeItems()[:1])
	_, cmd := updated.Update(key('s'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.StartSessionMsg{ProjectID: 1}, cmd())
}

func TestProjects_View_ContainsProjectNames(t *testing.T) {
	m := projects.New().SetItems(threeItems())
	view := m.View()
	assert.Contains(t, view, "Alpha")
	assert.Contains(t, view, "Beta")
	assert.Contains(t, view, "Gamma")
}

func TestProjects_View_ContainsDurations(t *testing.T) {
	m := projects.New().SetItems(threeItems())
	view := m.View()
	assert.Contains(t, view, "30m")
	assert.Contains(t, view, "45m")
	assert.Contains(t, view, "0m")
}

func TestProjects_View_EmptyState(t *testing.T) {
	m := projects.New()
	assert.Contains(t, m.View(), "No projects")
}

func TestProjects_Wipe_EmptyList(t *testing.T) {
	var m tea.Model = projects.New()
	_, cmd := m.Update(key('w'))
	assert.Nil(t, cmd)
}

func TestProjects_Wipe_FirstPressArmsConfirmation(t *testing.T) {
	m := projects.New().SetItems(threeItems())
	updated, cmd := m.Update(key('w'))
	assert.Nil(t, cmd)
	assert.Contains(t, updated.(projects.Model).View(), "confirm")
}

func TestProjects_Wipe_SecondPressEmitsMsg(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	m, _ = m.Update(key('w')) // arm
	_, cmd := m.Update(key('w')) // confirm
	require.NotNil(t, cmd)
	assert.Equal(t, projects.WipeSessionsMsg{ProjectID: 1}, cmd())
}

func TestProjects_Wipe_EscCancels(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	m, _ = m.Update(key('w')) // arm
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Nil(t, cmd)
	assert.NotContains(t, updated.(projects.Model).View(), "confirm")
}

func TestProjects_Wipe_NavigationCancels(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	m, _ = m.Update(key('w')) // arm
	updated, _ := m.Update(key('j'))
	assert.NotContains(t, updated.(projects.Model).View(), "confirm")
}

func TestProjects_Wipe_TargetsCurrentCursor(t *testing.T) {
	var m tea.Model = projects.New().SetItems(threeItems())
	m, _ = m.Update(key('j')) // move to Beta (ID 2)
	m, _ = m.Update(key('w'))
	_, cmd := m.Update(key('w'))
	require.NotNil(t, cmd)
	assert.Equal(t, projects.WipeSessionsMsg{ProjectID: 2}, cmd())
}
