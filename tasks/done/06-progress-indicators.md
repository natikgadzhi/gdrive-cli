# Task 06: Progress indicators

**Phase**: 2 — Commands
**Status**: done
**Assignee**: builder-07
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/5 (merged)
**Depends on**: 02
**Blocks**: 08, 09

## Description

Implement progress indicators that write to stderr: a spinner for long operations and a counter for data fetched. These integrate into search and fetch commands.

## Acceptance Criteria

- [ ] `internal/progress/spinner.go`:
  - `NewSpinner(message string) *Spinner` — creates a spinner that writes to stderr
  - `Start()`, `Stop()`, `UpdateMessage(msg string)`
  - Uses `briandowns/spinner` under the hood
  - Auto-disabled when stderr is not a TTY (`os.Stderr.Fd()` + `term.IsTerminal()`)
- [ ] `internal/progress/counter.go`:
  - `NewCounter(label string) *Counter` — "Searching... 5 results" or "Downloading... 1.2 MB"
  - `Increment(n int)`, `SetBytes(n int64)`
  - Writes to stderr, respects TTY detection
- [ ] Tests:
  - Non-TTY mode produces no output
  - Counter formatting (bytes → human-readable)

## Files to Create

- `internal/progress/spinner.go`
- `internal/progress/counter.go`
- `internal/progress/progress_test.go`

## Notes

- Keep it simple — no bubbletea or complex TUI
- The spinner/counter should be usable as: `s := progress.NewSpinner("Searching..."); s.Start(); defer s.Stop()`
- When integrated into search: update counter as each page of results arrives
- When integrated into fetch: update with bytes downloaded
