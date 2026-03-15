package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	drive "google.golang.org/api/drive/v3"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

// ExportFile exports a Google Workspace file to the specified MIME type
// and writes the result to outputPath. Parent directories are created
// automatically if they do not exist.
func ExportFile(svc *drive.Service, fileID string, mimeType string, outputPath string) error {
	config.DebugLog("Exporting file %s as %s to %s", fileID, mimeType, outputPath)

	resp, err := svc.Files.Export(fileID, mimeType).Download()
	if err != nil {
		return fmt.Errorf("drive export failed: %w", err)
	}
	defer resp.Body.Close()

	// Create parent directories if needed.
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", dir, err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}

	n, err := io.Copy(outFile, resp.Body)
	if err != nil {
		outFile.Close()
		os.Remove(outputPath)
		return fmt.Errorf("failed to write export data: %w", err)
	}

	if err := outFile.Close(); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("closing export file: %w", err)
	}

	config.DebugLog("Wrote %d bytes to %s", n, outputPath)
	return nil
}
