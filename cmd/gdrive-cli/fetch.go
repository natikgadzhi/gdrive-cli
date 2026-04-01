package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natikgadzhi/cli-kit/debug"
	"github.com/natikgadzhi/cli-kit/derived"
	clierrors "github.com/natikgadzhi/cli-kit/errors"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	cliprogress "github.com/natikgadzhi/cli-kit/progress"
	"github.com/natikgadzhi/cli-kit/table"
	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/cache"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/spf13/cobra"
	drive "google.golang.org/api/drive/v3"
)

var (
	fetchDest         string
	fetchExportFormat string
	fetchNoComments   bool
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

// RenderBorderedTable renders fetch results into a bordered table.
func (r fetchResult) RenderBorderedTable(t *table.Table) {
	t.Header("STATUS", "NAME", "TYPE", "SAVED TO")
	t.Row(r.Status, r.Name, r.Type, r.SavedTo)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <url>",
	Short: "Download a Google Doc, Sheet, or Slides file",
	Long: `Downloads a Google Doc, Sheet, or Slides file and saves it locally.

Use --export / -e to choose the file format for the download. If omitted,
the default native format is used.

Export formats per document type:
  Google Doc    : docx (default), md
  Google Sheet  : csv  (default)
  Google Slides : pptx (default), md, pdf

When --export md (or -e md) is used:
  - Google Docs are exported as HTML and converted to Markdown (.md)
  - Google Slides are exported as plain text and converted to structured
    Markdown with slide boundary markers (.md)

Use --dest / -f to control where the file is saved:
  - If omitted, the file is saved in the current directory with an
    auto-generated name based on the document title.
  - If set to a directory (path ends with / or is an existing directory),
    the file is saved in that directory with an auto-generated name.
  - Otherwise, it is used as the explicit output file path.`,
	Example: `  gdrive-cli fetch https://docs.google.com/document/d/1abc.../edit
  gdrive-cli fetch https://docs.google.com/spreadsheets/d/1abc.../edit -e csv
  gdrive-cli fetch https://docs.google.com/document/d/1abc.../edit -e md -f ./output/`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a Google Docs/Sheets/Slides URL\n\nUsage: gdrive-cli fetch <url> [-e FORMAT] [--dest PATH]\n\nSupported URL formats:\n  https://docs.google.com/document/d/<ID>/...\n  https://docs.google.com/spreadsheets/d/<ID>/...\n  https://docs.google.com/presentation/d/<ID>/...")
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		format := clioutput.Resolve(cmd)

		// Parse the Google Drive URL to extract the file ID.
		fileID, err := formatting.ParseGoogleURL(rawURL)
		if err != nil {
			return cliError(clierrors.ExitError, "%s", cmd, err)
		}
		debug.Log("Parsed file ID: %s", fileID)

		// Authenticate.
		token, oauthConfig, err := auth.GetCredentials(config.ConfigDir())
		if err != nil {
			return cliError(clierrors.ExitAuthError, "Authentication failed: %s", cmd, err)
		}

		// Create Drive service.
		svc, err := api.NewDriveService(token, oauthConfig)
		if err != nil {
			return cliError(clierrors.ExitError, "Failed to create Drive service: %s", cmd, err)
		}

		// Fetch file metadata.
		spin := cliprogress.NewSpinner("Fetching file metadata...", format)
		defer spin.Finish()

		metadata, err := api.GetFileMetadata(svc, fileID)
		if err != nil {
			return cliError(clierrors.ExitError, "Failed to get file metadata: %s", cmd, err)
		}
		debug.Log("File: %s (MIME: %s)", metadata.Name, metadata.MimeType)

		// Branch: native Google Workspace type vs. non-native (uploaded) file.
		if !formatting.IsNativeGoogleType(metadata.MimeType) {
			return fetchBinaryFile(cmd, svc, metadata, rawURL, format, spin)
		}

		// --- Native Google Workspace export path ---

		// Resolve the requested export format against the document type.
		resolved, err := formatting.ResolveExportFormat(metadata.MimeType, fetchExportFormat)
		if err != nil {
			return cliError(clierrors.ExitError, "%s", cmd, err)
		}

		exportMIME := resolved.ExportMIME
		extension := resolved.Extension
		typeLabel, _ := formatting.GetTypeLabel(metadata.MimeType)

		debug.Log("Export format: MIME=%s ext=%s markdown=%v", exportMIME, extension, resolved.NeedsMarkdownConversion)

		// Determine output path from the --dest flag.
		outputPath := resolveOutputPath(fetchDest, metadata.Name, extension)
		debug.Log("Output path: %s", outputPath)

		// Build slug for derived directory.
		slug := cache.GenerateSlug(metadata.Name, fileID)

		// When exporting as markdown, use the markdown export path which
		// handles HTML-to-markdown conversion for Docs.
		if resolved.NeedsMarkdownConversion {
			spin.SetLabel("Exporting as markdown...")

			mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
			if err != nil {
				// Try fallbacks for known export errors.
				if fallbackErr := handleExportFallback(cmd, svc, metadata, outputPath, rawURL, slug, typeLabel, format, spin, err); fallbackErr != nil {
					return fallbackErr
				}
				return nil
			}

			// Write the markdown content to the output file.
			dir := filepath.Dir(outputPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return cliError(clierrors.ExitError, "Failed to create output directory %s: %s", cmd, dir, err)
			}
			if err := os.WriteFile(outputPath, []byte(mdContent), 0o644); err != nil {
				return cliError(clierrors.ExitError, "Failed to write output file: %s", cmd, err)
			}

			// Cache to derived directory.
			var cachedTo string
			if !noCache {
				cachedTo = writeDerivedFile(cmd, slug, typeLabel, rawURL, mdContent)
			}

			// Fetch and cache comments.
			fetchAndCacheComments(cmd, svc, fileID, metadata.Name, slug, spin)

			spin.Finish()
			result := fetchResult{
				Status:   "ok",
				FileID:   fileID,
				Name:     metadata.Name,
				Type:     typeLabel,
				SavedTo:  outputPath,
				CachedTo: cachedTo,
			}
			return printResult(format, result, result)
		}

		// Native export path (docx/csv/pptx).
		spin.SetLabel("Downloading file...")

		if err := api.ExportFile(svc, fileID, exportMIME, outputPath); err != nil {
			// Try fallbacks for known export errors.
			if fallbackErr := handleExportFallback(cmd, svc, metadata, outputPath, rawURL, slug, typeLabel, format, spin, err); fallbackErr != nil {
				return fallbackErr
			}
			return nil
		}

		// Export as markdown/text for the derived directory.
		spin.SetLabel("Caching derived data...")

		mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
		if err != nil {
			// Cache failure is non-fatal -- log it and continue.
			debug.Log("Warning: failed to export markdown for cache: %v", err)
		}

		var cachedTo string
		if mdContent != "" && !noCache {
			cachedTo = writeDerivedFile(cmd, slug, typeLabel, rawURL, mdContent)
		}

		// Fetch and cache comments.
		fetchAndCacheComments(cmd, svc, fileID, metadata.Name, slug, spin)

		spin.Finish()
		result := fetchResult{
			Status:   "ok",
			FileID:   fileID,
			Name:     metadata.Name,
			Type:     typeLabel,
			SavedTo:  outputPath,
			CachedTo: cachedTo,
		}
		return printResult(format, result, result)
	},
}

// fetchAndCacheComments fetches comments for a file and writes them as a
// companion .comments.md file in the derived directory. Failures are non-fatal.
func fetchAndCacheComments(cmd *cobra.Command, svc *drive.Service, fileID, docName, slug string, spin cliprogress.Indicator) {
	if fetchNoComments || noCache {
		return
	}

	spin.SetLabel("Fetching comments...")

	threads, err := api.ListComments(svc, fileID)
	if err != nil {
		debug.Log("Warning: failed to fetch comments: %v", err)
		return
	}

	if len(threads) == 0 {
		debug.Log("No comments found for file %s", fileID)
		return
	}

	body := output.FormatCommentsMarkdown(docName, threads)

	derivedDir := derived.Resolve(cmd, "gdrive-cli")
	if err := derived.EnsureDir(derivedDir); err != nil {
		debug.Log("Warning: failed to create derived directory for comments: %v", err)
		return
	}

	fm := derived.NewFrontmatter("gdrive-cli", "comments", slug+".comments", "", "fetch")
	content := derived.FormatFile(fm, body)

	filePath := filepath.Join(derivedDir, slug+".comments.md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		debug.Log("Warning: failed to write comments file: %v", err)
		return
	}

	debug.Log("Cached %d comment threads to %s", len(threads), filePath)
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

// writeDerivedFile writes a markdown file to the derived directory with
// cli-kit frontmatter. Returns the path of the written file, or "" on failure.
func writeDerivedFile(cmd *cobra.Command, slug, typeLabel, sourceURL, body string) string {
	derivedDir := derived.Resolve(cmd, "gdrive-cli")
	if err := derived.EnsureDir(derivedDir); err != nil {
		debug.Log("Warning: failed to create derived directory: %v", err)
		return ""
	}

	fm := derived.NewFrontmatter("gdrive-cli", typeLabel, slug, sourceURL, "fetch")
	content := derived.FormatFile(fm, body)

	filePath := filepath.Join(derivedDir, slug+".md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		debug.Log("Warning: failed to write derived file: %v", err)
		return ""
	}
	debug.Log("Cached to: %s", filePath)
	return filePath
}

// fetchBinaryFile handles downloading non-native (uploaded) files from Google
// Drive via the alt=media binary download path. It determines the file extension
// from the MIME type or the original filename and saves the file to disk.
func fetchBinaryFile(
	cmd *cobra.Command,
	svc *drive.Service,
	metadata *api.FileMetadata,
	rawURL string,
	format string,
	spin cliprogress.Indicator,
) error {
	// Determine extension and label from known binary MIME types.
	extension, typeLabel, ok := formatting.GetBinaryTypeInfo(metadata.MimeType)
	if !ok {
		// Fall back to the file's original extension.
		extension = formatting.ExtensionFromFilename(metadata.Name)
		typeLabel = metadata.MimeType
	}

	// If the user specified --export, ignore it for non-native files and warn.
	if fetchExportFormat != "" {
		debug.Log("Warning: --export flag ignored for non-native file type %s", metadata.MimeType)
	}

	outputPath := resolveOutputPath(fetchDest, metadata.Name, extension)
	debug.Log("Binary download: extension=%s label=%s path=%s", extension, typeLabel, outputPath)

	spin.SetLabel("Downloading file...")

	if err := api.DownloadFile(svc, metadata.ID, outputPath); err != nil {
		return cliError(clierrors.ExitError, "Failed to download file: %s", cmd, err)
	}

	// Fetch and cache comments (even for non-native files).
	slug := cache.GenerateSlug(metadata.Name, metadata.ID)
	fetchAndCacheComments(cmd, svc, metadata.ID, metadata.Name, slug, spin)

	spin.Finish()
	result := fetchResult{
		Status:  "ok",
		FileID:  metadata.ID,
		Name:    metadata.Name,
		Type:    typeLabel,
		SavedTo: outputPath,
		// No document body cache for binary files (comments may still be cached).
	}
	return printResult(format, result, result)
}

// handleExportFallback handles fallback logic when a Google Workspace file
// export fails with specific API errors:
//
//   - cannotExportFile (403): Falls back to binary download (alt=media). If that
//     also fails, returns a helpful error with file metadata.
//   - exportSizeLimitExceeded (403): Falls back to plain text export, then binary
//     download. If all fail, returns a helpful error.
//
// Returns nil if a fallback succeeded (caller should return nil too), or a
// non-nil error to propagate.
func handleExportFallback(
	cmd *cobra.Command,
	svc *drive.Service,
	metadata *api.FileMetadata,
	outputPath, rawURL, slug, typeLabel string,
	format string,
	spin cliprogress.Indicator,
	exportErr error,
) error {
	fileID := metadata.ID

	if api.IsExportSizeLimitExceeded(exportErr) {
		debug.Log("Export size limit exceeded, trying text/plain fallback")
		spin.SetLabel("File too large, trying plain text export...")

		// Try exporting as text/plain (higher limit).
		plainTextPath := replaceExtension(outputPath, ".txt")
		if err := api.ExportFile(svc, fileID, "text/plain", plainTextPath); err != nil {
			debug.Log("Plain text export also failed: %v, trying binary download", err)

			// Try binary download as last resort.
			spin.SetLabel("Plain text export failed, trying binary download...")
			if dlErr := api.DownloadFile(svc, fileID, outputPath); dlErr != nil {
				return cliError(clierrors.ExitError,
					"File too large to export.\n\n"+
						"  File:  %s\n"+
						"  Type:  %s\n"+
						"  URL:   %s\n\n"+
						"All export methods failed:\n"+
						"  - Original export: %s\n"+
						"  - Plain text export: %s\n"+
						"  - Binary download: %s",
					cmd, metadata.Name, typeLabel, rawURL, exportErr, err, dlErr)
			}
			// Binary download succeeded.
			fetchAndCacheComments(cmd, svc, fileID, metadata.Name, slug, spin)
			spin.Finish()
			result := fetchResult{
				Status:  "ok",
				FileID:  fileID,
				Name:    metadata.Name,
				Type:    typeLabel,
				SavedTo: outputPath,
			}
			return printResult(format, result, result)
		}

		// Plain text export succeeded.
		var cachedTo string
		if !noCache {
			// Read back the plain text for derived cache.
			if data, err := os.ReadFile(plainTextPath); err == nil {
				cachedTo = writeDerivedFile(cmd, slug, typeLabel, rawURL, string(data))
			}
		}

		fetchAndCacheComments(cmd, svc, fileID, metadata.Name, slug, spin)
		spin.Finish()
		result := fetchResult{
			Status:   "ok",
			FileID:   fileID,
			Name:     metadata.Name,
			Type:     typeLabel,
			SavedTo:  plainTextPath,
			CachedTo: cachedTo,
		}
		return printResult(format, result, result)
	}

	if api.IsCannotExportFile(exportErr) {
		debug.Log("Cannot export file (likely view-only with downloads disabled), trying binary download")
		spin.SetLabel("Export blocked, trying binary download...")

		if dlErr := api.DownloadFile(svc, fileID, outputPath); dlErr != nil {
			return cliError(clierrors.ExitError,
				"Cannot export or download this file. The file owner may have disabled downloads.\n\n"+
					"  File:  %s\n"+
					"  Type:  %s\n"+
					"  URL:   %s\n\n"+
					"Export error: %s\n"+
					"Download error: %s",
				cmd, metadata.Name, typeLabel, rawURL, exportErr, dlErr)
		}

		// Binary download succeeded.
		fetchAndCacheComments(cmd, svc, fileID, metadata.Name, slug, spin)
		spin.Finish()
		result := fetchResult{
			Status:  "ok",
			FileID:  fileID,
			Name:    metadata.Name,
			Type:    typeLabel,
			SavedTo: outputPath,
		}
		return printResult(format, result, result)
	}

	// Not a recognized fallback-able error — return the original error.
	return cliError(clierrors.ExitError, "Failed to export file: %s", cmd, exportErr)
}

// replaceExtension replaces the file extension in a path with a new one.
func replaceExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + newExt
	}
	return path[:len(path)-len(ext)] + newExt
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchExportFormat, "export", "e", "", "Export format: docx, md, csv, pptx (depends on document type)")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "f", "", "Output path (file or directory; auto-generates filename if directory)")
	fetchCmd.Flags().BoolVar(&fetchNoComments, "no-comments", false, "Skip fetching document comments")
	derived.RegisterFlag(rootCmd, "gdrive-cli")
	rootCmd.AddCommand(fetchCmd)
}
