package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// --- Token save/load round-trip tests ---

func TestSaveAndLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	token := &oauth2.Token{
		AccessToken:  "access-123",
		TokenType:    "Bearer",
		RefreshToken: "refresh-456",
		Expiry:       time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
	}

	if err := SaveToken(token, path); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	loaded, err := LoadToken(path)
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, token.RefreshToken)
	}
	if loaded.TokenType != token.TokenType {
		t.Errorf("TokenType = %q, want %q", loaded.TokenType, token.TokenType)
	}
	if !loaded.Expiry.Equal(token.Expiry) {
		t.Errorf("Expiry = %v, want %v", loaded.Expiry, token.Expiry)
	}
}

func TestSaveTokenCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "token.json")

	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}

	if err := SaveToken(token, path); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("token file was not created: %v", err)
	}
}

func TestSaveTokenFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	token := &oauth2.Token{AccessToken: "test"}
	if err := SaveToken(token, path); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Token file should be readable/writable by owner only (0600).
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("token file permissions = %o, want 0600", perm)
	}
}

func TestLoadTokenMissingFile(t *testing.T) {
	_, err := LoadToken("/nonexistent/path/token.json")
	if err == nil {
		t.Fatal("LoadToken should return error for missing file")
	}
}

func TestLoadTokenInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	os.WriteFile(path, []byte("not json"), 0600)

	_, err := LoadToken(path)
	if err == nil {
		t.Fatal("LoadToken should return error for invalid JSON")
	}
}

// --- Credentials.json parsing tests ---

func TestLoadOAuthConfigValidDesktop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	creds := map[string]any{
		"installed": map[string]any{
			"client_id":     "test-client-id.apps.googleusercontent.com",
			"client_secret": "test-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"http://localhost"},
		},
	}

	data, _ := json.Marshal(creds)
	os.WriteFile(path, data, 0600)

	cfg, err := LoadOAuthConfig(path)
	if err != nil {
		t.Fatalf("LoadOAuthConfig failed: %v", err)
	}

	if cfg.ClientID != "test-client-id.apps.googleusercontent.com" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "test-client-id.apps.googleusercontent.com")
	}
	if cfg.ClientSecret != "test-secret" {
		t.Errorf("ClientSecret = %q, want %q", cfg.ClientSecret, "test-secret")
	}

	// Should have the drive.readonly scope.
	foundScope := false
	for _, s := range cfg.Scopes {
		if s == DriveReadonlyScope {
			foundScope = true
			break
		}
	}
	if !foundScope {
		t.Errorf("expected scope %q in config scopes: %v", DriveReadonlyScope, cfg.Scopes)
	}
}

func TestLoadOAuthConfigWebClientRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	creds := map[string]any{
		"web": map[string]any{
			"client_id":     "web-client-id.apps.googleusercontent.com",
			"client_secret": "web-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"https://example.com/callback"},
		},
	}

	data, _ := json.Marshal(creds)
	os.WriteFile(path, data, 0600)

	_, err := LoadOAuthConfig(path)
	if err == nil {
		t.Fatal("LoadOAuthConfig should reject web client credentials")
	}

	// Check error message mentions Web application.
	errMsg := err.Error()
	if !(contains(errMsg, "Web application") || contains(errMsg, "web")) {
		t.Errorf("error message should mention Web application, got: %s", errMsg)
	}
}

func TestLoadOAuthConfigMissingFile(t *testing.T) {
	_, err := LoadOAuthConfig("/nonexistent/credentials.json")
	if err == nil {
		t.Fatal("LoadOAuthConfig should return error for missing file")
	}

	errMsg := err.Error()
	if !contains(errMsg, "not found") {
		t.Errorf("error message should mention 'not found', got: %s", errMsg)
	}
	if !contains(errMsg, "console.cloud.google.com") {
		t.Errorf("error message should include Google Cloud console link, got: %s", errMsg)
	}
}

func TestLoadOAuthConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	os.WriteFile(path, []byte("not json"), 0600)

	_, err := LoadOAuthConfig(path)
	if err == nil {
		t.Fatal("LoadOAuthConfig should return error for invalid JSON")
	}
}

func TestLoadOAuthConfigEmptyJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	os.WriteFile(path, []byte("{}"), 0600)

	_, err := LoadOAuthConfig(path)
	if err == nil {
		t.Fatal("LoadOAuthConfig should return error for empty credentials")
	}
}

// --- IsAuthenticated tests ---

func TestIsAuthenticatedWithValidToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GDRIVE_CONFIG_DIR", dir)

	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}

	tokenPath := filepath.Join(dir, "token.json")
	if err := SaveToken(token, tokenPath); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	if !IsAuthenticated(dir) {
		t.Error("IsAuthenticated should return true when a valid token file exists")
	}
}

func TestIsAuthenticatedWithNoToken(t *testing.T) {
	dir := t.TempDir()

	if IsAuthenticated(dir) {
		t.Error("IsAuthenticated should return false when no token file exists")
	}
}

func TestIsAuthenticatedWithInvalidToken(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	os.WriteFile(tokenPath, []byte("not valid json"), 0600)

	if IsAuthenticated(dir) {
		t.Error("IsAuthenticated should return false when token file contains invalid JSON")
	}
}

// --- Login flow test with mock OAuth server ---

func TestLoginFlowWithMockServer(t *testing.T) {
	// Create a mock token endpoint that returns a valid token.
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "mock-access-token",
			"token_type":    "Bearer",
			"refresh_token": "mock-refresh-token",
			"expires_in":    3600,
		})
	}))
	defer tokenServer.Close()

	// Create a mock auth endpoint (not actually used in the test since we
	// simulate the callback directly, but needed for the config).
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer authServer.Close()

	// Write credentials.json with the mock server URLs.
	dir := t.TempDir()
	credsPath := filepath.Join(dir, "credentials.json")
	creds := map[string]any{
		"installed": map[string]any{
			"client_id":     "test-client-id.apps.googleusercontent.com",
			"client_secret": "test-secret",
			"auth_uri":      authServer.URL,
			"token_uri":     tokenServer.URL,
			"redirect_uris": []string{"http://localhost"},
		},
	}
	data, _ := json.Marshal(creds)
	os.WriteFile(credsPath, data, 0600)

	// We can't easily test the full Login flow (it opens a browser and
	// starts a server on a specific port), so instead we test the
	// individual components that make up the flow.

	// Test 1: LoadOAuthConfig works with our test credentials.
	cfg, err := LoadOAuthConfig(credsPath)
	if err != nil {
		t.Fatalf("LoadOAuthConfig failed: %v", err)
	}
	if cfg.ClientID != "test-client-id.apps.googleusercontent.com" {
		t.Errorf("unexpected client ID: %s", cfg.ClientID)
	}

	// Test 2: Simulate exchanging a code for a token using the mock server.
	cfg.Endpoint.TokenURL = tokenServer.URL
	cfg.RedirectURL = "http://localhost:8085/"

	token, err := cfg.Exchange(t.Context(), "mock-auth-code")
	if err != nil {
		t.Fatalf("token exchange failed: %v", err)
	}
	if token.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "mock-access-token")
	}

	// Test 3: Save and reload the token.
	tokenPath := filepath.Join(dir, "token.json")
	if err := SaveToken(token, tokenPath); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	loaded, err := LoadToken(tokenPath)
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}
	if loaded.AccessToken != "mock-access-token" {
		t.Errorf("reloaded AccessToken = %q, want %q", loaded.AccessToken, "mock-access-token")
	}
}

// --- GetCredentials tests ---

func TestGetCredentialsNotAuthenticated(t *testing.T) {
	dir := t.TempDir()

	// Write valid credentials.json but no token.
	credsPath := filepath.Join(dir, "credentials.json")
	creds := map[string]any{
		"installed": map[string]any{
			"client_id":     "test-client-id.apps.googleusercontent.com",
			"client_secret": "test-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"http://localhost"},
		},
	}
	data, _ := json.Marshal(creds)
	os.WriteFile(credsPath, data, 0600)

	_, _, err := GetCredentials(dir)
	if err == nil {
		t.Fatal("GetCredentials should return error when not authenticated")
	}
	if !contains(err.Error(), "not authenticated") {
		t.Errorf("error should mention 'not authenticated', got: %s", err.Error())
	}
}

func TestGetCredentialsWithValidToken(t *testing.T) {
	dir := t.TempDir()

	// Write valid credentials.json.
	credsPath := filepath.Join(dir, "credentials.json")
	creds := map[string]any{
		"installed": map[string]any{
			"client_id":     "test-client-id.apps.googleusercontent.com",
			"client_secret": "test-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"http://localhost"},
		},
	}
	data, _ := json.Marshal(creds)
	os.WriteFile(credsPath, data, 0600)

	// Write a valid, non-expired token.
	tokenPath := filepath.Join(dir, "token.json")
	token := &oauth2.Token{
		AccessToken:  "valid-access-token",
		TokenType:    "Bearer",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	SaveToken(token, tokenPath)

	got, cfg, err := GetCredentials(dir)
	if err != nil {
		t.Fatalf("GetCredentials failed: %v", err)
	}
	if got.AccessToken != "valid-access-token" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "valid-access-token")
	}
	if cfg == nil {
		t.Error("OAuth config should not be nil")
	}
}

func TestGetCredentialsMissingCredentialsFile(t *testing.T) {
	dir := t.TempDir()
	// No credentials.json at all.

	_, _, err := GetCredentials(dir)
	if err == nil {
		t.Fatal("GetCredentials should return error when credentials.json is missing")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %s", err.Error())
	}
}

// --- Helper ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Silence the unused import warning for fmt if we need it later.
var _ = fmt.Sprintf
