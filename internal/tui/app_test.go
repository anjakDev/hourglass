package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anjakDev/hourglass/internal/db"
	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui"
)

// step executes cmd and feeds the resulting message back into m,
// simulating one iteration of the Bubbletea runtime loop.
func step(m tea.Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if cmd == nil {
		return m, nil
	}
	return m.Update(cmd())
}

func pressRune(m tea.Model, r rune) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

func pressSpecial(m tea.Model, kt tea.KeyType) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: kt})
}

func newTestApp(t *testing.T) tui.App {
	t.Helper()
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return tui.New(
		repository.NewProjectRepo(conn),
		repository.NewSessionRepo(conn),
	)
}

// appWithProject returns an app that has one project loaded in the list,
// plus the session repo for direct DB assertions.
func appWithProject(t *testing.T, name string) (tea.Model, *repository.SessionRepo) {
	t.Helper()
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	pr := repository.NewProjectRepo(conn)
	sr := repository.NewSessionRepo(conn)
	_, err = pr.Create(name, "", "")
	require.NoError(t, err)

	app := tui.New(pr, sr)
	var m tea.Model = app
	m, _ = step(m, app.Init())
	require.Contains(t, m.View(), name)
	return m, sr
}

// appAtTimerView returns an app already on the active-timer screen (session
// started), plus the session repo for DB assertions.
func appAtTimerView(t *testing.T) (tea.Model, *repository.SessionRepo) {
	t.Helper()
	m, sr := appWithProject(t, "Work")

	// 's' → StartSessionMsg → sessionStartedMsg → viewActiveTimer
	var cmd tea.Cmd
	m, cmd = pressRune(m, 's')
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)

	require.Contains(t, m.View(), "Work")
	return m, sr
}

// ── existing tests ────────────────────────────────────────────────────────────

func TestApp_StartsAtProjectsView(t *testing.T) {
	app := newTestApp(t)
	assert.Contains(t, app.View(), "hourglass")
}

func TestApp_InitLoadsProjects(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	pr := repository.NewProjectRepo(conn)
	sr := repository.NewSessionRepo(conn)
	pid, err := pr.Create("Work", "", "")
	require.NoError(t, err)
	start := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	sid, err := sr.StartSession(pid, start)
	require.NoError(t, err)
	require.NoError(t, sr.StopSession(sid, start.Add(time.Hour), 0))

	app := tui.New(pr, sr)
	var m tea.Model = app
	m, _ = step(m, app.Init())
	assert.Contains(t, m.View(), "Work")
	assert.Contains(t, m.View(), "1h 0m")
}

func TestApp_PressN_SwitchesToNewProject(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m, cmd := pressRune(m, 'n')
	m, _ = step(m, cmd)
	assert.Contains(t, m.View(), "New project")
}

func TestApp_EscapeFromNewProject_ReturnsToProjects(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m, cmd := pressRune(m, 'n')
	m, _ = step(m, cmd)

	m, cmd = pressSpecial(m, tea.KeyEscape)
	m, _ = step(m, cmd)
	assert.Contains(t, m.View(), "hourglass")
}

func TestApp_CreateProject_AppearsInList(t *testing.T) {
	var m tea.Model = newTestApp(t)

	m, cmd := pressRune(m, 'n')
	m, _ = step(m, cmd)

	for _, r := range "Design" {
		m, _ = pressRune(m, r)
	}

	// Enter → CreatedMsg → projectCreatedMsg → projectsLoadedMsg
	m, cmd = pressSpecial(m, tea.KeyEnter)
	m, cmd = step(m, cmd)
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)

	assert.Contains(t, m.View(), "hourglass")
	assert.Contains(t, m.View(), "Design")
}

func TestApp_ArchiveProject_RemovesFromList(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	pr := repository.NewProjectRepo(conn)
	sr := repository.NewSessionRepo(conn)
	_, err = pr.Create("ToArchive", "", "")
	require.NoError(t, err)

	app := tui.New(pr, sr)
	var m tea.Model = app
	m, _ = step(m, app.Init())
	require.Contains(t, m.View(), "ToArchive")

	// 'a' → ArchiveMsg → projectArchivedMsg → projectsLoadedMsg
	m, cmd := pressRune(m, 'a')
	m, cmd = step(m, cmd)
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)

	assert.NotContains(t, m.View(), "ToArchive")
}

// ── session flow tests ────────────────────────────────────────────────────────

func TestApp_StartSession_SwitchesToTimerView(t *testing.T) {
	m, _ := appAtTimerView(t)
	// view should show the active-timer screen
	assert.Contains(t, m.View(), "Work")
	assert.Contains(t, m.View(), "00:00:") // clock at ~zero
	assert.Contains(t, m.View(), "[s]")    // stop hint present
}

func TestApp_StopSession_ReturnsToProjects(t *testing.T) {
	m, _ := appAtTimerView(t)

	// 's' → StopSessionMsg → sessionStoppedMsg → projectsLoadedMsg
	var cmd tea.Cmd
	m, cmd = pressRune(m, 's')
	m, cmd = step(m, cmd)
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)

	assert.Contains(t, m.View(), "hourglass")
}

func TestApp_StopSession_WritesToDB(t *testing.T) {
	m, sr := appAtTimerView(t)

	var cmd tea.Cmd
	m, cmd = pressRune(m, 's')
	m, cmd = step(m, cmd)
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)
	_ = m

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.NotNil(t, sessions[0].EndedAt, "session should be closed in DB")
}

func TestApp_ShowSessionLog_SwitchesToLogView(t *testing.T) {
	m, _ := appWithProject(t, "Alpha")

	// 'l' → ShowSessionLogMsg → sessionLogLoadedMsg → viewSessionLog
	var cmd tea.Cmd
	m, cmd = pressRune(m, 'l')
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)

	assert.Contains(t, m.View(), "Sessions")
	assert.Contains(t, m.View(), "Alpha")
}

func TestApp_SessionLog_Back_ReturnsToProjects(t *testing.T) {
	m, _ := appWithProject(t, "Alpha")

	// navigate to session log
	var cmd tea.Cmd
	m, cmd = pressRune(m, 'l')
	m, cmd = step(m, cmd)
	m, _ = step(m, cmd)
	require.Contains(t, m.View(), "Sessions")

	// esc → BackMsg → viewProjects
	m, cmd = pressSpecial(m, tea.KeyEscape)
	m, _ = step(m, cmd)

	assert.Contains(t, m.View(), "hourglass")
}
