package main

import (
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/spf13/cobra"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google Drive",
	Long:  "Runs the OAuth2 installed-app flow. Opens a browser for Google consent and saves credentials locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Login(config.ConfigDir()); err != nil {
			return output.PrintJSON(map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
		}
		return output.PrintJSON(map[string]string{
			"status":  "ok",
			"message": "Successfully authenticated with Google Drive.",
		})
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
