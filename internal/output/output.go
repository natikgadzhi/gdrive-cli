// Package output provides utilities for writing structured
// output to stdout and debug logs to stderr.
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// StatusMessage is the standard JSON envelope for commands that report
// success or failure with a human-readable message.
type StatusMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// OK prints a success StatusMessage to stdout.
func OK(message string) error {
	return PrintJSON(StatusMessage{Status: "ok", Message: message})
}

// Errorf prints an error StatusMessage to stdout.
// Arguments are handled in the manner of fmt.Sprintf.
func Errorf(format string, args ...any) error {
	return PrintJSON(StatusMessage{Status: "error", Message: fmt.Sprintf(format, args...)})
}

// PrintJSON encodes v as indented JSON to stdout without HTML escaping.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
