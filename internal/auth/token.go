package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// DriveReadonlyScope is the OAuth2 scope for read-only Google Drive access.
const DriveReadonlyScope = "https://www.googleapis.com/auth/drive.readonly"

// credentialsFileJSON represents the structure of a Google OAuth2 credentials.json file.
type credentialsFileJSON struct {
	Installed *credentialsConfig `json:"installed"`
	Web       *credentialsConfig `json:"web"`
}

// credentialsConfig holds the OAuth2 client configuration fields.
type credentialsConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

// SaveToken writes an OAuth2 token to the given path as JSON.
// Parent directories are created automatically if they don't exist.
func SaveToken(token *oauth2.Token, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadToken reads an OAuth2 token from the given path.
func LoadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token file: %w", err)
	}

	return &token, nil
}

// LoadOAuthConfig reads a Google OAuth2 credentials.json file and returns
// an oauth2.Config. It validates that the credentials are for a Desktop
// (installed) application, not a Web application.
func LoadOAuthConfig(credentialsPath string) (*oauth2.Config, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"OAuth client credentials not found at %s\n"+
					"Download your OAuth 2.0 Client ID (Desktop app) JSON from:\n"+
					"  https://console.cloud.google.com/apis/credentials\n"+
					"Save it as %s",
				credentialsPath, credentialsPath,
			)
		}
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	var creds credentialsFileJSON
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials file: %w", err)
	}

	if creds.Web != nil && creds.Installed == nil {
		return nil, fmt.Errorf(
			"your credentials.json is for a 'Web application' OAuth client.\n" +
				"Create a 'Desktop app' OAuth client instead at:\n" +
				"  https://console.cloud.google.com/apis/credentials",
		)
	}

	if creds.Installed == nil {
		return nil, fmt.Errorf(
			"credentials.json does not contain an 'installed' or 'web' client configuration.\n" +
				"Download a Desktop app OAuth 2.0 Client ID from:\n" +
				"  https://console.cloud.google.com/apis/credentials",
		)
	}

	// Use google's helper to parse the credentials file, which handles
	// endpoint resolution and other details automatically.
	config, err := google.ConfigFromJSON(data, DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("building OAuth config: %w", err)
	}

	return config, nil
}
