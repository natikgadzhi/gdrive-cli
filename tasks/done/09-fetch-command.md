# Task 09: Fetch command

**Phase**: 2 — Commands
**Depends on**: 03, 05, 06
**Blocks**: 12, 13, 14

## Description

Implement the `fetch` cobra command with file export, auto-generated filenames, and progress indication.

## Acceptance Criteria

- [ ] `cmd/gdrive-cli/fetch.go`:
  - `gdrive-cli fetch <url> [--output FILE] [--dir DIR]`
  - `url` as positional argument (required)
  - `--output` / `-o` flag (explicit output file path)
  - `--dir` / `-d` flag (output directory, default `.`)
  - Flow: parse URL → get metadata → export file → save to disk
- [ ] Export formats: Google Doc → .docx, Sheet → .csv, Slides → .pptx
- [ ] Auto-generated filename: `<sanitized-title><extension>` in output dir
- [ ] Creates output directory if it doesn't exist
- [ ] JSON output: `{"status": "ok", "file_id": "...", "name": "...", "type": "...", "saved_to": "..."}`
- [ ] Markdown output (`--format markdown`): frontmatter + status info
- [ ] Progress: spinner while downloading, byte counter
- [ ] Error handling:
  - Invalid URL → list supported formats
  - Unsupported MIME type → list supported types
- [ ] Command tests with mocked API:
  - Fetch Google Doc, Sheet, Slides
  - Invalid URL
  - Custom output path and directory

## Files to Create

- `cmd/gdrive-cli/fetch.go`
- `cmd/gdrive-cli/fetch_test.go`

## Reference

Port from `src/gdrive_cli/commands/fetch.py` and `tests/test_commands.py::TestFetchCommand`.
