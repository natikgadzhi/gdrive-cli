# Task 19: Improve Google Slides Support

**Phase**: 2 — Enhancement
**Depends on**: 09, 13

## Context

Basic Slides support already exists: fetch to `.pptx`, search includes Slides, URL parsing works, and caching is in place. However, the Markdown export is a plain-text passthrough (`text/plain` from the Drive API) with no structure — slide boundaries, titles, and speaker notes are lost.

## Changes Required

### 1. Structured Markdown Export

The current `ExportAsMarkdown()` in `internal/output/convert.go` exports Slides as `text/plain` and returns it as-is. Improve this to produce structured Markdown:

- Add slide boundary markers (e.g., `## Slide 1`, `## Slide 2`)
- Use the Google Slides API (`presentations.get`) to extract slide titles, speaker notes, and text content per slide
- Fall back to the current plain-text export if the Slides API is unavailable or errors out

### 2. Speaker Notes

- Extract speaker notes from each slide via the Slides API
- Include them in the Markdown export as blockquotes or a fenced section below each slide's content

### 3. Cache Format

- Update `internal/cache/store.go` to store slide count in YAML frontmatter for cached Slides
- Ensure cached Markdown uses the new structured format

### 4. Additional Export Formats

- Add `.pdf` export support for Slides (Google Drive API supports `application/pdf`)
- Update `internal/formatting/mime.go` with the new format mapping

### 5. Tests

- Unit tests for the new structured Markdown conversion
- Unit tests for speaker notes extraction
- Integration test for Slides → structured Markdown export
- Test fallback behavior when Slides API is unavailable

## Acceptance Criteria

- [ ] `gdrive-cli fetch <slides-url> -o md` produces Markdown with slide boundaries and titles
- [ ] Speaker notes are included in Markdown export when present
- [ ] `.pdf` export works for Slides
- [ ] Cache stores slide count in frontmatter
- [ ] Fallback to plain-text export works if Slides API errors
- [ ] All existing Slides tests continue to pass
- [ ] New tests cover structured Markdown output and speaker notes
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` all clean

## Files to Modify

- `internal/output/convert.go` — structured Markdown conversion
- `internal/formatting/mime.go` — add PDF export mapping
- `internal/cache/store.go` — slide count in frontmatter
- `internal/output/convert_test.go` — new tests
- `cmd/gdrive-cli/fetch_test.go` — new tests
- `cmd/gdrive-cli/integration_test.go` — new integration tests

## Notes

- The Google Slides API (`slides.googleapis.com`) is separate from the Drive API. This will require adding the `presentations.get` endpoint and possibly a new OAuth scope (`https://www.googleapis.com/auth/presentations.readonly`).
- Plain-text export via the Drive API is a fallback — the Slides API gives much richer data but requires additional permissions.
