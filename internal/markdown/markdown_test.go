package markdown

import (
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic text",
			input:    "Hello, World!",
			expected: "<p>Hello, World!</p>\n",
		},
		{
			name:     "Headers",
			input:    "# H1\n## H2\n### H3",
			expected: "<h1 id=\"h1\">H1</h1>\n<h2 id=\"h2\">H2</h2>\n<h3 id=\"h3\">H3</h3>\n",
		},
		{
			name:     "Bold and italic",
			input:    "**bold** and *italic*",
			expected: "<p><strong>bold</strong> and <em>italic</em></p>\n",
		},
		{
			name:     "Code blocks",
			input:    "```go\nfunc main() {}\n```",
			expected: "<pre><code class=\"language-go\">func main() {}\n</code></pre>\n",
		},
		{
			name:     "Lists",
			input:    "- Item 1\n- Item 2\n  - Subitem",
			expected: "<ul>\n<li>Item 1</li>\n<li>Item 2\n<ul>\n<li>Subitem</li>\n</ul>\n</li>\n</ul>\n",
		},
		{
			name:     "Links",
			input:    "[Google](https://google.com)",
			expected: "<p><a href=\"https://google.com\">Google</a></p>\n",
		},
		{
			name:     "Images",
			input:    "![Alt text](image.jpg)",
			expected: "<p><img src=\"image.jpg\" alt=\"Alt text\" /></p>\n",
		},
		{
			name:     "Blockquotes",
			input:    "> This is a quote",
			expected: "<blockquote>\n<p>This is a quote</p>\n</blockquote>\n",
		},
		{
			name:     "Tables",
			input:    "| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |",
			expected: "<table>\n<thead>\n<tr>\n<th>Header 1</th>\n<th>Header 2</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>Cell 1</td>\n<td>Cell 2</td>\n</tr>\n</tbody>\n</table>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.input)
			if err != nil {
				t.Errorf("Render() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("Render() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRenderEmpty(t *testing.T) {
	got, err := Render("")
	if err != nil {
		t.Errorf("Render() error = %v", err)
		return
	}
	if got != "" {
		t.Errorf("Render() = %v, want empty string", got)
	}
}