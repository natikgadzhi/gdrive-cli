import click

from gdrive_cli import auth as auth_module
from gdrive_cli.formatting import print_json


@click.group()
def auth():
    """Manage Google OAuth2 authentication."""
    pass


@auth.command()
def login():
    """Authenticate with Google (opens browser)."""
    auth_module.login()
    print_json({"status": "ok", "message": "Successfully authenticated with Google Drive."})


@auth.command()
def status():
    """Check authentication status."""
    try:
        auth_module.get_credentials()
        print_json({"status": "ok", "message": "Authenticated and credentials are valid."})
    except click.ClickException as e:
        print_json({"status": "error", "message": str(e.message)})
