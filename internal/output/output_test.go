package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

// captureStdout calls fn while redirecting os.Stdout to a pipe,
// then returns everything written to stdout as a byte slice.
func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	return buf.Bytes()
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{
		"status":  "ok",
		"message": "test message",
	}

	out := captureStdout(t, func() {
		if err := PrintJSON(data); err != nil {
			t.Fatalf("PrintJSON returned error: %v", err)
		}
	})

	var result map[string]string
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
	if result["message"] != "test message" {
		t.Errorf("expected message='test message', got %q", result["message"])
	}
}

func TestPrintJSONIndented(t *testing.T) {
	out := captureStdout(t, func() {
		if err := PrintJSON(map[string]string{"key": "value"}); err != nil {
			t.Fatalf("PrintJSON returned error: %v", err)
		}
	})

	// Indented JSON should contain 2-space indentation.
	expected := []byte(`  "key": "value"`)
	if !bytes.Contains(out, expected) {
		t.Errorf("expected indented output containing %q, got:\n%s", expected, out)
	}
}

func TestPrintJSONNoHTMLEscape(t *testing.T) {
	out := captureStdout(t, func() {
		if err := PrintJSON(map[string]string{"url": "https://example.com?a=1&b=2"}); err != nil {
			t.Fatalf("PrintJSON returned error: %v", err)
		}
	})

	// Without HTML escaping, & should remain as & not \u0026.
	if bytes.Contains(out, []byte(`\u0026`)) {
		t.Errorf("expected no HTML escaping, but found \\u0026 in output:\n%s", out)
	}
}

func TestOK(t *testing.T) {
	out := captureStdout(t, func() {
		if err := OK("it worked"); err != nil {
			t.Fatalf("OK returned error: %v", err)
		}
	})

	var result StatusMessage
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if result.Status != "ok" {
		t.Errorf("expected status=ok, got %q", result.Status)
	}
	if result.Message != "it worked" {
		t.Errorf("expected message='it worked', got %q", result.Message)
	}
}

func TestErrorf(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Errorf("failed: %s", "bad input"); err != nil {
			t.Fatalf("Errorf returned error: %v", err)
		}
	})

	var result StatusMessage
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if result.Status != "error" {
		t.Errorf("expected status=error, got %q", result.Status)
	}
	if result.Message != "failed: bad input" {
		t.Errorf("expected message='failed: bad input', got %q", result.Message)
	}
}
