// Package cache provides a local caching layer for fetched Google Drive
// documents. Each cached file is stored with YAML frontmatter metadata
// followed by the document body.
package cache

import "time"

// CacheEntry holds the metadata and body content for a cached file.
type CacheEntry struct {
	// Tool is always "gdrive-cli".
	Tool string `yaml:"tool"`
	// Name is the original Google Drive file name.
	Name string `yaml:"name"`
	// Slug is the sanitized filename used on disk.
	Slug string `yaml:"slug"`
	// Type is the human-readable file type (e.g. "Google Doc").
	Type string `yaml:"type"`
	// FileID is the Google Drive file ID.
	FileID string `yaml:"file_id"`
	// SourceURL is the web URL for the original file.
	SourceURL string `yaml:"source_url"`
	// CreatedAt is when the cache entry was first written.
	CreatedAt time.Time `yaml:"created_at"`
	// UpdatedAt is when the cache entry was last updated.
	UpdatedAt time.Time `yaml:"updated_at"`
	// RequestedBy records who or what triggered the fetch (e.g. "cli").
	RequestedBy string `yaml:"requested_by"`
	// Body is the cached content (markdown, CSV, or plain text).
	// It is not serialized in the YAML frontmatter; it follows the "---" delimiter.
	Body string `yaml:"-"`
}
