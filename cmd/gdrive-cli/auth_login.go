package main

import (
	clierrors "github.com/natikgadzhi/cli-kit/errors"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/spf13/cobra"
)

// authLoginResult is the JSON/table output for a successful login.
type authLoginResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (r authLoginResult) RenderTable(t *clioutput.Table) {
	t.Header("Status", "Message")
	t.Row(r.Status, r.Message)
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google Drive",
	Long:  "Runs the OAuth2 installed-app flow. Opens a browser for Google consent and saves credentials locally.",
	Example: `  gdrive-cli auth login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Login(config.ConfigDir()); err != nil {
			return cliError(clierrors.ExitError, "%s", cmd, err)
		}
		format := clioutput.Resolve(cmd)
		result := authLoginResult{
			Status:  "ok",
			Message: "Successfully authenticated with Google Drive.",
		}
		return clioutput.Print(format, result, result)
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
