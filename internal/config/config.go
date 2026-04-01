// Package config manages application configuration,
// including config directory resolution and credential file paths.
package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the directory for configuration files.
// It checks the GDRIVE_CONFIG_DIR environment variable first,
// falling back to ~/.config/gdrive-cli.
func ConfigDir() string {
	if dir := os.Getenv("GDRIVE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to current directory if home is unresolvable.
		return filepath.Join(".", ".config", "gdrive-cli")
	}
	return filepath.Join(home, ".config", "gdrive-cli")
}

// CredentialsFileIn returns the path to the OAuth client credentials JSON file
// within the given config directory.
func CredentialsFileIn(configDir string) string {
	return filepath.Join(configDir, "credentials.json")
}

// CredentialsFile returns the path to the OAuth client credentials JSON file
// in the default config directory.
func CredentialsFile() string {
	return CredentialsFileIn(ConfigDir())
}

// TokenFileIn returns the path to the stored OAuth token JSON file
// within the given config directory.
func TokenFileIn(configDir string) string {
	return filepath.Join(configDir, "token.json")
}

// TokenFile returns the path to the stored OAuth token JSON file
// in the default config directory.
func TokenFile() string {
	return TokenFileIn(ConfigDir())
}

