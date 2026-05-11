package editsession_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anjakDev/hourglass/internal/tui/views/editsession"
)

const timeFormat = "2006-01-02 15:04"

// fixedTimes returns a same-day UTC start/end used by most tests.
func fixedTimes() (start, end time.Time) {
	start = time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	end = time.Date(2026, 5, 7, 14, 0, 0, 0, time.UTC)
	return
}

func key(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// typeString clears the current input then types each character of s.
// Space is sent as tea.KeySpace to match real terminal behaviour.
func typeString(m tea.Model, s string) tea.Model {
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	for _, r := range s {
		if r == ' ' {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
		} else {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}
	return m
}

// localFmt formats t the same way the view does.
func localFmt(t time.Time) string {
	return t.In(time.Local).Format(timeFormat)
}

func TestEditSession_View_ShowsProjectName(t *testing.T) {
	start, end := fixedTimes()
	m := editsession.New("Alpha", start, end)
	assert.Contains(t, m.View(), "Alpha")
}

func TestEditSession_View_ShowsStartTime(t *testing.T) {
	start, end := fixedTimes()
	m := editsession.New("Alpha", start, end)
	assert.Contains(t, m.View(), localFmt(start))
}

func TestEditSession_View_ShowsEndTime(t *testing.T) {
	start, end := fixedTimes()
	m := editsession.New("Alpha", start, end)
	assert.Contains(t, m.View(), localFmt(end))
}

func TestEditSession_View_ShowsDuration(t *testing.T) {
	start := time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)
	m := editsession.New("Alpha", start, end)
	assert.Contains(t, m.View(), "1h 30m")
}

// TestEditSession_View_ShowsOvernightScenario verifies the view correctly
// shows times that span midnight (the primary use case for this feature).
func TestEditSession_View_ShowsOvernightScenario(t *testing.T) {
	start := time.Date(2026, 5, 7, 22, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 8, 8, 0, 0, 0, time.UTC)
	m := editsession.New("Alpha", start, end)
	view := m.View()
	assert.Contains(t, view, localFmt(start))
	assert.Contains(t, view, localFmt(end))
	assert.Contains(t, view, "10h")
}

func TestEditSession_Save_EmitsMsg(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	sm, ok := cmd().(editsession.SaveMsg)
	require.True(t, ok)
	assert.True(t, start.Equal(sm.StartedAt))
	assert.True(t, end.Equal(sm.EndedAt))
}

func TestEditSession_Cancel_EmitsOriginalTimes(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	sm, ok := cmd().(editsession.SaveMsg)
	require.True(t, ok)
	assert.True(t, start.Equal(sm.StartedAt))
	assert.True(t, end.Equal(sm.EndedAt))
}

func TestEditSession_EditMode_EntersOnE(t *testing.T) {
	start, end := fixedTimes()
	m, _ := editsession.New("Alpha", start, end).Update(key('e'))
	assert.True(t, m.(editsession.Model).IsEditing())
}

func TestEditSession_EditMode_InputValueIsCurrentStart(t *testing.T) {
	start, end := fixedTimes()
	m, _ := editsession.New("Alpha", start, end).Update(key('e'))
	assert.Equal(t, localFmt(start), m.(editsession.Model).InputValue())
}

func TestEditSession_EditMode_EscDiscards(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	em := m.(editsession.Model)
	assert.False(t, em.IsEditing())
	assert.True(t, start.Equal(em.StartedAt()))
}

func TestEditSession_EditStart_UpdatesTime(t *testing.T) {
	start, end := fixedTimes()
	// newStart is 30 min before end, definitely valid in any timezone
	newStart := end.Add(-30 * time.Minute)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(newStart))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	em := m.(editsession.Model)
	assert.False(t, em.IsEditing())
	assert.True(t, newStart.Equal(em.StartedAt()), "want %v got %v", newStart, em.StartedAt())
	assert.Equal(t, "", em.InputErr())
}

func TestEditSession_EditEnd_UpdatesTime(t *testing.T) {
	start, end := fixedTimes()
	// newEnd is 30 min after start, definitely valid in any timezone
	newEnd := start.Add(30 * time.Minute)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(newEnd))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	em := m.(editsession.Model)
	assert.False(t, em.IsEditing())
	assert.True(t, newEnd.Equal(em.EndedAt()), "want %v got %v", newEnd, em.EndedAt())
	assert.Equal(t, "", em.InputErr())
}

func TestEditSession_Edit_InvalidFormat_ShowsError(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, "not a time")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	em := m.(editsession.Model)
	assert.True(t, em.IsEditing())
	assert.NotEmpty(t, em.InputErr())
	assert.True(t, start.Equal(em.StartedAt()))
}

func TestEditSession_Edit_StartAfterEnd_ShowsError(t *testing.T) {
	start, end := fixedTimes()
	// badStart is after end, definitely invalid
	badStart := end.Add(time.Hour)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(badStart))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	em := m.(editsession.Model)
	assert.True(t, em.IsEditing())
	assert.NotEmpty(t, em.InputErr())
	assert.True(t, start.Equal(em.StartedAt()))
}

func TestEditSession_Edit_EndBeforeStart_ShowsError(t *testing.T) {
	start, end := fixedTimes()
	// badEnd is before start, definitely invalid
	badEnd := start.Add(-time.Hour)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(badEnd))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	em := m.(editsession.Model)
	assert.True(t, em.IsEditing())
	assert.NotEmpty(t, em.InputErr())
	assert.True(t, end.Equal(em.EndedAt()))
}

func TestEditSession_CursorMoves_Tab(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(key('e'))
	assert.Equal(t, localFmt(end), m.(editsession.Model).InputValue())
}

func TestEditSession_CursorMoves_J(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('j'))
	m, _ = m.Update(key('e'))
	assert.Equal(t, localFmt(end), m.(editsession.Model).InputValue())
}

func TestEditSession_CursorMoves_K(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('j')) // start→end
	m, _ = m.Update(key('k')) // end→start
	m, _ = m.Update(key('e'))
	assert.Equal(t, localFmt(start), m.(editsession.Model).InputValue())
}

func TestEditSession_EditStart_SaveEmitsUpdatedTime(t *testing.T) {
	start, end := fixedTimes()
	newStart := end.Add(-30 * time.Minute)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(newStart))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm edit

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // save
	require.NotNil(t, cmd)
	sm, ok := cmd().(editsession.SaveMsg)
	require.True(t, ok)
	assert.True(t, newStart.Equal(sm.StartedAt))
	assert.True(t, end.Equal(sm.EndedAt))
}

func TestEditSession_Cancel_AfterEdit_EmitsOriginalTimes(t *testing.T) {
	start, end := fixedTimes()
	newStart := end.Add(-30 * time.Minute)

	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, localFmt(newStart))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm edit — startedAt is now newStart

	// Esc in summary mode reverts to original
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	sm, ok := cmd().(editsession.SaveMsg)
	require.True(t, ok)
	assert.True(t, start.Equal(sm.StartedAt), "want original %v got %v", start, sm.StartedAt)
	assert.True(t, end.Equal(sm.EndedAt))
}

func TestEditSession_View_ShowsEditingHint(t *testing.T) {
	start, end := fixedTimes()
	m, _ := editsession.New("Alpha", start, end).Update(key('e'))
	assert.Contains(t, m.(editsession.Model).View(), "confirm")
}

func TestEditSession_EditMode_SpaceCharacterTyped(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Equal(t, " ", m.(editsession.Model).InputValue())
}

func TestEditSession_View_ShowsError(t *testing.T) {
	start, end := fixedTimes()
	var m tea.Model = editsession.New("Alpha", start, end)
	m, _ = m.Update(key('e'))
	m = typeString(m, "bad input")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, m.(editsession.Model).View(), "invalid")
}
