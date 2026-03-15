// Package formatting handles URL parsing, MIME type mappings,
// filename sanitization, and query escaping for Google Drive operations.
package formatting

import (
	"fmt"
	"sort"
	"strings"
)

// Google Workspace MIME types.
const (
	MIMEGoogleDoc    = "application/vnd.google-apps.document"
	MIMEGoogleSheet  = "application/vnd.google-apps.spreadsheet"
	MIMEGoogleSlides = "application/vnd.google-apps.presentation"
)

// ExportFormatInfo holds the export MIME type and file extension for a given
// user-facing export format string (e.g. "docx", "md", "csv", "pptx").
type ExportFormatInfo struct {
	ExportMIME string
	Extension  string
	// NeedsMarkdownConversion is true when the export MIME is HTML and the
	// result must be converted to Markdown before saving (Google Docs as md).
	NeedsMarkdownConversion bool
}

// mimeInfo holds all export metadata for a single Google Workspace MIME type.
type mimeInfo struct {
	ExportMIME         string
	Extension          string
	Label              string
	MarkdownExportMIME string
	// DefaultExportFormat is the default user-facing export format (e.g. "docx").
	DefaultExportFormat string
	// ValidExportFormats maps user-facing format strings to their export info.
	ValidExportFormats map[string]ExportFormatInfo
}

// mimeRegistry maps each Google Workspace MIME type to its export metadata.
// All per-type data lives in one place, so new types only need a single entry.
var mimeRegistry = map[string]mimeInfo{
	MIMEGoogleDoc: {
		ExportMIME:          "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Extension:           ".docx",
		Label:               "Google Doc",
		MarkdownExportMIME:  "text/html",
		DefaultExportFormat: "docx",
		ValidExportFormats: map[string]ExportFormatInfo{
			"docx": {
				ExportMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				Extension:  ".docx",
			},
			"md": {
				ExportMIME:              "text/html",
				Extension:               ".md",
				NeedsMarkdownConversion: true,
			},
		},
	},
	MIMEGoogleSheet: {
		ExportMIME:          "text/csv",
		Extension:           ".csv",
		Label:               "Google Sheet",
		MarkdownExportMIME:  "text/csv",
		DefaultExportFormat: "csv",
		ValidExportFormats: map[string]ExportFormatInfo{
			"csv": {
				ExportMIME: "text/csv",
				Extension:  ".csv",
			},
		},
	},
	MIMEGoogleSlides: {
		ExportMIME:          "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Extension:           ".pptx",
		Label:               "Google Slides",
		MarkdownExportMIME:  "text/plain",
		DefaultExportFormat: "pptx",
		ValidExportFormats: map[string]ExportFormatInfo{
			"pptx": {
				ExportMIME: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
				Extension:  ".pptx",
			},
			"md": {
				ExportMIME: "text/plain",
				Extension:  ".md",
			},
		},
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

// DefaultExportFormat returns the default export format string for a Google
// Workspace MIME type (e.g. "docx" for Google Docs). Returns "" and false if
// the MIME type is unknown.
func DefaultExportFormat(mime string) (string, bool) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return "", false
	}
	return info.DefaultExportFormat, true
}

// ResolveExportFormat resolves a user-provided export format string against a
// Google Workspace MIME type. If exportFormat is empty, the default for that
// type is used. Returns the resolved ExportFormatInfo or an error describing
// which formats are valid for the given type.
func ResolveExportFormat(mime, exportFormat string) (ExportFormatInfo, error) {
	info, ok := mimeRegistry[mime]
	if !ok {
		return ExportFormatInfo{}, fmt.Errorf("unsupported MIME type: %s", mime)
	}

	// Use the default if no format was specified.
	if exportFormat == "" {
		exportFormat = info.DefaultExportFormat
	}

	fmtInfo, ok := info.ValidExportFormats[exportFormat]
	if !ok {
		return ExportFormatInfo{}, fmt.Errorf(
			"%s can be exported as %s",
			info.Label, validFormatsString(info.ValidExportFormats),
		)
	}
	return fmtInfo, nil
}

// validFormatsString returns a human-readable list of valid export format keys,
// sorted alphabetically (e.g. "csv" or "docx or md").
func validFormatsString(formats map[string]ExportFormatInfo) string {
	keys := make([]string, 0, len(formats))
	for k := range formats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, " or ")
}
