package progress

import (
	"fmt"
	"os"
	"sync"
)

// Counter tracks a numeric count or byte total and displays it on stderr.
// Useful for showing "Searching... 5 results" or "Downloading... 1.2 MB".
// Output is suppressed when stderr is not a TTY.
type Counter struct {
	mu      sync.Mutex
	label   string
	count   int
	bytes   int64
	mode    counterMode
	enabled bool
}

type counterMode int

const (
	modeCount counterMode = iota
	modeBytes
)

// NewCounter creates a new counter with the given label.
// The counter writes to stderr and is automatically disabled
// when stderr is not a terminal.
func NewCounter(label string) *Counter {
	return &Counter{
		label:   label,
		enabled: isTTY(),
		mode:    modeCount,
	}
}

// Increment adds n to the current count and refreshes the display.
func (c *Counter) Increment(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = modeCount
	c.count += n
	c.render()
}

// SetBytes sets the byte total and refreshes the display.
func (c *Counter) SetBytes(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = modeBytes
	c.bytes = n
	c.render()
}

// String returns the current display string without any terminal control characters.
func (c *Counter) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.display()
}

// display returns the formatted display string (must be called with lock held).
func (c *Counter) display() string {
	switch c.mode {
	case modeBytes:
		return fmt.Sprintf("%s %s", c.label, FormatBytes(c.bytes))
	default:
		return fmt.Sprintf("%s %d results", c.label, c.count)
	}
}

// render writes the current display to stderr with a carriage return
// so the line is overwritten on each update (must be called with lock held).
func (c *Counter) render() {
	if !c.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "\r%s", c.display())
}

// Done finalizes the counter display by writing a newline to stderr.
// Call this when the operation is complete to avoid overwriting the last line.
func (c *Counter) Done() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.enabled {
		return
	}
	fmt.Fprintln(os.Stderr)
}

// FormatBytes converts a byte count into a human-readable string.
// Uses decimal SI-like prefixes: B, KB, MB, GB.
func FormatBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
