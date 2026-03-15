# Task 08: Search command

**Phase**: 2 — Commands
**Depends on**: 03, 05, 06
**Blocks**: 12, 14

## Description

Implement the `search` cobra command with JSON and markdown output support and progress indication.

## Acceptance Criteria

- [ ] `cmd/gdrive-cli/search.go`:
  - `gdrive-cli search <query> [--count N]`
  - `query` as positional argument (required)
  - `--count` / `-n` flag, default 20
  - Authenticates, calls `api.SearchFiles()`, formats output
- [ ] JSON output (default): `{"query": "...", "count": N, "results": [{"name": "...", "type": "...", "url": "...", "modified": "..."}]}`
- [ ] Markdown output (`--format markdown`): table with columns Name, Type, URL, Modified
- [ ] Progress: spinner while searching, disabled in non-TTY
- [ ] Command tests with mocked API:
  - Returns results (verify JSON structure)
  - Empty results
  - Error handling (auth failure, API error)

## Files to Create

- `cmd/gdrive-cli/search.go`
- `cmd/gdrive-cli/search_test.go`

## Reference

Port from `src/gdrive_cli/commands/search.py` and `tests/test_commands.py::TestSearchCommand`.
