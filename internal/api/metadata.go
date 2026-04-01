package api

import (
	"fmt"

	"github.com/natikgadzhi/cli-kit/debug"
	drive "google.golang.org/api/drive/v3"
)

// GetFileMetadata fetches metadata for a single file from Google Drive.
// It returns the file's ID, name, MIME type, and web view link.
func GetFileMetadata(svc *drive.Service, fileID string) (*FileMetadata, error) {
	debug.Log("Fetching metadata for file %s", fileID)

	file, err := svc.Files.Get(fileID).
		SupportsAllDrives(true).
		Fields("id,name,mimeType,webViewLink").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return &FileMetadata{
		ID:          file.Id,
		Name:        file.Name,
		MimeType:    file.MimeType,
		WebViewLink: file.WebViewLink,
	}, nil
}
