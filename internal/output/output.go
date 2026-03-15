// Package output provides utilities for writing structured
// output to stdout and debug logs to stderr.
package output

import (
	"encoding/json"
	"os"
)

// PrintJSON encodes v as indented JSON to stdout without HTML escaping.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
