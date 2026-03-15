package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]string{
			"version": Version,
			"commit":  Commit,
			"date":    Date,
		}
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling version info: %w", err)
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
