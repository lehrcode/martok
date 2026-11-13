package main

import (
	"html/template"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		outputFilename string
		htmlInput      bool
		want           string
		wantErr        bool
	}{
		{name: "default is html", want: formatHTML},
		{name: "html output filename", outputFilename: "out.html", want: formatHTML},
		{name: "pdf output filename", outputFilename: "out.pdf", want: formatPDF},
		{name: "pdf output filename in upper case", outputFilename: "OUT.PDF", want: formatPDF},
		{name: "explicit pdf", format: formatPDF, want: formatPDF},
		{name: "explicit format wins over the output filename", format: formatHTML, outputFilename: "out.pdf", want: formatHTML},
		{name: "html input defaults to pdf", htmlInput: true, want: formatPDF},
		{name: "html input with pdf output filename", outputFilename: "out.pdf", htmlInput: true, want: formatPDF},
		{name: "html input with other output filename", outputFilename: "out", htmlInput: true, want: formatPDF},
		{name: "html input with html output filename", outputFilename: "out.html", htmlInput: true, wantErr: true},
		{name: "html input with explicit html", format: formatHTML, htmlInput: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFormat(tt.format, tt.outputFilename, tt.htmlInput)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderPageFooter(t *testing.T) {
	t.Cleanup(func() { outputTemplate = nil })

	// The embedded template defines the footer, but it is not part of the
	// HTML document.
	outputTemplate = template.Must(template.New("template.gohtml").Funcs(tplFuncs).Parse(outputTemplateSource))
	if footer := renderPageFooter(); !strings.Contains(footer, `class="pageNumber"`) {
		t.Errorf("renderPageFooter() = %q, want the page number", footer)
	}
	var html strings.Builder
	if err := outputTemplate.Execute(&html, Document{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html.String(), "pageNumber") {
		t.Error("the HTML document contains the page footer")
	}

	// A template without the footer template has no footer.
	outputTemplate = template.Must(template.New("custom.gohtml").Funcs(tplFuncs).Parse("{{ .Body }}"))
	if footer := renderPageFooter(); footer != "" {
		t.Errorf("renderPageFooter() = %q, want no footer", footer)
	}
}

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
