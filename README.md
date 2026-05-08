# hourglass
<p align='center'><img src="logo_simple.png" width="200" alt="hourglass logo"></br></p>
A keyboard-driven terminal time tracker. Create projects, start a timer, take breaks, stop — all without leaving your terminal. Session history and daily totals are persisted locally in SQLite.

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| TUI framework | [Bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| UI components | [Bubbles](https://github.com/charmbracelet/bubbles) (textinput) |
| Database | SQLite via [go-sqlite3](https://github.com/mattn/go-sqlite3) (CGO) |
| Testing | [testify](https://github.com/stretchr/testify) |

## Requirements

- Go 1.22+
- A C compiler (required by `go-sqlite3` — `gcc` or `clang`)

## Build

```sh
go build -o hourglass ./cmd/hourglass
```

## Run

```sh
./hourglass
```

On first run, the database is created at `$XDG_DATA_HOME/hourglass/hourglass.db` (defaults to `~/.local/share/hourglass/hourglass.db`).

## Tests

```sh
go test ./...
```

All tests use an in-memory SQLite database — no setup required.

## MVP features

### Project management
- Create named projects
- Archive projects (removes them from the active list)
- Browse projects with `j`/`k` or arrow keys

### Timer
- Start a session for the selected project (`s` or `enter`)
- Toggle break / resume (`b`) while a session is running
- Stop the session (`s`) — opens the session editor before saving

### Session editor
Shown automatically after stopping a session. Useful for correcting sessions that were started but forgotten to be stopped.
- Displays start time, end time, and net duration
- Navigate between start and end fields; press `e` to edit the highlighted field
- Duration recalculates automatically after each edit
- Press `enter` to save (with or without edits), or `esc` to revert to the original times

### Session log
- View today's sessions for any project (`l`)
- Each row shows start time, end time, work duration, and break time if applicable
- Total work time for the day shown at the bottom

### Daily totals
- The project list shows today's total tracked time per project at a glance

## Keybindings

### Project list
| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `s` / `enter` | Start session |
| `n` | New project |
| `a` | Archive selected project |
| `l` | View today's session log |
| `q` / `ctrl+c` | Quit |

### Active timer
| Key | Action |
|---|---|
| `b` | Toggle break / resume |
| `s` | Stop session |

### Session editor
| Key | Action |
|---|---|
| `tab` / `j` / `k` | Switch between start and end fields |
| `e` | Edit the selected field |
| `enter` | Save session |
| `esc` | Revert to original times and save |

**While editing a field:**
| Key | Action |
|---|---|
| `enter` | Confirm the new time (`YYYY-MM-DD HH:MM`) |
| `esc` | Discard edit, return to summary |
| `ctrl+u` | Clear input |
| `backspace` | Delete last character |

### Session log
| Key | Action |
|---|---|
| `esc` / `q` | Back to project list |

## Data storage

Sessions are stored locally in SQLite. The schema uses two tables:

- `projects` — name, description, color, archived flag
- `sessions` — project reference, start/end timestamps, break duration in seconds

Migrations run automatically on startup from embedded SQL files — no manual schema setup needed.

## Roadmap

### Phase 2 — Reports
- Daily / weekly / monthly summaries per project
- Unicode sparkline charts in the TUI
- Goal tracking: set a target hours/week per project, show a progress bar
- Export to CSV, JSON, or Markdown

### Phase 3 — Metrics
- `hourglass metrics` subcommand: runs a lightweight Prometheus `/metrics` HTTP endpoint
- Exposes gauges: `hourglass_session_seconds_total{project}`, `hourglass_sessions_count{project}`
- Point Grafana at `localhost:9091` via a Prometheus datasource — no external runtime dependency unless metrics mode is active

### Phase 4 — Remote mode
- `hourglass serve`: starts a REST API server backed by SQLite or Postgres
- `~/.config/hourglass/config.toml`: `mode = "remote"`, `server_url`, `api_key`
- TUI detects remote config and routes all reads/writes through an HTTP client instead of local SQLite
- Repository layer abstracted behind an interface so local and remote are interchangeable
