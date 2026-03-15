package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// newTestService creates a Drive service pointed at the given httptest server.
func newTestService(t *testing.T, server *httptest.Server) *drive.Service {
	t.Helper()
	svc, err := drive.NewService(t.Context(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("failed to create test Drive service: %v", err)
	}
	return svc
}

// jsonErrorHandler returns an http.HandlerFunc that responds with the given
// HTTP status code and a JSON error body. Useful for API error tests.
func jsonErrorHandler(code int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		resp := map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// --- Search tests ---

func TestSearchFiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a files list request.
		// With a custom endpoint the client may strip the /drive/v3 prefix.
		if !strings.HasSuffix(r.URL.Path, "/files") && !strings.HasPrefix(r.URL.Path, "/drive/v3/files") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Verify query parameter is present and contains our search term.
		q := r.URL.Query().Get("q")
		if q == "" {
			t.Error("expected q parameter in request")
		}
		if !strings.Contains(q, "budget") {
			t.Errorf("expected query to contain 'budget', got: %s", q)
		}
		// Verify MIME type filter is present.
		if !strings.Contains(q, "mimeType=") {
			t.Errorf("expected query to contain mimeType filter, got: %s", q)
		}
		// Verify trashed filter is present.
		if !strings.Contains(q, "trashed = false") {
			t.Errorf("expected query to contain 'trashed = false', got: %s", q)
		}

		// Verify orderBy.
		orderBy := r.URL.Query().Get("orderBy")
		if orderBy != "modifiedTime desc" {
			t.Errorf("expected orderBy='modifiedTime desc', got: %s", orderBy)
		}

		// Verify fields.
		fields := r.URL.Query().Get("fields")
		if !strings.Contains(fields, "files(id,name,mimeType,webViewLink,modifiedTime)") {
			t.Errorf("expected fields to contain file fields, got: %s", fields)
		}

		// Return mock response.
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":           "abc123",
					"name":         "Q1 Budget",
					"mimeType":     "application/vnd.google-apps.spreadsheet",
					"webViewLink":  "https://docs.google.com/spreadsheets/d/abc123/edit",
					"modifiedTime": "2025-03-01T10:00:00.000Z",
				},
				{
					"id":           "def456",
					"name":         "Budget Planning",
					"mimeType":     "application/vnd.google-apps.document",
					"webViewLink":  "https://docs.google.com/document/d/def456/edit",
					"modifiedTime": "2025-02-15T08:30:00.000Z",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	results, err := SearchFiles(svc, "budget", 20)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify first result.
	if results[0].Name != "Q1 Budget" {
		t.Errorf("expected name 'Q1 Budget', got '%s'", results[0].Name)
	}
	if results[0].Type != "Google Sheet" {
		t.Errorf("expected type 'Google Sheet', got '%s'", results[0].Type)
	}
	if results[0].URL != "https://docs.google.com/spreadsheets/d/abc123/edit" {
		t.Errorf("unexpected URL: %s", results[0].URL)
	}
	if results[0].Modified != "2025-03-01T10:00:00.000Z" {
		t.Errorf("unexpected modified time: %s", results[0].Modified)
	}

	// Verify second result.
	if results[1].Name != "Budget Planning" {
		t.Errorf("expected name 'Budget Planning', got '%s'", results[1].Name)
	}
	if results[1].Type != "Google Doc" {
		t.Errorf("expected type 'Google Doc', got '%s'", results[1].Type)
	}
}

func TestSearchFiles_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"files": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	results, err := SearchFiles(svc, "nonexistent", 10)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchFiles_QueryEscaping(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query().Get("q")
		resp := map[string]any{"files": []map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	_, err := SearchFiles(svc, "it's a test", 10)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// The query should contain the escaped single quote.
	if !strings.Contains(capturedQuery, `it\'s a test`) {
		t.Errorf("expected escaped query, got: %s", capturedQuery)
	}
	// Verify trashed filter is present.
	if !strings.Contains(capturedQuery, "trashed = false") {
		t.Errorf("expected query to contain 'trashed = false', got: %s", capturedQuery)
	}
}

func TestSearchFiles_APIError(t *testing.T) {
	server := httptest.NewServer(jsonErrorHandler(http.StatusForbidden, "Insufficient Permission"))
	defer server.Close()

	svc := newTestService(t, server)

	_, err := SearchFiles(svc, "test", 10)
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "drive search failed") {
		t.Errorf("expected 'drive search failed' in error, got: %v", err)
	}
}

// --- Export tests ---

func TestExportFile_Success(t *testing.T) {
	exportContent := []byte("column1,column2\nvalue1,value2\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the export path.
		// With a custom endpoint the client may strip the /drive/v3 prefix.
		if !strings.Contains(r.URL.Path, "files/abc123/export") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Verify mimeType query parameter.
		mt := r.URL.Query().Get("mimeType")
		if mt != "text/csv" {
			t.Errorf("expected mimeType 'text/csv', got '%s'", mt)
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Write(exportContent)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	// Create a temp dir for output.
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "subdir", "output.csv")

	err := ExportFile(svc, "abc123", "text/csv", outputPath)
	if err != nil {
		t.Fatalf("ExportFile failed: %v", err)
	}

	// Verify the file was written.
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != string(exportContent) {
		t.Errorf("expected content %q, got %q", string(exportContent), string(data))
	}
}

func TestExportFile_CreatesParentDirs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	svc := newTestService(t, server)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "a", "b", "c", "output.docx")

	err := ExportFile(svc, "xyz789", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", outputPath)
	if err != nil {
		t.Fatalf("ExportFile failed: %v", err)
	}

	// Verify parent directories were created.
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		t.Error("expected parent directories to be created")
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestExportFile_APIError(t *testing.T) {
	server := httptest.NewServer(jsonErrorHandler(http.StatusNotFound, "File not found"))
	defer server.Close()

	svc := newTestService(t, server)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	err := ExportFile(svc, "notfound", "text/csv", outputPath)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "drive export failed") {
		t.Errorf("expected 'drive export failed' in error, got: %v", err)
	}
}

// --- Metadata tests ---

func TestGetFileMetadata_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a get request for the right file.
		// With a custom endpoint the client may strip the /drive/v3 prefix.
		if !strings.Contains(r.URL.Path, "files/abc123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Should not be an export request (no /export suffix).
		if strings.Contains(r.URL.Path, "export") {
			t.Error("metadata request should not hit export endpoint")
			http.NotFound(w, r)
			return
		}

		// Verify fields parameter.
		fields := r.URL.Query().Get("fields")
		if !strings.Contains(fields, "id") || !strings.Contains(fields, "name") ||
			!strings.Contains(fields, "mimeType") || !strings.Contains(fields, "webViewLink") {
			t.Errorf("expected fields to contain id,name,mimeType,webViewLink, got: %s", fields)
		}

		resp := map[string]any{
			"id":          "abc123",
			"name":        "My Document",
			"mimeType":    "application/vnd.google-apps.document",
			"webViewLink": "https://docs.google.com/document/d/abc123/edit",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	meta, err := GetFileMetadata(svc, "abc123")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	if meta.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got '%s'", meta.ID)
	}
	if meta.Name != "My Document" {
		t.Errorf("expected Name 'My Document', got '%s'", meta.Name)
	}
	if meta.MimeType != "application/vnd.google-apps.document" {
		t.Errorf("expected MimeType 'application/vnd.google-apps.document', got '%s'", meta.MimeType)
	}
	if meta.WebViewLink != "https://docs.google.com/document/d/abc123/edit" {
		t.Errorf("expected WebViewLink URL, got '%s'", meta.WebViewLink)
	}
}

func TestGetFileMetadata_NotFound(t *testing.T) {
	server := httptest.NewServer(jsonErrorHandler(http.StatusNotFound, "File not found: notfound"))
	defer server.Close()

	svc := newTestService(t, server)

	_, err := GetFileMetadata(svc, "notfound")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get file metadata") {
		t.Errorf("expected 'failed to get file metadata' in error, got: %v", err)
	}
}

func TestGetFileMetadata_Forbidden(t *testing.T) {
	server := httptest.NewServer(jsonErrorHandler(http.StatusForbidden, "Insufficient Permission"))
	defer server.Close()

	svc := newTestService(t, server)

	_, err := GetFileMetadata(svc, "forbidden123")
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
}

// --- Client tests ---

func TestNewDriveServiceWithClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc, err := NewDriveServiceWithClient(server.Client())
	if err != nil {
		t.Fatalf("NewDriveServiceWithClient failed: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

// --- Type tests ---

func TestFileResult_Fields(t *testing.T) {
	r := FileResult{
		Name:     "Test Doc",
		Type:     "Google Doc",
		URL:      "https://docs.google.com/document/d/123/edit",
		Modified: "2025-01-01T00:00:00.000Z",
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal FileResult: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FileResult: %v", err)
	}

	if decoded["name"] != "Test Doc" {
		t.Errorf("expected name 'Test Doc', got '%s'", decoded["name"])
	}
	if decoded["type"] != "Google Doc" {
		t.Errorf("expected type 'Google Doc', got '%s'", decoded["type"])
	}
	if decoded["url"] != "https://docs.google.com/document/d/123/edit" {
		t.Errorf("unexpected url: %s", decoded["url"])
	}
}

func TestFileMetadata_Fields(t *testing.T) {
	m := FileMetadata{
		ID:          "abc123",
		Name:        "Test Sheet",
		MimeType:    "application/vnd.google-apps.spreadsheet",
		WebViewLink: "https://docs.google.com/spreadsheets/d/abc123/edit",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal FileMetadata: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FileMetadata: %v", err)
	}

	if decoded["id"] != "abc123" {
		t.Errorf("expected id 'abc123', got '%s'", decoded["id"])
	}
	if decoded["mimeType"] != "application/vnd.google-apps.spreadsheet" {
		t.Errorf("expected mimeType for spreadsheet, got '%s'", decoded["mimeType"])
	}
}

