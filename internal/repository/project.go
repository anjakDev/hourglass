package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// Project represents a trackable project.
type Project struct {
	ID          int64
	Name        string
	Description string
	Color       string
	Archived    bool
	CreatedAt   time.Time
}

// ProjectRepo provides persistence operations for projects.
type ProjectRepo struct {
	db *sql.DB
}

// NewProjectRepo creates a new ProjectRepo backed by the given database.
func NewProjectRepo(db *sql.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

// Create inserts a new project and returns its ID.
func (r *ProjectRepo) Create(name, description, color string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO projects (name, description, color) VALUES (?, ?, ?)`,
		name, description, color,
	)
	if err != nil {
		return 0, fmt.Errorf("create project: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// GetByID returns the project with the given ID.
func (r *ProjectRepo) GetByID(id int64) (Project, error) {
	row := r.db.QueryRow(
		`SELECT id, name, description, color, archived, created_at FROM projects WHERE id = ?`, id,
	)
	return scanProject(row)
}

// List returns all non-archived projects ordered by name.
func (r *ProjectRepo) List() ([]Project, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, color, archived, created_at
		 FROM projects WHERE archived = 0 ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

// ListAll returns all projects including archived ones.
func (r *ProjectRepo) ListAll() ([]Project, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, color, archived, created_at
		 FROM projects ORDER BY archived, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all projects: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

// Archive marks the project as archived.
func (r *ProjectRepo) Archive(id int64) error {
	res, err := r.db.Exec(`UPDATE projects SET archived = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("archive project: %w", err)
	}
	return requireAffected(res)
}

// Delete permanently removes a project. Sessions must be deleted first.
func (r *ProjectRepo) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return requireAffected(res)
}

// fillProject populates p from a single scanned row.
// scan is either (*sql.Row).Scan or (*sql.Rows).Scan.
func fillProject(p *Project, scan func(...any) error) error {
	var archived int
	var createdUnix int64
	if err := scan(&p.ID, &p.Name, &p.Description, &p.Color, &archived, &createdUnix); err != nil {
		return err
	}
	p.Archived = archived == 1
	p.CreatedAt = time.Unix(createdUnix, 0).UTC()
	return nil
}

func scanProject(row *sql.Row) (Project, error) {
	var p Project
	if err := fillProject(&p, row.Scan); errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	} else if err != nil {
		return Project{}, fmt.Errorf("scan project: %w", err)
	}
	return p, nil
}

func scanProjects(rows *sql.Rows) ([]Project, error) {
	var projects []Project
	for rows.Next() {
		var p Project
		if err := fillProject(&p, rows.Scan); err != nil {
			return nil, fmt.Errorf("scan project row: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
