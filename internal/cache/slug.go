package cache

import (
	"regexp"
	"strings"
)

var (
	// slugUnsafe matches characters that are not lowercase alphanumeric, hyphens, or underscores.
	slugUnsafe = regexp.MustCompile(`[^a-z0-9_-]+`)
	// multiHyphen collapses consecutive hyphens into one.
	multiHyphen = regexp.MustCompile(`-{2,}`)
)

// GenerateSlug produces a filesystem-safe slug from a document name and file ID.
// The slug is lowercase, with unsafe characters replaced by hyphens, and has the
// first 8 characters of the file ID appended for uniqueness.
//
// Example: GenerateSlug("Q1 Budget Report", "1aBcDeFgHiJkL") → "q1-budget-report-1abcdefg"
func GenerateSlug(name string, fileID string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugUnsafe.ReplaceAllString(s, "-")
	s = multiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	// Append first 8 chars of the file ID (or the whole thing if shorter).
	idSuffix := strings.ToLower(fileID)
	if len(idSuffix) > 8 {
		idSuffix = idSuffix[:8]
	}

	if s == "" {
		return idSuffix
	}
	return s + "-" + idSuffix
}
