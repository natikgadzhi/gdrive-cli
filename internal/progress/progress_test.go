package progress

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{2469606195, "2.3 GB"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCounterIncrement(t *testing.T) {
	c := NewCounter("Searching...")
	// Force disabled so render() is a no-op (no TTY in tests).
	c.enabled = false

	c.Increment(3)
	if c.count != 3 {
		t.Errorf("count = %d, want 3", c.count)
	}
	c.Increment(2)
	if c.count != 5 {
		t.Errorf("count = %d, want 5", c.count)
	}

	want := "Searching... 5 results"
	got := c.String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestCounterSetBytes(t *testing.T) {
	c := NewCounter("Downloading...")
	c.enabled = false

	c.SetBytes(1572864) // 1.5 MB
	want := "Downloading... 1.5 MB"
	got := c.String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestCounterModeSwitching(t *testing.T) {
	c := NewCounter("Progress...")
	c.enabled = false

	// Start in count mode.
	c.Increment(10)
	got := c.String()
	if got != "Progress... 10 results" {
		t.Errorf("after Increment, String() = %q", got)
	}

	// Switch to bytes mode.
	c.SetBytes(2048)
	got = c.String()
	if got != "Progress... 2.0 KB" {
		t.Errorf("after SetBytes, String() = %q", got)
	}
}

func TestSpinnerNonTTY(t *testing.T) {
	// In test environments stderr is not a TTY, so the spinner
	// should be disabled. Start/Stop/UpdateMessage should be no-ops.
	s := NewSpinner("Loading...")
	if s.enabled {
		t.Skip("stderr is a TTY in this test environment; skipping non-TTY test")
	}
	// These should not panic or produce output.
	s.Start()
	s.UpdateMessage("Still loading...")
	s.Stop()
}

func TestCounterNonTTYNoOutput(t *testing.T) {
	// When stderr is not a TTY (the case in tests), render() should
	// be a no-op. We verify by checking that enabled is false.
	c := NewCounter("Test...")
	if c.enabled {
		t.Skip("stderr is a TTY in this test environment; skipping non-TTY test")
	}
	// These should not panic or produce output to stderr.
	c.Increment(1)
	c.SetBytes(1024)
	c.Done()
}

func TestCounterZeroState(t *testing.T) {
	c := NewCounter("Items...")
	c.enabled = false
	got := c.String()
	want := "Items... 0 results"
	if got != want {
		t.Errorf("zero-state String() = %q, want %q", got, want)
	}
}
