# Task 11+12: JSON/Markdown output formatters + Cache storage

**Phase**: 3 — Output & Cache
**Depends on**: 08, 09, 13 (markdown conversion)
**Blocks**: 14 (e2e tests)
**Consolidated from**: tasks 11 and 12

## Description

Implement the `--format` flag, cache layer, and wire them together. The cache IS the markdown output — when fetching a document, the fetch command writes both the native export (docx/csv/pptx) and a cached markdown version with YAML frontmatter metadata.

## Acceptance Criteria

### Format flag
- [ ] `internal/output/format.go`:
  - `Format` type: `FormatJSON`, `FormatMarkdown`
  - `ParseFormat(s string) (Format, error)` — "json" | "markdown"
- [ ] `--format` persistent flag on root command, default "json"

### Cache layer
- [ ] `internal/cache/entry.go`:
  - `CacheEntry` struct: Tool, Name, Slug, Type, FileID, SourceURL, CreatedAt, UpdatedAt, RequestedBy, Body
- [ ] `internal/cache/slug.go`:
  - `GenerateSlug(name string, fileID string) string` — lowercase sanitized name + first 8 chars of file ID
- [ ] `internal/cache/store.go`:
  - `Store(cacheDir string, entry CacheEntry) (string, error)` — writes file with YAML frontmatter, returns path
  - `Load(cacheDir string, slug string) (*CacheEntry, error)` — reads and parses cached file
  - `Exists(cacheDir string, slug string) bool`
  - `List(cacheDir string) ([]CacheEntry, error)` — list all cached entries (frontmatter only)
  - Directory structure: `documents/<slug>.md`, `spreadsheets/<slug>.csv`, `presentations/<slug>.md`
- [ ] YAML frontmatter serialization with `gopkg.in/yaml.v3`

### Wiring into commands
- [ ] Fetch command: after native export, also write to cache using `ExportAsMarkdown` + cache.Store
  - JSON output unchanged by default
  - With `--format markdown`: print the cached markdown to stdout instead of JSON
- [ ] Search command: `--format markdown` renders results as a markdown table
- [ ] Auth commands: always JSON (ignore format flag)

### Tests
- [ ] Slug generation (various names, special chars)
- [ ] Store + Load round-trip
- [ ] Exists check, List cached entries
- [ ] Directory auto-creation
- [ ] Format flag parsing
- [ ] Frontmatter round-trip (write then parse)

## Files to Create/Modify

- `internal/output/format.go` (new)
- `internal/cache/entry.go` (new)
- `internal/cache/slug.go` (new)
- `internal/cache/store.go` (new)
- `internal/cache/cache_test.go` (new)
- `cmd/gdrive-cli/root.go` (add --format flag)
- `cmd/gdrive-cli/fetch.go` (add cache write + markdown output mode)
- `cmd/gdrive-cli/search.go` (add markdown table output mode)

## Notes

- Cache location: `~/.local/share/gdrive-cli/cache/` (from `config.CacheDir()`)
- If cache entry already exists for the same slug, update it (overwrite with new UpdatedAt)
- Spreadsheets cached as .csv, docs/slides as .md with frontmatter
- The `ExportAsMarkdown` function (already in internal/output/convert.go) feeds into the cache
