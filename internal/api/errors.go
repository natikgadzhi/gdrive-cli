package api

import (
	"errors"
	"strings"

	"google.golang.org/api/googleapi"
)

// IsCannotExportFile reports whether err is a Google Drive "cannotExportFile" error.
// This typically occurs when a user has view-only access and the owner has disabled
// downloads/exports.
func IsCannotExportFile(err error) bool {
	return hasErrorReason(err, "cannotExportFile")
}

// IsExportSizeLimitExceeded reports whether err is a Google Drive
// "exportSizeLimitExceeded" error. This occurs when a document is too large
// for the requested export format.
func IsExportSizeLimitExceeded(err error) bool {
	return hasErrorReason(err, "exportSizeLimitExceeded")
}

// hasErrorReason checks whether the given error is a Google API error containing
// a specific reason string. It checks both the structured Errors slice and the
// top-level message as a fallback.
func hasErrorReason(err error, reason string) bool {
	if err == nil {
		return false
	}

	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	for _, e := range apiErr.Errors {
		if e.Reason == reason {
			return true
		}
	}

	// Also check the message as a fallback — some errors embed the reason
	// in the message text rather than in structured error entries.
	return strings.Contains(apiErr.Message, reason)
}
