package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirDefault(t *testing.T) {
	// Ensure the env var is not set so we get the default.
	t.Setenv("GDRIVE_CONFIG_DIR", "")

	dir := ConfigDir()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir failed: %v", err)
	}
	expected := filepath.Join(home, ".config", "gdrive-cli")
	if dir != expected {
		t.Errorf("ConfigDir() = %q, want %q", dir, expected)
	}
}

func TestConfigDirEnvOverride(t *testing.T) {
	t.Setenv("GDRIVE_CONFIG_DIR", "/tmp/custom-config")

	dir := ConfigDir()
	if dir != "/tmp/custom-config" {
		t.Errorf("ConfigDir() = %q, want %q", dir, "/tmp/custom-config")
	}
}

func TestCredentialsFile(t *testing.T) {
	t.Setenv("GDRIVE_CONFIG_DIR", "/tmp/test-config")

	got := CredentialsFile()
	want := filepath.Join("/tmp/test-config", "credentials.json")
	if got != want {
		t.Errorf("CredentialsFile() = %q, want %q", got, want)
	}
}

func TestTokenFile(t *testing.T) {
	t.Setenv("GDRIVE_CONFIG_DIR", "/tmp/test-config")

	got := TokenFile()
	want := filepath.Join("/tmp/test-config", "token.json")
	if got != want {
		t.Errorf("TokenFile() = %q, want %q", got, want)
	}
}

func TestCredentialsFileDefaultSuffix(t *testing.T) {
	t.Setenv("GDRIVE_CONFIG_DIR", "")

	got := CredentialsFile()
	if !strings.HasSuffix(got, filepath.Join("gdrive-cli", "credentials.json")) {
		t.Errorf("CredentialsFile() = %q, expected suffix %q", got, "gdrive-cli/credentials.json")
	}
}

func TestTokenFileDefaultSuffix(t *testing.T) {
	t.Setenv("GDRIVE_CONFIG_DIR", "")

	got := TokenFile()
	if !strings.HasSuffix(got, filepath.Join("gdrive-cli", "token.json")) {
		t.Errorf("TokenFile() = %q, expected suffix %q", got, "gdrive-cli/token.json")
	}
}

