package main

import (
	"fmt"
	"os"

	clierrors "github.com/natikgadzhi/cli-kit/errors"
	clioutput "github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/cli-kit/version"
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
	debug   bool
	noCache bool
)

// versionInfo is populated in init() from ldflags variables.
var versionInfo = &version.Info{
	Version: "dev",
	Commit:  "dev",
	Date:    "unknown",
}

// SilentError is a sentinel error returned after an error has already been
// printed. Cobra's RunE should return this to signal a non-zero exit code
// without Cobra printing its own error message.
type SilentError struct {
	Message  string
	ExitCode int
}

func (e *SilentError) Error() string {
	return e.Message
}

// IsSilentError reports whether err is (or wraps) a SilentError.
func IsSilentError(err error) bool {
	_, ok := err.(*SilentError)
	return ok
}

// cliError creates a CLIError, prints it to stderr, and returns a SilentError
// that carries the appropriate exit code.
func cliError(exitCode int, format string, cmd *cobra.Command, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	outputFormat := clioutput.Resolve(cmd)
	cliErr := clierrors.NewCLIError(exitCode, msg)
	clierrors.PrintError(cliErr, clioutput.IsJSON(outputFormat))
	return &SilentError{Message: msg, ExitCode: exitCode}
}

var rootCmd = &cobra.Command{
	Use:   "gdrive-cli",
	Short: "CLI tool for Google Drive",
	Long:  "A command-line tool to search and download Google Docs, Sheets, and Slides via the Google Drive API.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config.SetDebug(debug)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Populate versionInfo from ldflags.
	versionInfo.Version = Version
	versionInfo.Commit = Commit
	versionInfo.Date = Date

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print verbose debug logs to stderr")
	rootCmd.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Skip writing to derived directory")

	// Register cli-kit flags.
	clioutput.RegisterFlag(rootCmd)

	// Register version command and --version flag.
	rootCmd.AddCommand(version.NewCommand(versionInfo))
	version.SetupFlag(rootCmd, versionInfo)
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		if se, ok := err.(*SilentError); ok {
			if se.ExitCode != 0 {
				os.Exit(se.ExitCode)
			}
			return err
		}
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
