package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	clierrors "github.com/natikgadzhi/cli-kit/errors"
	"github.com/natikgadzhi/cli-kit/derived"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	cliprogress "github.com/natikgadzhi/cli-kit/progress"
	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/cache"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/spf13/cobra"
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

// RenderTable implements cli-kit TableRenderer.
func (r fetchResult) RenderTable(t *clioutput.Table) {
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
		config.DebugLog("Parsed file ID: %s", fileID)

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
		config.DebugLog("File: %s (MIME: %s)", metadata.Name, metadata.MimeType)

		// Check that this is a supported Google Workspace type at all.
		_, supportedType := formatting.GetExportMIME(metadata.MimeType)
		if !supportedType {
			return cliError(clierrors.ExitError,
				"Unsupported file type: %s\n\nSupported types:\n"+
					"  Google Doc    (application/vnd.google-apps.document)\n"+
					"  Google Sheet  (application/vnd.google-apps.spreadsheet)\n"+
					"  Google Slides (application/vnd.google-apps.presentation)",
				cmd, metadata.MimeType,
			)
		}

		// Resolve the requested export format against the document type.
		resolved, err := formatting.ResolveExportFormat(metadata.MimeType, fetchExportFormat)
		if err != nil {
			return cliError(clierrors.ExitError, "%s", cmd, err)
		}

		exportMIME := resolved.ExportMIME
		extension := resolved.Extension
		typeLabel, _ := formatting.GetTypeLabel(metadata.MimeType)

		config.DebugLog("Export format: MIME=%s ext=%s markdown=%v", exportMIME, extension, resolved.NeedsMarkdownConversion)

		// Determine output path from the --dest flag.
		outputPath := resolveOutputPath(fetchDest, metadata.Name, extension)
		config.DebugLog("Output path: %s", outputPath)

		// Build slug for derived directory.
		slug := cache.GenerateSlug(metadata.Name, fileID)

		// When exporting as markdown, use the markdown export path which
		// handles HTML-to-markdown conversion for Docs.
		if resolved.NeedsMarkdownConversion {
			spin.SetMessage("Exporting as markdown...")

			mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
			if err != nil {
				return cliError(clierrors.ExitError, "Failed to export as markdown: %s", cmd, err)
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

			spin.Finish()
			result := fetchResult{
				Status:   "ok",
				FileID:   fileID,
				Name:     metadata.Name,
				Type:     typeLabel,
				SavedTo:  outputPath,
				CachedTo: cachedTo,
			}
			return clioutput.Print(format, result, result)
		}

		// Native export path (docx/csv/pptx).
		spin.SetMessage("Downloading file...")

		if err := api.ExportFile(svc, fileID, exportMIME, outputPath); err != nil {
			return cliError(clierrors.ExitError, "Failed to export file: %s", cmd, err)
		}

		// Export as markdown/text for the derived directory.
		spin.SetMessage("Caching derived data...")

		mdContent, err := output.ExportAsMarkdown(svc, fileID, metadata.MimeType)
		if err != nil {
			// Cache failure is non-fatal -- log it and continue.
			config.DebugLog("Warning: failed to export markdown for cache: %v", err)
		}

		var cachedTo string
		if mdContent != "" && !noCache {
			cachedTo = writeDerivedFile(cmd, slug, typeLabel, rawURL, mdContent)
		}

		spin.Finish()
		result := fetchResult{
			Status:   "ok",
			FileID:   fileID,
			Name:     metadata.Name,
			Type:     typeLabel,
			SavedTo:  outputPath,
			CachedTo: cachedTo,
		}
		return clioutput.Print(format, result, result)
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

// writeDerivedFile writes a markdown file to the derived directory with
// cli-kit frontmatter. Returns the path of the written file, or "" on failure.
func writeDerivedFile(cmd *cobra.Command, slug, typeLabel, sourceURL, body string) string {
	derivedDir := derived.Resolve(cmd, "gdrive-cli")
	if err := derived.EnsureDir(derivedDir); err != nil {
		config.DebugLog("Warning: failed to create derived directory: %v", err)
		return ""
	}

	fm := derived.NewFrontmatter("gdrive-cli", typeLabel, slug, sourceURL, "fetch")
	content := derived.FormatFile(fm, body)

	filePath := filepath.Join(derivedDir, slug+".md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		config.DebugLog("Warning: failed to write derived file: %v", err)
		return ""
	}
	config.DebugLog("Cached to: %s", filePath)
	return filePath
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchExportFormat, "export", "e", "", "Export format: docx, md, csv, pptx (depends on document type)")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "f", "", "Output path (file or directory; auto-generates filename if directory)")
	derived.RegisterFlag(rootCmd, "gdrive-cli")
	rootCmd.AddCommand(fetchCmd)
}
