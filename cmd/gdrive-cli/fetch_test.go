package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
)

func TestFetchCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "fetch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'fetch' command to be registered on rootCmd")
	}
}

func TestFetchCommandFlags(t *testing.T) {
	flags := map[string]struct {
		shorthand string
		defValue  string
	}{
		"output": {shorthand: "o", defValue: ""},
		"dest":   {shorthand: "f", defValue: ""},
	}

	for name, want := range flags {
		f := fetchCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q to be registered on fetchCmd", name)
			continue
		}
		if f.Shorthand != want.shorthand {
			t.Errorf("flag %q: shorthand = %q, want %q", name, f.Shorthand, want.shorthand)
		}
		if f.DefValue != want.defValue {
			t.Errorf("flag %q: default = %q, want %q", name, f.DefValue, want.defValue)
		}
	}
}

func TestFetchCommandDirFlagRemoved(t *testing.T) {
	f := fetchCmd.Flags().Lookup("dir")
	if f != nil {
		t.Error("expected --dir flag to be removed from fetchCmd, but it is still registered")
	}
}

func TestFetchCommandRequiresArgs(t *testing.T) {
	// The fetch command requires exactly one argument (the URL).
	if fetchCmd.Args == nil {
		t.Fatal("expected fetchCmd.Args to be set (cobra.ExactArgs(1))")
	}
}

func TestFetchCommandHasHelp(t *testing.T) {
	if fetchCmd.Short == "" {
		t.Error("fetch command should have a short description")
	}
	if fetchCmd.Long == "" {
		t.Error("fetch command should have a long description")
	}
}

func TestResolveOutputPath_Empty(t *testing.T) {
	// No flag value: auto-generate in current directory.
	// SanitizeFilename only replaces / \ : * ? " < > | -- spaces are preserved.
	got := resolveOutputPath("", "My Document", ".docx")
	if got != "My Document.docx" {
		t.Errorf("resolveOutputPath(\"\", ...) = %q, want %q", got, "My Document.docx")
	}
}

func TestResolveOutputPath_ExplicitFile(t *testing.T) {
	got := resolveOutputPath("/tmp/custom.docx", "My Document", ".docx")
	if got != "/tmp/custom.docx" {
		t.Errorf("resolveOutputPath explicit file = %q, want %q", got, "/tmp/custom.docx")
	}
}

func TestResolveOutputPath_TrailingSeparator(t *testing.T) {
	got := resolveOutputPath("/tmp/downloads/", "Budget Report", ".csv")
	want := filepath.Join("/tmp/downloads/", "Budget Report.csv")
	if got != want {
		t.Errorf("resolveOutputPath trailing separator = %q, want %q", got, want)
	}
}

func TestResolveOutputPath_ExistingDir(t *testing.T) {
	dir := t.TempDir()
	got := resolveOutputPath(dir, "Slides Deck", ".pptx")
	want := filepath.Join(dir, "Slides Deck.pptx")
	if got != want {
		t.Errorf("resolveOutputPath existing dir = %q, want %q", got, want)
	}
}

func TestResolveOutputPath_MarkdownExtension(t *testing.T) {
	got := resolveOutputPath("", "My Document", ".md")
	if got != "My Document.md" {
		t.Errorf("resolveOutputPath md = %q, want %q", got, "My Document.md")
	}
}

// --- Export format validation tests ---

func TestResolveExportFormat_DocDefault(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleDoc, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Extension != ".docx" {
		t.Errorf("expected .docx extension, got %q", info.Extension)
	}
}

func TestResolveExportFormat_DocMarkdown(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleDoc, "md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Extension != ".md" {
		t.Errorf("expected .md extension, got %q", info.Extension)
	}
	if !info.NeedsMarkdownConversion {
		t.Error("expected NeedsMarkdownConversion=true for doc md export")
	}
}

func TestResolveExportFormat_DocInvalid(t *testing.T) {
	_, err := formatting.ResolveExportFormat(formatting.MIMEGoogleDoc, "pptx")
	if err == nil {
		t.Fatal("expected error for invalid format pptx on Google Doc")
	}
}

func TestResolveExportFormat_SheetDefault(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSheet, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Extension != ".csv" {
		t.Errorf("expected .csv extension, got %q", info.Extension)
	}
}

func TestResolveExportFormat_SheetInvalidDocx(t *testing.T) {
	_, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSheet, "docx")
	if err == nil {
		t.Fatal("expected error for invalid format docx on Google Sheet")
	}
}

func TestResolveExportFormat_SheetInvalidMd(t *testing.T) {
	_, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSheet, "md")
	if err == nil {
		t.Fatal("expected error for invalid format md on Google Sheet")
	}
}

func TestResolveExportFormat_SlidesDefault(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSlides, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Extension != ".pptx" {
		t.Errorf("expected .pptx extension, got %q", info.Extension)
	}
}

func TestResolveExportFormat_SlidesMarkdown(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSlides, "md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Extension != ".md" {
		t.Errorf("expected .md extension, got %q", info.Extension)
	}
}

func TestResolveExportFormat_SlidesInvalidCsv(t *testing.T) {
	_, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSlides, "csv")
	if err == nil {
		t.Fatal("expected error for invalid format csv on Google Slides")
	}
}

// --- Integration-style tests for markdown export ---

func TestFetchResolveExportFormat_DocAsMarkdown(t *testing.T) {
	// Verify that when fetching a doc with --output md, the resolved format
	// produces an .md extension and uses HTML export MIME for conversion.
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleDoc, "md")
	if err != nil {
		t.Fatalf("ResolveExportFormat failed: %v", err)
	}
	if info.Extension != ".md" {
		t.Errorf("expected .md, got %q", info.Extension)
	}
	if info.ExportMIME != "text/html" {
		t.Errorf("expected text/html export MIME for doc->md, got %q", info.ExportMIME)
	}

	// Verify the output path would be correct.
	tmpDir := t.TempDir()
	outPath := resolveOutputPath(tmpDir, "My Doc", ".md")
	want := filepath.Join(tmpDir, "My Doc.md")
	if outPath != want {
		t.Errorf("output path = %q, want %q", outPath, want)
	}

	// Write a test file to verify the path works.
	if err := os.WriteFile(outPath, []byte("# Test"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	if string(data) != "# Test" {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestFetchResolveExportFormat_SheetDefaultCsv(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSheet, "csv")
	if err != nil {
		t.Fatalf("ResolveExportFormat(sheet, csv) failed: %v", err)
	}
	if info.Extension != ".csv" {
		t.Errorf("expected .csv, got %q", info.Extension)
	}
}

func TestFetchResolveExportFormat_SlidesAsMarkdown(t *testing.T) {
	info, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSlides, "md")
	if err != nil {
		t.Fatalf("ResolveExportFormat(slides, md) failed: %v", err)
	}
	if info.Extension != ".md" {
		t.Errorf("expected .md, got %q", info.Extension)
	}
	if info.ExportMIME != "text/plain" {
		t.Errorf("expected text/plain export MIME for slides->md, got %q", info.ExportMIME)
	}
}
