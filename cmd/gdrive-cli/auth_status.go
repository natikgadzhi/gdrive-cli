package main

import (
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  "Checks whether stored credentials exist and are valid (or refreshable).",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := auth.GetCredentials(config.ConfigDir())
		if err != nil {
			return output.PrintJSON(map[string]string{
				"status":  "error",
				"message": "Not authenticated. Run `gdrive-cli auth login` first.",
			})
		}
		return output.PrintJSON(map[string]string{
			"status":  "ok",
			"message": "Authenticated and credentials are valid.",
		})
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
