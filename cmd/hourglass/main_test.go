package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpFlag(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"--help"}},
		{[]string{"-h"}},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		code := run(tt.args, &buf)
		out := buf.String()

		assert.Equal(t, 0, code)
		assert.True(t, strings.Contains(out, "hourglass"), "missing app name")
		assert.True(t, strings.Contains(out, "Keybindings"), "missing keybindings section")
		assert.True(t, strings.Contains(out, "[s]"), "missing start key hint")
		assert.True(t, strings.Contains(out, "[b]"), "missing break key hint")
		assert.True(t, strings.Contains(out, "[q]"), "missing quit key hint")
	}
}

func TestNoArgs(t *testing.T) {
	// Without --help, run returns -1 to signal "launch TUI".
	var buf bytes.Buffer
	code := run([]string{}, &buf)
	assert.Equal(t, -1, code)
}
