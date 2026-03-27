package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/natikgadzhi/gdrive-cli/internal/api"
)

// FormatCommentsMarkdown renders a slice of comment threads as a Markdown
// document. Threads are grouped into "Open Threads" and "Resolved Threads"
// sections. Each thread shows the author, date, quoted text (if any),
// and all replies with indentation.
func FormatCommentsMarkdown(docName string, threads []api.CommentThread) string {
	if len(threads) == 0 {
		return fmt.Sprintf("# Comments on %q\n\nNo comments.\n", docName)
	}

	var open, resolved []api.CommentThread
	for _, t := range threads {
		if t.Resolved {
			resolved = append(resolved, t)
		} else {
			open = append(open, t)
		}
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "# Comments on %q\n", docName)

	if len(open) > 0 {
		buf.WriteString("\n## Open Threads\n")
		for i, t := range open {
			writeThread(&buf, i+1, t)
		}
	}

	if len(resolved) > 0 {
		buf.WriteString("\n## Resolved Threads\n")
		for i, t := range resolved {
			writeThread(&buf, i+1, t)
		}
	}

	return buf.String()
}

// writeThread writes a single comment thread to the buffer.
func writeThread(buf *strings.Builder, num int, t api.CommentThread) {
	date := formatDate(t.CreatedTime)

	// Thread header with resolved indicator.
	header := fmt.Sprintf("%s (%s)", t.Author.DisplayName, date)
	if t.Resolved {
		resolver := resolvedBy(t)
		if resolver != "" {
			header += " — resolved by " + resolver
		} else {
			header += " — resolved"
		}
	}
	fmt.Fprintf(buf, "\n### Thread %d — %s\n", num, header)

	// Quoted text the comment is anchored to.
	if t.QuotedText != "" {
		fmt.Fprintf(buf, "\n> **Quoted text:** %q\n", t.QuotedText)
	}

	// Original comment.
	fmt.Fprintf(buf, "\n**%s** (%s):\n%s\n", t.Author.DisplayName, date, t.Content)

	// Replies, indented.
	for _, r := range t.Replies {
		rDate := formatDate(r.CreatedTime)
		action := ""
		if r.Action == "resolve" {
			action = " [resolved]"
		} else if r.Action == "reopen" {
			action = " [reopened]"
		}
		fmt.Fprintf(buf, "\n> **%s** (%s)%s:\n> %s\n",
			r.Author.DisplayName, rDate, action,
			strings.ReplaceAll(r.Content, "\n", "\n> "))
	}

	buf.WriteString("\n---\n")
}

// resolvedBy returns the display name of the person who resolved a thread,
// identified by the last reply with action "resolve".
func resolvedBy(t api.CommentThread) string {
	for i := len(t.Replies) - 1; i >= 0; i-- {
		if t.Replies[i].Action == "resolve" {
			return t.Replies[i].Author.DisplayName
		}
	}
	return ""
}

// formatDate converts an RFC 3339 timestamp to a human-readable short date.
// Returns the original string if parsing fails.
func formatDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		// Try with fractional seconds (Drive API sometimes uses this format).
		t, err = time.Parse("2006-01-02T15:04:05.000Z", rfc3339)
		if err != nil {
			return rfc3339
		}
	}
	return t.Format("Jan 2, 2006")
}
