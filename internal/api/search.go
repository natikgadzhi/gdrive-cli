package api

import (
	"fmt"
	"strings"

	drive "google.golang.org/api/drive/v3"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
	"github.com/natikgadzhi/gdrive-cli/internal/formatting"
)

// SearchFiles searches Google Drive for Docs, Sheets, and Slides files
// matching the given query string. Results are ordered by modifiedTime desc.
// maxResults controls the maximum number of results returned.
func SearchFiles(svc *drive.Service, query string, maxResults int) ([]FileResult, error) {
	escaped := formatting.EscapeQuery(query)

	// Build MIME type filter.
	mimeTypes := formatting.SupportedMIMETypes()
	typeClauses := make([]string, len(mimeTypes))
	for i, mt := range mimeTypes {
		typeClauses[i] = fmt.Sprintf("mimeType='%s'", mt)
	}
	typeFilter := strings.Join(typeClauses, " or ")

	fullQuery := fmt.Sprintf("(%s) and (name contains '%s' or fullText contains '%s')",
		typeFilter, escaped, escaped)

	config.DebugLog("Drive API query: %s", fullQuery)

	call := svc.Files.List().
		Q(fullQuery).
		PageSize(int64(maxResults)).
		Fields("files(id,name,mimeType,webViewLink,modifiedTime)").
		OrderBy("modifiedTime desc")

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("drive search failed: %w", err)
	}

	results := make([]FileResult, 0, len(resp.Files))
	for _, f := range resp.Files {
		label, _ := formatting.GetTypeLabel(f.MimeType)
		results = append(results, FileResult{
			Name:     f.Name,
			Type:     label,
			URL:      f.WebViewLink,
			Modified: f.ModifiedTime,
		})
	}

	config.DebugLog("Search returned %d results", len(results))
	return results, nil
}
