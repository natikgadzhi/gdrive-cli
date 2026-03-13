from pathlib import Path

from googleapiclient.discovery import build

from gdrive_cli.auth import get_credentials
from gdrive_cli.config import debug_log as _log


def drive_service():
    """Build and return a Google Drive API v3 service."""
    creds = get_credentials()
    return build("drive", "v3", credentials=creds)


def export_file(file_id: str, mime_type: str, output_path: Path) -> None:
    """Export a Google Workspace file to the given MIME type and save to disk."""
    _log(f"Exporting file {file_id} as {mime_type}")
    service = drive_service()
    content = service.files().export(fileId=file_id, mimeType=mime_type).execute()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_bytes(content)
    _log(f"Saved {len(content)} bytes to {output_path}")


def get_file_metadata(file_id: str) -> dict:
    """Get file metadata (name, mimeType)."""
    _log(f"Fetching metadata for file {file_id}")
    service = drive_service()
    metadata = service.files().get(fileId=file_id, fields="id,name,mimeType,webViewLink").execute()
    _log(f"File: {metadata.get('name')} ({metadata.get('mimeType')})")
    return metadata


def search_files(query: str, max_results: int = 20) -> list[dict]:
    """Search Google Drive for files matching the query. Returns docs, sheets, and slides only."""
    service = drive_service()
    supported_types = [
        "application/vnd.google-apps.document",
        "application/vnd.google-apps.spreadsheet",
        "application/vnd.google-apps.presentation",
    ]
    type_filter = " or ".join(f"mimeType='{t}'" for t in supported_types)
    escaped = _escape_query(query)
    full_query = (
        f"({type_filter}) and "
        f"(name contains '{escaped}' or fullText contains '{escaped}')"
    )

    _log(f"Search query: {full_query}")
    results = (
        service.files()
        .list(
            q=full_query,
            pageSize=max_results,
            fields="files(id,name,mimeType,webViewLink,modifiedTime)",
            orderBy="modifiedTime desc",
        )
        .execute()
    )
    files = results.get("files", [])
    _log(f"Search returned {len(files)} result(s)")
    return files


def _escape_query(q: str) -> str:
    """Escape single quotes for Drive API query strings."""
    return q.replace("\\", "\\\\").replace("'", "\\'")
