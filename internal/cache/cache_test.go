package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Slug generation tests ---

func TestGenerateSlug_BasicName(t *testing.T) {
	slug := GenerateSlug("Q1 Budget Report", "1aBcDeFgHiJkL")
	want := "q1-budget-report-1abcdefg"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_SpecialCharacters(t *testing.T) {
	slug := GenerateSlug("Hello/World: A Test?", "ABCDEFGH12345")
	// Special chars become hyphens, lowercased, first 8 of ID.
	want := "hello-world-a-test-abcdefgh"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_UnicodeChars(t *testing.T) {
	slug := GenerateSlug("Cafe Resume", "xyzABCD1234")
	want := "cafe-resume-xyzabcd1"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_EmptyName(t *testing.T) {
	slug := GenerateSlug("", "ABC12345678")
	want := "abc12345"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_ShortFileID(t *testing.T) {
	slug := GenerateSlug("Doc", "abc")
	want := "doc-abc"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_LeadingTrailingSpaces(t *testing.T) {
	slug := GenerateSlug("  Trimmed  ", "12345678")
	want := "trimmed-12345678"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

func TestGenerateSlug_ConsecutiveSpecialChars(t *testing.T) {
	slug := GenerateSlug("A---B___C", "ABCDEFGH")
	want := "a-b___c-abcdefgh"
	if slug != want {
		t.Errorf("GenerateSlug() = %q, want %q", slug, want)
	}
}

// --- Store and Load round-trip tests ---

func newTestEntry() CacheEntry {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	return CacheEntry{
		Tool:        "gdrive-cli",
		Name:        "Test Document",
		Slug:        "test-document-abc12345",
		Type:        "Google Doc",
		FileID:      "abc123456789",
		SourceURL:   "https://docs.google.com/document/d/abc123456789/edit",
		CreatedAt:   now,
		UpdatedAt:   now,
		RequestedBy: "cli",
		Body:        "# Test Document\n\nHello, world!\n",
	}
}

func TestStoreAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	entry := newTestEntry()

	path, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	// Verify path is in the right subdirectory.
	if !strings.Contains(path, filepath.Join("documents", entry.Slug+".md")) {
		t.Errorf("expected path to contain documents/%s.md, got %s", entry.Slug, path)
	}

	// Verify file exists on disk.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file at %s, but it does not exist", path)
	}

	// Load it back.
	loaded, err := Load(tmpDir, entry.Slug)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Tool != entry.Tool {
		t.Errorf("Tool = %q, want %q", loaded.Tool, entry.Tool)
	}
	if loaded.Name != entry.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, entry.Name)
	}
	if loaded.Slug != entry.Slug {
		t.Errorf("Slug = %q, want %q", loaded.Slug, entry.Slug)
	}
	if loaded.Type != entry.Type {
		t.Errorf("Type = %q, want %q", loaded.Type, entry.Type)
	}
	if loaded.FileID != entry.FileID {
		t.Errorf("FileID = %q, want %q", loaded.FileID, entry.FileID)
	}
	if loaded.SourceURL != entry.SourceURL {
		t.Errorf("SourceURL = %q, want %q", loaded.SourceURL, entry.SourceURL)
	}
	if loaded.RequestedBy != entry.RequestedBy {
		t.Errorf("RequestedBy = %q, want %q", loaded.RequestedBy, entry.RequestedBy)
	}
	if loaded.Body != entry.Body {
		t.Errorf("Body = %q, want %q", loaded.Body, entry.Body)
	}
	if !loaded.CreatedAt.Equal(entry.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", loaded.CreatedAt, entry.CreatedAt)
	}
	if !loaded.UpdatedAt.Equal(entry.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", loaded.UpdatedAt, entry.UpdatedAt)
	}
}

func TestStore_Spreadsheet(t *testing.T) {
	tmpDir := t.TempDir()
	entry := CacheEntry{
		Tool:        "gdrive-cli",
		Name:        "Q1 Budget",
		Slug:        "q1-budget-abc12345",
		Type:        "Google Sheet",
		FileID:      "abc123456789",
		SourceURL:   "https://docs.google.com/spreadsheets/d/abc123456789/edit",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RequestedBy: "cli",
		Body:        "Name,Value\nAlpha,1\nBeta,2\n",
	}

	path, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	// Spreadsheets go into spreadsheets/ with .csv extension.
	if !strings.Contains(path, filepath.Join("spreadsheets", entry.Slug+".csv")) {
		t.Errorf("expected spreadsheet path, got %s", path)
	}

	loaded, err := Load(tmpDir, entry.Slug)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Body != entry.Body {
		t.Errorf("Body = %q, want %q", loaded.Body, entry.Body)
	}
}

func TestStore_Presentation(t *testing.T) {
	tmpDir := t.TempDir()
	entry := CacheEntry{
		Tool:        "gdrive-cli",
		Name:        "Team Slides",
		Slug:        "team-slides-abc12345",
		Type:        "Google Slides",
		FileID:      "abc123456789",
		SourceURL:   "https://docs.google.com/presentation/d/abc123456789/edit",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RequestedBy: "cli",
		Body:        "Slide 1: Introduction\nSlide 2: Details\n",
	}

	path, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if !strings.Contains(path, filepath.Join("presentations", entry.Slug+".md")) {
		t.Errorf("expected presentations path, got %s", path)
	}

	loaded, err := Load(tmpDir, entry.Slug)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Body != entry.Body {
		t.Errorf("Body = %q, want %q", loaded.Body, entry.Body)
	}
}

func TestStore_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	entry := newTestEntry()

	_, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("first Store() error: %v", err)
	}

	// Update the entry and store again.
	entry.Body = "# Updated\n\nNew content.\n"
	entry.UpdatedAt = entry.UpdatedAt.Add(time.Hour)

	_, err = Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("second Store() error: %v", err)
	}

	loaded, err := Load(tmpDir, entry.Slug)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Body != "# Updated\n\nNew content.\n" {
		t.Errorf("Body was not updated, got: %q", loaded.Body)
	}
}

func TestStoreAndLoad_EmptyBody(t *testing.T) {
	tmpDir := t.TempDir()
	entry := newTestEntry()
	entry.Body = ""

	path, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file at %s, but it does not exist", path)
	}

	loaded, err := Load(tmpDir, entry.Slug)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Body != "" {
		t.Errorf("Body = %q, want empty string", loaded.Body)
	}
	if loaded.Name != entry.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, entry.Name)
	}
	if loaded.FileID != entry.FileID {
		t.Errorf("FileID = %q, want %q", loaded.FileID, entry.FileID)
	}
}

// --- Exists tests ---

func TestExists_Found(t *testing.T) {
	tmpDir := t.TempDir()
	entry := newTestEntry()

	if _, err := Store(tmpDir, entry); err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if !Exists(tmpDir, entry.Slug) {
		t.Error("Exists() returned false, want true")
	}
}

func TestExists_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	if Exists(tmpDir, "nonexistent-slug") {
		t.Error("Exists() returned true for nonexistent slug, want false")
	}
}

// --- List tests ---

func TestList_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()

	doc := newTestEntry()
	sheet := CacheEntry{
		Tool:        "gdrive-cli",
		Name:        "Budget",
		Slug:        "budget-def12345",
		Type:        "Google Sheet",
		FileID:      "def123456789",
		SourceURL:   "https://docs.google.com/spreadsheets/d/def123456789/edit",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RequestedBy: "cli",
		Body:        "A,B\n1,2\n",
	}

	if _, err := Store(tmpDir, doc); err != nil {
		t.Fatalf("Store(doc) error: %v", err)
	}
	if _, err := Store(tmpDir, sheet); err != nil {
		t.Fatalf("Store(sheet) error: %v", err)
	}

	entries, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("List() returned %d entries, want 2", len(entries))
	}

	// Bodies should be empty in list results.
	for _, e := range entries {
		if e.Body != "" {
			t.Errorf("List() entry %q has non-empty Body", e.Slug)
		}
	}
}

func TestList_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	entries, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("List() returned %d entries for empty dir, want 0", len(entries))
	}
}

// --- Directory auto-creation tests ---

func TestStore_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "deep", "nested", "cache")

	entry := newTestEntry()
	path, err := Store(nested, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it does not exist", path)
	}
}

// --- Load error cases ---

func TestLoad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Load(tmpDir, "nonexistent")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent slug, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

// --- Frontmatter round-trip test ---

func TestFrontmatterRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	entry := newTestEntry()

	path, err := Store(tmpDir, entry)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	// Read raw file content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	content := string(data)

	// Verify it starts with "---".
	if !strings.HasPrefix(content, "---\n") {
		t.Error("file should start with '---' frontmatter delimiter")
	}

	// Verify frontmatter contains expected YAML keys.
	for _, key := range []string{"tool:", "name:", "slug:", "type:", "file_id:", "source_url:", "created_at:", "updated_at:", "requested_by:"} {
		if !strings.Contains(content, key) {
			t.Errorf("frontmatter should contain %q", key)
		}
	}

	// Verify body follows the closing delimiter.
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts split by '---', got %d", len(parts))
	}
	if parts[2] != entry.Body {
		t.Errorf("body = %q, want %q", parts[2], entry.Body)
	}
}

// --- Slug edge-case tests ---

func TestGenerateSlug_EdgeCases(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		fileID string
		want   string
	}{
		{"simple", "My Doc", "ABCDEFGH", "my-doc-abcdefgh"},
		{"with numbers", "Report 2025 Q1", "12345678", "report-2025-q1-12345678"},
		{"only special chars", "***", "ABCDEFGH", "abcdefgh"},
		{"mixed case ID", "Doc", "AbCdEfGh", "doc-abcdefgh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.input, tt.fileID)
			if got != tt.want {
				t.Errorf("GenerateSlug(%q, %q) = %q, want %q", tt.input, tt.fileID, got, tt.want)
			}
		})
	}
}
