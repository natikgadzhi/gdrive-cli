package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// subdirectory maps a human-readable type label to the cache subdirectory.
func subdirectory(typeLabel string) string {
	switch typeLabel {
	case "Google Sheet":
		return "spreadsheets"
	case "Google Slides":
		return "presentations"
	default:
		// Google Doc and anything else goes into documents.
		return "documents"
	}
}

// extension returns the cache file extension based on the type label.
func extension(typeLabel string) string {
	switch typeLabel {
	case "Google Sheet":
		return ".csv"
	default:
		return ".md"
	}
}

// entryPath returns the full path for a cache entry file.
func entryPath(cacheDir string, entry CacheEntry) string {
	return filepath.Join(cacheDir, subdirectory(entry.Type), entry.Slug+extension(entry.Type))
}

// entryPathForSlug searches all subdirectories for a file matching the slug.
// It tries all known subdirectories and extensions.
func entryPathForSlug(cacheDir string, slug string) (string, bool) {
	candidates := []struct {
		sub string
		ext string
	}{
		{"documents", ".md"},
		{"spreadsheets", ".csv"},
		{"presentations", ".md"},
	}
	for _, c := range candidates {
		p := filepath.Join(cacheDir, c.sub, slug+c.ext)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

// Store writes a CacheEntry to disk as a file with YAML frontmatter followed
// by the body content. It creates the necessary subdirectories automatically.
// Returns the path the file was written to.
func Store(cacheDir string, entry CacheEntry) (string, error) {
	p := entryPath(cacheDir, entry)

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", fmt.Errorf("creating cache directory: %w", err)
	}

	frontmatter, err := yaml.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(frontmatter)
	buf.WriteString("---\n")
	buf.WriteString(entry.Body)

	if err := os.WriteFile(p, []byte(buf.String()), 0o644); err != nil {
		return "", fmt.Errorf("writing cache file: %w", err)
	}

	return p, nil
}

// Load reads a cached file by slug, parses the YAML frontmatter, and returns
// the full CacheEntry including the body.
func Load(cacheDir string, slug string) (*CacheEntry, error) {
	p, found := entryPathForSlug(cacheDir, slug)
	if !found {
		return nil, fmt.Errorf("cache entry not found for slug %q", slug)
	}

	return loadFromPath(p)
}

// loadFromPath reads and parses a single cache file.
func loadFromPath(p string) (*CacheEntry, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	return parseFrontmatter(string(data))
}

// parseFrontmatter splits a file into YAML frontmatter and body,
// then unmarshals the frontmatter into a CacheEntry.
func parseFrontmatter(content string) (*CacheEntry, error) {
	// The file format is: "---\n<yaml>\n---\n<body>"
	// We split on the delimiters to extract frontmatter and body.
	const delim = "---\n"

	if !strings.HasPrefix(content, delim) {
		return nil, fmt.Errorf("cache file missing opening frontmatter delimiter")
	}

	// Find the closing delimiter after the opening one.
	rest := content[len(delim):]
	closingIdx := strings.Index(rest, "\n"+delim)
	if closingIdx == -1 {
		// Try without trailing newline (closing delimiter at the very end).
		if strings.HasSuffix(rest, "\n---") {
			closingIdx = len(rest) - 4
		} else {
			return nil, fmt.Errorf("cache file missing closing frontmatter delimiter")
		}
	}

	fmContent := rest[:closingIdx]
	body := rest[closingIdx+1+len(delim):]

	var entry CacheEntry
	if err := yaml.Unmarshal([]byte(fmContent), &entry); err != nil {
		return nil, fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	entry.Body = body

	return &entry, nil
}

// Exists reports whether a cache entry exists for the given slug.
func Exists(cacheDir string, slug string) bool {
	_, found := entryPathForSlug(cacheDir, slug)
	return found
}

// List returns all cached entries in the cache directory. Each file is read
// and parsed in full, but Body is cleared before returning for efficiency.
func List(cacheDir string) ([]CacheEntry, error) {
	var entries []CacheEntry

	subdirs := []struct {
		sub string
		ext string
	}{
		{"documents", ".md"},
		{"spreadsheets", ".csv"},
		{"presentations", ".md"},
	}

	for _, sd := range subdirs {
		dir := filepath.Join(cacheDir, sd.sub)
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading cache directory %s: %w", dir, err)
		}
		for _, de := range dirEntries {
			if de.IsDir() || !strings.HasSuffix(de.Name(), sd.ext) {
				continue
			}
			p := filepath.Join(dir, de.Name())
			entry, err := loadFromPath(p)
			if err != nil {
				// Skip malformed cache files.
				continue
			}
			// List returns frontmatter only — clear the body.
			entry.Body = ""
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}
