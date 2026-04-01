package api

import (
	"context"
	"net/http"

	"github.com/natikgadzhi/cli-kit/debug"
	"golang.org/x/oauth2"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// NewDriveService builds an authenticated Google Drive v3 service
// using the provided OAuth2 token and config.
func NewDriveService(token *oauth2.Token, oauthConfig *oauth2.Config) (*drive.Service, error) {
	debug.Log("Creating Drive service with OAuth token")
	client := oauthConfig.Client(context.Background(), token)
	return drive.NewService(context.Background(), option.WithHTTPClient(client))
}

// NewDriveServiceWithClient builds a Google Drive v3 service using the
// provided HTTP client. This is useful for testing with httptest servers.
func NewDriveServiceWithClient(client *http.Client) (*drive.Service, error) {
	return drive.NewService(context.Background(), option.WithHTTPClient(client))
}
