package timer

import (
	"errors"
	"time"
)

// State represents the current state of a timer.
type State int

const (
	StateIdle    State = iota // no active session
	StateRunning              // session is running
	StateOnBreak              // session paused for a break
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateOnBreak:
		return "on_break"
	default:
		return "unknown"
	}
}

var (
	ErrAlreadyRunning = errors.New("timer is already running")
	ErrNotRunning     = errors.New("timer is not running")
	ErrAlreadyOnBreak = errors.New("timer is already on break")
	ErrNotOnBreak     = errors.New("timer is not on break")
)

// Timer holds the in-memory state of the active tracking session.
// All times are stored in UTC.
type Timer struct {
	state      State
	startedAt  time.Time
	breakStart time.Time
	totalBreak time.Duration
}

// New returns a new idle Timer.
func New() *Timer {
	return &Timer{state: StateIdle}
}

// State returns the current state.
func (t *Timer) State() State { return t.state }

// Start transitions the timer from Idle → Running.
// Returns ErrAlreadyRunning if the timer is not idle.
func (t *Timer) Start(now time.Time) error {
	if t.state != StateIdle {
		return ErrAlreadyRunning
	}
	t.startedAt = now.UTC()
	t.totalBreak = 0
	t.state = StateRunning
	return nil
}

// Stop transitions the timer from Running or OnBreak → Idle and returns the
// completed Session summary.
func (t *Timer) Stop(now time.Time) (Session, error) {
	if t.state == StateIdle {
		return Session{}, ErrNotRunning
	}
	// If we stop while on a break, close the break first.
	if t.state == StateOnBreak {
		t.totalBreak += now.UTC().Sub(t.breakStart)
	}

	sess := Session{
		StartedAt:  t.startedAt,
		EndedAt:    now.UTC(),
		TotalBreak: t.totalBreak,
	}
	t.state = StateIdle
	t.startedAt = time.Time{}
	t.breakStart = time.Time{}
	t.totalBreak = 0
	return sess, nil
}

// Break transitions the timer from Running → OnBreak.
func (t *Timer) Break(now time.Time) error {
	if t.state != StateRunning {
		if t.state == StateOnBreak {
			return ErrAlreadyOnBreak
		}
		return ErrNotRunning
	}
	t.breakStart = now.UTC()
	t.state = StateOnBreak
	return nil
}

// Resume transitions the timer from OnBreak → Running.
func (t *Timer) Resume(now time.Time) error {
	if t.state != StateOnBreak {
		return ErrNotOnBreak
	}
	t.totalBreak += now.UTC().Sub(t.breakStart)
	t.breakStart = time.Time{}
	t.state = StateRunning
	return nil
}

// Elapsed returns the total wall-clock duration since the session started,
// including any break time.
func (t *Timer) Elapsed(now time.Time) time.Duration {
	if t.state == StateIdle {
		return 0
	}
	return now.UTC().Sub(t.startedAt)
}

// WorkDuration returns elapsed time minus accumulated break time (up to now).
func (t *Timer) WorkDuration(now time.Time) time.Duration {
	if t.state == StateIdle {
		return 0
	}
	elapsed := now.UTC().Sub(t.startedAt)
	breakSoFar := t.totalBreak
	if t.state == StateOnBreak {
		breakSoFar += now.UTC().Sub(t.breakStart)
	}
	return elapsed - breakSoFar
}

// BreakDuration returns the accumulated break time up to now.
func (t *Timer) BreakDuration(now time.Time) time.Duration {
	if t.state == StateIdle {
		return 0
	}
	d := t.totalBreak
	if t.state == StateOnBreak {
		d += now.UTC().Sub(t.breakStart)
	}
	return d
}

// Session is the summary produced when a timer is stopped.
type Session struct {
	StartedAt  time.Time
	EndedAt    time.Time
	TotalBreak time.Duration
}

// WorkDuration returns the net work time (elapsed minus breaks).
func (s Session) WorkDuration() time.Duration {
	return s.EndedAt.Sub(s.StartedAt) - s.TotalBreak
}
