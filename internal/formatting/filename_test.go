package formatting

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean filename",
			input: "my-document",
			want:  "my-document",
		},
		{
			name:  "forward slash",
			input: "path/to/file",
			want:  "path_to_file",
		},
		{
			name:  "backslash",
			input: `path\to\file`,
			want:  "path_to_file",
		},
		{
			name:  "colon",
			input: "file:name",
			want:  "file_name",
		},
		{
			name:  "asterisk",
			input: "file*name",
			want:  "file_name",
		},
		{
			name:  "question mark",
			input: "file?name",
			want:  "file_name",
		},
		{
			name:  "double quotes",
			input: `file"name"`,
			want:  "file_name_",
		},
		{
			name:  "angle brackets",
			input: "file<name>",
			want:  "file_name_",
		},
		{
			name:  "pipe",
			input: "file|name",
			want:  "file_name",
		},
		{
			name:  "multiple special characters",
			input: `Q1 Budget: "Final" <2025>`,
			want:  `Q1 Budget_ _Final_ _2025_`,
		},
		{
			name:  "leading and trailing whitespace",
			input: "  my document  ",
			want:  "my document",
		},
		{
			name:  "all special characters",
			input: `/\:*?"<>|`,
			want:  "_________",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
