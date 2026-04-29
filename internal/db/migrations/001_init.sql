CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    name        TEXT     NOT NULL UNIQUE,
    description TEXT     NOT NULL DEFAULT '',
    color       TEXT     NOT NULL DEFAULT '',
    archived    INTEGER  NOT NULL DEFAULT 0,
    created_at  INTEGER  NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS sessions (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id             INTEGER NOT NULL REFERENCES projects(id),
    started_at             INTEGER NOT NULL,           -- Unix seconds UTC
    ended_at               INTEGER,                    -- Unix seconds UTC; NULL when open
    break_duration_seconds INTEGER NOT NULL DEFAULT 0,
    created_at             INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS breaks (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id),
    started_at INTEGER NOT NULL,   -- Unix seconds UTC
    ended_at   INTEGER            -- Unix seconds UTC; NULL when open
);
