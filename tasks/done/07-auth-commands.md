# Task 07: Auth commands (login + status)

**Phase**: 2 — Commands
**Status**: done
**Assignee**: builder-08
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/7 (merged)
**Depends on**: 03, 04
**Blocks**: 14

## Description

Implement the `auth login` and `auth status` cobra subcommands, wiring up the auth package to the CLI.

## Acceptance Criteria

- [ ] `cmd/gdrive-cli/auth.go`:
  - `auth` command group (cobra parent command)
- [ ] `cmd/gdrive-cli/auth_login.go`:
  - Calls `auth.Login(config.ConfigDir())`
  - On success: prints `{"status": "ok", "message": "Successfully authenticated with Google Drive."}`
  - On error: prints `{"status": "error", "message": "..."}` with descriptive message
- [ ] `cmd/gdrive-cli/auth_status.go`:
  - Calls `auth.GetCredentials(config.ConfigDir())`
  - Authenticated: `{"status": "ok", "message": "Authenticated and credentials are valid."}`
  - Not authenticated: `{"status": "error", "message": "Not authenticated. Run 'gdrive-cli auth login' first."}`
- [ ] All output is JSON to stdout, errors to stderr
- [ ] Command tests verifying JSON output for both success and error paths

## Files to Create

- `cmd/gdrive-cli/auth.go`
- `cmd/gdrive-cli/auth_login.go`
- `cmd/gdrive-cli/auth_status.go`
- `cmd/gdrive-cli/auth_test.go`

## Reference

Port from `src/gdrive_cli/commands/auth.py`. Output JSON must match the existing format.
