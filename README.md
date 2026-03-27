# gdrive-cli

A command-line tool to search and download Google Docs, Sheets, and Slides via the Google Drive API. Output is JSON or table (auto-detected based on TTY); debug logs go to stderr.

## Setup

### 1. Google Cloud credentials

Create an OAuth 2.0 Client ID of type **Desktop app** at [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials). Download the JSON and save it to:

```
~/.config/gdrive-cli/credentials.json
```

Override the config directory with `GDRIVE_CONFIG_DIR`:

```sh
export GDRIVE_CONFIG_DIR=/path/to/config
```

The tool uses the `drive.readonly` scope only.

### 2. Install

**Homebrew:**

```sh
brew install natikgadzhi/homebrew-taps/gdrive-cli
```

**From source:**

```sh
go install github.com/natikgadzhi/gdrive-cli/cmd/gdrive-cli@latest
```

**Or build from a local checkout:**

```sh
go build -o bin/gdrive-cli ./cmd/gdrive-cli
```

### 3. Authenticate

```sh
gdrive-cli auth login
```

Opens a browser for Google OAuth consent. On success, saves a token to `~/.config/gdrive-cli/token.json`. Tokens are refreshed automatically on subsequent runs.

---

## Global flags

```
gdrive-cli [--debug] [-o json|table] [--no-cache] <command>
```

| Flag | Description |
|------|-------------|
| `--debug` | Print verbose debug logs to stderr |
| `-o` / `--output` | Output format: `json` or `table` (default: auto-detected; table in TTY, json when piped) |
| `--no-cache` | Skip writing to the derived data directory |
| `-d` / `--derived` | Derived data directory (default: `~/.local/share/lambdal/derived/gdrive-cli`) |

---

## Commands

### `auth login`

```sh
gdrive-cli auth login
```

Runs the OAuth2 installed-app flow. Opens a browser for Google consent and saves credentials locally.

**JSON output:**
```json
{ "status": "ok", "message": "Successfully authenticated with Google Drive." }
```

**Errors:**
- `credentials.json` not found -- prints path and Google Cloud console link
- `credentials.json` is for a Web application client (not Desktop) -- tells you to create a Desktop app client

---

### `auth check`

```sh
gdrive-cli auth check
```

Checks whether stored credentials exist and are valid (or refreshable).

**JSON output (authenticated):**
```json
{ "status": "ok", "message": "Authenticated and credentials are valid." }
```

**Error (not authenticated):**
Prints an error to stderr indicating credentials are missing or invalid.

---

### `search`

```sh
gdrive-cli search <query> [--limit N] [-o json|table]
```

Searches Google Drive for Docs, Sheets, and Slides matching `query`. Matches on both file name and full text content. Results are ordered by `modifiedTime desc`.

| Argument / Option | Default | Description |
|---|---|---|
| `query` | required | Search string |
| `--limit` / `-n` | `20` | Max results to return |
| `-o` / `--output` | auto | Output format: `json` or `table` |

**JSON output:**
```json
{
  "query": "budget 2025",
  "count": 2,
  "results": [
    {
      "name": "Q1 Budget",
      "type": "Google Sheet",
      "url": "https://docs.google.com/spreadsheets/d/...",
      "modified": "2025-03-01T10:00:00.000Z"
    }
  ]
}
```

**Table output** (default in TTY):

Prints an aligned table with columns: NAME, TYPE, MODIFIED, URL.

**Notes:**
- Only returns Docs, Sheets, and Slides -- no other Drive files.
- Single quotes in `query` are escaped for the Drive API query syntax.
- `--limit` maps to `pageSize` in the Drive API; the API may return fewer results than requested.

---

### `fetch`

```sh
gdrive-cli fetch <url> [--export FORMAT] [--dest PATH] [--no-comments] [-o json|table]
```

Downloads a Google Doc, Sheet, or Slides file and saves it locally. Also supports non-native files uploaded to Google Drive (`.docx`, `.xlsx`, `.pptx`, `.pdf`, etc.) via direct binary download.

| Argument / Option | Default | Description |
|---|---|---|
| `url` | required | Full Google Docs/Sheets/Slides URL |
| `--export` / `-e` | type default | Export format: `docx`, `md`, `csv`, `pptx` (depends on document type; ignored for non-native files) |
| `--dest` / `-f` | auto-generated | Output file path (or directory; auto-generates filename if directory) |
| `--no-comments` | `false` | Skip fetching document comments |
| `-o` / `--output` | auto | Output format: `json` or `table` |

**Default export formats (native Google Workspace files):**

| Source type | Default | Other |
|---|---|---|
| Google Doc | `.docx` | `.md` |
| Google Sheet | `.csv` | |
| Google Slides | `.pptx` | `.md` |

**Non-native files** (uploaded `.docx`, `.xlsx`, `.pdf`, etc.) are downloaded directly in their original format. The `--export` flag is ignored for these files.

**Auto-generated filename:** `<sanitized-title><extension>` in the output directory. Characters `/ \ : * ? " < > |` in the title are replaced with `_`.

**Accepted URL formats:**
```
https://docs.google.com/document/d/<ID>/...
https://docs.google.com/spreadsheets/d/<ID>/...
https://docs.google.com/presentation/d/<ID>/...
```

URL-encoded characters are decoded before ID extraction.

**JSON output:**
```json
{
  "status": "ok",
  "file_id": "1aBcDe...",
  "name": "Q1 Budget",
  "type": "Google Sheet",
  "saved_to": "./Q1_Budget.csv",
  "cached_to": "~/.local/share/lambdal/derived/gdrive-cli/q1-budget-1aBcDe.md"
}
```

**Fallback behavior:**

When exporting native Google Workspace files, the tool handles common API errors gracefully:
- **View-only files (cannotExportFile)**: Falls back to binary download. If the file owner has disabled downloads entirely, a helpful error is shown with the file name, type, and URL.
- **Oversized files (exportSizeLimitExceeded)**: Falls back to plain text export (which has a higher size limit), then to binary download if that also fails.

**Comments:**

By default, `fetch` also retrieves all comment threads from the document and saves them as a companion Markdown file in the derived data directory:

```
~/.local/share/lambdal/derived/gdrive-cli/<slug>.comments.md
```

The comments file includes:
- All open and resolved comment threads, grouped by status
- Author names and dates for each comment and reply
- Quoted text showing what part of the document each comment refers to
- Resolve/reopen actions attributed to the acting user

Use `--no-comments` to skip fetching comments. Comments are also skipped when `--no-cache` is set.

Note: the Google Drive API does not expose comment author email addresses -- only display names are available.

**Errors:**
- Unrecognized URL format -- lists supported formats
- Output directory is created automatically if it does not exist

---

### `version`

```sh
gdrive-cli version
```

Prints version, commit, and build date information as JSON.

**Output:**
```json
{ "version": "0.1.0", "commit": "abc1234", "date": "2025-01-01T00:00:00Z" }
```

Also available as `gdrive-cli --version`.

---

## Config reference

| Path | Purpose |
|---|---|
| `~/.config/gdrive-cli/credentials.json` | OAuth client credentials (you provide) |
| `~/.config/gdrive-cli/token.json` | OAuth token (written after login) |
| `$GDRIVE_CONFIG_DIR` | Override for the config directory |
| `~/.local/share/lambdal/derived/gdrive-cli/` | Derived Markdown exports of fetched documents |
| `$GDRIVE_CLI_DERIVED_DIR` | Override for the derived directory |
| `$LAMBDAL_DERIVED_DIR` | Override for the base derived directory (tool name appended) |
