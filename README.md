# gdrive-cli

A command-line tool to search and download Google Docs, Sheets, and Slides via the Google Drive API. All output is JSON to stdout; debug logs go to stderr.

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

```sh
pip install -e .
# or with uv
uv pip install -e .
```

### 3. Authenticate

```sh
gdrive-cli auth login
```

Opens a browser for Google OAuth consent. On success, saves a token to `~/.config/gdrive-cli/token.json`. Tokens are refreshed automatically on subsequent runs.

---

## Global flags

```
gdrive-cli [--debug] <command>
```

| Flag | Description |
|------|-------------|
| `--debug` | Print verbose debug logs to stderr |

---

## Commands

### `auth login`

```sh
gdrive-cli auth login
```

Runs the OAuth2 installed-app flow. Opens a browser; starts a local WSGI server on `localhost:8085` to receive the redirect. Saves credentials to `token.json`.

**Output:**
```json
{ "status": "ok", "message": "Successfully authenticated with Google Drive." }
```

**Errors:**
- `credentials.json` not found → prints path and Google Cloud console link
- `credentials.json` is for a Web application client (not Desktop) → tells you to create a Desktop app client
- Port 8085 already in use → raises error with OS message
- No OAuth callback received after 20 requests → raises error

---

### `auth status`

```sh
gdrive-cli auth status
```

Checks whether stored credentials exist and are valid (or refreshable).

**Output (authenticated):**
```json
{ "status": "ok", "message": "Authenticated and credentials are valid." }
```

**Output (not authenticated):**
```json
{ "status": "error", "message": "Not authenticated. Run `gdrive-cli auth login` first." }
```

---

### `search`

```sh
gdrive-cli search <query> [--count N]
```

Searches Google Drive for Docs, Sheets, and Slides matching `query`. Matches on both file name and full text content. Results are ordered by `modifiedTime desc`.

| Argument / Option | Default | Description |
|---|---|---|
| `query` | required | Search string |
| `--count` / `-n` | `20` | Max results to return |

**Output:**
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

**Notes:**
- Only returns Docs, Sheets, and Slides — no other Drive files.
- Single quotes in `query` are escaped for the Drive API query syntax.
- `--count` maps to `pageSize` in the Drive API; the API may return fewer results than requested.

---

### `fetch`

```sh
gdrive-cli fetch <url> [--output FILE] [--dir DIR]
```

Downloads a Google Doc, Sheet, or Slides file and saves it locally.

| Argument / Option | Default | Description |
|---|---|---|
| `url` | required | Full Google Docs/Sheets/Slides URL |
| `--output` / `-o` | auto-generated | Explicit output file path |
| `--dir` / `-d` | `.` (current dir) | Output directory (used when `--output` is not set) |

**Export formats:**

| Source type | Saved as |
|---|---|
| Google Doc | `.docx` |
| Google Sheet | `.csv` |
| Google Slides | `.pptx` |

**Auto-generated filename:** `<sanitized-title><extension>` in the output directory. Characters `/ \ : * ? " < > |` in the title are replaced with `_`.

**Accepted URL formats:**
```
https://docs.google.com/document/d/<ID>/...
https://docs.google.com/spreadsheets/d/<ID>/...
https://docs.google.com/presentation/d/<ID>/...
```

URL-encoded characters are decoded before ID extraction.

**Output:**
```json
{
  "status": "ok",
  "file_id": "1aBcDe...",
  "name": "Q1 Budget",
  "type": "Google Sheet",
  "saved_to": "./Q1_Budget.csv"
}
```

**Errors:**
- Unrecognized URL format → lists supported formats
- Unsupported MIME type (e.g., a plain Drive file, not a Workspace doc) → lists supported types
- Output directory is created automatically if it does not exist

---

## Config reference

| Path | Purpose |
|---|---|
| `~/.config/gdrive-cli/credentials.json` | OAuth client credentials (you provide) |
| `~/.config/gdrive-cli/token.json` | OAuth token (written after login) |
| `$GDRIVE_CONFIG_DIR` | Override for the config directory |
