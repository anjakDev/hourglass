package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session represents a completed (or in-progress) work session.
type Session struct {
	ID                   int64
	ProjectID            int64
	StartedAt            time.Time
	EndedAt              *time.Time // nil when still running
	BreakDurationSeconds int64
	CreatedAt            time.Time
}

// WorkDuration returns net work time (elapsed minus breaks).
// Returns 0 for an open session.
func (s Session) WorkDuration() time.Duration {
	if s.EndedAt == nil {
		return 0
	}
	elapsed := s.EndedAt.Sub(s.StartedAt)
	return elapsed - time.Duration(s.BreakDurationSeconds)*time.Second
}

// ProjectTotal holds the aggregate work time for a project on a given day.
type ProjectTotal struct {
	ProjectID   int64
	ProjectName string
	Total       time.Duration
}

// SessionRepo provides persistence operations for sessions.
type SessionRepo struct {
	db *sql.DB
}

// NewSessionRepo creates a new SessionRepo backed by the given database.
func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// StartSession opens a new session for the given project and returns its ID.
func (r *SessionRepo) StartSession(projectID int64, startedAt time.Time) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO sessions (project_id, started_at) VALUES (?, ?)`,
		projectID, startedAt.UTC().Unix(),
	)
	if err != nil {
		return 0, fmt.Errorf("start session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// StopSession closes an open session by setting ended_at and the accumulated
// break duration.
func (r *SessionRepo) StopSession(id int64, endedAt time.Time, breakSeconds int64) error {
	res, err := r.db.Exec(
		`UPDATE sessions SET ended_at = ?, break_duration_seconds = ? WHERE id = ? AND ended_at IS NULL`,
		endedAt.UTC().Unix(), breakSeconds, id,
	)
	if err != nil {
		return fmt.Errorf("stop session: %w", err)
	}
	return requireAffected(res)
}

// ActiveSession returns the currently open session (ended_at IS NULL), if any.
func (r *SessionRepo) ActiveSession() (Session, error) {
	row := r.db.QueryRow(
		`SELECT id, project_id, started_at, ended_at, break_duration_seconds, created_at
		 FROM sessions WHERE ended_at IS NULL LIMIT 1`,
	)
	return scanSession(row)
}

// todayBounds returns [start, end) Unix timestamps (seconds) for today in UTC.
func todayBounds() (start, end int64) {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return midnight.Unix(), midnight.Add(24 * time.Hour).Unix()
}

// ListToday returns all sessions (open or closed) that started today (UTC date).
func (r *SessionRepo) ListToday() ([]Session, error) {
	todayStart, todayEnd := todayBounds()

	rows, err := r.db.Query(
		`SELECT id, project_id, started_at, ended_at, break_duration_seconds, created_at
		 FROM sessions
		 WHERE started_at >= ? AND started_at < ?
		 ORDER BY started_at`,
		todayStart, todayEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("list today sessions: %w", err)
	}
	defer rows.Close()
	return scanSessions(rows)
}

// TodayTotalsByProject returns the net work duration per project for today.
func (r *SessionRepo) TodayTotalsByProject() ([]ProjectTotal, error) {
	todayStart, todayEnd := todayBounds()
	nowUnix := time.Now().UTC().Unix()

	rows, err := r.db.Query(
		`SELECT s.project_id, p.name,
		        SUM(COALESCE(s.ended_at, ?) - s.started_at - s.break_duration_seconds) AS work_seconds
		 FROM sessions s
		 JOIN projects p ON p.id = s.project_id
		 WHERE s.started_at >= ? AND s.started_at < ?
		 GROUP BY s.project_id, p.name
		 ORDER BY p.name`,
		nowUnix, todayStart, todayEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("today totals: %w", err)
	}
	defer rows.Close()

	var totals []ProjectTotal
	for rows.Next() {
		var pt ProjectTotal
		var seconds int64
		if err := rows.Scan(&pt.ProjectID, &pt.ProjectName, &seconds); err != nil {
			return nil, fmt.Errorf("scan total row: %w", err)
		}
		pt.Total = time.Duration(seconds) * time.Second
		totals = append(totals, pt)
	}
	return totals, rows.Err()
}

// UpdateSession overwrites the start time, end time, and break duration for a
// session. Used when the user edits a just-stopped session.
func (r *SessionRepo) UpdateSession(id int64, startedAt, endedAt time.Time, breakSeconds int64) error {
	res, err := r.db.Exec(
		`UPDATE sessions SET started_at = ?, ended_at = ?, break_duration_seconds = ? WHERE id = ?`,
		startedAt.UTC().Unix(), endedAt.UTC().Unix(), breakSeconds, id,
	)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return requireAffected(res)
}

// DeleteAllSessions removes every session (open or closed) for the given project.
func (r *SessionRepo) DeleteAllSessions(projectID int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE project_id = ?`, projectID)
	if err != nil {
		return fmt.Errorf("delete all sessions: %w", err)
	}
	return nil
}

// fillSession populates s from a single scanned row.
// scan is either (*sql.Row).Scan or (*sql.Rows).Scan.
func fillSession(s *Session, scan func(...any) error) error {
	var startedUnix, createdUnix int64
	var endedUnix sql.NullInt64
	if err := scan(&s.ID, &s.ProjectID, &startedUnix, &endedUnix, &s.BreakDurationSeconds, &createdUnix); err != nil {
		return err
	}
	s.StartedAt = time.Unix(startedUnix, 0).UTC()
	if endedUnix.Valid {
		t := time.Unix(endedUnix.Int64, 0).UTC()
		s.EndedAt = &t
	}
	s.CreatedAt = time.Unix(createdUnix, 0).UTC()
	return nil
}

func scanSession(row *sql.Row) (Session, error) {
	var s Session
	if err := fillSession(&s, row.Scan); errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	} else if err != nil {
		return Session{}, fmt.Errorf("scan session: %w", err)
	}
	return s, nil
}

func scanSessions(rows *sql.Rows) ([]Session, error) {
	var sessions []Session
	for rows.Next() {
		var s Session
		if err := fillSession(&s, rows.Scan); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
