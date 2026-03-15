package main

import (
	"fmt"
	"os"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/spf13/cobra"
)

// Version, Commit, and Date are set via ldflags at build time.
// Example: go build -ldflags "-X main.Version=1.0.0 -X main.Commit=abc123 -X main.Date=2025-01-01"
var (
	Version = "dev"
	Commit  = "dev"
	Date    = "unknown"
)

var (
	debug  bool
	format string
)

var rootCmd = &cobra.Command{
	Use:   "gdrive-cli",
	Short: "CLI tool for Google Drive",
	Long:  "A command-line tool to search and download Google Docs, Sheets, and Slides via the Google Drive API.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config.SetDebug(debug)
		if format != "json" && format != "markdown" {
			return fmt.Errorf("invalid format %q: must be \"json\" or \"markdown\"", format)
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print verbose debug logs to stderr")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json or markdown")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
