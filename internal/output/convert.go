package output

import (
	"fmt"
	"io"
	"strings"

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

// PlainTextToSlideMarkdown converts plain text exported from the Google Drive
// API (text/plain for Slides) into structured Markdown with slide boundary
// markers. The Drive API separates slides with blank lines in the text/plain
// export. This function detects those boundaries and wraps each slide's content
// in a Markdown heading (## Slide N).
//
// If the input is empty or contains only whitespace, an empty string is returned.
func PlainTextToSlideMarkdown(plainText string) string {
	trimmed := strings.TrimSpace(plainText)
	if trimmed == "" {
		return ""
	}

	// The Google Drive text/plain export for Slides separates slides with
	// one or more blank lines. Split on runs of 2+ newlines to find slide
	// boundaries.
	rawSlides := splitSlides(trimmed)

	var buf strings.Builder
	for i, slide := range rawSlides {
		content := strings.TrimSpace(slide)
		if content == "" {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		fmt.Fprintf(&buf, "## Slide %d\n\n%s\n", i+1, content)
	}

	return buf.String()
}

// splitSlides splits plain text into slide chunks. The Drive API text/plain
// export for Slides uses two or more consecutive newlines as slide boundaries.
func splitSlides(text string) []string {
	// Normalise line endings to \n.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Split on two or more consecutive newlines (blank line boundaries).
	var slides []string
	var current strings.Builder
	lines := strings.Split(text, "\n")

	blankRun := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankRun++
			if blankRun == 2 && current.Len() > 0 {
				slides = append(slides, current.String())
				current.Reset()
				blankRun = 0
			}
			continue
		}

		// If we had exactly one blank line, it's a paragraph break within
		// the same slide — preserve it.
		if blankRun == 1 && current.Len() > 0 {
			current.WriteString("\n\n")
		}
		blankRun = 0

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)
	}

	if current.Len() > 0 {
		slides = append(slides, current.String())
	}

	return slides
}

// ExportAsMarkdown exports a Google Workspace file as a text representation.
//
// For Google Docs, the file is exported as HTML and then converted to Markdown.
// For Google Slides, the file is exported as plain text and then converted to
// structured Markdown with slide boundary markers.
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

	switch mimeType {
	case formatting.MIMEGoogleDoc:
		return HTMLToMarkdown(body)
	case formatting.MIMEGoogleSlides:
		return PlainTextToSlideMarkdown(string(body)), nil
	default:
		// Sheets (CSV) and any future types are returned as-is.
		return string(body), nil
	}
}
