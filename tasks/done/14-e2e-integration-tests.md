# Task 14: End-to-end integration tests

**Phase**: 4 — Integration & Release
**Depends on**: 07, 08, 09, 10, 11, 12, 13
**Blocks**: 15

## Description

Write end-to-end tests that exercise the full CLI against the real Google Drive API using user-provided OAuth tokens.

## Acceptance Criteria

- [ ] `tests/integration/setup_test.go`:
  - Build tag: `//go:build integration`
  - `TestMain()` — loads credentials from `GDRIVE_TOKEN_JSON` (JSON string) or `GDRIVE_CONFIG_DIR` (directory path)
  - Skip all tests if neither env var is set
  - Create temp config dir, write token.json
- [ ] `tests/integration/auth_test.go`:
  - Test `auth status` returns ok when token is valid
- [ ] `tests/integration/search_test.go`:
  - Test `search "test"` returns results (at least 1 result expected)
  - Test `search "test" --count 1` returns at most 1 result
  - Test `search "test" --format markdown` returns markdown table
  - Test search with non-existent query returns empty results
- [ ] `tests/integration/fetch_test.go`:
  - Test fetch a known Google Doc (env var `GDRIVE_TEST_DOC_URL`)
  - Test fetch a known Google Sheet (env var `GDRIVE_TEST_SHEET_URL`)
  - Verify files are saved to disk with correct extension
  - Verify JSON output contains expected fields
  - Verify cache entry is created
  - Clean up downloaded files after test
- [ ] `tests/integration/ratelimit_test.go`:
  - Test rapid successive search queries don't fail
  - Verify rate limiter is active (requests are spaced)
- [ ] All tests use `t.TempDir()` for output files
- [ ] Tests are skippable in CI without credentials

## Files to Create

- `tests/integration/setup_test.go`
- `tests/integration/auth_test.go`
- `tests/integration/search_test.go`
- `tests/integration/fetch_test.go`
- `tests/integration/ratelimit_test.go`

## Environment Variables for Tests

- `GDRIVE_TOKEN_JSON` — full token.json content as string
- `GDRIVE_CONFIG_DIR` — path to directory containing credentials.json + token.json
- `GDRIVE_TEST_DOC_URL` — URL of a Google Doc to use in fetch tests
- `GDRIVE_TEST_SHEET_URL` — URL of a Google Sheet to use in fetch tests

## Notes

- Run with: `go test -tags integration ./tests/integration/...`
- These tests make real API calls — they may be slow and rate-limited
- Keep test data read-only (use existing docs, don't create/modify)
