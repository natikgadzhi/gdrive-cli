package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"golang.org/x/oauth2"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

const (
	// callbackPort is the port the local HTTP server listens on for the OAuth redirect.
	callbackPort = 8085
	// maxRequests is the maximum number of HTTP requests to handle before giving up.
	// Browsers may send favicon or preflight requests before the actual callback.
	maxRequests = 20
	// serverTimeout is the read/write timeout for the callback server.
	serverTimeout = 120 * time.Second
)

// Login runs the full OAuth2 installed-app flow:
//  1. Reads and validates credentials.json
//  2. Starts a local HTTP server on port 8085
//  3. Opens the browser to the Google consent page
//  4. Handles the redirect callback to extract the auth code
//  5. Exchanges the code for a token
//  6. Saves the token to token.json
func Login(configDir string) error {
	credsPath := config.CredentialsFileIn(configDir)
	tokenPath := config.TokenFileIn(configDir)

	oauthConfig, err := LoadOAuthConfig(credsPath)
	if err != nil {
		return err
	}

	config.DebugLog("client_id: %s", oauthConfig.ClientID)
	config.DebugLog("scopes: %v", oauthConfig.Scopes)

	// Set the redirect URI to our local callback server.
	redirectURI := fmt.Sprintf("http://localhost:%d/", callbackPort)
	oauthConfig.RedirectURL = redirectURI
	config.DebugLog("redirect_uri set to: %s", redirectURI)

	// Generate a random state parameter for CSRF protection.
	state, err := randomState()
	if err != nil {
		return fmt.Errorf("generating state parameter: %w", err)
	}

	// Build the authorization URL.
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	config.DebugLog("OAuth state: %s", state)
	config.DebugLog("Full auth URL:\n  %s", authURL)

	// Start the local callback server.
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	var requestCount atomic.Int32
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		config.DebugLog("HTTP request #%d: %s %s from %s", count, r.Method, r.URL.String(), r.RemoteAddr)

		// Enforce max request limit.
		if int(count) > maxRequests {
			w.WriteHeader(http.StatusServiceUnavailable)
			errChan <- fmt.Errorf("handled %d requests without receiving OAuth callback, giving up", maxRequests)
			return
		}

		// Check for the authorization code in the query parameters.
		code := r.URL.Query().Get("code")
		returnedState := r.URL.Query().Get("state")
		errParam := r.URL.Query().Get("error")

		if errParam != "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>Error: %s</p><p>You can close this tab.</p></body></html>", html.EscapeString(errParam))
			errChan <- fmt.Errorf("OAuth error: %s", errParam)
			return
		}

		if code == "" {
			// This is likely a favicon or preflight request; ignore it.
			config.DebugLog("No auth code in request #%d, ignoring", count)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Validate state parameter.
		if returnedState != state {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h2>Authentication failed</h2><p>State mismatch.</p><p>You can close this tab.</p></body></html>")
			errChan <- fmt.Errorf("OAuth state mismatch: expected %q, got %q", state, returnedState)
			return
		}

		// Success! Return the code.
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab.</p></body></html>")
		config.DebugLog("Received auth code after %d request(s)", count)
		codeChan <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", callbackPort))
	if err != nil {
		return fmt.Errorf("could not start local server on localhost:%d: %w", callbackPort, err)
	}

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
	}

	config.DebugLog("Local server listening on localhost:%d", callbackPort)

	// Run server in a goroutine.
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", serveErr)
		}
	}()

	// Open the browser.
	fmt.Fprintf(os.Stderr, "Opening browser for authentication...\n")
	fmt.Fprintf(os.Stderr, "If the browser doesn't open, visit:\n  %s\n", authURL)
	openBrowser(authURL)

	// Wait for the auth code or an error.
	config.DebugLog("Waiting for OAuth callback...")
	var code string
	select {
	case code = <-codeChan:
		config.DebugLog("Got authorization code")
	case err := <-errChan:
		_ = server.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		_ = server.Shutdown(context.Background())
		return fmt.Errorf("timed out waiting for OAuth callback after 5 minutes")
	}

	// Shut down the server.
	_ = server.Shutdown(context.Background())

	// Exchange the authorization code for a token.
	config.DebugLog("Exchanging code for token...")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}
	config.DebugLog("Token exchange successful!")

	// Save the token.
	if err := SaveToken(token, tokenPath); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}
	config.DebugLog("Token saved to %s", tokenPath)

	return nil
}

// randomState generates a random hex string for OAuth2 state parameter.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// openBrowser opens the given URL in the default browser.
// It does not return an error; failure to open is non-fatal since the user
// can manually navigate to the URL.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		config.DebugLog("Unsupported platform for browser opening: %s", runtime.GOOS)
		return
	}
	if err := cmd.Start(); err != nil {
		config.DebugLog("Failed to open browser: %v", err)
	}
}
