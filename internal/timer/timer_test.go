package timer_test

import (
	"testing"
	"time"

	"github.com/anjakDev/hourglass/internal/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// base is a fixed reference time used across all tests.
var base = time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

func TestTimer_InitialState(t *testing.T) {
	tm := timer.New()
	assert.Equal(t, timer.StateIdle, tm.State())
	assert.Equal(t, time.Duration(0), tm.Elapsed(base))
	assert.Equal(t, time.Duration(0), tm.WorkDuration(base))
}

func TestTimer_Start(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	assert.Equal(t, timer.StateRunning, tm.State())
}

func TestTimer_Start_AlreadyRunning(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	err := tm.Start(base.Add(time.Minute))
	assert.ErrorIs(t, err, timer.ErrAlreadyRunning)
}

func TestTimer_Stop_WhenIdle(t *testing.T) {
	tm := timer.New()
	_, err := tm.Stop(base)
	assert.ErrorIs(t, err, timer.ErrNotRunning)
}

func TestTimer_Elapsed(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	now := base.Add(30 * time.Minute)
	assert.Equal(t, 30*time.Minute, tm.Elapsed(now))
}

func TestTimer_WorkDuration_NoBreaks(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	now := base.Add(45 * time.Minute)
	assert.Equal(t, 45*time.Minute, tm.WorkDuration(now))
}

func TestTimer_Break(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(20*time.Minute)))
	assert.Equal(t, timer.StateOnBreak, tm.State())
}

func TestTimer_Break_WhenIdle(t *testing.T) {
	tm := timer.New()
	err := tm.Break(base)
	assert.ErrorIs(t, err, timer.ErrNotRunning)
}

func TestTimer_Break_WhenAlreadyOnBreak(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(10*time.Minute)))
	err := tm.Break(base.Add(15 * time.Minute))
	assert.ErrorIs(t, err, timer.ErrAlreadyOnBreak)
}

func TestTimer_Resume(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(20*time.Minute)))
	require.NoError(t, tm.Resume(base.Add(30*time.Minute)))
	assert.Equal(t, timer.StateRunning, tm.State())
}

func TestTimer_Resume_WhenRunning(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	err := tm.Resume(base.Add(5 * time.Minute))
	assert.ErrorIs(t, err, timer.ErrNotOnBreak)
}

func TestTimer_WorkDuration_WithBreak(t *testing.T) {
	// start at T+0, break at T+20, resume at T+30, now = T+50
	// work = 50 - 10(break) = 40 min
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(20*time.Minute)))
	require.NoError(t, tm.Resume(base.Add(30*time.Minute)))

	now := base.Add(50 * time.Minute)
	assert.Equal(t, 40*time.Minute, tm.WorkDuration(now))
	assert.Equal(t, 10*time.Minute, tm.BreakDuration(now))
}

func TestTimer_WorkDuration_WhileOnBreak(t *testing.T) {
	// start at T+0, break at T+20, check at T+35 (still on break)
	// work so far = 20min, break so far = 15min
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(20*time.Minute)))

	now := base.Add(35 * time.Minute)
	assert.Equal(t, 20*time.Minute, tm.WorkDuration(now))
	assert.Equal(t, 15*time.Minute, tm.BreakDuration(now))
}

func TestTimer_Stop_ReturnsSession(t *testing.T) {
	// 60min total, 15min break → 45min work
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(20*time.Minute)))
	require.NoError(t, tm.Resume(base.Add(35*time.Minute)))

	sess, err := tm.Stop(base.Add(60 * time.Minute))
	require.NoError(t, err)
	assert.Equal(t, timer.StateIdle, tm.State())
	assert.Equal(t, base, sess.StartedAt)
	assert.Equal(t, base.Add(60*time.Minute), sess.EndedAt)
	assert.Equal(t, 15*time.Minute, sess.TotalBreak)
	assert.Equal(t, 45*time.Minute, sess.WorkDuration())
}

func TestTimer_Stop_WhileOnBreak(t *testing.T) {
	// stop mid-break: break counts up to stop time
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(30*time.Minute)))

	sess, err := tm.Stop(base.Add(40 * time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 10*time.Minute, sess.TotalBreak)
	assert.Equal(t, 30*time.Minute, sess.WorkDuration())
}

func TestTimer_ReusableAfterStop(t *testing.T) {
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	_, err := tm.Stop(base.Add(time.Hour))
	require.NoError(t, err)

	// should be able to start a new session
	require.NoError(t, tm.Start(base.Add(2*time.Hour)))
	assert.Equal(t, timer.StateRunning, tm.State())
}

func TestTimer_MultipleBreaks(t *testing.T) {
	// Two breaks of 5min each → 10min total break in a 60min session → 50min work
	tm := timer.New()
	require.NoError(t, tm.Start(base))
	require.NoError(t, tm.Break(base.Add(10*time.Minute)))
	require.NoError(t, tm.Resume(base.Add(15*time.Minute)))
	require.NoError(t, tm.Break(base.Add(40*time.Minute)))
	require.NoError(t, tm.Resume(base.Add(45*time.Minute)))

	sess, err := tm.Stop(base.Add(60 * time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 10*time.Minute, sess.TotalBreak)
	assert.Equal(t, 50*time.Minute, sess.WorkDuration())
}
