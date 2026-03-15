package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
	"github.com/natikgadzhi/gdrive-cli/internal/output"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// binaryPath is the path to the compiled gdrive-cli binary, built once in TestMain.
var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once for all integration tests that use exec.Command.
	tmpDir, err := os.MkdirTemp("", "gdrive-cli-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "gdrive-cli")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runBinary executes the compiled binary with the given arguments and environment.
// Returns stdout, stderr, and any error from the process.
func runBinary(t *testing.T, env []string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Env = append(os.Environ(), env...)
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// newMockDriveService creates a Drive service pointing at the given httptest server.
func newMockDriveService(t *testing.T, server *httptest.Server) *drive.Service {
	t.Helper()
	svc, err := drive.NewService(t.Context(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("failed to create mock Drive service: %v", err)
	}
	return svc
}

// --- Binary-level tests (exec.Command) ---

func TestIntegration_AuthStatus_NotAuthenticated(t *testing.T) {
	// Point GDRIVE_CONFIG_DIR to an empty temp directory (no token.json).
	tmpDir := t.TempDir()
	env := []string{"GDRIVE_CONFIG_DIR=" + tmpDir}

	stdout, _, err := runBinary(t, env, "auth", "status")
	if err == nil {
		t.Fatal("expected non-zero exit code when not authenticated")
	}

	var result output.StatusMessage
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw stdout: %s", jsonErr, stdout)
	}

	if result.Status != "error" {
		t.Errorf("expected status %q, got %q", "error", result.Status)
	}
	if !strings.Contains(result.Message, "Not authenticated") {
		t.Errorf("expected message to contain 'Not authenticated', got: %s", result.Message)
	}
}

func TestIntegration_Version(t *testing.T) {
	stdout, _, err := runBinary(t, nil, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("failed to parse version JSON: %v\nraw: %s", err, stdout)
	}

	// Verify exactly 3 keys exist.
	expectedKeys := []string{"version", "commit", "date"}
	if len(parsed) != len(expectedKeys) {
		t.Errorf("expected exactly %d keys in version output, got %d", len(expectedKeys), len(parsed))
	}

	// Verify each key exists, is a string, and has the correct default value.
	expectedValues := map[string]string{
		"version": "dev",
		"commit":  "dev",
		"date":    "unknown",
	}
	for _, key := range expectedKeys {
		val, ok := parsed[key]
		if !ok {
			t.Errorf("missing key %q in version output", key)
			continue
		}
		str, isString := val.(string)
		if !isString {
			t.Errorf("expected key %q to be a string, got %T", key, val)
			continue
		}
		if str != expectedValues[key] {
			t.Errorf("expected %s %q, got %q", key, expectedValues[key], str)
		}
	}
}

func TestIntegration_InvalidFormatFlag(t *testing.T) {
	stdout, _, err := runBinary(t, nil, "--format", "xml", "version")
	if err == nil {
		t.Fatal("expected non-zero exit code for invalid format flag")
	}

	var result output.StatusMessage
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw stdout: %s", jsonErr, stdout)
	}

	if result.Status != "error" {
		t.Errorf("expected status %q, got %q", "error", result.Status)
	}
	if !strings.Contains(result.Message, "invalid format") {
		t.Errorf("expected message to contain 'invalid format', got: %s", result.Message)
	}
}

// --- In-process tests with httptest mock servers ---
//
// These tests use httptest to mock the Google Drive API and verify
// the full pipeline: API call -> result transformation -> output formatting.

func TestIntegration_Search_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/files") {
			http.NotFound(w, r)
			return
		}
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":           "doc123",
					"name":         "Project Proposal",
					"mimeType":     "application/vnd.google-apps.document",
					"webViewLink":  "https://docs.google.com/document/d/doc123/edit",
					"modifiedTime": "2025-06-15T14:30:00.000Z",
				},
				{
					"id":           "sheet456",
					"name":         "Budget 2025",
					"mimeType":     "application/vnd.google-apps.spreadsheet",
					"webViewLink":  "https://docs.google.com/spreadsheets/d/sheet456/edit",
					"modifiedTime": "2025-05-20T09:00:00.000Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	results, err := api.SearchFiles(svc, "project", 20)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// Build the same JSON response the search command would produce.
	response := searchResponse{
		Query:   "project",
		Count:   len(results),
		Results: results,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Verify the JSON structure by parsing it back.
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse generated JSON: %v", err)
	}

	if parsed["query"] != "project" {
		t.Errorf("expected query %q, got %v", "project", parsed["query"])
	}
	if int(parsed["count"].(float64)) != 2 {
		t.Errorf("expected count 2, got %v", parsed["count"])
	}

	resultsArr, ok := parsed["results"].([]any)
	if !ok {
		t.Fatal("expected 'results' to be an array")
	}
	if len(resultsArr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resultsArr))
	}

	first := resultsArr[0].(map[string]any)
	if first["name"] != "Project Proposal" {
		t.Errorf("expected first result name %q, got %v", "Project Proposal", first["name"])
	}
	if first["type"] != "Google Doc" {
		t.Errorf("expected first result type %q, got %v", "Google Doc", first["type"])
	}
	if first["url"] != "https://docs.google.com/document/d/doc123/edit" {
		t.Errorf("unexpected first result URL: %v", first["url"])
	}
	if first["modified"] != "2025-06-15T14:30:00.000Z" {
		t.Errorf("unexpected first result modified: %v", first["modified"])
	}

	second := resultsArr[1].(map[string]any)
	if second["name"] != "Budget 2025" {
		t.Errorf("expected second result name %q, got %v", "Budget 2025", second["name"])
	}
	if second["type"] != "Google Sheet" {
		t.Errorf("expected second result type %q, got %v", "Google Sheet", second["type"])
	}
}

func TestIntegration_Search_Markdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":           "slides789",
					"name":         "Q3 Review",
					"mimeType":     "application/vnd.google-apps.presentation",
					"webViewLink":  "https://docs.google.com/presentation/d/slides789/edit",
					"modifiedTime": "2025-09-01T12:00:00.000Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	results, err := api.SearchFiles(svc, "review", 20)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// Build the same markdown output the search command produces.
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Search: %s\n\n", "review")
	fmt.Fprintf(&buf, "**%d results**\n\n", len(results))
	if len(results) > 0 {
		fmt.Fprintln(&buf, "| Name | Type | Modified | URL |")
		fmt.Fprintln(&buf, "|------|------|----------|-----|")
		for _, r := range results {
			fmt.Fprintf(&buf, "| %s | %s | %s | %s |\n",
				r.Name, r.Type, r.Modified, r.URL)
		}
	}

	md := buf.String()

	if !strings.Contains(md, "# Search: review") {
		t.Error("expected markdown to contain search heading")
	}
	if !strings.Contains(md, "**1 results**") {
		t.Error("expected markdown to contain result count")
	}
	if !strings.Contains(md, "| Name | Type | Modified | URL |") {
		t.Error("expected markdown to contain table header")
	}
	if !strings.Contains(md, "| Q3 Review | Google Slides |") {
		t.Error("expected markdown to contain result row")
	}
	if !strings.Contains(md, "https://docs.google.com/presentation/d/slides789/edit") {
		t.Error("expected markdown to contain result URL")
	}
}

func TestIntegration_Search_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"files": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	results, err := api.SearchFiles(svc, "nonexistent-query-xyz", 20)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// Build the JSON response.
	response := searchResponse{
		Query:   "nonexistent-query-xyz",
		Count:   len(results),
		Results: results,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if int(parsed["count"].(float64)) != 0 {
		t.Errorf("expected count 0, got %v", parsed["count"])
	}

	resultsArr, ok := parsed["results"].([]any)
	if !ok {
		t.Fatal("expected 'results' to be an array")
	}
	if len(resultsArr) != 0 {
		t.Errorf("expected empty results array, got %d items", len(resultsArr))
	}
}

func TestIntegration_Fetch_JSON(t *testing.T) {
	// Mock server that handles both metadata and export requests.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Metadata request: GET /drive/v3/files/{id} (no /export suffix).
		if strings.Contains(path, "files/testdoc123") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "testdoc123",
				"name":        "Integration Test Doc",
				"mimeType":    "application/vnd.google-apps.document",
				"webViewLink": "https://docs.google.com/document/d/testdoc123/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Export request: GET /drive/v3/files/{id}/export.
		if strings.Contains(path, "files/testdoc123/export") {
			mimeType := r.URL.Query().Get("mimeType")
			switch mimeType {
			case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
				// Simulated .docx export content.
				w.Header().Set("Content-Type", mimeType)
				w.Write([]byte("fake-docx-binary-content"))
			case "text/html":
				// Simulated HTML export for markdown cache.
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte("<h1>Integration Test Doc</h1><p>Some content here.</p>"))
			default:
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("exported content"))
			}
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)
	tmpDir := t.TempDir()

	// Step 1: Get metadata (as the fetch command does).
	meta, err := api.GetFileMetadata(svc, "testdoc123")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	if meta.Name != "Integration Test Doc" {
		t.Errorf("expected name %q, got %q", "Integration Test Doc", meta.Name)
	}
	if meta.MimeType != "application/vnd.google-apps.document" {
		t.Errorf("expected mimeType for Google Doc, got %q", meta.MimeType)
	}

	// Step 2: Export the file (as the fetch command does).
	outputPath := filepath.Join(tmpDir, "Integration_Test_Doc.docx")
	exportMIME := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	err = api.ExportFile(svc, "testdoc123", exportMIME, outputPath)
	if err != nil {
		t.Fatalf("ExportFile failed: %v", err)
	}

	// Verify the file was created.
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// Verify file content.
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != "fake-docx-binary-content" {
		t.Errorf("unexpected file content: %q", string(content))
	}

	// Step 3: Build the JSON output as the fetch command would.
	result := fetchResult{
		Status:  "ok",
		FileID:  "testdoc123",
		Name:    meta.Name,
		Type:    "Google Doc",
		SavedTo: outputPath,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal fetch result: %v", err)
	}

	// Parse and verify structure.
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["status"] != "ok" {
		t.Errorf("expected status %q, got %v", "ok", parsed["status"])
	}
	if parsed["file_id"] != "testdoc123" {
		t.Errorf("expected file_id %q, got %v", "testdoc123", parsed["file_id"])
	}
	if parsed["name"] != "Integration Test Doc" {
		t.Errorf("expected name %q, got %v", "Integration Test Doc", parsed["name"])
	}
	if parsed["type"] != "Google Doc" {
		t.Errorf("expected type %q, got %v", "Google Doc", parsed["type"])
	}
	if parsed["saved_to"] != outputPath {
		t.Errorf("expected saved_to %q, got %v", outputPath, parsed["saved_to"])
	}
}

func TestIntegration_Fetch_Markdown(t *testing.T) {
	// Mock server for metadata + HTML export (for markdown conversion).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "files/mddoc456") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "mddoc456",
				"name":        "Markdown Test Doc",
				"mimeType":    "application/vnd.google-apps.document",
				"webViewLink": "https://docs.google.com/document/d/mddoc456/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(path, "files/mddoc456/export") {
			mimeType := r.URL.Query().Get("mimeType")
			switch mimeType {
			case "text/html":
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte("<h1>Heading</h1><p>Paragraph with <strong>bold</strong> text.</p>"))
			default:
				w.Header().Set("Content-Type", mimeType)
				w.Write([]byte("exported content"))
			}
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	// Get metadata.
	meta, err := api.GetFileMetadata(svc, "mddoc456")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	// Export as markdown (the HTML -> markdown conversion path).
	mdContent, err := output.ExportAsMarkdown(svc, "mddoc456", meta.MimeType)
	if err != nil {
		t.Fatalf("ExportAsMarkdown failed: %v", err)
	}

	// Verify the markdown conversion worked.
	if mdContent == "" {
		t.Fatal("expected non-empty markdown content")
	}
	if !strings.Contains(mdContent, "Heading") {
		t.Errorf("expected markdown to contain 'Heading', got: %s", mdContent)
	}
	if !strings.Contains(mdContent, "bold") {
		t.Errorf("expected markdown to contain 'bold', got: %s", mdContent)
	}

	// Verify YAML frontmatter can be built (as the fetch command does with --format markdown).
	// We just verify the fields are set correctly; YAML serialization is tested elsewhere.
	if meta.Name != "Markdown Test Doc" {
		t.Errorf("expected name %q for frontmatter, got %q", "Markdown Test Doc", meta.Name)
	}
}

func TestIntegration_Fetch_Sheet(t *testing.T) {
	csvContent := "Name,Score\nAlice,95\nBob,87\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "files/sheet789") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "sheet789",
				"name":        "Test Scores",
				"mimeType":    "application/vnd.google-apps.spreadsheet",
				"webViewLink": "https://docs.google.com/spreadsheets/d/sheet789/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(path, "files/sheet789/export") {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(csvContent))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)
	tmpDir := t.TempDir()

	meta, err := api.GetFileMetadata(svc, "sheet789")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	if meta.MimeType != "application/vnd.google-apps.spreadsheet" {
		t.Errorf("expected spreadsheet MIME type, got %q", meta.MimeType)
	}

	outputPath := filepath.Join(tmpDir, "Test_Scores.csv")
	err = api.ExportFile(svc, "sheet789", "text/csv", outputPath)
	if err != nil {
		t.Fatalf("ExportFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if string(data) != csvContent {
		t.Errorf("expected CSV content %q, got %q", csvContent, string(data))
	}

	// Verify the fetch result JSON shape.
	result := fetchResult{
		Status:  "ok",
		FileID:  "sheet789",
		Name:    meta.Name,
		Type:    "Google Sheet",
		SavedTo: outputPath,
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(resultJSON, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["type"] != "Google Sheet" {
		t.Errorf("expected type %q, got %v", "Google Sheet", parsed["type"])
	}
	if !strings.HasSuffix(parsed["saved_to"].(string), ".csv") {
		t.Errorf("expected saved_to to end with .csv, got %v", parsed["saved_to"])
	}
}

func TestIntegration_Search_CountFlag(t *testing.T) {
	var capturedPageSize string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPageSize = r.URL.Query().Get("pageSize")
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":           "single1",
					"name":         "Only Result",
					"mimeType":     "application/vnd.google-apps.document",
					"webViewLink":  "https://docs.google.com/document/d/single1/edit",
					"modifiedTime": "2025-01-01T00:00:00.000Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	results, err := api.SearchFiles(svc, "test", 5)
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	// Verify the page size was passed to the API.
	if capturedPageSize != "5" {
		t.Errorf("expected pageSize=5 in API request, got %q", capturedPageSize)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestIntegration_Fetch_UnsupportedMIME(t *testing.T) {
	// Mock server that returns a non-Workspace file type.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":          "pdf999",
			"name":        "Some PDF",
			"mimeType":    "application/pdf",
			"webViewLink": "https://drive.google.com/file/d/pdf999/view",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	meta, err := api.GetFileMetadata(svc, "pdf999")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	// The fetch command checks if the MIME type is supported before exporting.
	// Use the real formatting package instead of duplicating the logic.
	_, ok := formatting.GetExportMIME(meta.MimeType)
	if ok {
		t.Errorf("expected application/pdf to be unsupported, but it was accepted")
	}
}

func TestIntegration_Fetch_Slides(t *testing.T) {
	pptxContent := "fake-pptx-binary-content-for-slides"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "files/pres101") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "pres101",
				"name":        "Team Standup",
				"mimeType":    "application/vnd.google-apps.presentation",
				"webViewLink": "https://docs.google.com/presentation/d/pres101/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(path, "files/pres101/export") {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
			w.Write([]byte(pptxContent))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)
	tmpDir := t.TempDir()

	meta, err := api.GetFileMetadata(svc, "pres101")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	if meta.Name != "Team Standup" {
		t.Errorf("expected name %q, got %q", "Team Standup", meta.Name)
	}

	outputPath := filepath.Join(tmpDir, "Team_Standup.pptx")
	exportMIME := "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	err = api.ExportFile(svc, "pres101", exportMIME, outputPath)
	if err != nil {
		t.Fatalf("ExportFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if string(data) != pptxContent {
		t.Errorf("expected pptx content, got %q", string(data))
	}

	result := fetchResult{
		Status:  "ok",
		FileID:  "pres101",
		Name:    "Team Standup",
		Type:    "Google Slides",
		SavedTo: outputPath,
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(resultJSON, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["type"] != "Google Slides" {
		t.Errorf("expected type %q, got %v", "Google Slides", parsed["type"])
	}
	if !strings.HasSuffix(parsed["saved_to"].(string), ".pptx") {
		t.Errorf("expected saved_to to end with .pptx, got %v", parsed["saved_to"])
	}
}

func TestIntegration_Search_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "Insufficient Permission",
			},
		})
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	_, err := api.SearchFiles(svc, "test", 10)
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "drive search failed") {
		t.Errorf("expected 'drive search failed' in error, got: %v", err)
	}
}

func TestIntegration_Fetch_MetadataNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    404,
				"message": "File not found",
			},
		})
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)

	_, err := api.GetFileMetadata(svc, "doesnotexist")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "failed to get file metadata") {
		t.Errorf("expected 'failed to get file metadata' in error, got: %v", err)
	}
}

func TestIntegration_Fetch_DocAsMarkdown(t *testing.T) {
	// Mock server that returns HTML for doc export, which gets converted to markdown.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "files/mddoc789") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "mddoc789",
				"name":        "Export MD Doc",
				"mimeType":    "application/vnd.google-apps.document",
				"webViewLink": "https://docs.google.com/document/d/mddoc789/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(path, "files/mddoc789/export") {
			mimeType := r.URL.Query().Get("mimeType")
			if mimeType == "text/html" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte("<h1>Hello World</h1><p>This is <strong>bold</strong>.</p>"))
				return
			}
			w.Header().Set("Content-Type", mimeType)
			w.Write([]byte("exported"))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)
	tmpDir := t.TempDir()

	// Verify metadata retrieval.
	meta, err := api.GetFileMetadata(svc, "mddoc789")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}
	if meta.MimeType != formatting.MIMEGoogleDoc {
		t.Fatalf("expected Google Doc MIME, got %q", meta.MimeType)
	}

	// Resolve export format to md.
	resolved, err := formatting.ResolveExportFormat(meta.MimeType, "md")
	if err != nil {
		t.Fatalf("ResolveExportFormat failed: %v", err)
	}
	if resolved.Extension != ".md" {
		t.Errorf("expected .md extension, got %q", resolved.Extension)
	}
	if !resolved.NeedsMarkdownConversion {
		t.Error("expected NeedsMarkdownConversion=true")
	}

	// Export as markdown (same path as fetch --output md).
	mdContent, err := output.ExportAsMarkdown(svc, "mddoc789", meta.MimeType)
	if err != nil {
		t.Fatalf("ExportAsMarkdown failed: %v", err)
	}

	// Verify the content was converted from HTML to markdown.
	if !strings.Contains(mdContent, "Hello World") {
		t.Errorf("expected markdown to contain 'Hello World', got: %s", mdContent)
	}
	if !strings.Contains(mdContent, "bold") {
		t.Errorf("expected markdown to contain 'bold', got: %s", mdContent)
	}

	// Write to file (as fetch command would).
	outputPath := filepath.Join(tmpDir, "Export_MD_Doc.md")
	if err := os.WriteFile(outputPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), "Hello World") {
		t.Errorf("written file missing expected content")
	}

	// Verify the JSON result shape.
	result := fetchResult{
		Status:  "ok",
		FileID:  "mddoc789",
		Name:    meta.Name,
		Type:    "Google Doc",
		SavedTo: outputPath,
	}
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(resultJSON, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !strings.HasSuffix(parsed["saved_to"].(string), ".md") {
		t.Errorf("expected saved_to to end with .md, got %v", parsed["saved_to"])
	}
}

func TestIntegration_Fetch_SlidesAsMarkdown(t *testing.T) {
	plainTextContent := "Slide 1: Introduction\nSlide 2: Details\nSlide 3: Conclusion"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "files/pres202") && !strings.Contains(path, "export") {
			resp := map[string]any{
				"id":          "pres202",
				"name":        "Team Slides",
				"mimeType":    "application/vnd.google-apps.presentation",
				"webViewLink": "https://docs.google.com/presentation/d/pres202/edit",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(path, "files/pres202/export") {
			mimeType := r.URL.Query().Get("mimeType")
			if mimeType == "text/plain" {
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte(plainTextContent))
				return
			}
			w.Header().Set("Content-Type", mimeType)
			w.Write([]byte("pptx-content"))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := newMockDriveService(t, server)
	tmpDir := t.TempDir()

	meta, err := api.GetFileMetadata(svc, "pres202")
	if err != nil {
		t.Fatalf("GetFileMetadata failed: %v", err)
	}

	// Resolve to md format.
	resolved, err := formatting.ResolveExportFormat(meta.MimeType, "md")
	if err != nil {
		t.Fatalf("ResolveExportFormat failed: %v", err)
	}
	if resolved.Extension != ".md" {
		t.Errorf("expected .md, got %q", resolved.Extension)
	}

	// Export as markdown (plain text for slides).
	mdContent, err := output.ExportAsMarkdown(svc, "pres202", meta.MimeType)
	if err != nil {
		t.Fatalf("ExportAsMarkdown failed: %v", err)
	}
	if mdContent != plainTextContent {
		t.Errorf("expected plain text content, got %q", mdContent)
	}

	// Write to .md file.
	outputPath := filepath.Join(tmpDir, "Team_Slides.md")
	if err := os.WriteFile(outputPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != plainTextContent {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestIntegration_Fetch_SheetRejectsDocx(t *testing.T) {
	// Verifying that ResolveExportFormat rejects invalid formats
	// (this is what the fetch command does before making any API calls).
	_, err := formatting.ResolveExportFormat(formatting.MIMEGoogleSheet, "docx")
	if err == nil {
		t.Fatal("expected error when requesting docx export for a Google Sheet")
	}
	if !strings.Contains(err.Error(), "Google Sheet") {
		t.Errorf("error should mention 'Google Sheet', got: %v", err)
	}
	if !strings.Contains(err.Error(), "csv") {
		t.Errorf("error should mention valid format 'csv', got: %v", err)
	}
}
