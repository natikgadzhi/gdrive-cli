package api

import (
	"fmt"

	drive "google.golang.org/api/drive/v3"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

// GetFileMetadata fetches metadata for a single file from Google Drive.
// It returns the file's ID, name, MIME type, and web view link.
func GetFileMetadata(svc *drive.Service, fileID string) (*FileMetadata, error) {
	config.DebugLog("Fetching metadata for file %s", fileID)

	file, err := svc.Files.Get(fileID).
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
