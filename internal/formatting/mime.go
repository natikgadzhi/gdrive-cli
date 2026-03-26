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
			"pdf": {
				ExportMIME: "application/pdf",
				Extension:  ".pdf",
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

	// Strip a leading dot so that ".md" is treated the same as "md".
	exportFormat = strings.TrimPrefix(exportFormat, ".")

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

// IsNativeGoogleType reports whether the MIME type is a native Google Workspace
// type that requires Files.Export() for download.
func IsNativeGoogleType(mime string) bool {
	_, ok := mimeRegistry[mime]
	return ok
}

// binaryTypeInfo holds the file extension and label for a non-native MIME type.
type binaryTypeInfo struct {
	Extension string
	Label     string
}

// binaryMIMEExtensions maps common non-native MIME types that can be downloaded
// directly via alt=media to their file extension and human-readable label.
var binaryMIMEExtensions = map[string]binaryTypeInfo{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {".docx", "Word Document"},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {".xlsx", "Excel Spreadsheet"},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {".pptx", "PowerPoint Presentation"},
	"application/pdf":                {".pdf", "PDF"},
	"application/msword":             {".doc", "Word Document (Legacy)"},
	"application/vnd.ms-excel":       {".xls", "Excel Spreadsheet (Legacy)"},
	"application/vnd.ms-powerpoint":  {".ppt", "PowerPoint (Legacy)"},
	"text/plain":                     {".txt", "Text File"},
	"text/csv":                       {".csv", "CSV File"},
	"application/zip":                {".zip", "ZIP Archive"},
	"image/png":                      {".png", "PNG Image"},
	"image/jpeg":                     {".jpg", "JPEG Image"},
	"application/octet-stream":       {"", "Binary File"},
}

// GetBinaryTypeInfo returns the extension and label for a non-native MIME type
// that can be downloaded directly. Returns false if the MIME type is not in
// the known binary type map.
func GetBinaryTypeInfo(mime string) (extension string, label string, ok bool) {
	info, ok := binaryMIMEExtensions[mime]
	if !ok {
		return "", "", false
	}
	return info.Extension, info.Label, true
}

// ExtensionFromFilename extracts the file extension from a filename.
// Returns empty string if no extension is found.
func ExtensionFromFilename(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
		if name[i] == '/' || name[i] == '\\' {
			break
		}
	}
	return ""
}
