from pathlib import Path

import click

from gdrive_cli.api import export_file, get_file_metadata
from gdrive_cli.formatting import EXPORT_MAP, TYPE_LABELS, parse_google_url, print_json, sanitize_filename


@click.command()
@click.argument("url")
@click.option("--output", "-o", type=click.Path(), default=None, help="Output file path. Auto-generated if omitted.")
@click.option("--dir", "-d", "output_dir", type=click.Path(), default=".", help="Output directory.")
def fetch(url: str, output: str | None, output_dir: str):
    """Fetch a Google Doc, Sheet, or Slides presentation and save to disk.

    URL should be a direct link to a Google Docs, Sheets, or Slides file.
    Docs are saved as .docx, Sheets as .csv, Slides as .pptx.
    """
    file_id = parse_google_url(url)
    metadata = get_file_metadata(file_id)
    mime_type = metadata["mimeType"]

    if mime_type not in EXPORT_MAP:
        raise click.ClickException(
            f"Unsupported file type: {mime_type}\n"
            f"Supported types: {', '.join(TYPE_LABELS.values())}"
        )

    export_mime, extension = EXPORT_MAP[mime_type]

    if output:
        output_path = Path(output)
    else:
        safe_name = sanitize_filename(metadata["name"])
        output_path = Path(output_dir) / f"{safe_name}{extension}"

    export_file(file_id, export_mime, output_path)

    print_json({
        "status": "ok",
        "file_id": file_id,
        "name": metadata["name"],
        "type": TYPE_LABELS.get(mime_type, mime_type),
        "saved_to": str(output_path),
    })
