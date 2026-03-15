# Task 10: Rate limiting and HTTP 429 handling

**Phase**: 2 — Commands
**Depends on**: 05
**Blocks**: 14

## Description

Implement rate limiting for Drive API requests and fix the 429 bug: when rate-limited during search, return all results collected so far instead of nothing.

## Acceptance Criteria

- [ ] `internal/ratelimit/transport.go`:
  - `NewRateLimitedTransport(base http.RoundTripper, rps float64) http.RoundTripper`
  - Wraps any `http.RoundTripper` with `golang.org/x/time/rate` token bucket
  - Default: 10 requests/second (configurable)
- [ ] `internal/ratelimit/backoff.go`:
  - On HTTP 429: exponential backoff with jitter
  - Reads `Retry-After` header if present
  - Max retries: 5
  - Base delay: 1s, max delay: 60s
- [ ] `internal/api/client.go` updated:
  - `NewDriveService()` wraps the HTTP client transport with rate limiter
- [ ] `internal/api/search.go` updated:
  - Paginated search: collect results page by page
  - On 429 after retries exhausted: return partial results + warning (not error)
  - Warning goes to stderr (or as a field in the response)
- [ ] Tests:
  - Rate limiter enforces delay between requests
  - 429 with Retry-After triggers correct backoff
  - Partial results returned on unrecoverable 429
  - Successful retry after transient 429

## Files to Create/Modify

- `internal/ratelimit/transport.go`
- `internal/ratelimit/backoff.go`
- `internal/ratelimit/ratelimit_test.go`
- Modify: `internal/api/client.go`, `internal/api/search.go`

## Notes

- The rate limiter wraps `http.RoundTripper` so it's transparent to the Google API client
- Search pagination: use the `NextPageToken` from Drive API list response
- Partial result response should include `"warning": "Rate limited. Returning N of expected M results."`
