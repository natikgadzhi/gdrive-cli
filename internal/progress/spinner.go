// Package progress provides terminal progress indicators
// for long-running operations like searches and downloads.
// All output goes to stderr to preserve stdout for JSON output.
// Indicators are automatically disabled when stderr is not a TTY
// (e.g., in CI, pipes, or non-interactive environments).
package progress

import (
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"golang.org/x/term"
)

// isTTY reports whether stderr is connected to a terminal.
func isTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// Spinner displays an animated spinner on stderr for long-running operations.
// It is a no-op when stderr is not a TTY.
type Spinner struct {
	mu      sync.Mutex
	inner   *spinner.Spinner
	active  bool
	enabled bool
}

// NewSpinner creates a new spinner with the given message.
// The spinner writes to stderr and is automatically disabled
// when stderr is not a terminal.
func NewSpinner(message string) *Spinner {
	enabled := isTTY()
	s := &Spinner{enabled: enabled}
	if enabled {
		s.inner = spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		s.inner.Suffix = " " + message
	}
	return s
}

// Start begins the spinner animation.
// It is a no-op if stderr is not a TTY or if the spinner is already active.
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.active {
		return
	}
	s.inner.Start()
	s.active = true
}

// Stop halts the spinner animation and clears the line.
// It is a no-op if stderr is not a TTY or if the spinner is not active.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || !s.active {
		return
	}
	s.inner.Stop()
	s.active = false
}

// UpdateMessage changes the spinner's display message while it is running.
// It is a no-op if stderr is not a TTY.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled {
		return
	}
	s.inner.Suffix = " " + msg
}
