package main

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRenderPageTemplate(t *testing.T) {
	savedDocument := document
	t.Cleanup(func() { outputTemplate, document = nil, savedDocument })
	document = Document{
		Author:       "Erika Mustermann",
		LastModified: time.Date(2026, time.March, 7, 23, 59, 0, 0, time.Local),
		PDFMargin:    0.8,
	}

	// The embedded template defines the header and the footer, but they are
	// not part of the HTML document.
	outputTemplate = template.Must(template.New("template.gohtml").Funcs(tplFuncs).Parse(outputTemplateSource))
	if header := renderPageTemplate("page-header"); header != "<span></span>" {
		t.Errorf(`renderPageTemplate("page-header") = %q, want an empty span`, header)
	}
	footer := renderPageTemplate("page-footer")
	for _, want := range []string{"Erika Mustermann", `class="pageNumber"`, `class="totalPages"`, "2026-03-07", "padding: 0 0.8in"} {
		if !strings.Contains(footer, want) {
			t.Errorf(`renderPageTemplate("page-footer") = %q, want it to contain %q`, footer, want)
		}
	}
	if strings.Contains(footer, "<img") {
		t.Errorf(`renderPageTemplate("page-footer") = %q, want no image`, footer)
	}

	// An AI generated document has the image in front of the author.
	document.AIGenerated = true
	footer = renderPageTemplate("page-footer")
	image, author := strings.Index(footer, `<img src="data:image/svg`), strings.Index(footer, "Erika Mustermann")
	if image < 0 || image > author {
		t.Errorf("the page footer has the image at %d and the author at %d, want the image in front of the author", image, author)
	}
	var html strings.Builder
	if err := outputTemplate.Execute(&html, document); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html.String(), "pageNumber") {
		t.Error("the HTML document contains the page footer")
	}
	if !strings.Contains(html.String(), `<meta name="author" content="Erika Mustermann">`) {
		t.Error("the HTML document does not name the author")
	}

	// A template without the header and footer templates has none.
	outputTemplate = template.Must(template.New("custom.gohtml").Funcs(tplFuncs).Parse("{{ .Body }}"))
	for _, name := range []string{"page-header", "page-footer"} {
		if rendered := renderPageTemplate(name); rendered != "" {
			t.Errorf("renderPageTemplate(%q) = %q, want an empty string", name, rendered)
		}
	}
}

func TestRenderHTMLDiagrams(t *testing.T) {
	savedDocument, savedInputFiles := document, inputFiles
	t.Cleanup(func() {
		outputTemplate, document, inputFiles = nil, savedDocument, savedInputFiles
		mermaidModuleURL, plantUMLURLPrefix = "", ""
	})
	outputTemplate = template.Must(template.New("custom.gohtml").Funcs(tplFuncs).Parse("{{ .Body }}"))
	input := filepath.Join(t.TempDir(), "diagrams.md")
	source := "```mermaid\ngraph TD; A-->B;\n```\n\n```plantuml\n@startuml\nBob -> Alice : hello\n@enduml\n```\n"
	if err := os.WriteFile(input, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	inputFiles = []string{input}

	tests := []struct {
		name              string
		mermaidModuleURL  string
		plantUMLURLPrefix string
		want              []string
	}{
		{
			name: "defaults",
			want: []string{
				"import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@latest/dist/mermaid.esm.min.mjs';",
				`<img src="https://www.plantuml.com/plantuml/svg/`,
			},
		},
		{
			name:              "custom URLs",
			mermaidModuleURL:  "/js/mermaid.esm.min.mjs",
			plantUMLURLPrefix: "http://localhost:8080/png/",
			want: []string{
				"import mermaid from '/js/mermaid.esm.min.mjs';",
				`<img src="http://localhost:8080/png/`,
			},
		},
		{
			name:              "PlantUML URL prefix without a trailing slash",
			plantUMLURLPrefix: "http://localhost:8080/png",
			want:              []string{`<img src="http://localhost:8080/png/`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document = Document{}
			mermaidModuleURL, plantUMLURLPrefix = tt.mermaidModuleURL, tt.plantUMLURLPrefix
			html := string(renderHTML())
			for _, want := range append(tt.want, `<pre class="mermaid">`) {
				if !strings.Contains(html, want) {
					t.Errorf("renderHTML() = %q, want it to contain %q", html, want)
				}
			}
		})
	}
}

func TestLastModified(t *testing.T) {
	dir := t.TempDir()
	older, newer := filepath.Join(dir, "older.md"), filepath.Join(dir, "newer.md")
	want := time.Date(2026, time.March, 7, 12, 0, 0, 0, time.Local)
	for filename, modified := range map[string]time.Time{older: want.Add(-48 * time.Hour), newer: want} {
		writeTestFile(t, filename, "text")
		if err := os.Chtimes(filename, modified, modified); err != nil {
			t.Fatal(err)
		}
	}

	for _, filenames := range [][]string{{older, newer}, {newer, older}} {
		got, err := lastModified(filenames)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(want) {
			t.Errorf("lastModified() = %v, want %v", got, want)
		}
	}
	if _, err := lastModified([]string{filepath.Join(dir, "missing.md")}); err == nil {
		t.Error("lastModified() of a missing file: want an error")
	}
}

func TestFileOwners(t *testing.T) {
	dir := t.TempDir()
	first, second := filepath.Join(dir, "first.md"), filepath.Join(dir, "second.md")
	writeTestFile(t, first, "text")
	writeTestFile(t, second, "text")

	owner, err := fileOwner(first)
	if err != nil {
		t.Skipf("owner of the file not available: %v", err)
	}
	if owner == "" {
		t.Fatal("fileOwner() returned no username")
	}
	// Both files have the same owner, who is named once. A file without a
	// known owner is skipped.
	if got := fileOwners([]string{first, second, filepath.Join(dir, "missing.md")}); got != owner {
		t.Errorf("fileOwners() = %q, want %q", got, owner)
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
