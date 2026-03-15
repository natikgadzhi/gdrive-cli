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

func TestGetExportMIME(t *testing.T) {
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
			got, ok := GetExportMIME(tt.mime)
			if !ok {
				t.Fatalf("GetExportMIME(%q) not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("GetExportMIME(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestGetExportMIME_Unknown(t *testing.T) {
	_, ok := GetExportMIME("application/pdf")
	if ok {
		t.Error("GetExportMIME(application/pdf) should return false for unknown MIME type")
	}
}

func TestGetExportExtension(t *testing.T) {
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
			got, ok := GetExportExtension(tt.mime)
			if !ok {
				t.Fatalf("GetExportExtension(%q) not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("GetExportExtension(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestGetExportExtension_Unknown(t *testing.T) {
	_, ok := GetExportExtension("application/pdf")
	if ok {
		t.Error("GetExportExtension(application/pdf) should return false for unknown MIME type")
	}
}

func TestGetTypeLabel(t *testing.T) {
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
			got, ok := GetTypeLabel(tt.mime)
			if !ok {
				t.Fatalf("GetTypeLabel(%q) not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("GetTypeLabel(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestGetTypeLabel_Unknown(t *testing.T) {
	_, ok := GetTypeLabel("application/pdf")
	if ok {
		t.Error("GetTypeLabel(application/pdf) should return false for unknown MIME type")
	}
}

func TestGetMarkdownExportMIME(t *testing.T) {
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
			got, ok := GetMarkdownExportMIME(tt.mime)
			if !ok {
				t.Fatalf("GetMarkdownExportMIME(%q) not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("GetMarkdownExportMIME(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestGetMarkdownExportMIME_Unknown(t *testing.T) {
	_, ok := GetMarkdownExportMIME("application/pdf")
	if ok {
		t.Error("GetMarkdownExportMIME(application/pdf) should return false for unknown MIME type")
	}
}

func TestSupportedMIMETypes(t *testing.T) {
	types := SupportedMIMETypes()
	if len(types) != 3 {
		t.Fatalf("SupportedMIMETypes() returned %d types, want 3", len(types))
	}
	want := map[string]bool{
		MIMEGoogleDoc:    true,
		MIMEGoogleSheet:  true,
		MIMEGoogleSlides: true,
	}
	for _, mt := range types {
		if !want[mt] {
			t.Errorf("SupportedMIMETypes() contains unexpected type %q", mt)
		}
	}
}
