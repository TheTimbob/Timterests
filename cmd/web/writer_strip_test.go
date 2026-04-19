package web_test

import (
	"testing"
	"timterests/cmd/web"
)

func TestStripDocumentHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips title and subtitle leaving body",
			input:    "# My Title\n## My Subtitle\n\nbody content here",
			expected: "body content here",
		},
		{
			name:     "strips headers from multiline body",
			input:    "# Title\n## Subtitle\n\nline one\nline two\nline three",
			expected: "line one\nline two\nline three",
		},
		{
			name:     "returns empty string when body is absent",
			input:    "# Title\n## Subtitle\n\n",
			expected: "",
		},
		{
			name:     "returns input unchanged when no h1 prefix",
			input:    "just some body text",
			expected: "just some body text",
		},
		{
			name:     "returns input unchanged when only h1 present",
			input:    "# Title\nbody without subtitle",
			expected: "# Title\nbody without subtitle",
		},
		{
			name:     "returns input unchanged when h1 but no h2",
			input:    "# Title\nsome content\n## not-a-subtitle",
			expected: "# Title\nsome content\n## not-a-subtitle",
		},
		{
			name:     "returns input unchanged when empty",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := web.StripDocumentHeaders(tc.input)
			if got != tc.expected {
				t.Errorf("StripDocumentHeaders(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
