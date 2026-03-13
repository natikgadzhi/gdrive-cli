import json
import re
from urllib.parse import unquote

import click

# Maps Google Workspace MIME types to (export MIME type, file extension)
_DOCX_MIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
_PPTX_MIME = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

EXPORT_MAP = {
    "application/vnd.google-apps.document": (_DOCX_MIME, ".docx"),
    "application/vnd.google-apps.spreadsheet": ("text/csv", ".csv"),
    "application/vnd.google-apps.presentation": (_PPTX_MIME, ".pptx"),
}

TYPE_LABELS = {
    "application/vnd.google-apps.document": "Google Doc",
    "application/vnd.google-apps.spreadsheet": "Google Sheet",
    "application/vnd.google-apps.presentation": "Google Slides",
}

# Patterns for Google Docs/Sheets/Slides URLs
_URL_PATTERNS = [
    re.compile(r"docs\.google\.com/document/d/([a-zA-Z0-9_-]+)"),
    re.compile(r"docs\.google\.com/spreadsheets/d/([a-zA-Z0-9_-]+)"),
    re.compile(r"docs\.google\.com/presentation/d/([a-zA-Z0-9_-]+)"),
]


def parse_google_url(url: str) -> str:
    """Extract a Google file ID from a docs/sheets/slides URL. Raises ClickException on failure."""
    url = unquote(url)
    for pattern in _URL_PATTERNS:
        match = pattern.search(url)
        if match:
            return match.group(1)
    raise click.ClickException(
        f"Could not extract file ID from URL: {url}\n"
        "Supported URL formats:\n"
        "  https://docs.google.com/document/d/<ID>/...\n"
        "  https://docs.google.com/spreadsheets/d/<ID>/...\n"
        "  https://docs.google.com/presentation/d/<ID>/..."
    )


def sanitize_filename(name: str) -> str:
    """Remove characters that are problematic in filenames."""
    return re.sub(r'[/\\:*?"<>|]', "_", name).strip()


def print_json(data: object) -> None:
    click.echo(json.dumps(data, indent=2, ensure_ascii=False))
