# Task 02: Config package and testing infrastructure

**Phase**: 0 — Bootstrap
**Status**: done
**Assignee**: builder-02
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/3 (merged)
**Depends on**: 01
**Blocks**: 03, 04, 05

## Description

Implement the config package (paths, env var overrides, debug logging) and establish the testing conventions for the project.

## Acceptance Criteria

- [ ] `internal/config/config.go`:
  - `ConfigDir() string` — returns `~/.config/gdrive-cli`, overridable via `GDRIVE_CONFIG_DIR`
  - `CredentialsFile() string` — `ConfigDir()/credentials.json`
  - `TokenFile() string` — `ConfigDir()/token.json`
  - `CacheDir() string` — `~/.local/share/gdrive-cli/cache`, overridable via `GDRIVE_CACHE_DIR`
  - `DebugLog(msg string, args ...any)` — prints to stderr when debug mode is on
  - `SetDebug(enabled bool)` — toggle debug mode
- [ ] `internal/config/config_test.go`:
  - Test default paths
  - Test env var overrides (set env, check result, unset)
  - Test debug logging (capture stderr, verify output)
- [ ] Integration test build tag convention established: `//go:build integration`
- [ ] `go test ./...` passes

## Files to Create

- `internal/config/config.go`
- `internal/config/config_test.go`

## Notes

- Use `os.UserHomeDir()` for `~` expansion
- Use `t.Setenv()` in tests for env var overrides (auto-cleaned up)
