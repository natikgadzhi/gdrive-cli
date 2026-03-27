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

// CommentAuthor represents the author of a comment or reply.
// Note: the Drive API does not populate email addresses for comment authors.
type CommentAuthor struct {
	DisplayName string `json:"display_name"`
	IsMe        bool   `json:"is_me,omitempty"`
}

// CommentReply represents a single reply within a comment thread.
// Deleted replies are filtered out by ListComments and never returned.
type CommentReply struct {
	ID          string        `json:"id"`
	Author      CommentAuthor `json:"author"`
	Content     string        `json:"content"`
	CreatedTime string        `json:"created_time"`
	Action      string        `json:"action,omitempty"` // "resolve" or "reopen"
}

// CommentThread represents a top-level comment and its replies.
// Deleted comments are filtered out by ListComments and never returned.
type CommentThread struct {
	ID           string         `json:"id"`
	Author       CommentAuthor  `json:"author"`
	Content      string         `json:"content"`
	QuotedText   string         `json:"quoted_text,omitempty"`
	CreatedTime  string         `json:"created_time"`
	ModifiedTime string         `json:"modified_time"`
	Resolved     bool           `json:"resolved"`
	Replies      []CommentReply `json:"replies,omitempty"`
}
