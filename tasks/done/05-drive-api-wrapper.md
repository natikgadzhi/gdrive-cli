# Task 05: Drive API wrapper

**Phase**: 1 — Core
**Status**: done
**Assignee**: builder-06
**PR**: https://github.com/natikgadzhi/gdrive-cli/pull/8 (merged)
**Depends on**: 04
**Blocks**: 08, 09, 10, 13

## Description

Implement the Google Drive API v3 wrapper: building an authenticated service, searching files, exporting files, and fetching metadata.

## Acceptance Criteria

- [ ] `internal/api/types.go`:
  - `FileResult` struct: Name, Type (label), URL, Modified (time string)
  - `FileMetadata` struct: ID, Name, MimeType, WebViewLink
- [ ] `internal/api/client.go`:
  - `NewDriveService(token *oauth2.Token, config *oauth2.Config) (*drive.Service, error)` — builds authenticated Drive v3 client
  - Accepts optional `http.Client` for testing
- [ ] `internal/api/search.go`:
  - `SearchFiles(svc *drive.Service, query string, maxResults int) ([]FileResult, error)`
  - Builds query: `(mimeType=doc OR sheet OR slides) AND (name contains 'q' OR fullText contains 'q')`
  - Orders by `modifiedTime desc`
  - Returns typed results
- [ ] `internal/api/export.go`:
  - `ExportFile(svc *drive.Service, fileID string, mimeType string, outputPath string) error`
  - Writes exported content to disk
- [ ] `internal/api/metadata.go`:
  - `GetFileMetadata(svc *drive.Service, fileID string) (*FileMetadata, error)`
  - Fetches id, name, mimeType, webViewLink
- [ ] Tests with httptest:
  - Mock Drive API list endpoint, verify query construction
  - Mock export endpoint, verify file is written
  - Mock metadata endpoint
  - Test error cases (404, 403, network error)

## Files to Create

- `internal/api/types.go`
- `internal/api/client.go`
- `internal/api/search.go`
- `internal/api/export.go`
- `internal/api/metadata.go`
- `internal/api/api_test.go`

## Reference

Port from `src/gdrive_cli/api.py`. Key details:
- MIME type filter in query: `application/vnd.google-apps.document`, `.spreadsheet`, `.presentation`
- Fields: `files(id, name, mimeType, webViewLink, modifiedTime)`
- Single quotes in search query escaped with backslash
