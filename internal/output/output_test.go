package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	// Capture stdout.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	data := map[string]string{
		"status":  "ok",
		"message": "test message",
	}
	if err := PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
	if result["message"] != "test message" {
		t.Errorf("expected message='test message', got %q", result["message"])
	}
}

func TestPrintJSONIndented(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	data := map[string]string{"key": "value"}
	if err := PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	output := buf.String()
	// Indented JSON should contain newlines and spaces.
	if len(output) < 10 {
		t.Fatalf("output too short for indented JSON: %q", output)
	}
	// Should contain 2-space indentation.
	expected := "  \"key\": \"value\""
	if !bytes.Contains(buf.Bytes(), []byte(expected)) {
		t.Errorf("expected indented output containing %q, got:\n%s", expected, output)
	}
}

func TestPrintJSONNoHTMLEscape(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	data := map[string]string{"url": "https://example.com?a=1&b=2"}
	if err := PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	output := buf.String()
	// Without HTML escaping, & should remain as & not \u0026.
	if bytes.Contains(buf.Bytes(), []byte(`\u0026`)) {
		t.Errorf("expected no HTML escaping, but found \\u0026 in output:\n%s", output)
	}
}
