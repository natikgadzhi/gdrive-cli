package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/cache"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/natikgadzhi/gdrive-cli/internal/progress"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	fetchDest         string
	fetchExportFormat string
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

Use --export / -o to choose the file format for the download. If omitted,
the default native format is used.

Export formats per document type:
  Google Doc    : docx (default), md
  Google Sheet  : csv  (default)
  Google Slides : pptx (default), md

When --export md (or -o md) is used:
  - Google Docs are exported as HTML and converted to Markdown (.md)
  - Google Slides are exported as plain text and saved as .md

Use --dest / -f to control where the file is saved:
  - If omitted, the file is saved in the current directory with an
    auto-generated name based on the document title.
  - If set to a directory (path ends with / or is an existing directory),
    the file is saved in that directory with an auto-generated name.
  - Otherwise, it is used as the explicit output file path.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a Google Docs/Sheets/Slides URL\n\nUsage: gdrive-cli fetch <url> [-o FORMAT] [--dest PATH]\n\nSupported URL formats:\n  https://docs.google.com/document/d/<ID>/...\n  https://docs.google.com/spreadsheets/d/<ID>/...\n  https://docs.google.com/presentation/d/<ID>/...")
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
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

		// Check that this is a supported Google Workspace type at all.
		_, supportedType := formatting.GetExportMIME(metadata.MimeType)
		if !supportedType {
			return output.Errorf(
				"Unsupported file type: %s\n\nSupported types:\n"+
					"  Google Doc    (application/vnd.google-apps.document)\n"+
					"  Google Sheet  (application/vnd.google-apps.spreadsheet)\n"+
					"  Google Slides (application/vnd.google-apps.presentation)",
				metadata.MimeType,
			)
		}

		// Resolve the requested export format against the document type.
		resolved, err := formatting.ResolveExportFormat(metadata.MimeType, fetchExportFormat)
		if err != nil {
			return output.Errorf("%s", err)
		}

		exportMIME := resolved.ExportMIME
		extension := resolved.Extension
		typeLabel, _ := formatting.GetTypeLabel(metadata.MimeType)

		config.DebugLog("Export format: MIME=%s ext=%s markdown=%v", exportMIME, extension, resolved.NeedsMarkdownConversion)

		// Determine output path from the --dest flag.
		outputPath := resolveOutputPath(fetchDest, metadata.Name, extension)
		config.DebugLog("Output path: %s", outputPath)

		// Build the CacheEntry once so both writeCacheEntry and
		// printMarkdownOutput share the same timestamps.
		now := time.Now().UTC()
		slug := cache.GenerateSlug(metadata.Name, fileID)
		cacheEntry := cache.CacheEntry{
			Tool:        "gdrive-cli",
			Name:        metadata.Name,
			Slug:        slug,
			Type:        typeLabel,
			FileID:      fileID,
			SourceURL:   rawURL,
			CreatedAt:   now,
			UpdatedAt:   now,
			RequestedBy: "cli",
		}

		// When exporting as markdown, use the markdown export path which
		// handles HTML-to-markdown conversion for Docs.
		if resolved.NeedsMarkdownConversion {
			spin.UpdateMessage("Exporting " + metadata.Name + " as markdown...")

			mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
			if err != nil {
				return output.Errorf("Failed to export as markdown: %s", err)
			}

			// Write the markdown content to the output file.
			dir := filepath.Dir(outputPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return output.Errorf("Failed to create output directory %s: %s", dir, err)
			}
			if err := os.WriteFile(outputPath, []byte(mdContent), 0o644); err != nil {
				return output.Errorf("Failed to write output file: %s", err)
			}

			// Cache the markdown content (same content as the file we just wrote).
			cacheEntry.Body = mdContent
			spin.UpdateMessage("Caching " + metadata.Name + "...")
			cachedTo := writeCacheEntry(cacheEntry)

			// If the global --format is markdown, print frontmatter + content to stdout.
			if outputFormat == output.FormatMarkdown {
				spin.Stop()
				printMarkdownOutput(cacheEntry)
				return nil
			}

			spin.Stop()
			return output.PrintJSON(fetchResult{
				Status:   "ok",
				FileID:   fileID,
				Name:     metadata.Name,
				Type:     typeLabel,
				SavedTo:  outputPath,
				CachedTo: cachedTo,
			})
		}

		// Native export path (docx/csv/pptx).
		spin.UpdateMessage("Downloading " + metadata.Name + "...")

		if err := api.ExportFile(svc, fileID, exportMIME, outputPath); err != nil {
			return output.Errorf("Failed to export file: %s", err)
		}

		// Export as markdown/text for the cache.
		spin.UpdateMessage("Caching " + metadata.Name + "...")

		mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
		if err != nil {
			// Cache failure is non-fatal -- log it and continue.
			config.DebugLog("Warning: failed to export markdown for cache: %v", err)
		}

		var cachedTo string
		if mdContent != "" {
			cacheEntry.Body = mdContent
			cachedTo = writeCacheEntry(cacheEntry)

			// If markdown format requested, print the cached content to stdout.
			if outputFormat == output.FormatMarkdown {
				spin.Stop()
				printMarkdownOutput(cacheEntry)
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

// resolveOutputPath determines the output file path from the --dest flag value.
//   - If flagValue is empty, auto-generates a filename in the current directory.
//   - If flagValue ends with a path separator or is an existing directory,
//     auto-generates a filename inside that directory.
//   - Otherwise, uses flagValue as the explicit file path.
func resolveOutputPath(flagValue, docTitle, extension string) string {
	safeName := formatting.SanitizeFilename(docTitle)
	autoName := safeName + extension

	if flagValue == "" {
		return autoName
	}

	// Treat paths ending in a separator as directories.
	if strings.HasSuffix(flagValue, string(filepath.Separator)) {
		return filepath.Join(flagValue, autoName)
	}

	// If the path is an existing directory, put the file inside it.
	info, err := os.Stat(flagValue)
	if err == nil && info.IsDir() {
		return filepath.Join(flagValue, autoName)
	}

	return flagValue
}

// writeCacheEntry writes a markdown cache entry and returns the cached file path.
// Returns "" if caching fails (failures are logged but non-fatal).
func writeCacheEntry(entry cache.CacheEntry) string {
	cacheDir := config.CacheDir()
	cachedPath, err := cache.Store(cacheDir, entry)
	if err != nil {
		config.DebugLog("Warning: failed to write cache: %v", err)
		return ""
	}
	config.DebugLog("Cached to: %s", cachedPath)
	return cachedPath
}

// printMarkdownOutput prints frontmatter + markdown content to stdout.
func printMarkdownOutput(entry cache.CacheEntry) {
	fm, err := yaml.Marshal(entry)
	if err != nil {
		// Best-effort: skip frontmatter if marshaling fails.
		config.DebugLog("Warning: failed to marshal frontmatter: %v", err)
	} else {
		fmt.Fprint(os.Stdout, "---\n")
		os.Stdout.Write(fm)
		fmt.Fprint(os.Stdout, "---\n")
	}
	fmt.Fprint(os.Stdout, entry.Body)
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchExportFormat, "export", "o", "", "Export format: docx, md, csv, pptx (depends on document type)")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "f", "", "Output path (file or directory; auto-generates filename if directory)")
	rootCmd.AddCommand(fetchCmd)
}
