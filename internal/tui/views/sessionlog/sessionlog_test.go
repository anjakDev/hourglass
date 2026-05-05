package sessionlog_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui/views/sessionlog"
)

func closedSession(start, end time.Time, breakSec int64) repository.Session {
	e := end.UTC()
	return repository.Session{
		StartedAt:            start.UTC(),
		EndedAt:              &e,
		BreakDurationSeconds: breakSec,
	}
}

func openSession(start time.Time) repository.Session {
	return repository.Session{StartedAt: start.UTC()}
}

func TestSessionLog_View_ContainsProjectName(t *testing.T) {
	m := sessionlog.New("Alpha", nil)
	assert.Contains(t, m.View(), "Alpha")
}

func TestSessionLog_View_Empty(t *testing.T) {
	m := sessionlog.New("Alpha", nil)
	assert.Contains(t, m.View(), "No sessions")
}

func TestSessionLog_View_ShowsStartTime(t *testing.T) {
	start := time.Now().Add(-time.Hour).Truncate(time.Second)
	end := start.Add(time.Hour)
	m := sessionlog.New("Alpha", []repository.Session{closedSession(start, end, 0)})
	assert.Contains(t, m.View(), start.Local().Format("15:04"))
}

func TestSessionLog_View_ShowsEndTime(t *testing.T) {
	start := time.Now().Add(-time.Hour).Truncate(time.Second)
	end := start.Add(time.Hour)
	m := sessionlog.New("Alpha", []repository.Session{closedSession(start, end, 0)})
	assert.Contains(t, m.View(), end.Local().Format("15:04"))
}

func TestSessionLog_View_ShowsDuration(t *testing.T) {
	start := time.Now().Add(-time.Hour).Truncate(time.Second)
	end := start.Add(time.Hour)
	m := sessionlog.New("Alpha", []repository.Session{closedSession(start, end, 0)})
	assert.Contains(t, m.View(), "1h 0m")
}

func TestSessionLog_View_ShowsBreak(t *testing.T) {
	start := time.Now().Add(-time.Hour).Truncate(time.Second)
	end := start.Add(time.Hour)
	m := sessionlog.New("Alpha", []repository.Session{
		closedSession(start, end, int64((10*time.Minute).Seconds())),
	})
	view := m.View()
	assert.Contains(t, view, "10m") // break shown
	assert.Contains(t, view, "50m") // net work (60m - 10m)
}

func TestSessionLog_View_OpenSession_ShowsRunning(t *testing.T) {
	start := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
	m := sessionlog.New("Alpha", []repository.Session{openSession(start)})
	assert.Contains(t, m.View(), "running")
}

func TestSessionLog_View_ShowsTotal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	s1 := closedSession(now.Add(-2*time.Hour), now.Add(-time.Hour), 0) // 1h
	s2 := closedSession(now.Add(-30*time.Minute), now, 0)               // 30m
	m := sessionlog.New("Alpha", []repository.Session{s1, s2})
	assert.Contains(t, m.View(), "1h 30m")
}

func TestSessionLog_Esc_ReturnsBackMsg(t *testing.T) {
	var m tea.Model = sessionlog.New("Alpha", nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	require.NotNil(t, cmd)
	assert.Equal(t, sessionlog.BackMsg{}, cmd())
}

func TestSessionLog_Q_ReturnsBackMsg(t *testing.T) {
	var m tea.Model = sessionlog.New("Alpha", nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd)
	assert.Equal(t, sessionlog.BackMsg{}, cmd())
}
