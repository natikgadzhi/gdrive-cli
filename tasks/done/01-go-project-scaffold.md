# Task 01: Initialize Go project structure and dependencies

**Phase**: 0 — Bootstrap
**Status**: done
**Assignee**: builder-01
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/1 (merged)
**Depends on**: none
**Blocks**: 02

## Description

Create the Go module, directory structure, build tooling, and minimal cobra root command that compiles and runs.

## Acceptance Criteria

- [ ] `go.mod` with module `github.com/natikgadzhi/gdrive-cli`
- [ ] Directory tree created:
  - `cmd/gdrive-cli/` (main.go, root.go)
  - `internal/{auth,api,config,formatting,output,cache,ratelimit,progress}/`
- [ ] Dependencies added: cobra, golang.org/x/oauth2, google.golang.org/api, golang.org/x/time, gopkg.in/yaml.v3, briandowns/spinner, html-to-markdown
- [ ] `cmd/gdrive-cli/main.go` — calls `root.Execute()`
- [ ] `cmd/gdrive-cli/root.go` — cobra root command with `--debug` persistent flag and `--format` persistent flag (json|markdown, default json)
- [ ] `Makefile` with targets: `build`, `test`, `vet`, `lint`, `run`
- [ ] `.goreleaser.yml` skeleton (can be minimal, completed in task 16)
- [ ] `.gitignore` updated for Go binaries, vendor/
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean

## Files to Create

- `go.mod`, `go.sum`
- `cmd/gdrive-cli/main.go`
- `cmd/gdrive-cli/root.go`
- `Makefile` (replace existing Python one)
- `.goreleaser.yml`
- `.gitignore` (update)

## Notes

- Keep root command minimal — just the flag definitions and basic Execute
- Version variable should be declared but populated later via ldflags
