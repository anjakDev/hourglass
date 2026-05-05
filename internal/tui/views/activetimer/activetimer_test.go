package activetimer_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmr "github.com/anjakDev/hourglass/internal/timer"
	"github.com/anjakDev/hourglass/internal/tui/views/activetimer"
)

// newRunningModel creates a model with an already-started timer.
func newRunningModel(projectName string, start time.Time) (activetimer.Model, *tmr.Timer) {
	t := tmr.New()
	_ = t.Start(start)
	return activetimer.New(projectName, t, start), t
}

func TestActiveTimer_View_ContainsProjectName(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	m, _ := newRunningModel("Alpha", start)
	assert.Contains(t, m.View(), "Alpha")
}

func TestActiveTimer_View_ContainsFormattedDuration(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	m, _ := newRunningModel("Alpha", start)
	// advance 90 seconds via a tick
	tick := start.Add(90 * time.Second)
	m2, _ := m.Update(activetimer.TickMsg(tick))
	// FormatClock(90s) → "00:01:30"
	assert.Contains(t, m2.(activetimer.Model).View(), "00:01:30")
}

func TestActiveTimer_Tick_ReturnsTick(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	m, _ := newRunningModel("Alpha", start)
	_, cmd := m.Update(activetimer.TickMsg(start.Add(time.Second)))
	assert.NotNil(t, cmd, "tick should schedule the next tick")
}

func TestActiveTimer_Break_TogglesState(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	var m tea.Model
	var timer *tmr.Timer
	m, timer = newRunningModel("Alpha", start)

	// press b → running→on_break
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, tmr.StateOnBreak, timer.State())

	// press b again → on_break→running
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	_ = m
	assert.Equal(t, tmr.StateRunning, timer.State())
}

func TestActiveTimer_Break_NotRunning_NoOp(t *testing.T) {
	// Timer is idle (never started) — 'b' should be a no-op.
	t2 := tmr.New()
	m := activetimer.New("Alpha", t2, time.Now())
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, tmr.StateIdle, t2.State())
}

func TestActiveTimer_Stop_ReturnsStopMsg(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	m, _ := newRunningModel("Alpha", start)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	require.NotNil(t, cmd)
	assert.Equal(t, activetimer.StopSessionMsg{}, cmd())
}

func TestActiveTimer_View_ShowsBreakIndicator(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	var m tea.Model
	m, _ = newRunningModel("Alpha", start)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Contains(t, m.(activetimer.Model).View(), "break")
}

func TestActiveTimer_View_KeyHints_Running(t *testing.T) {
	start := time.Now().Truncate(time.Second)
	m, _ := newRunningModel("Alpha", start)
	view := m.View()
	assert.Contains(t, view, "[b]")
	assert.Contains(t, view, "[s]")
}
