package processor

import (
	"io"
	"strings"
	"testing"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple HTML",
			html:     "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "nested HTML",
			html:     "<div><span>Test</span> <p>Content</p></div>",
			expected: "Test Content",
		},
		{
			name:     "multiple spaces",
			html:     "<p>Hello    World</p>",
			expected: "Hello World",
		},
		{
			name:     "newlines removed",
			html:     "<div>\nHello\nWorld</div>",
			expected: "Hello World",
		},
		{
			name:     "empty input",
			html:     "",
			expected: "",
		},
		{
			name:     "only tags",
			html:     "<div><span></span></div>",
			expected: "",
		},
		{
			name:     "complex real HTML",
			html:     "<html><head><title>Test</title></head><body><h1>Title</h1><p>Paragraph with <strong>bold</strong> text.</p></body></html>",
			expected: "Test Title Paragraph with bold text.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractText(tt.html)
			if result != tt.expected {
				t.Errorf("ExtractText(%q) = %q, want %q", tt.html, result, tt.expected)
			}
		})
	}
}

func TestExtractText_RegexCompiled(t *testing.T) {
	if tagRE == nil {
		t.Error("tagRE should not be nil")
	}
	if spaceRE == nil {
		t.Error("spaceRE should not be nil")
	}

	_ = io.EOF
	_ = strings.TrimSpace
}
