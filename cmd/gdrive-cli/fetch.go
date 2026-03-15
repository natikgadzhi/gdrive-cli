package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/cache"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/natikgadzhi/gdrive-cli/internal/progress"
	"github.com/spf13/cobra"
)

var (
	fetchOutput string
	fetchDir    string
)

// fetchResult is the JSON output for a successful fetch.
type fetchResult struct {
	Status   string `json:"status"`
	FileID   string `json:"file_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	SavedTo  string `json:"saved_to"`
	CachedTo string `json:"cached_to,omitempty"`
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <url>",
	Short: "Download a Google Doc, Sheet, or Slides file",
	Long: `Downloads a Google Doc, Sheet, or Slides file and saves it locally.

Export formats:
  Google Doc    → .docx
  Google Sheet  → .csv
  Google Slides → .pptx

The output filename is auto-generated from the document title unless
--output is specified.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]

		// Parse the Google Drive URL to extract the file ID.
		fileID, err := formatting.ParseGoogleURL(rawURL)
		if err != nil {
			return output.Errorf("%s", err)
		}
		config.DebugLog("Parsed file ID: %s", fileID)

		// Authenticate.
		token, oauthConfig, err := auth.GetCredentials(config.ConfigDir())
		if err != nil {
			return output.Errorf("Authentication failed: %s", err)
		}

		// Create Drive service.
		svc, err := api.NewDriveService(token, oauthConfig)
		if err != nil {
			return output.Errorf("Failed to create Drive service: %s", err)
		}

		// Fetch file metadata.
		spin := progress.NewSpinner("Fetching file metadata...")
		spin.Start()
		defer spin.Stop()

		metadata, err := api.GetFileMetadata(svc, fileID)
		if err != nil {
			return output.Errorf("Failed to get file metadata: %s", err)
		}
		config.DebugLog("File: %s (MIME: %s)", metadata.Name, metadata.MimeType)

		// Determine export format.
		exportMIME, ok := formatting.GetExportMIME(metadata.MimeType)
		if !ok {
			return output.Errorf(
				"Unsupported file type: %s\n\nSupported types:\n"+
					"  Google Doc    (application/vnd.google-apps.document)\n"+
					"  Google Sheet  (application/vnd.google-apps.spreadsheet)\n"+
					"  Google Slides (application/vnd.google-apps.presentation)",
				metadata.MimeType,
			)
		}

		extension, _ := formatting.GetExportExtension(metadata.MimeType)
		typeLabel, _ := formatting.GetTypeLabel(metadata.MimeType)

		// Determine output path.
		var outputPath string
		if fetchOutput != "" {
			outputPath = fetchOutput
		} else {
			safeName := formatting.SanitizeFilename(metadata.Name)
			outputPath = filepath.Join(fetchDir, safeName+extension)
		}
		config.DebugLog("Output path: %s", outputPath)

		// Export the file in its native format (docx/csv/pptx).
		spin.UpdateMessage("Downloading " + metadata.Name + "...")

		if err := api.ExportFile(svc, fileID, exportMIME, outputPath); err != nil {
			return output.Errorf("Failed to export file: %s", err)
		}

		// Export as markdown/text for the cache.
		spin.UpdateMessage("Caching " + metadata.Name + "...")

		mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
		if err != nil {
			// Cache failure is non-fatal — log it and continue.
			config.DebugLog("Warning: failed to export markdown for cache: %v", err)
		}

		var cachedTo string
		if mdContent != "" {
			now := time.Now().UTC()
			slug := cache.GenerateSlug(metadata.Name, fileID)
			entry := cache.CacheEntry{
				Tool:        "gdrive-cli",
				Name:        metadata.Name,
				Slug:        slug,
				Type:        typeLabel,
				FileID:      fileID,
				SourceURL:   rawURL,
				CreatedAt:   now,
				UpdatedAt:   now,
				RequestedBy: "cli",
				Body:        mdContent,
			}

			cacheDir := config.CacheDir()
			cachedPath, err := cache.Store(cacheDir, entry)
			if err != nil {
				config.DebugLog("Warning: failed to write cache: %v", err)
			} else {
				cachedTo = cachedPath
				config.DebugLog("Cached to: %s", cachedTo)
			}

			// If markdown format requested, print the cached content to stdout.
			if format == "markdown" {
				spin.Stop()
				fmt.Fprint(os.Stdout, "---\n")
				fmt.Fprintf(os.Stdout, "tool: %s\n", entry.Tool)
				fmt.Fprintf(os.Stdout, "name: %s\n", entry.Name)
				fmt.Fprintf(os.Stdout, "type: %s\n", entry.Type)
				fmt.Fprintf(os.Stdout, "file_id: %s\n", entry.FileID)
				fmt.Fprintf(os.Stdout, "source_url: %s\n", entry.SourceURL)
				fmt.Fprint(os.Stdout, "---\n")
				fmt.Fprint(os.Stdout, mdContent)
				return nil
			}
		}

		// Stop spinner before printing to clear the terminal line.
		spin.Stop()

		return output.PrintJSON(fetchResult{
			Status:   "ok",
			FileID:   fileID,
			Name:     metadata.Name,
			Type:     typeLabel,
			SavedTo:  outputPath,
			CachedTo: cachedTo,
		})
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchOutput, "output", "o", "", "Explicit output file path")
	fetchCmd.Flags().StringVarP(&fetchDir, "dir", "d", ".", "Output directory (used when --output is not set)")
	rootCmd.AddCommand(fetchCmd)
}
