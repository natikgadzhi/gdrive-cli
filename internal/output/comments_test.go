package output

import (
	"strings"
	"testing"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
)

func TestFormatCommentsMarkdown_Empty(t *testing.T) {
	result := FormatCommentsMarkdown("Test Doc", nil)
	if !strings.Contains(result, "No comments.") {
		t.Errorf("expected 'No comments.' for empty threads, got:\n%s", result)
	}
	if !strings.Contains(result, `"Test Doc"`) {
		t.Errorf("expected document name in output, got:\n%s", result)
	}
}

func TestFormatCommentsMarkdown_OpenThreads(t *testing.T) {
	threads := []api.CommentThread{
		{
			ID:           "1",
			Author:       api.CommentAuthor{DisplayName: "Alice"},
			Content:      "This needs work",
			QuotedText:   "some quoted text",
			CreatedTime:  "2026-03-10T10:00:00Z",
			ModifiedTime: "2026-03-12T10:00:00Z",
			Resolved:     false,
			Replies: []api.CommentReply{
				{
					ID:          "r1",
					Author:      api.CommentAuthor{DisplayName: "Bob"},
					Content:     "I'll fix it",
					CreatedTime: "2026-03-11T09:00:00Z",
				},
			},
		},
	}

	result := FormatCommentsMarkdown("My Doc", threads)

	// Should have Open Threads section.
	if !strings.Contains(result, "## Open Threads") {
		t.Error("expected '## Open Threads' section")
	}
	// Should NOT have Resolved Threads section.
	if strings.Contains(result, "## Resolved Threads") {
		t.Error("did not expect '## Resolved Threads' section")
	}
	// Check thread header.
	if !strings.Contains(result, "Alice") {
		t.Error("expected author name 'Alice'")
	}
	if !strings.Contains(result, "Mar 10, 2026") {
		t.Error("expected formatted date 'Mar 10, 2026'")
	}
	// Check quoted text.
	if !strings.Contains(result, "Quoted text:") {
		t.Error("expected quoted text block")
	}
	if !strings.Contains(result, "some quoted text") {
		t.Error("expected quoted text content")
	}
	// Check reply.
	if !strings.Contains(result, "Bob") {
		t.Error("expected reply author 'Bob'")
	}
	if !strings.Contains(result, "I'll fix it") {
		t.Error("expected reply content")
	}
}

func TestFormatCommentsMarkdown_ResolvedThreads(t *testing.T) {
	threads := []api.CommentThread{
		{
			ID:           "1",
			Author:       api.CommentAuthor{DisplayName: "Carol"},
			Content:      "Please fix the typo",
			CreatedTime:  "2026-02-28T10:00:00Z",
			ModifiedTime: "2026-03-01T12:00:00Z",
			Resolved:     true,
			Replies: []api.CommentReply{
				{
					ID:          "r1",
					Author:      api.CommentAuthor{DisplayName: "Dave"},
					Content:     "Fixed",
					CreatedTime: "2026-03-01T12:00:00Z",
					Action:      "resolve",
				},
			},
		},
	}

	result := FormatCommentsMarkdown("My Doc", threads)

	// Should have Resolved Threads section.
	if !strings.Contains(result, "## Resolved Threads") {
		t.Error("expected '## Resolved Threads' section")
	}
	// Should NOT have Open Threads section.
	if strings.Contains(result, "## Open Threads") {
		t.Error("did not expect '## Open Threads' section")
	}
	// Check resolved-by attribution.
	if !strings.Contains(result, "resolved by Dave") {
		t.Error("expected 'resolved by Dave' in thread header")
	}
	// Check resolve action tag.
	if !strings.Contains(result, "[resolved]") {
		t.Error("expected '[resolved]' action tag on reply")
	}
}

func TestFormatCommentsMarkdown_MixedThreads(t *testing.T) {
	threads := []api.CommentThread{
		{
			ID:          "1",
			Author:      api.CommentAuthor{DisplayName: "Alice"},
			Content:     "Open thread",
			CreatedTime: "2026-03-10T10:00:00Z",
			Resolved:    false,
		},
		{
			ID:          "2",
			Author:      api.CommentAuthor{DisplayName: "Bob"},
			Content:     "Resolved thread",
			CreatedTime: "2026-03-11T10:00:00Z",
			Resolved:    true,
		},
	}

	result := FormatCommentsMarkdown("My Doc", threads)

	if !strings.Contains(result, "## Open Threads") {
		t.Error("expected '## Open Threads' section")
	}
	if !strings.Contains(result, "## Resolved Threads") {
		t.Error("expected '## Resolved Threads' section")
	}

	// Verify open section comes before resolved.
	openIdx := strings.Index(result, "## Open Threads")
	resolvedIdx := strings.Index(result, "## Resolved Threads")
	if openIdx >= resolvedIdx {
		t.Error("expected Open Threads section to come before Resolved Threads")
	}
}

func TestFormatCommentsMarkdown_ReopenAction(t *testing.T) {
	threads := []api.CommentThread{
		{
			ID:          "1",
			Author:      api.CommentAuthor{DisplayName: "Alice"},
			Content:     "Check this",
			CreatedTime: "2026-03-10T10:00:00Z",
			Resolved:    false,
			Replies: []api.CommentReply{
				{
					ID:          "r1",
					Author:      api.CommentAuthor{DisplayName: "Bob"},
					Content:     "Done",
					CreatedTime: "2026-03-11T10:00:00Z",
					Action:      "resolve",
				},
				{
					ID:          "r2",
					Author:      api.CommentAuthor{DisplayName: "Alice"},
					Content:     "Not quite",
					CreatedTime: "2026-03-12T10:00:00Z",
					Action:      "reopen",
				},
			},
		},
	}

	result := FormatCommentsMarkdown("My Doc", threads)

	if !strings.Contains(result, "[reopened]") {
		t.Error("expected '[reopened]' action tag on reply")
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-03-10T10:00:00Z", "Mar 10, 2026"},
		{"2026-01-15T08:30:00Z", "Jan 15, 2026"},
		{"2026-03-10T10:00:00.000Z", "Mar 10, 2026"},
		{"not-a-date", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatDate(tt.input)
			if got != tt.want {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvedBy(t *testing.T) {
	// Thread with a resolve action.
	thread := api.CommentThread{
		Replies: []api.CommentReply{
			{Author: api.CommentAuthor{DisplayName: "Alice"}, Action: ""},
			{Author: api.CommentAuthor{DisplayName: "Bob"}, Action: "resolve"},
		},
	}
	if got := resolvedBy(thread); got != "Bob" {
		t.Errorf("resolvedBy = %q, want %q", got, "Bob")
	}

	// Thread with no resolve action.
	thread2 := api.CommentThread{
		Replies: []api.CommentReply{
			{Author: api.CommentAuthor{DisplayName: "Alice"}, Action: ""},
		},
	}
	if got := resolvedBy(thread2); got != "" {
		t.Errorf("resolvedBy = %q, want empty string", got)
	}

	// Thread with no replies.
	thread3 := api.CommentThread{}
	if got := resolvedBy(thread3); got != "" {
		t.Errorf("resolvedBy = %q, want empty string", got)
	}
}
