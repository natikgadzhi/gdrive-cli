package formatting

import "strings"

// EscapeQuery escapes a string for use in Google Drive API query syntax.
// It escapes backslashes first, then single quotes.
func EscapeQuery(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `'`, `\'`)
	return q
}
