package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/db"
	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui"
)

const helpText = `hourglass — keyboard-driven terminal time tracker

Usage:
  hourglass [flags]

Flags:
  -h, --help    Show this help message

Keybindings:

  Project list
    j / ↓       Move down
    k / ↑       Move up
    [s] / enter Start session
    [n]         New project
    [a]         Archive selected project
    [l]         View today's session log
    [q] / ctrl+c  Quit

  Active timer
    [b]         Toggle break / resume
    [s]         Stop session

  Session log
    esc / [q]   Back to project list

Data is stored in SQLite at $XDG_DATA_HOME/hourglass/hourglass.db
(defaults to ~/.local/share/hourglass/hourglass.db)
`

// run parses args and returns an exit code.
// -1 means "no flag matched — launch the TUI".
// 0 means help was printed successfully.
func run(args []string, out io.Writer) int {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(out, helpText)
			return 0
		}
	}
	return -1
}

func main() {
	if code := run(os.Args[1:], os.Stdout); code != -1 {
		os.Exit(code)
	}

	dbPath, err := defaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "hourglass: %v\n", err)
		os.Exit(1)
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hourglass: open database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	pr := repository.NewProjectRepo(conn)
	sr := repository.NewSessionRepo(conn)
	app := tui.New(pr, sr)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "hourglass: %v\n", err)
		os.Exit(1)
	}
}

func defaultDBPath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home dir: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(dataHome, "hourglass")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}
	return filepath.Join(dir, "hourglass.db"), nil
}
