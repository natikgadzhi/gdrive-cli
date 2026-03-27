package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListComments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "files/abc123/comments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Verify fields parameter is set.
		fields := r.URL.Query().Get("fields")
		if fields == "" {
			t.Error("expected fields parameter to be set")
		}

		resp := map[string]any{
			"comments": []map[string]any{
				{
					"id":           "comment1",
					"content":      "This needs updating",
					"createdTime":  "2026-03-10T10:00:00Z",
					"modifiedTime": "2026-03-12T15:00:00Z",
					"resolved":     false,
					"author": map[string]any{
						"displayName": "Alice Smith",
						"me":          false,
					},
					"quotedFileContent": map[string]any{
						"mimeType": "text/html",
						"value":    "old text here",
					},
					"replies": []map[string]any{
						{
							"id":          "reply1",
							"content":     "I agree, let me fix it",
							"createdTime": "2026-03-11T09:00:00Z",
							"author": map[string]any{
								"displayName": "Bob Jones",
								"me":          true,
							},
						},
					},
				},
				{
					"id":           "comment2",
					"content":      "Looks good now",
					"createdTime":  "2026-03-15T08:00:00Z",
					"modifiedTime": "2026-03-16T12:00:00Z",
					"resolved":     true,
					"author": map[string]any{
						"displayName": "Carol Chen",
						"me":          false,
					},
					"replies": []map[string]any{
						{
							"id":          "reply2",
							"content":     "",
							"createdTime": "2026-03-16T12:00:00Z",
							"action":      "resolve",
							"author": map[string]any{
								"displayName": "Carol Chen",
								"me":          false,
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	threads, err := ListComments(svc, "abc123")
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}

	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}

	// Verify first thread.
	if threads[0].ID != "comment1" {
		t.Errorf("expected ID 'comment1', got %q", threads[0].ID)
	}
	if threads[0].Content != "This needs updating" {
		t.Errorf("expected content 'This needs updating', got %q", threads[0].Content)
	}
	if threads[0].Author.DisplayName != "Alice Smith" {
		t.Errorf("expected author 'Alice Smith', got %q", threads[0].Author.DisplayName)
	}
	if threads[0].QuotedText != "old text here" {
		t.Errorf("expected quoted text 'old text here', got %q", threads[0].QuotedText)
	}
	if threads[0].Resolved {
		t.Error("expected first thread to not be resolved")
	}
	if len(threads[0].Replies) != 1 {
		t.Fatalf("expected 1 reply on first thread, got %d", len(threads[0].Replies))
	}
	if threads[0].Replies[0].Author.DisplayName != "Bob Jones" {
		t.Errorf("expected reply author 'Bob Jones', got %q", threads[0].Replies[0].Author.DisplayName)
	}
	if !threads[0].Replies[0].Author.IsMe {
		t.Error("expected reply author.IsMe to be true")
	}

	// Verify second thread.
	if !threads[1].Resolved {
		t.Error("expected second thread to be resolved")
	}
	if threads[1].Replies[0].Action != "resolve" {
		t.Errorf("expected reply action 'resolve', got %q", threads[1].Replies[0].Action)
	}
}

func TestListComments_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"comments": []map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	threads, err := ListComments(svc, "abc123")
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}
	if len(threads) != 0 {
		t.Errorf("expected 0 threads, got %d", len(threads))
	}
}

func TestListComments_SkipsDeletedComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"comments": []map[string]any{
				{
					"id":           "comment1",
					"content":      "Visible comment",
					"createdTime":  "2026-03-10T10:00:00Z",
					"modifiedTime": "2026-03-10T10:00:00Z",
					"resolved":     false,
					"deleted":      false,
					"author":       map[string]any{"displayName": "Alice"},
				},
				{
					"id":           "comment2",
					"content":      "",
					"createdTime":  "2026-03-11T10:00:00Z",
					"modifiedTime": "2026-03-11T10:00:00Z",
					"resolved":     false,
					"deleted":      true,
					"author":       map[string]any{"displayName": "Bob"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	threads, err := ListComments(svc, "abc123")
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread (deleted skipped), got %d", len(threads))
	}
	if threads[0].ID != "comment1" {
		t.Errorf("expected surviving comment to be 'comment1', got %q", threads[0].ID)
	}
}

func TestListComments_SkipsDeletedReplies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"comments": []map[string]any{
				{
					"id":           "comment1",
					"content":      "A comment",
					"createdTime":  "2026-03-10T10:00:00Z",
					"modifiedTime": "2026-03-12T10:00:00Z",
					"resolved":     false,
					"author":       map[string]any{"displayName": "Alice"},
					"replies": []map[string]any{
						{
							"id":          "reply1",
							"content":     "Visible reply",
							"createdTime": "2026-03-11T10:00:00Z",
							"deleted":     false,
							"author":      map[string]any{"displayName": "Bob"},
						},
						{
							"id":          "reply2",
							"content":     "",
							"createdTime": "2026-03-12T10:00:00Z",
							"deleted":     true,
							"author":      map[string]any{"displayName": "Carol"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server)

	threads, err := ListComments(svc, "abc123")
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if len(threads[0].Replies) != 1 {
		t.Fatalf("expected 1 reply (deleted skipped), got %d", len(threads[0].Replies))
	}
	if threads[0].Replies[0].ID != "reply1" {
		t.Errorf("expected surviving reply to be 'reply1', got %q", threads[0].Replies[0].ID)
	}
}

func TestListComments_APIError(t *testing.T) {
	server := httptest.NewServer(jsonErrorHandler(http.StatusForbidden, "Insufficient Permission"))
	defer server.Close()

	svc := newTestService(t, server)

	_, err := ListComments(svc, "abc123")
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list comments") {
		t.Errorf("expected 'failed to list comments' in error, got: %v", err)
	}
}

func TestListComments_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		pageToken := r.URL.Query().Get("pageToken")

		if pageToken == "" {
			// First page.
			resp := map[string]any{
				"comments": []map[string]any{
					{
						"id":           "comment1",
						"content":      "First page comment",
						"createdTime":  "2026-03-10T10:00:00Z",
						"modifiedTime": "2026-03-10T10:00:00Z",
						"resolved":     false,
						"author":       map[string]any{"displayName": "Alice"},
					},
				},
				"nextPageToken": "page2token",
			}
			json.NewEncoder(w).Encode(resp)
		} else if pageToken == "page2token" {
			// Second page.
			resp := map[string]any{
				"comments": []map[string]any{
					{
						"id":           "comment2",
						"content":      "Second page comment",
						"createdTime":  "2026-03-11T10:00:00Z",
						"modifiedTime": "2026-03-11T10:00:00Z",
						"resolved":     true,
						"author":       map[string]any{"displayName": "Bob"},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			t.Errorf("unexpected pageToken: %s", pageToken)
			http.Error(w, "bad token", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	svc := newTestService(t, server)

	threads, err := ListComments(svc, "abc123")
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}

	if len(threads) != 2 {
		t.Fatalf("expected 2 threads across pages, got %d", len(threads))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
	if threads[0].Content != "First page comment" {
		t.Errorf("expected first comment content, got %q", threads[0].Content)
	}
	if threads[1].Content != "Second page comment" {
		t.Errorf("expected second comment content, got %q", threads[1].Content)
	}
}
