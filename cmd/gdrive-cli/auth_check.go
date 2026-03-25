package main

import (
	clierrors "github.com/natikgadzhi/cli-kit/errors"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/gdrive-cli/internal/auth"
	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/spf13/cobra"
)

// authCheckResult is the JSON/table output for an auth check.
type authCheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (r authCheckResult) RenderTable(t *clioutput.Table) {
	t.Header("Status", "Message")
	t.Row(r.Status, r.Message)
}

var authCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check authentication status",
	Long:  "Checks whether stored credentials exist and are valid (or refreshable).",
	Example: `  gdrive-cli auth check`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := auth.GetCredentials(config.ConfigDir())
		if err != nil {
			return cliError(clierrors.ExitAuthError, "Not authenticated. Run `gdrive-cli auth login` first.", cmd)
		}
		format := clioutput.Resolve(cmd)
		result := authCheckResult{
			Status:  "ok",
			Message: "Authenticated and credentials are valid.",
		}
		return clioutput.Print(format, result, result)
	},
}

func init() {
	authCmd.AddCommand(authCheckCmd)
}
