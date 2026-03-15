package formatting

import (
	"regexp"
	"strings"
)

var unsafeChars = regexp.MustCompile(`[/\\:*?"<>|]`)

// SanitizeFilename replaces characters that are unsafe in filenames
// (/ \ : * ? " < > |) with underscores and strips leading/trailing whitespace.
func SanitizeFilename(name string) string {
	return strings.TrimSpace(unsafeChars.ReplaceAllString(name, "_"))
}
