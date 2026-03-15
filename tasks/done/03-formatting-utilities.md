# Task 03: Formatting and URL parsing utilities

**Phase**: 1 — Core
**Status**: done
**Assignee**: builder-04
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/4 (merged)
**Depends on**: 02
**Blocks**: 07, 08, 09

## Description

Port the Python formatting module to Go: URL parsing, MIME type mappings, filename sanitization, and query escaping.

## Acceptance Criteria

- [ ] `internal/formatting/url.go`:
  - `ParseGoogleURL(url string) (fileID string, err error)` — regex extraction from document/spreadsheets/presentation URLs
  - Handles URL-encoded characters (decode before extraction)
  - Returns descriptive error with supported URL formats on failure
- [ ] `internal/formatting/mime.go`:
  - `ExportMIME` map: Google Doc → docx, Sheet → csv, Slides → pptx MIME types
  - `ExportExtension` map: Google MIME → file extension
  - `TypeLabel` map: Google MIME → human label ("Google Doc", "Google Sheet", "Google Slides")
  - `MarkdownExportMIME` map: Google Doc → text/html, Sheet → text/csv, Slides → text/plain
- [ ] `internal/formatting/filename.go`:
  - `SanitizeFilename(name string) string` — replaces `/ \ : * ? " < > |` with `_`
- [ ] `internal/formatting/query.go`:
  - `EscapeQuery(q string) string` — escapes single quotes for Drive API query syntax
- [ ] Full table-driven tests for all functions covering:
  - All three URL types (doc, sheet, slides)
  - URL-encoded URLs
  - Invalid URLs
  - Filenames with special characters
  - Queries with single quotes

## Files to Create

- `internal/formatting/url.go`
- `internal/formatting/url_test.go`
- `internal/formatting/mime.go`
- `internal/formatting/mime_test.go`
- `internal/formatting/filename.go`
- `internal/formatting/filename_test.go`
- `internal/formatting/query.go`
- `internal/formatting/query_test.go`

## Reference

Port from `src/gdrive_cli/formatting.py` and `tests/test_formatting.py`.
