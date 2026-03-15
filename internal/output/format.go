package output

import "fmt"

// Format represents an output format for CLI commands.
type Format int

const (
	// FormatJSON is the default output format — structured JSON to stdout.
	FormatJSON Format = iota
	// FormatMarkdown outputs human-readable markdown.
	FormatMarkdown
)

// ParseFormat converts a string ("json" or "markdown") into a Format value.
// Returns an error for unrecognized strings.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "json":
		return FormatJSON, nil
	case "markdown":
		return FormatMarkdown, nil
	default:
		return FormatJSON, fmt.Errorf("invalid format %q: must be \"json\" or \"markdown\"", s)
	}
}

// String returns the string representation of a Format.
func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatMarkdown:
		return "markdown"
	default:
		return "json"
	}
}
