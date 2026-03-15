package output

import (
	"fmt"
	"io"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"google.golang.org/api/drive/v3"
)

// HTMLToMarkdown converts an HTML byte slice to a Markdown string.
// It preserves headings, bold, italic, links, lists, and code blocks,
// while stripping scripts, styles, and non-content elements.
func HTMLToMarkdown(htmlContent []byte) (string, error) {
	md, err := htmltomarkdown.ConvertString(string(htmlContent))
	if err != nil {
		return "", fmt.Errorf("html-to-markdown conversion failed: %w", err)
	}
	return md, nil
}

// ExportAsMarkdown exports a Google Workspace file as a text representation.
//
// For Google Docs, the file is exported as HTML and then converted to Markdown.
// For Google Slides, the file is exported as plain text (returned as-is).
// For Google Sheets, the file is exported as CSV (returned as-is).
//
// Returns an error if the MIME type is unsupported or the Drive API call fails.
func ExportAsMarkdown(svc *drive.Service, fileID string, mimeType string) (string, error) {
	exportMIME, ok := formatting.GetMarkdownExportMIME(mimeType)
	if !ok {
		return "", fmt.Errorf("unsupported MIME type for markdown export: %s", mimeType)
	}

	resp, err := svc.Files.Export(fileID, exportMIME).Download()
	if err != nil {
		return "", fmt.Errorf("drive export failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading export response: %w", err)
	}

	// Google Docs are exported as HTML and converted to Markdown.
	// Sheets (CSV) and Slides (plain text) are returned as-is.
	if mimeType == formatting.MIMEGoogleDoc {
		return HTMLToMarkdown(body)
	}

	return string(body), nil
}
