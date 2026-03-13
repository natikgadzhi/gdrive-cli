import click

from gdrive_cli.api import search_files
from gdrive_cli.formatting import TYPE_LABELS, print_json


@click.command()
@click.argument("query")
@click.option("--count", "-n", default=20, help="Maximum number of results.")
def search(query: str, count: int):
    """Search Google Drive for docs, sheets, and slides matching QUERY."""
    files = search_files(query, max_results=count)

    results = []
    for f in files:
        results.append({
            "name": f["name"],
            "type": TYPE_LABELS.get(f["mimeType"], f["mimeType"]),
            "url": f.get("webViewLink", ""),
            "modified": f.get("modifiedTime", ""),
        })

    print_json({"query": query, "count": len(results), "results": results})
