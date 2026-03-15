// Package api provides a client for the Google Drive API,
// wrapping search, metadata retrieval, and file export operations.
package api

// FileResult represents a search result from the Google Drive API.
type FileResult struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Modified string `json:"modified"`
}

// FileMetadata holds metadata for a single Google Drive file.
type FileMetadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MimeType    string `json:"mimeType"`
	WebViewLink string `json:"webViewLink"`
}
