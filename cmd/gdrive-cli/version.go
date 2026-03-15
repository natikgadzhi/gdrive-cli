package main

import (
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	"github.com/spf13/cobra"
)

// versionInfo is the JSON envelope for version output.
type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.PrintJSON(versionInfo{
			Version: Version,
			Commit:  Commit,
			Date:    Date,
		})
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
