# Task 04: OAuth2 authentication

**Phase**: 1 — Core
**Status**: done
**Assignee**: builder-05
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/6 (merged)
**Depends on**: 02
**Blocks**: 05, 07

## Description

Implement the OAuth2 installed-app flow: loading credentials, running the local server for the OAuth redirect, exchanging codes for tokens, saving/loading/refreshing tokens.

## Acceptance Criteria

- [ ] `internal/auth/auth.go`:
  - `GetCredentials(configDir string) (*oauth2.Token, *oauth2.Config, error)` — loads token.json, refreshes if expired, returns error if not authenticated
  - `IsAuthenticated(configDir string) bool` — quick check without refresh
- [ ] `internal/auth/login.go`:
  - `Login(configDir string) error` — full OAuth2 installed-app flow:
    - Read credentials.json, validate it's a Desktop app client (not Web)
    - Start local HTTP server on port 8085
    - Open browser to consent URL (use `os/exec` with `open`/`xdg-open`)
    - Handle redirect, extract auth code
    - Exchange code for token
    - Save token to token.json
  - Error cases: missing credentials.json, wrong client type, port in use, timeout
- [ ] `internal/auth/token.go`:
  - `SaveToken(token *oauth2.Token, path string) error`
  - `LoadToken(path string) (*oauth2.Token, error)`
  - `LoadOAuthConfig(credentialsPath string) (*oauth2.Config, error)` — parses credentials.json
- [ ] Tests:
  - Token save/load round-trip (use temp directory)
  - Credentials.json parsing (valid desktop, invalid web, missing file)
  - Login flow with httptest mock (simulate token exchange)

## Files to Create

- `internal/auth/auth.go`
- `internal/auth/login.go`
- `internal/auth/token.go`
- `internal/auth/auth_test.go`

## Reference

Port from `src/gdrive_cli/auth.py`. Key behaviors:
- Scope: `https://www.googleapis.com/auth/drive.readonly`
- Port 8085 for local redirect
- Max 20 requests handled before timeout
- credentials.json has `installed.client_id` for Desktop apps vs `web.client_id` for Web apps

## Notes

- **Breaking change**: Go's `oauth2.Token` JSON format differs from Python's `google.oauth2.credentials.Credentials`. Users will need to re-authenticate after upgrade.
