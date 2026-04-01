// Package auth handles Google OAuth2 authentication flows,
// including the installed-app flow, token storage, and refresh.
package auth

import (
	"context"
	"fmt"

	"github.com/natikgadzhi/cli-kit/debug"
	"golang.org/x/oauth2"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

// GetCredentials loads the saved OAuth2 token, refreshes it if expired,
// and returns the token along with the OAuth2 config.
// Returns an error if not authenticated or if the token cannot be refreshed.
func GetCredentials(configDir string) (*oauth2.Token, *oauth2.Config, error) {
	tokenPath := config.TokenFileIn(configDir)
	credsPath := config.CredentialsFileIn(configDir)

	oauthConfig, err := LoadOAuthConfig(credsPath)
	if err != nil {
		return nil, nil, err
	}

	token, err := LoadToken(tokenPath)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"not authenticated. Run `gdrive-cli auth login` first.\n"+
				"Ensure your OAuth credentials are at %s", credsPath,
		)
	}

	// If token is still valid, return it as-is.
	if token.Valid() {
		debug.Log("Token is valid, no refresh needed")
		return token, oauthConfig, nil
	}

	// Try to refresh the token.
	debug.Log("Token expired, attempting refresh")
	tokenSource := oauthConfig.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to refresh token: %w. Run `gdrive-cli auth login` to re-authenticate", err,
		)
	}

	// Save the refreshed token if it changed.
	if newToken.AccessToken != token.AccessToken {
		debug.Log("Token was refreshed, saving new token")
		if err := SaveToken(newToken, tokenPath); err != nil {
			debug.Log("Warning: failed to save refreshed token: %v", err)
		}
	}

	return newToken, oauthConfig, nil
}

// IsAuthenticated returns true if a token file exists and can be loaded.
// It does not attempt to refresh the token or validate it against Google.
func IsAuthenticated(configDir string) bool {
	_, err := LoadToken(config.TokenFileIn(configDir))
	return err == nil
}
