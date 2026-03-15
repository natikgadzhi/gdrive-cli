// Package output provides utilities for writing structured
// output to stdout and debug logs to stderr.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// SilentError is a sentinel error returned by Errorf after
// it has already written the JSON error message to stdout.
// Cobra's RunE should return this to signal a non-zero exit code
// without Cobra printing its own error message (requires
// SilenceErrors: true on the root command).
type SilentError struct {
	// Message is the error message that was already printed as JSON.
	Message string
}

func (e *SilentError) Error() string {
	return e.Message
}

// IsSilentError reports whether err is (or wraps) a SilentError.
func IsSilentError(err error) bool {
	var se *SilentError
	return errors.As(err, &se)
}

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

// Errorf prints an error StatusMessage to stdout and returns a SilentError.
// The JSON error has already been written, so the caller should return
// the error to Cobra without additional printing.
// Arguments are handled in the manner of fmt.Sprintf.
func Errorf(format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	_ = PrintJSON(StatusMessage{Status: "error", Message: msg})
	return &SilentError{Message: msg}
}

// PrintJSON encodes v as indented JSON to stdout without HTML escaping.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
