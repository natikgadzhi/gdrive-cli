package main

import (
	"fmt"

	clierrors "github.com/natikgadzhi/cli-kit/errors"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	cliprogress "github.com/natikgadzhi/cli-kit/progress"
	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/spf13/cobra"
)

var searchLimit int

// searchResponse is the JSON envelope for search results.
type searchResponse struct {
	Query   string           `json:"query"`
	Count   int              `json:"count"`
	Results []api.FileResult `json:"results"`
}

// RenderTable implements cli-kit TableRenderer.
func (s searchResponse) RenderTable(t *clioutput.Table) {
	t.Header("Name", "Type", "Modified", "URL")
	for _, r := range s.Results {
		t.Row(r.Name, r.Type, r.Modified, r.URL)
	}
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Google Drive for documents",
	Long:  "Searches Google Drive for Docs, Sheets, and Slides matching the query. Matches on both file name and full text content.",
	Example: `  gdrive-cli search "budget 2025"
  gdrive-cli search "project proposal" -n 5
  gdrive-cli search "Q1 report" -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a search query\n\nUsage: gdrive-cli search <query> [--limit N]\n\nExample: gdrive-cli search \"budget 2025\"")
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		format := clioutput.Resolve(cmd)

		if query == "" {
			return cliError(clierrors.ExitError, "search query must not be empty", cmd)
		}

		config.DebugLog("Searching for %q with limit=%d", query, searchLimit)

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

		// Show spinner while searching.
		spinner := cliprogress.NewSpinner("Searching Google Drive...", format)
		defer spinner.Finish()

		results, err := api.SearchFiles(svc, query, searchLimit)
		if err != nil {
			return cliError(clierrors.ExitError, "Search failed: %s", cmd, err)
		}

		spinner.Finish()

		resp := searchResponse{
			Query:   query,
			Count:   len(results),
			Results: results,
		}
		return clioutput.Print(format, resp, resp)
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 20, "Maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}
