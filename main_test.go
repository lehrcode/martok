package main

import (
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

func TestFirstHeadingText(t *testing.T) {
	tests := []struct {
		name    string
		sources []string
		want    string
	}{
		{
			name:    "first heading of any level without markup",
			sources: []string{"intro\n\n## The *first* `heading` &amp; [more](https://example.com)\n\n# Second\n"},
			want:    "The first heading & more",
		},
		{
			name:    "setext heading in a container",
			sources: []string{"> Quoted\n> heading\n> ======\n\n# Second\n"},
			want:    "Quoted heading",
		},
		{
			name:    "headings without text are skipped",
			sources: []string{"#\n\n## ![](image.png)\n\n### Third\n"},
			want:    "Third",
		},
		{
			name:    "heading in a following file",
			sources: []string{"just text\n", "", "# Appendix\n"},
			want:    "Appendix",
		},
		{
			name:    "no heading",
			sources: []string{"just text\n", "- list\n"},
			want:    "",
		},
	}

	mdParser := parser.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				sources   [][]byte
				documents []ast.Node
			)
			for _, source := range tt.sources {
				sources = append(sources, []byte(source))
				documents = append(documents, mdParser.Parse([]byte(source)))
			}
			if got := firstHeadingText(documents, sources); got != tt.want {
				t.Errorf("firstHeadingText() = %q, want %q", got, tt.want)
			}
		})
	}
}
