# Task 05b: Add trashed=false filter to Drive search query

**Phase**: Follow-up (from PR #8 review)
**Depends on**: 05 (Drive API wrapper)
**Blocks**: none
**Priority**: low

## Description

The Drive API search query in `internal/api/search.go` does not explicitly filter out trashed files. While the Drive API defaults to excluding trashed files in most cases, being explicit with `trashed = false` prevents edge cases where trashed files could appear in results.

## Origin

Reviewer feedback on PR #8 (Drive API wrapper). Non-blocking, deferred.

## Acceptance Criteria

- [ ] `SearchFiles` query includes `and trashed = false` clause
- [ ] Existing search tests updated to verify the trashed filter is present
- [ ] `go test ./internal/api/...` passes

## Files to Modify

- `internal/api/search.go`
- `internal/api/api_test.go`

## Status
backlog
