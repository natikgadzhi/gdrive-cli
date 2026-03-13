
import json
import wsgiref.simple_server

import click
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow, _RedirectWSGIApp, _WSGIRequestHandler

from gdrive_cli.config import CREDENTIALS_FILE, SCOPES, TOKEN_FILE
from gdrive_cli.config import debug_log as _log


def get_credentials() -> Credentials:
    """Load or refresh OAuth2 credentials. Raises ClickException if not authenticated."""
    creds: Credentials | None = None

    if TOKEN_FILE.exists():
        creds = Credentials.from_authorized_user_file(str(TOKEN_FILE), SCOPES)

    if creds and creds.valid:
        return creds

    if creds and creds.expired and creds.refresh_token:
        try:
            creds.refresh(Request())
            _save_token(creds)
            return creds
        except Exception as e:
            raise click.ClickException(f"Failed to refresh token: {e}. Run `gdrive-cli auth login` to re-authenticate.")

    raise click.ClickException(
        "Not authenticated. Run `gdrive-cli auth login` first.\n"
        f"Ensure your OAuth credentials are at {CREDENTIALS_FILE}"
    )


def login() -> Credentials:
    """Run the OAuth2 installed-app flow. Opens a browser for consent."""
    if not CREDENTIALS_FILE.exists():
        raise click.ClickException(
            f"OAuth client credentials not found at {CREDENTIALS_FILE}\n"
            "Download your OAuth 2.0 Client ID (Desktop app) JSON from:\n"
            "  https://console.cloud.google.com/apis/credentials\n"
            f"Save it as {CREDENTIALS_FILE}"
        )

    TOKEN_FILE.parent.mkdir(parents=True, exist_ok=True)

    # Read credentials to check type
    with open(CREDENTIALS_FILE) as f:
        client_config = json.load(f)

    client_type = "installed" if "installed" in client_config else "web" if "web" in client_config else "unknown"
    _log(f"credentials.json client type: {client_type}")

    if client_type == "web":
        raise click.ClickException(
            "Your credentials.json is for a 'Web application' OAuth client.\n"
            "Create a 'Desktop app' OAuth client instead at:\n"
            "  https://console.cloud.google.com/apis/credentials"
        )

    client_info = client_config.get("installed", {})
    _log(f"client_id: {client_info.get('client_id', 'MISSING')}")
    _log(f"redirect_uris in credentials: {client_info.get('redirect_uris', 'MISSING')}")
    _log(f"auth_uri: {client_info.get('auth_uri', 'MISSING')}")
    _log(f"token_uri: {client_info.get('token_uri', 'MISSING')}")

    flow = InstalledAppFlow.from_client_secrets_file(str(CREDENTIALS_FILE), SCOPES)
    _log(f"scopes: {SCOPES}")

    # Manual local server flow so we can handle multiple requests
    # (browsers often send favicon/preflight requests that eat the single handle_request)
    host, port = "localhost", 8085

    class _DebugWSGIApp(_RedirectWSGIApp):
        """Wraps the redirect WSGI app to log every incoming request."""

        def __call__(self, environ, start_response):
            method = environ.get("REQUEST_METHOD", "?")
            path = environ.get("PATH_INFO", "?")
            query = environ.get("QUERY_STRING", "")
            remote = environ.get("REMOTE_ADDR", "?")
            _log(f"WSGI request: {method} {path}{'?' + query if query else ''} from {remote}")
            result = super().__call__(environ, start_response)
            _log(f"last_request_uri after this request: {self.last_request_uri}")
            return result

    wsgi_app = _DebugWSGIApp("Authentication successful! You can close this tab.")
    wsgiref.simple_server.WSGIServer.allow_reuse_address = True

    _log(f"Starting local server on {host}:{port}...")
    try:
        local_server = wsgiref.simple_server.make_server(
            host, port, wsgi_app, handler_class=_WSGIRequestHandler
        )
    except OSError as e:
        raise click.ClickException(f"Could not start local server on {host}:{port}: {e}")

    _log(f"Local server listening on {host}:{local_server.server_port}")

    flow.redirect_uri = f"http://{host}:{port}/"
    _log(f"redirect_uri set to: {flow.redirect_uri}")

    auth_url, state = flow.authorization_url(access_type="offline", prompt="consent")
    _log(f"OAuth state: {state}")
    _log(f"Full auth URL:\n  {auth_url}")

    click.echo("Opening browser for authentication...")
    click.echo(f"If the browser doesn't open, visit:\n  {auth_url}")

    import webbrowser

    webbrowser.open(auth_url, new=1, autoraise=True)

    # Handle requests in a loop until we get the auth code
    _log("Waiting for OAuth callback (120s timeout per request)...")
    local_server.timeout = 120
    request_count = 0
    while True:
        _log(f"Calling handle_request() (request #{request_count + 1})...")
        local_server.handle_request()
        request_count += 1
        _log(f"handle_request() returned. Total requests handled: {request_count}")
        _log(f"last_request_uri: {wsgi_app.last_request_uri}")
        if wsgi_app.last_request_uri:
            break
        if request_count > 20:
            raise click.ClickException("Handled 20 requests without receiving OAuth callback. Giving up.")

    _log(f"Got authorization response after {request_count} request(s)")
    authorization_response = wsgi_app.last_request_uri.replace("http", "https")
    _log("Exchanging code for token...")

    try:
        flow.fetch_token(authorization_response=authorization_response)
    except Exception as e:
        _log(f"Token exchange failed: {e}")
        raise click.ClickException(f"Token exchange failed: {e}")

    _log("Token exchange successful!")
    local_server.server_close()

    creds = flow.credentials
    _save_token(creds)
    _log(f"Token saved to {TOKEN_FILE}")
    return creds


def _save_token(creds: Credentials) -> None:
    TOKEN_FILE.parent.mkdir(parents=True, exist_ok=True)
    TOKEN_FILE.write_text(creds.to_json())
