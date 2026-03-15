package formatting

import (
	"strings"
	"testing"
)

func TestParseGoogleURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "Google Doc URL",
			url:  "https://docs.google.com/document/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit",
			want: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ",
		},
		{
			name: "Google Sheet URL",
			url:  "https://docs.google.com/spreadsheets/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit#gid=0",
			want: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ",
		},
		{
			name: "Google Slides URL",
			url:  "https://docs.google.com/presentation/d/1aBcDeFgHiJkLmNoPqRsTuVwXyZ/edit",
			want: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ",
		},
		{
			name: "URL with hyphens and underscores in ID",
			url:  "https://docs.google.com/document/d/abc-123_DEF/edit",
			want: "abc-123_DEF",
		},
		{
			name: "URL-encoded document URL",
			url:  "https://docs.google.com/document/d/1aBcDeFg%48iJkLmNoPqRsTuVwXyZ/edit",
			want: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ",
		},
		{
			name: "URL-encoded spreadsheet URL",
			url:  "https%3A%2F%2Fdocs.google.com%2Fspreadsheets%2Fd%2F1aBcDeFgHiJkLmNoPqRsTuVwXyZ%2Fedit",
			want: "1aBcDeFgHiJkLmNoPqRsTuVwXyZ",
		},
		{
			name:    "invalid URL - random website",
			url:     "https://example.com/some/path",
			wantErr: true,
		},
		{
			name:    "invalid URL - empty string",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL - Google Drive but not a doc",
			url:     "https://drive.google.com/file/d/1aBcDeFg/view",
			wantErr: true,
		},
		{
			name:    "invalid URL - no ID after /d/",
			url:     "https://docs.google.com/document/d/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGoogleURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseGoogleURL(%q) expected error, got %q", tt.url, got)
				}
				// Verify the error message includes supported formats.
				errMsg := err.Error()
				if !strings.Contains(errMsg, "docs.google.com/document/d/") {
					t.Errorf("error message should list supported formats, got: %s", errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseGoogleURL(%q) unexpected error: %v", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("ParseGoogleURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
