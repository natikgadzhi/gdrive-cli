package formatting

import "testing"

func TestEscapeQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "budget 2025",
			want:  "budget 2025",
		},
		{
			name:  "single quote",
			input: "it's a test",
			want:  `it\'s a test`,
		},
		{
			name:  "multiple single quotes",
			input: "it's Bob's file",
			want:  `it\'s Bob\'s file`,
		},
		{
			name:  "backslash",
			input: `path\to\file`,
			want:  `path\\to\\file`,
		},
		{
			name:  "backslash and single quote",
			input: `it's a path\to\file`,
			want:  `it\'s a path\\to\\file`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only backslashes",
			input: `\\`,
			want:  `\\\\`,
		},
		{
			name:  "only single quotes",
			input: "'''",
			want:  `\'\'\'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeQuery(tt.input)
			if got != tt.want {
				t.Errorf("EscapeQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
