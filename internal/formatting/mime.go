// Package formatting handles URL parsing, MIME type mappings,
// filename sanitization, and query escaping for Google Drive operations.
package formatting

// Google Workspace MIME types.
const (
	MIMEGoogleDoc    = "application/vnd.google-apps.document"
	MIMEGoogleSheet  = "application/vnd.google-apps.spreadsheet"
	MIMEGoogleSlides = "application/vnd.google-apps.presentation"
)

// mimeInfo holds all export metadata for a single Google Workspace MIME type.
type mimeInfo struct {
	ExportMIME         string
	Extension          string
	Label              string
	MarkdownExportMIME string
}

// mimeRegistry maps each Google Workspace MIME type to its export metadata.
// All per-type data lives in one place, so new types only need a single entry.
var mimeRegistry = map[string]mimeInfo{
	MIMEGoogleDoc: {
		ExportMIME:         "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Extension:          ".docx",
		Label:              "Google Doc",
		MarkdownExportMIME: "text/html",
	},
	MIMEGoogleSheet: {
		ExportMIME:         "text/csv",
		Extension:          ".csv",
		Label:              "Google Sheet",
		MarkdownExportMIME: "text/csv",
	},
	MIMEGoogleSlides: {
		ExportMIME:         "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Extension:          ".pptx",
		Label:              "Google Slides",
		MarkdownExportMIME: "text/plain",
	},
}

// GetExportMIME returns the export MIME type for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetExportMIME(mime string) (string, bool) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return "", false
	}
	return info.ExportMIME, true
}

// GetExportExtension returns the file extension for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetExportExtension(mime string) (string, bool) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return "", false
	}
	return info.Extension, true
}

// GetTypeLabel returns the human-readable label for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetTypeLabel(mime string) (string, bool) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return "", false
	}
	return info.Label, true
}

// GetMarkdownExportMIME returns the markdown/text export MIME type for a
// Google Workspace MIME type. The second return value reports whether the key
// was found.
func GetMarkdownExportMIME(mime string) (string, bool) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return "", false
	}
	return info.MarkdownExportMIME, true
}

// SupportedMIMETypes returns the set of Google Workspace MIME types that this
// package knows how to export. Useful for building Drive API query filters.
func SupportedMIMETypes() []string {
	return []string{MIMEGoogleDoc, MIMEGoogleSheet, MIMEGoogleSlides}
}
