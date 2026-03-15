# Task 02b: GitHub Actions CI workflow for Go

## Phase
Phase 0 — Bootstrap

## Depends on
- 01 (Go project scaffold must exist)

## Blocks
- None directly, but should be in place before Phase 1 PRs

## Description
Add a GitHub Actions workflow that runs on all pull requests and pushes to `main`.

## Requirements

### Workflow file: `.github/workflows/ci.yml`
- **Trigger**: `push` to `main`, `pull_request` to `main`
- **Go version**: Use matrix with latest stable Go (1.24.x)
- **Steps**:
  1. Checkout code
  2. Set up Go (with caching)
  3. `go build ./...`
  4. `go vet ./...`
  5. `go test ./...` (unit tests only, no integration tag)
- **OS**: `ubuntu-latest`

### Optional nice-to-haves (only if trivial)
- Cache `~/go/pkg/mod` via `actions/setup-go` built-in caching
- Separate job for `golangci-lint` if the linter is already configured in task 01

## Acceptance Criteria
- [ ] `.github/workflows/ci.yml` exists and is valid YAML
- [ ] Workflow triggers on PRs to main and pushes to main
- [ ] Runs `go build`, `go vet`, `go test` successfully
- [ ] Does NOT run integration tests (no `-tags integration`)

## Status
done (builder-03)
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/2 (merged)
