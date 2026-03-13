import os
from pathlib import Path

import click

CONFIG_DIR = Path(os.environ.get("GDRIVE_CONFIG_DIR", Path.home() / ".config" / "gdrive-cli"))
CREDENTIALS_FILE = CONFIG_DIR / "credentials.json"
TOKEN_FILE = CONFIG_DIR / "token.json"

SCOPES = [
    "https://www.googleapis.com/auth/drive.readonly",
]


def debug_log(msg: str) -> None:
    """Log a debug message to stderr if --debug is enabled."""
    ctx = click.get_current_context(silent=True)
    if ctx and ctx.find_root().obj and ctx.find_root().obj.get("debug"):
        click.echo(f"[DEBUG] {msg}", err=True)
