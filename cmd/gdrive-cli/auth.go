package main

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google Drive authentication",
	Long:  "Commands for authenticating with Google Drive and checking auth status.",
}

func init() {
	rootCmd.AddCommand(authCmd)
}
