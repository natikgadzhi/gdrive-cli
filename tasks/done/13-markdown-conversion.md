# Task 13: Markdown conversion for Google Docs and Slides

**Phase**: 3 — Output & Cache
**Depends on**: 05
**Blocks**: 14

## Description

Implement conversion of Google Docs and Slides content to Markdown for cache storage. Uses Google's HTML export and converts to Markdown.

## Acceptance Criteria

- [ ] `internal/output/convert.go`:
  - `HTMLToMarkdown(html []byte) (string, error)` — converts HTML to Markdown using `html-to-markdown` library
  - `ExportAsMarkdown(svc *drive.Service, fileID string, mimeType string) (string, error)`:
    - Google Doc: export as HTML → convert to Markdown
    - Google Slides: export as plain text (text/plain)
    - Google Sheet: export as CSV (no conversion, returned as-is)
- [ ] Configuration of html-to-markdown converter:
  - Preserve headings, bold, italic, links, lists, code blocks
  - Strip scripts, styles, and non-content elements
- [ ] Tests:
  - HTML → Markdown with sample HTML (headings, paragraphs, lists, bold/italic, links)
  - Verify Slides plain text export path
  - Verify Sheet CSV pass-through

## Files to Create

- `internal/output/convert.go`
- `internal/output/convert_test.go`

## Notes

- This is used by the cache storage (task 12) to generate the markdown body
- Quality is "best effort" — the primary export (docx/csv/pptx) remains authoritative
- The HTML export uses `text/html` MIME type in the Drive API export
