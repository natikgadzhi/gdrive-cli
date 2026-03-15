# Task 17: Documentation update and Python cleanup

**Phase**: 4 — Integration & Release
**Depends on**: 16
**Blocks**: none (final task)

## Description

Update all documentation for the Go version, remove Python source code and tooling, and do a final verification pass.

## Acceptance Criteria

- [ ] `README.md` updated:
  - Installation via Homebrew: `brew tap natikgadzhi/taps && brew install gdrive-cli`
  - Installation from source: `go install github.com/natikgadzhi/gdrive-cli/cmd/gdrive-cli@latest`
  - Updated command documentation (add `--format` flag, `version` command)
  - Remove Python-specific instructions (pip, uv)
- [ ] `CLAUDE.md` updated:
  - Build commands: `go build ./...`, `go vet ./...`, `go test ./...`
  - Integration test command: `go test -tags integration ./tests/integration/...`
- [ ] `--help` text for all commands matches README documentation
- [ ] Python files removed:
  - `src/` directory (entire)
  - `tests/` directory (Python tests — Go tests live elsewhere)
  - `pyproject.toml`
  - `uv.lock`
  - `.python-version`
  - Any `__pycache__` directories
- [ ] Final checks:
  - `go build ./...` — clean
  - `go vet ./...` — clean
  - `go test ./...` — all pass
  - `goreleaser build --snapshot --clean` — builds all targets
  - Binary runs: `./dist/gdrive-cli --help`

## Files to Remove

- `src/gdrive_cli/` (entire directory tree)
- `tests/__init__.py`, `tests/test_commands.py`, `tests/test_formatting.py`
- `pyproject.toml`
- `uv.lock`
- `.python-version` (if exists)

## Files to Modify

- `README.md`
- `CLAUDE.md`
- `Makefile` (Go targets only)
- `.gitignore` (remove Python patterns, keep Go patterns)

## Notes

- Do NOT remove `PROJECT_PROMPT.md` — that's the user's spec
- Keep `docs/PLAN.md` for reference
- This is the final task — after this, the repo is fully Go
