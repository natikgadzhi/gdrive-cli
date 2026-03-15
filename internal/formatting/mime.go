// Package formatting handles URL parsing, MIME type mappings,
// filename sanitization, and query escaping for Google Drive operations.
package formatting

// Google Workspace MIME types.
const (
	MIMEGoogleDoc    = "application/vnd.google-apps.document"
	MIMEGoogleSheet  = "application/vnd.google-apps.spreadsheet"
	MIMEGoogleSlides = "application/vnd.google-apps.presentation"
)

// unexported maps — callers use the accessor functions below.
var exportMIME = map[string]string{
	MIMEGoogleDoc:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	MIMEGoogleSheet:  "text/csv",
	MIMEGoogleSlides: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
}

var exportExtension = map[string]string{
	MIMEGoogleDoc:    ".docx",
	MIMEGoogleSheet:  ".csv",
	MIMEGoogleSlides: ".pptx",
}

var typeLabel = map[string]string{
	MIMEGoogleDoc:    "Google Doc",
	MIMEGoogleSheet:  "Google Sheet",
	MIMEGoogleSlides: "Google Slides",
}

var markdownExportMIME = map[string]string{
	MIMEGoogleDoc:    "text/html",
	MIMEGoogleSheet:  "text/csv",
	MIMEGoogleSlides: "text/plain",
}

// GetExportMIME returns the export MIME type for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetExportMIME(mime string) (string, bool) {
	v, ok := exportMIME[mime]
	return v, ok
}

// GetExportExtension returns the file extension for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetExportExtension(mime string) (string, bool) {
	v, ok := exportExtension[mime]
	return v, ok
}

// GetTypeLabel returns the human-readable label for a Google Workspace MIME type.
// The second return value reports whether the key was found.
func GetTypeLabel(mime string) (string, bool) {
	v, ok := typeLabel[mime]
	return v, ok
}

// GetMarkdownExportMIME returns the markdown/text export MIME type for a
// Google Workspace MIME type. The second return value reports whether the key
// was found.
func GetMarkdownExportMIME(mime string) (string, bool) {
	v, ok := markdownExportMIME[mime]
	return v, ok
}

// SupportedMIMETypes returns the set of Google Workspace MIME types that this
// package knows how to export. Useful for building Drive API query filters.
func SupportedMIMETypes() []string {
	return []string{MIMEGoogleDoc, MIMEGoogleSheet, MIMEGoogleSlides}
}
