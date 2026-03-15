# Task 15: Error handling audit and edge cases

**Phase**: 4 — Integration & Release
**Depends on**: 14
**Blocks**: 16

## Description

Review all error paths across the codebase. Ensure consistent error output, comprehensive debug logging, and proper stdout/stderr separation.

## Acceptance Criteria

- [ ] All user-facing errors output as JSON: `{"status": "error", "message": "..."}`
- [ ] All errors go to stdout (as JSON), debug info to stderr
- [ ] Error messages are descriptive and actionable:
  - Missing credentials.json → print path and Google Cloud console link
  - Invalid URL → list supported URL formats
  - Unsupported MIME type → list supported types
  - Network error → suggest checking connectivity
  - Permission denied → suggest checking Drive sharing settings
- [ ] Debug logging (`--debug`) covers:
  - Config paths being used
  - API requests being made (URL, method)
  - Token refresh events
  - Rate limiter wait times
  - Cache read/write operations
- [ ] Exit codes: 0 for success, 1 for errors
- [ ] No panics — all errors handled gracefully
- [ ] Tests for each error scenario

## Files to Modify

- All `cmd/gdrive-cli/*.go` command files
- `internal/auth/*.go`
- `internal/api/*.go`

## Notes

- Review against Python implementation for any error cases we might have missed
- Grep for `TODO`, `FIXME`, unhandled errors
