package output

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	var retErr error
	out := captureStdout(t, func() {
		retErr = Errorf("failed: %s", "bad input")
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

	// Errorf must return a non-nil SilentError so Cobra exits with code 1.
	if retErr == nil {
		t.Fatal("Errorf should return a non-nil error")
	}
	if !IsSilentError(retErr) {
		t.Errorf("Errorf should return a SilentError, got %T", retErr)
	}
	if retErr.Error() != "failed: bad input" {
		t.Errorf("SilentError.Error() = %q, want %q", retErr.Error(), "failed: bad input")
	}
}

func TestSilentError(t *testing.T) {
	err := &SilentError{Message: "test error"}

	if err.Error() != "test error" {
		t.Errorf("SilentError.Error() = %q, want %q", err.Error(), "test error")
	}

	if !IsSilentError(err) {
		t.Error("IsSilentError should return true for *SilentError")
	}

	if IsSilentError(nil) {
		t.Error("IsSilentError should return false for nil")
	}

	otherErr := fmt.Errorf("some other error")
	if IsSilentError(otherErr) {
		t.Error("IsSilentError should return false for non-SilentError")
	}
}
