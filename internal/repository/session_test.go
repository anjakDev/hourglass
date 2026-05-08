package repository_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/anjakDev/hourglass/internal/db"
	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSessionTestDB returns both repos sharing one in-memory DB.
func newSessionTestDB(t *testing.T) (*sql.DB, *repository.ProjectRepo, *repository.SessionRepo) {
	t.Helper()
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn, repository.NewProjectRepo(conn), repository.NewSessionRepo(conn)
}

// seedProject creates a project and returns its ID.
func seedProject(t *testing.T, pr *repository.ProjectRepo, name string) int64 {
	t.Helper()
	id, err := pr.Create(name, "", "")
	require.NoError(t, err)
	return id
}

func TestSessionRepo_StartStop(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	start := time.Now().UTC().Truncate(time.Second)
	id, err := sr.StartSession(pid, start)
	require.NoError(t, err)
	assert.Positive(t, id)

	end := start.Add(30 * time.Minute)
	err = sr.StopSession(id, end, 0)
	require.NoError(t, err)
}

func TestSessionRepo_ActiveSession_None(t *testing.T) {
	_, _, sr := newSessionTestDB(t)
	_, err := sr.ActiveSession()
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSessionRepo_ActiveSession_Found(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	start := time.Now().UTC().Truncate(time.Second)
	id, err := sr.StartSession(pid, start)
	require.NoError(t, err)

	sess, err := sr.ActiveSession()
	require.NoError(t, err)
	assert.Equal(t, id, sess.ID)
	assert.Equal(t, pid, sess.ProjectID)
	assert.Nil(t, sess.EndedAt)
}

func TestSessionRepo_StopSession_AlreadyStopped(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	start := time.Now().UTC().Truncate(time.Second)
	id, _ := sr.StartSession(pid, start)
	require.NoError(t, sr.StopSession(id, start.Add(time.Hour), 0))

	// stopping again should fail
	err := sr.StopSession(id, start.Add(2*time.Hour), 0)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSessionRepo_ListToday(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	now := time.Now().UTC().Truncate(time.Second)

	// two sessions today
	id1, _ := sr.StartSession(pid, now)
	sr.StopSession(id1, now.Add(30*time.Minute), 0)
	id2, _ := sr.StartSession(pid, now.Add(time.Hour))
	sr.StopSession(id2, now.Add(90*time.Minute), 0)

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSessionRepo_ListToday_IncludesOpenSession(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	now := time.Now().UTC().Truncate(time.Second)
	sr.StartSession(pid, now) // open, not stopped

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Nil(t, sessions[0].EndedAt)
}

func TestSessionRepo_WorkDuration(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	start := time.Now().UTC().Truncate(time.Second)
	id, _ := sr.StartSession(pid, start)
	end := start.Add(60 * time.Minute)
	require.NoError(t, sr.StopSession(id, end, int64((10*time.Minute).Seconds())))

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, 50*time.Minute, sessions[0].WorkDuration())
}

func TestSessionRepo_TodayTotalsByProject(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid1 := seedProject(t, pr, "Alpha")
	pid2 := seedProject(t, pr, "Beta")

	now := time.Now().UTC().Truncate(time.Second)

	// Alpha: 30min work
	id1, _ := sr.StartSession(pid1, now)
	sr.StopSession(id1, now.Add(30*time.Minute), 0)

	// Beta: 45min work (60min - 15min break)
	id2, _ := sr.StartSession(pid2, now.Add(time.Hour))
	sr.StopSession(id2, now.Add(2*time.Hour), int64((15 * time.Minute).Seconds()))

	totals, err := sr.TodayTotalsByProject()
	require.NoError(t, err)
	require.Len(t, totals, 2)

	assert.Equal(t, "Alpha", totals[0].ProjectName)
	assert.Equal(t, 30*time.Minute, totals[0].Total)
	assert.Equal(t, "Beta", totals[1].ProjectName)
	assert.Equal(t, 45*time.Minute, totals[1].Total)
}

func TestSessionRepo_TodayTotalsByProject_Empty(t *testing.T) {
	_, _, sr := newSessionTestDB(t)
	totals, err := sr.TodayTotalsByProject()
	require.NoError(t, err)
	assert.Empty(t, totals)
}

func TestSessionRepo_TodayTotalsByProject_IncludesOpenSession(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	// Start a minute ago so the query sees positive elapsed time even at sub-second resolution.
	now := time.Now().UTC().Truncate(time.Second)
	sr.StartSession(pid, now.Add(-time.Minute)) // still running — totals should count it up to now

	totals, err := sr.TodayTotalsByProject()
	require.NoError(t, err)
	require.Len(t, totals, 1)
	assert.Positive(t, totals[0].Total)
}

func TestSessionRepo_DeleteAllSessions(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	now := time.Now().UTC().Truncate(time.Second)
	id1, _ := sr.StartSession(pid, now)
	sr.StopSession(id1, now.Add(30*time.Minute), 0)
	id2, _ := sr.StartSession(pid, now.Add(time.Hour))
	sr.StopSession(id2, now.Add(90*time.Minute), 0)

	err := sr.DeleteAllSessions(pid)
	require.NoError(t, err)

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_DeleteAllSessions_NoSessions(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Empty")

	err := sr.DeleteAllSessions(pid)
	assert.NoError(t, err)
}

func TestSessionRepo_DeleteAllSessions_OnlyAffectsTargetProject(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid1 := seedProject(t, pr, "Work")
	pid2 := seedProject(t, pr, "Side")

	now := time.Now().UTC().Truncate(time.Second)
	id1, _ := sr.StartSession(pid1, now)
	sr.StopSession(id1, now.Add(30*time.Minute), 0)
	id2, _ := sr.StartSession(pid2, now.Add(time.Hour))
	sr.StopSession(id2, now.Add(90*time.Minute), 0)

	require.NoError(t, sr.DeleteAllSessions(pid1))

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, pid2, sessions[0].ProjectID)
}

func TestSessionRepo_UpdateSession_UpdatesBothTimes(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	base := time.Now().UTC().Truncate(time.Second)
	id, err := sr.StartSession(pid, base)
	require.NoError(t, err)

	newStart := base.Add(-30 * time.Minute)
	newEnd := base.Add(60 * time.Minute)
	require.NoError(t, sr.UpdateSession(id, newStart, newEnd, 0))

	sessions, err := sr.ListToday()
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.True(t, newStart.Equal(sessions[0].StartedAt), "started_at mismatch")
	require.NotNil(t, sessions[0].EndedAt)
	assert.True(t, newEnd.Equal(*sessions[0].EndedAt), "ended_at mismatch")
}

func TestSessionRepo_UpdateSession_NotFound(t *testing.T) {
	_, _, sr := newSessionTestDB(t)
	err := sr.UpdateSession(9999, time.Now(), time.Now().Add(time.Hour), 0)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSessionRepo_DeleteAllSessions_IncludesOpenSession(t *testing.T) {
	_, pr, sr := newSessionTestDB(t)
	pid := seedProject(t, pr, "Work")

	now := time.Now().UTC().Truncate(time.Second)
	sr.StartSession(pid, now) // open, not stopped

	require.NoError(t, sr.DeleteAllSessions(pid))

	_, err := sr.ActiveSession()
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
