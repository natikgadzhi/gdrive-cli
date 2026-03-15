# gdrive-cli: Go Rewrite Plan

## Overview

Rewrite gdrive-cli from Python to Go, preserving all existing functionality while adding: rate limiting with backoff, progress indicators, JSON/Markdown output formats, markdown caching with frontmatter, and Homebrew tap distribution.

## Design Decisions

### CLI Framework: cobra
Maps directly to existing Click-based command hierarchy (`gdrive-cli auth login`, `search`, `fetch`). Provides subcommand nesting, persistent flags, shell completion. De facto Go CLI standard (used by kubectl, gh, docker).

### Google API Client: official Go SDK
`google.golang.org/api/drive/v3` + `golang.org/x/oauth2`. Typed structs, pagination, clean OAuth2 integration. No reason to use raw HTTP.

### Project Layout
```
cmd/gdrive-cli/          # Entrypoint + cobra commands
  main.go
  root.go                # Root command, --debug, --format flags
  auth.go                # auth command group
  auth_login.go          # auth login
  auth_status.go         # auth status
  search.go              # search command
  fetch.go               # fetch command
  version.go             # version command
internal/
  auth/                  # OAuth2 flow, token storage, credential loading
  api/                   # Drive API wrapper (search, export, metadata)
  config/                # Config paths, env vars, debug logging
  formatting/            # URL parsing, MIME maps, filename sanitization
  output/                # JSON + Markdown formatters
  cache/                 # Markdown cache with YAML frontmatter
  ratelimit/             # Rate limiter + 429 backoff
  progress/              # Spinner + byte counter on stderr
```

### OAuth2 Authentication
Local HTTP server on port 8085 (matching Python behavior). Uses `golang.org/x/oauth2` for token exchange/refresh. **Breaking change**: new token.json format — users re-authenticate once after upgrade.

### Output Format Flag
`--format`/`-f` globally (not `-o`, which `fetch` uses for `--output` file path). Values: `json` (default), `markdown`. Tabular spreadsheet data uses `csv` instead of `markdown`.

### Rate Limiting
`golang.org/x/time/rate` token bucket + custom exponential backoff with jitter on HTTP 429. Reads `Retry-After` header. Wraps `http.RoundTripper` so all API calls are rate-limited transparently.

**Key fix**: On 429 during search, return partial results collected so far instead of returning nothing.

### Progress Indicators
`briandowns/spinner` on stderr for operations in progress. Byte counter for downloads. Auto-disabled when stderr is not a TTY (piped output).

### Markdown Cache
Location: `~/.local/share/gdrive-cli/cache/` (XDG-ish). Override: `GDRIVE_CACHE_DIR` env var.

```
cache/
  documents/<slug>.md
  spreadsheets/<slug>.csv
  presentations/<slug>.md
```

YAML frontmatter:
```yaml
---
tool: gdrive-cli
name: "Q1 Budget Report"
slug: "q1-budget-report-1aBcDeFg"
type: "Google Doc"
file_id: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ"
source_url: "https://docs.google.com/document/d/1aBcDeFg/edit"
created_at: "2026-03-14T10:00:00Z"
updated_at: "2026-03-14T10:00:00Z"
requested_by: "gdrive-cli fetch"
---
```

### Markdown Conversion
Google HTML export (`text/html`) → Markdown via `github.com/JohannesKaufmann/html-to-markdown`. For Slides: plain text export. For Sheets: CSV (no markdown conversion).

### Build & Release
GoReleaser for cross-compilation (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64). Homebrew tap formula published to `github.com/natikgadzhi/taps`. Version embedded via ldflags.

### Testing Strategy
- **Unit tests**: Table-driven for all pure functions
- **API mocking**: `net/http/httptest` fake Drive API servers, injected via custom `http.Client`
- **Command tests**: cobra `Execute()` with captured stdout/stderr
- **E2E tests**: Build-tagged `//go:build integration`, use real OAuth tokens from env vars (`GDRIVE_TOKEN_JSON` or `GDRIVE_CONFIG_DIR`)

## Go Dependencies

```
github.com/spf13/cobra
golang.org/x/oauth2
golang.org/x/oauth2/google
google.golang.org/api/drive/v3
google.golang.org/api/option
golang.org/x/time/rate
gopkg.in/yaml.v3
github.com/briandowns/spinner
github.com/JohannesKaufmann/html-to-markdown
```

## Phases

### Phase 0: Bootstrap
Go module init, directory structure, Makefile, minimal cobra root command, .goreleaser skeleton.

### Phase 1: Core
Config paths + env vars, OAuth2 auth (login/token/refresh), Drive API wrapper (search/export/metadata), URL parsing + formatting utilities.

### Phase 2: Commands + Features
Cobra command tree, auth login/status, search, fetch, rate limiting + 429 handling, progress indicators.

### Phase 3: Output & Cache
JSON/Markdown output formatters, markdown cache storage with frontmatter, HTML→Markdown doc conversion.

### Phase 4: Integration & Release
E2E tests, error handling audit, GoReleaser + Homebrew tap, documentation update, Python cleanup.

## Dependency Graph

```
Phase 0: [0.1] → [0.2]
Phase 1: [0.2] → [1.1] [1.2] [1.3] (1.2 needs 1.1, 1.3 needs 1.2)
Phase 2: [1.*] → [2.1] [2.2] [2.3] [2.4] [2.5] (2.1-2.3 parallelizable, 2.4 needs 1.3, 2.5 independent)
Phase 3: [2.*] → [3.1] [3.2] [3.3] (all parallelizable)
Phase 4: [3.*] → [4.1] → [4.2] → [4.3] → [4.4]
```

## Risks

1. **OAuth token format break** — users must re-auth once after upgrade. Document clearly.
2. **Markdown conversion fidelity** — HTML→MD is best-effort. Primary export (docx/csv/pptx) remains authoritative.
3. **429 partial results** — requires paginated search collection with per-page error handling.
4. **Progress on stderr** — must detect non-TTY and disable to avoid corrupting piped output.
