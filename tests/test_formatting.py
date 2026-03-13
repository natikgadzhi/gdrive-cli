import pytest

from gdrive_cli.formatting import parse_google_url, sanitize_filename


class TestParseGoogleUrl:
    def test_google_doc_url(self):
        url = "https://docs.google.com/document/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit"
        assert parse_google_url(url) == "1aBcDeFgHiJkLmNoPqRsTuVwXyZ"

    def test_google_sheet_url(self):
        url = "https://docs.google.com/spreadsheets/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit#gid=0"
        assert parse_google_url(url) == "1aBcDeFgHiJkLmNoPqRsTuVwXyZ"

    def test_google_slides_url(self):
        url = "https://docs.google.com/presentation/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit"
        assert parse_google_url(url) == "1aBcDeFgHiJkLmNoPqRsTuVwXyZ"

    def test_url_with_extra_path(self):
        url = "https://docs.google.com/document/d/abc123_-XYZ/edit?tab=t.0"
        assert parse_google_url(url) == "abc123_-XYZ"

    def test_invalid_url_raises(self):
        with pytest.raises(Exception):
            parse_google_url("https://example.com/not-a-google-doc")

    def test_url_encoded(self):
        url = "https://docs.google.com/document/d/1aBcDeFg%48iJk/edit"
        result = parse_google_url(url)
        assert result == "1aBcDeFgHiJk"


class TestSanitizeFilename:
    def test_clean_name(self):
        assert sanitize_filename("My Document") == "My Document"

    def test_strips_slashes(self):
        assert sanitize_filename("Q1/Q2 Report") == "Q1_Q2 Report"

    def test_strips_special_chars(self):
        assert sanitize_filename('File: "test" <draft>') == "File_ _test_ _draft_"
