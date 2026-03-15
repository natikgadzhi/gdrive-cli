package formatting

import "testing"

func TestMIMEConstants(t *testing.T) {
	// Verify the constants have the expected values.
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"MIMEGoogleDoc", MIMEGoogleDoc, "application/vnd.google-apps.document"},
		{"MIMEGoogleSheet", MIMEGoogleSheet, "application/vnd.google-apps.spreadsheet"},
		{"MIMEGoogleSlides", MIMEGoogleSlides, "application/vnd.google-apps.presentation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestExportMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{MIMEGoogleDoc, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{MIMEGoogleSheet, "text/csv"},
		{MIMEGoogleSlides, "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got, ok := ExportMIME[tt.mime]
			if !ok {
				t.Fatalf("ExportMIME[%q] not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("ExportMIME[%q] = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestExportExtension(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{MIMEGoogleDoc, ".docx"},
		{MIMEGoogleSheet, ".csv"},
		{MIMEGoogleSlides, ".pptx"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got, ok := ExportExtension[tt.mime]
			if !ok {
				t.Fatalf("ExportExtension[%q] not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("ExportExtension[%q] = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestTypeLabel(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{MIMEGoogleDoc, "Google Doc"},
		{MIMEGoogleSheet, "Google Sheet"},
		{MIMEGoogleSlides, "Google Slides"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got, ok := TypeLabel[tt.mime]
			if !ok {
				t.Fatalf("TypeLabel[%q] not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("TypeLabel[%q] = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestMarkdownExportMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{MIMEGoogleDoc, "text/html"},
		{MIMEGoogleSheet, "text/csv"},
		{MIMEGoogleSlides, "text/plain"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got, ok := MarkdownExportMIME[tt.mime]
			if !ok {
				t.Fatalf("MarkdownExportMIME[%q] not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("MarkdownExportMIME[%q] = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}
