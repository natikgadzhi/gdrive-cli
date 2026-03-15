package config

import (
	"bytes"
	"fmt"
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

func TestCacheDirDefault(t *testing.T) {
	t.Setenv("GDRIVE_CACHE_DIR", "")

	dir := CacheDir()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir failed: %v", err)
	}
	expected := filepath.Join(home, ".local", "share", "gdrive-cli", "cache")
	if dir != expected {
		t.Errorf("CacheDir() = %q, want %q", dir, expected)
	}
}

func TestCacheDirEnvOverride(t *testing.T) {
	t.Setenv("GDRIVE_CACHE_DIR", "/tmp/custom-cache")

	dir := CacheDir()
	if dir != "/tmp/custom-cache" {
		t.Errorf("CacheDir() = %q, want %q", dir, "/tmp/custom-cache")
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

func TestSetDebugAndDebug(t *testing.T) {
	// Start with debug off.
	SetDebug(false)
	if Debug() {
		t.Error("Debug() = true after SetDebug(false)")
	}

	SetDebug(true)
	if !Debug() {
		t.Error("Debug() = false after SetDebug(true)")
	}

	// Clean up.
	SetDebug(false)
}

func TestDebugLogWhenEnabled(t *testing.T) {
	SetDebug(true)
	defer SetDebug(false)

	// Capture stderr by temporarily replacing os.Stderr.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	DebugLog("hello %s", "world")

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	want := "[DEBUG] hello world\n"
	if got != want {
		t.Errorf("DebugLog output = %q, want %q", got, want)
	}
}

func TestDebugLogWhenDisabled(t *testing.T) {
	SetDebug(false)

	// Capture stderr.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	DebugLog("should not appear")

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	if got != "" {
		t.Errorf("DebugLog with debug off produced output: %q", got)
	}
}

func TestDebugLogFormatting(t *testing.T) {
	SetDebug(true)
	defer SetDebug(false)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	DebugLog("count=%d name=%s", 42, "test")

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	want := fmt.Sprintf("[DEBUG] count=%d name=%s\n", 42, "test")
	if got != want {
		t.Errorf("DebugLog output = %q, want %q", got, want)
	}
}
