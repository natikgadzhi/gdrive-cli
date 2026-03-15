package main

import (
	"fmt"
	"os"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/natikgadzhi/gdrive-cli/internal/progress"
	"github.com/spf13/cobra"
)

var searchCount int

// searchResponse is the JSON envelope for search results.
type searchResponse struct {
	Query   string           `json:"query"`
	Count   int              `json:"count"`
	Results []api.FileResult `json:"results"`
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Google Drive for documents",
	Long:  "Searches Google Drive for Docs, Sheets, and Slides matching the query. Matches on both file name and full text content.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a search query\n\nUsage: gdrive-cli search <query> [--count N]\n\nExample: gdrive-cli search \"budget 2025\"")
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		if query == "" {
			return output.Errorf("search query must not be empty")
		}

		config.DebugLog("Searching for %q with count=%d", query, searchCount)

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

		// Show spinner while searching.
		spinner := progress.NewSpinner("Searching Google Drive...")
		spinner.Start()
		defer spinner.Stop()

		results, err := api.SearchFiles(svc, query, searchCount)
		if err != nil {
			return output.Errorf("Search failed: %s", err)
		}

		// Markdown table output.
		if outputFormat == output.FormatMarkdown {
			spinner.Stop()
			fmt.Fprintf(os.Stdout, "# Search: %s\n\n", query)
			fmt.Fprintf(os.Stdout, "**%d results**\n\n", len(results))
			if len(results) > 0 {
				fmt.Fprintln(os.Stdout, "| Name | Type | Modified | URL |")
				fmt.Fprintln(os.Stdout, "|------|------|----------|-----|")
				for _, r := range results {
					fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
						r.Name, r.Type, r.Modified, r.URL)
				}
			}
			return nil
		}

		return output.PrintJSON(searchResponse{
			Query:   query,
			Count:   len(results),
			Results: results,
		})
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchCount, "count", "n", 20, "Maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}
