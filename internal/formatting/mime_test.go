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

func TestDefaultExportFormat(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{MIMEGoogleDoc, "docx"},
		{MIMEGoogleSheet, "csv"},
		{MIMEGoogleSlides, "pptx"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got, ok := DefaultExportFormat(tt.mime)
			if !ok {
				t.Fatalf("DefaultExportFormat(%q) not found", tt.mime)
			}
			if got != tt.want {
				t.Errorf("DefaultExportFormat(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}

func TestDefaultExportFormat_Unknown(t *testing.T) {
	_, ok := DefaultExportFormat("application/pdf")
	if ok {
		t.Error("DefaultExportFormat(application/pdf) should return false")
	}
}

func TestResolveExportFormat_Defaults(t *testing.T) {
	tests := []struct {
		mime    string
		wantExt string
	}{
		{MIMEGoogleDoc, ".docx"},
		{MIMEGoogleSheet, ".csv"},
		{MIMEGoogleSlides, ".pptx"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			info, err := ResolveExportFormat(tt.mime, "")
			if err != nil {
				t.Fatalf("ResolveExportFormat(%q, \"\") error: %v", tt.mime, err)
			}
			if info.Extension != tt.wantExt {
				t.Errorf("extension = %q, want %q", info.Extension, tt.wantExt)
			}
		})
	}
}

func TestResolveExportFormat_ValidFormats(t *testing.T) {
	tests := []struct {
		mime       string
		format     string
		wantExt    string
		wantMdConv bool
	}{
		{MIMEGoogleDoc, "docx", ".docx", false},
		{MIMEGoogleDoc, "md", ".md", true},
		{MIMEGoogleSheet, "csv", ".csv", false},
		{MIMEGoogleSlides, "pptx", ".pptx", false},
		{MIMEGoogleSlides, "md", ".md", false},
		{MIMEGoogleSlides, "pdf", ".pdf", false},
	}
	for _, tt := range tests {
		t.Run(tt.mime+"/"+tt.format, func(t *testing.T) {
			info, err := ResolveExportFormat(tt.mime, tt.format)
			if err != nil {
				t.Fatalf("ResolveExportFormat(%q, %q) error: %v", tt.mime, tt.format, err)
			}
			if info.Extension != tt.wantExt {
				t.Errorf("extension = %q, want %q", info.Extension, tt.wantExt)
			}
			if info.NeedsMarkdownConversion != tt.wantMdConv {
				t.Errorf("NeedsMarkdownConversion = %v, want %v", info.NeedsMarkdownConversion, tt.wantMdConv)
			}
		})
	}
}

func TestResolveExportFormat_DotPrefixedFormats(t *testing.T) {
	tests := []struct {
		mime    string
		format  string
		wantExt string
	}{
		{MIMEGoogleDoc, ".docx", ".docx"},
		{MIMEGoogleDoc, ".md", ".md"},
		{MIMEGoogleSheet, ".csv", ".csv"},
		{MIMEGoogleSlides, ".pptx", ".pptx"},
		{MIMEGoogleSlides, ".md", ".md"},
		{MIMEGoogleSlides, ".pdf", ".pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.mime+"/"+tt.format, func(t *testing.T) {
			info, err := ResolveExportFormat(tt.mime, tt.format)
			if err != nil {
				t.Fatalf("ResolveExportFormat(%q, %q) error: %v", tt.mime, tt.format, err)
			}
			if info.Extension != tt.wantExt {
				t.Errorf("extension = %q, want %q", info.Extension, tt.wantExt)
			}
		})
	}
}

func TestResolveExportFormat_InvalidFormats(t *testing.T) {
	tests := []struct {
		mime   string
		format string
	}{
		{MIMEGoogleDoc, "pptx"},
		{MIMEGoogleDoc, "csv"},
		{MIMEGoogleSheet, "docx"},
		{MIMEGoogleSheet, "md"},
		{MIMEGoogleSheet, "pptx"},
		{MIMEGoogleSlides, "docx"},
		{MIMEGoogleSlides, "csv"},
	}
	for _, tt := range tests {
		t.Run(tt.mime+"/"+tt.format, func(t *testing.T) {
			_, err := ResolveExportFormat(tt.mime, tt.format)
			if err == nil {
				t.Errorf("ResolveExportFormat(%q, %q) expected error, got nil", tt.mime, tt.format)
			}
		})
	}
}

func TestResolveExportFormat_UnsupportedMIME(t *testing.T) {
	_, err := ResolveExportFormat("application/pdf", "docx")
	if err == nil {
		t.Error("ResolveExportFormat(application/pdf, docx) expected error, got nil")
	}
}

func TestResolveExportFormat_SlidesPDF(t *testing.T) {
	info, err := ResolveExportFormat(MIMEGoogleSlides, "pdf")
	if err != nil {
		t.Fatalf("ResolveExportFormat(Slides, pdf) error: %v", err)
	}
	if info.Extension != ".pdf" {
		t.Errorf("extension = %q, want %q", info.Extension, ".pdf")
	}
	if info.ExportMIME != "application/pdf" {
		t.Errorf("ExportMIME = %q, want %q", info.ExportMIME, "application/pdf")
	}
	if info.NeedsMarkdownConversion {
		t.Error("NeedsMarkdownConversion should be false for PDF export")
	}
}
