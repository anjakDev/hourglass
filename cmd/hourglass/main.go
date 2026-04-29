package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anjakDev/hourglass/internal/db"
)

func main() {
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

	fmt.Println("hourglass — database ready at", dbPath)
	// TODO: start Bubbletea TUI
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
