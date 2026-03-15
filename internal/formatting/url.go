package formatting

import (
	"fmt"
	"net/url"
	"regexp"
)

var urlPatterns = []*regexp.Regexp{
	regexp.MustCompile(`docs\.google\.com/document/d/([a-zA-Z0-9_-]+)`),
	regexp.MustCompile(`docs\.google\.com/spreadsheets/d/([a-zA-Z0-9_-]+)`),
	regexp.MustCompile(`docs\.google\.com/presentation/d/([a-zA-Z0-9_-]+)`),
}

// ParseGoogleURL extracts the file ID from a Google Docs, Sheets, or Slides URL.
// It handles URL-encoded characters by decoding the URL before extraction.
// Returns a descriptive error listing supported URL formats on failure.
func ParseGoogleURL(rawURL string) (string, error) {
	decoded, err := url.QueryUnescape(rawURL)
	if err != nil {
		decoded = rawURL
	}

	for _, pattern := range urlPatterns {
		match := pattern.FindStringSubmatch(decoded)
		if match != nil {
			return match[1], nil
		}
	}

	return "", fmt.Errorf(
		"unrecognized Google Drive URL: %s\n\nSupported URL formats:\n"+
			"  https://docs.google.com/document/d/<ID>/...\n"+
			"  https://docs.google.com/spreadsheets/d/<ID>/...\n"+
			"  https://docs.google.com/presentation/d/<ID>/...",
		rawURL,
	)
}
