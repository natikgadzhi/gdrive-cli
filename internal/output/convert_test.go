package output

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func TestHTMLToMarkdown_HeadingsAndParagraphs(t *testing.T) {
	html := []byte(`<h1>Title</h1><p>Hello world.</p><h2>Subtitle</h2><p>More text.</p>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	for _, want := range []string{"# Title", "Hello world.", "## Subtitle", "More text."} {
		if !strings.Contains(md, want) {
			t.Errorf("expected markdown to contain %q, got:\n%s", want, md)
		}
	}
}

func TestHTMLToMarkdown_BoldAndItalic(t *testing.T) {
	html := []byte(`<p><strong>bold</strong> and <em>italic</em></p>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected bold markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "*italic*") {
		t.Errorf("expected italic markdown, got:\n%s", md)
	}
}

func TestHTMLToMarkdown_Links(t *testing.T) {
	html := []byte(`<p>Visit <a href="https://example.com">Example</a>.</p>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	if !strings.Contains(md, "[Example](https://example.com)") {
		t.Errorf("expected markdown link, got:\n%s", md)
	}
}

func TestHTMLToMarkdown_Lists(t *testing.T) {
	html := []byte(`<ul><li>One</li><li>Two</li><li>Three</li></ul>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	for _, item := range []string{"One", "Two", "Three"} {
		// List items should appear prefixed with - or *
		if !strings.Contains(md, item) {
			t.Errorf("expected list item %q in markdown, got:\n%s", item, md)
		}
	}

	// Verify at least one list marker exists
	if !strings.Contains(md, "- ") && !strings.Contains(md, "* ") {
		t.Errorf("expected list markers (- or *) in markdown, got:\n%s", md)
	}
}

func TestHTMLToMarkdown_CodeBlock(t *testing.T) {
	html := []byte(`<pre><code>func main() {}</code></pre>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	if !strings.Contains(md, "func main() {}") {
		t.Errorf("expected code content in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "```") {
		t.Errorf("expected code fence in markdown, got:\n%s", md)
	}
}

func TestHTMLToMarkdown_StripsScriptsAndStyles(t *testing.T) {
	html := []byte(`<html><head><style>body{color:red}</style></head><body><script>alert('x')</script><p>Content</p></body></html>`)

	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error: %v", err)
	}

	if strings.Contains(md, "alert") {
		t.Errorf("expected scripts to be stripped, got:\n%s", md)
	}
	if strings.Contains(md, "color:red") {
		t.Errorf("expected styles to be stripped, got:\n%s", md)
	}
	if !strings.Contains(md, "Content") {
		t.Errorf("expected content to be preserved, got:\n%s", md)
	}
}

func TestHTMLToMarkdown_EmptyInput(t *testing.T) {
	md, err := HTMLToMarkdown([]byte(""))
	if err != nil {
		t.Fatalf("HTMLToMarkdown returned error on empty input: %v", err)
	}
	// Empty HTML should produce empty or whitespace-only markdown.
	if strings.TrimSpace(md) != "" {
		t.Errorf("expected empty markdown for empty HTML, got: %q", md)
	}
}

// newTestDriveService creates a Drive service backed by the given test server.
func newTestDriveService(t *testing.T, server *httptest.Server) *drive.Service {
	t.Helper()

	svc, err := drive.NewService(
		t.Context(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("creating test drive service: %v", err)
	}
	return svc
}

func TestExportAsMarkdown_GoogleDoc(t *testing.T) {
	// Mock server returns HTML for a Google Doc export.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the correct export MIME type is requested.
		if got := r.URL.Query().Get("mimeType"); got != "text/html" {
			t.Errorf("expected mimeType=text/html, got %q", got)
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<h1>My Document</h1><p>Some text with <strong>bold</strong>.</p>`))
	}))
	defer server.Close()

	svc := newTestDriveService(t, server)

	md, err := ExportAsMarkdown(svc, "file-123", formatting.MIMEGoogleDoc)
	if err != nil {
		t.Fatalf("ExportAsMarkdown returned error: %v", err)
	}

	if !strings.Contains(md, "# My Document") {
		t.Errorf("expected heading in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected bold in markdown, got:\n%s", md)
	}
}

func TestExportAsMarkdown_GoogleSlides(t *testing.T) {
	// Drive API text/plain export separates slides with double newlines.
	plainText := "Introduction\nWelcome to the talk\n\n\nDetails\nSome important info\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the correct export MIME type is requested.
		if got := r.URL.Query().Get("mimeType"); got != "text/plain" {
			t.Errorf("expected mimeType=text/plain, got %q", got)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(plainText))
	}))
	defer server.Close()

	svc := newTestDriveService(t, server)

	result, err := ExportAsMarkdown(svc, "file-456", formatting.MIMEGoogleSlides)
	if err != nil {
		t.Fatalf("ExportAsMarkdown returned error: %v", err)
	}

	// The result should now be structured Markdown with slide markers.
	if !strings.Contains(result, "## Slide 1") {
		t.Errorf("expected '## Slide 1' in result, got:\n%s", result)
	}
	if !strings.Contains(result, "## Slide 2") {
		t.Errorf("expected '## Slide 2' in result, got:\n%s", result)
	}
	if !strings.Contains(result, "Introduction") {
		t.Errorf("expected 'Introduction' in result, got:\n%s", result)
	}
	if !strings.Contains(result, "Details") {
		t.Errorf("expected 'Details' in result, got:\n%s", result)
	}
}

func TestExportAsMarkdown_GoogleSheet(t *testing.T) {
	csvContent := "Name,Value\nAlpha,1\nBeta,2\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the correct export MIME type is requested.
		if got := r.URL.Query().Get("mimeType"); got != "text/csv" {
			t.Errorf("expected mimeType=text/csv, got %q", got)
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte(csvContent))
	}))
	defer server.Close()

	svc := newTestDriveService(t, server)

	result, err := ExportAsMarkdown(svc, "file-789", formatting.MIMEGoogleSheet)
	if err != nil {
		t.Fatalf("ExportAsMarkdown returned error: %v", err)
	}

	if result != csvContent {
		t.Errorf("expected CSV pass-through, got:\n%q\nwant:\n%q", result, csvContent)
	}
}

func TestExportAsMarkdown_UnsupportedMIME(t *testing.T) {
	// The server is never called because ExportAsMarkdown returns early
	// for unsupported MIME types, but we use it to create a valid service.
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	svc := newTestDriveService(t, server)

	_, err := ExportAsMarkdown(svc, "file-000", "application/pdf")
	if err == nil {
		t.Fatal("expected error for unsupported MIME type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported MIME type") {
		t.Errorf("expected 'unsupported MIME type' in error, got: %v", err)
	}
}

func TestExportAsMarkdown_DriveAPIError(t *testing.T) {
	// Mock server returns an error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := newTestDriveService(t, server)

	_, err := ExportAsMarkdown(svc, "file-err", formatting.MIMEGoogleDoc)
	if err == nil {
		t.Fatal("expected error from Drive API failure, got nil")
	}
	if !strings.Contains(err.Error(), "drive export failed") {
		t.Errorf("expected 'drive export failed' in error, got: %v", err)
	}
}

// --- PlainTextToSlideMarkdown tests ---

func TestPlainTextToSlideMarkdown_MultipleSlides(t *testing.T) {
	// Simulate the Drive API text/plain export: slides separated by double newlines.
	input := "Title Slide\nPresentation by Author\n\n\nAgenda\nItem 1\nItem 2\n\n\nConclusion\nThank you\n"

	result := PlainTextToSlideMarkdown(input)

	if !strings.Contains(result, "## Slide 1") {
		t.Errorf("expected '## Slide 1', got:\n%s", result)
	}
	if !strings.Contains(result, "## Slide 2") {
		t.Errorf("expected '## Slide 2', got:\n%s", result)
	}
	if !strings.Contains(result, "## Slide 3") {
		t.Errorf("expected '## Slide 3', got:\n%s", result)
	}
	if !strings.Contains(result, "Title Slide") {
		t.Errorf("expected 'Title Slide' content, got:\n%s", result)
	}
	if !strings.Contains(result, "Agenda") {
		t.Errorf("expected 'Agenda' content, got:\n%s", result)
	}
	if !strings.Contains(result, "Conclusion") {
		t.Errorf("expected 'Conclusion' content, got:\n%s", result)
	}
}

func TestPlainTextToSlideMarkdown_SingleSlide(t *testing.T) {
	input := "Only one slide\nWith some text\n"

	result := PlainTextToSlideMarkdown(input)

	if !strings.Contains(result, "## Slide 1") {
		t.Errorf("expected '## Slide 1', got:\n%s", result)
	}
	if strings.Contains(result, "## Slide 2") {
		t.Errorf("should not contain '## Slide 2', got:\n%s", result)
	}
	if !strings.Contains(result, "Only one slide") {
		t.Errorf("expected slide content, got:\n%s", result)
	}
}

func TestPlainTextToSlideMarkdown_Empty(t *testing.T) {
	result := PlainTextToSlideMarkdown("")
	if result != "" {
		t.Errorf("expected empty string for empty input, got: %q", result)
	}
}

func TestPlainTextToSlideMarkdown_WhitespaceOnly(t *testing.T) {
	result := PlainTextToSlideMarkdown("   \n\n\n   \n")
	if result != "" {
		t.Errorf("expected empty string for whitespace-only input, got: %q", result)
	}
}

func TestPlainTextToSlideMarkdown_PreservesParagraphBreaks(t *testing.T) {
	// A single blank line within a slide should be preserved as a paragraph break,
	// while double blank lines mark slide boundaries.
	input := "First paragraph\n\nSecond paragraph\n\n\nNext slide content\n"

	result := PlainTextToSlideMarkdown(input)

	if !strings.Contains(result, "## Slide 1") {
		t.Errorf("expected '## Slide 1', got:\n%s", result)
	}
	if !strings.Contains(result, "## Slide 2") {
		t.Errorf("expected '## Slide 2', got:\n%s", result)
	}
	if !strings.Contains(result, "First paragraph") {
		t.Errorf("expected first paragraph, got:\n%s", result)
	}
	if !strings.Contains(result, "Second paragraph") {
		t.Errorf("expected second paragraph, got:\n%s", result)
	}
}

func TestPlainTextToSlideMarkdown_WindowsLineEndings(t *testing.T) {
	input := "Slide one\r\n\r\n\r\nSlide two\r\n"

	result := PlainTextToSlideMarkdown(input)

	if !strings.Contains(result, "## Slide 1") {
		t.Errorf("expected '## Slide 1', got:\n%s", result)
	}
	if !strings.Contains(result, "## Slide 2") {
		t.Errorf("expected '## Slide 2', got:\n%s", result)
	}
}
