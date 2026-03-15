// Package config manages application configuration,
// including config directory resolution and credential file paths.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	debugEnabled bool
	debugMu      sync.RWMutex
)

// SetDebug enables or disables debug logging.
func SetDebug(enabled bool) {
	debugMu.Lock()
	defer debugMu.Unlock()
	debugEnabled = enabled
}

// Debug returns whether debug mode is currently enabled.
func Debug() bool {
	debugMu.RLock()
	defer debugMu.RUnlock()
	return debugEnabled
}

// DebugLog prints a formatted message to stderr when debug mode is on.
// Arguments are handled in the manner of fmt.Fprintf.
func DebugLog(msg string, args ...any) {
	if !Debug() {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] "+msg+"\n", args...)
}

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

// CredentialsFile returns the path to the OAuth client credentials JSON file.
func CredentialsFile() string {
	return filepath.Join(ConfigDir(), "credentials.json")
}

// TokenFile returns the path to the stored OAuth token JSON file.
func TokenFile() string {
	return filepath.Join(ConfigDir(), "token.json")
}

// CacheDir returns the directory for cached data.
// It checks the GDRIVE_CACHE_DIR environment variable first,
// falling back to ~/.local/share/gdrive-cli/cache.
func CacheDir() string {
	if dir := os.Getenv("GDRIVE_CACHE_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".local", "share", "gdrive-cli", "cache")
	}
	return filepath.Join(home, ".local", "share", "gdrive-cli", "cache")
}
