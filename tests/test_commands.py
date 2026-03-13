import json
from unittest.mock import patch

from click.testing import CliRunner

from gdrive_cli.commands.fetch import fetch
from gdrive_cli.commands.search import search


class TestSearchCommand:
    def test_search_returns_results(self):
        mock_files = [
            {
                "name": "Test Doc",
                "mimeType": "application/vnd.google-apps.document",
                "webViewLink": "https://docs.google.com/document/d/123/edit",
                "modifiedTime": "2026-01-01T00:00:00Z",
            },
        ]

        runner = CliRunner()
        with patch("gdrive_cli.commands.search.search_files", return_value=mock_files):
            result = runner.invoke(search, ["test query"])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert data["query"] == "test query"
        assert data["count"] == 1
        assert data["results"][0]["name"] == "Test Doc"
        assert data["results"][0]["type"] == "Google Doc"

    def test_search_empty_results(self):
        runner = CliRunner()
        with patch("gdrive_cli.commands.search.search_files", return_value=[]):
            result = runner.invoke(search, ["nonexistent"])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert data["count"] == 0
        assert data["results"] == []


class TestFetchCommand:
    def test_fetch_google_doc(self, tmp_path):
        mock_metadata = {
            "id": "abc123",
            "name": "Test Document",
            "mimeType": "application/vnd.google-apps.document",
            "webViewLink": "https://docs.google.com/document/d/abc123/edit",
        }

        runner = CliRunner()
        with (
            patch("gdrive_cli.commands.fetch.get_file_metadata", return_value=mock_metadata),
            patch("gdrive_cli.commands.fetch.export_file"),
        ):
            result = runner.invoke(fetch, [
                "https://docs.google.com/document/d/abc123/edit",
                "--dir", str(tmp_path),
            ])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert data["status"] == "ok"
        assert data["name"] == "Test Document"
        assert data["type"] == "Google Doc"
        assert data["saved_to"].endswith(".docx")

    def test_fetch_google_sheet(self, tmp_path):
        mock_metadata = {
            "id": "xyz789",
            "name": "Budget 2026",
            "mimeType": "application/vnd.google-apps.spreadsheet",
            "webViewLink": "https://docs.google.com/spreadsheets/d/xyz789/edit",
        }

        runner = CliRunner()
        with (
            patch("gdrive_cli.commands.fetch.get_file_metadata", return_value=mock_metadata),
            patch("gdrive_cli.commands.fetch.export_file"),
        ):
            result = runner.invoke(fetch, [
                "https://docs.google.com/spreadsheets/d/xyz789/edit",
                "--dir", str(tmp_path),
            ])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert data["type"] == "Google Sheet"
        assert data["saved_to"].endswith(".csv")

    def test_fetch_invalid_url(self):
        runner = CliRunner()
        result = runner.invoke(fetch, ["https://example.com/not-a-doc"])
        assert result.exit_code != 0
