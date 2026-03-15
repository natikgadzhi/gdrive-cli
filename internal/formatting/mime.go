// Package formatting handles URL parsing, MIME type mappings,
// filename sanitization, and query escaping for Google Drive operations.
package formatting

// Google Workspace MIME types.
const (
	MIMEGoogleDoc    = "application/vnd.google-apps.document"
	MIMEGoogleSheet  = "application/vnd.google-apps.spreadsheet"
	MIMEGoogleSlides = "application/vnd.google-apps.presentation"
)

// ExportMIME maps Google Workspace MIME types to their export MIME types.
var ExportMIME = map[string]string{
	MIMEGoogleDoc:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	MIMEGoogleSheet:  "text/csv",
	MIMEGoogleSlides: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
}

// ExportExtension maps Google Workspace MIME types to file extensions.
var ExportExtension = map[string]string{
	MIMEGoogleDoc:    ".docx",
	MIMEGoogleSheet:  ".csv",
	MIMEGoogleSlides: ".pptx",
}

// TypeLabel maps Google Workspace MIME types to human-readable labels.
var TypeLabel = map[string]string{
	MIMEGoogleDoc:    "Google Doc",
	MIMEGoogleSheet:  "Google Sheet",
	MIMEGoogleSlides: "Google Slides",
}

// MarkdownExportMIME maps Google Workspace MIME types to MIME types
// suitable for markdown/text export.
var MarkdownExportMIME = map[string]string{
	MIMEGoogleDoc:    "text/html",
	MIMEGoogleSheet:  "text/csv",
	MIMEGoogleSlides: "text/plain",
}
