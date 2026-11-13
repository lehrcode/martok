package toc_test

import (
	"strings"
	"testing"

	"github.com/lehrcode/mdtohtml/toc"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

// convert parses all sources with a shared context and renders them into one
// HTML document, like mdtohtml does.
func convert(t *testing.T, parserOptions []parser.Option, tocOptions []toc.ParserOption, sources ...string) string {
	t.Helper()
	parserOptions = append(parserOptions, parser.WithExtensions(toc.NewParser(tocOptions...)))
	var (
		mdParser   = parser.New(parserOptions...)
		mdRenderer = html.New(html.WithExtensions(toc.NewHTMLRenderer()))
		context    = parser.NewContext()
	)
	documents := make([]func(*strings.Builder) error, 0, len(sources))
	for _, source := range sources {
		document := mdParser.ParseStringSource(source, parser.WithContext(context))
		documents = append(documents, func(out *strings.Builder) error {
			return mdRenderer.RenderStringSource(out, source, document)
		})
	}
	var out strings.Builder
	for _, render := range documents {
		if err := render(&out); err != nil {
			t.Fatal(err)
		}
	}
	return out.String()
}

func TestTableOfContents(t *testing.T) {
	autoHeadingID := []parser.Option{parser.WithAutoHeadingID()}

	tests := []struct {
		name          string
		parserOptions []parser.Option
		tocOptions    []toc.ParserOption
		sources       []string
		want          string
	}{
		{
			name:          "nested headings",
			parserOptions: autoHeadingID,
			tocOptions:    []toc.ParserOption{toc.WithTitle("Contents")},
			sources:       []string{"# One\n\ntext\n\n## One A\n\n### One A I\n\n## One B\n\n# Two\n"},
			want: `<nav class="toc" aria-labelledby="toc-title">
<h1 id="toc-title">Contents</h1>
<ul id="toc-list">
<li><a href="#one">One</a>
<ul>
<li><a href="#one-a">One A</a>
<ul>
<li><a href="#one-a-i">One A I</a></li>
</ul>
</li>
<li><a href="#one-b">One B</a></li>
</ul>
</li>
<li><a href="#two">Two</a></li>
</ul>
</nav>
<h1 id="one">One</h1>
<p>text</p>
<h2 id="one-a">One A</h2>
<h3 id="one-a-i">One A I</h3>
<h2 id="one-b">One B</h2>
<h1 id="two">Two</h1>
`,
		},
		{
			name:          "without title",
			parserOptions: autoHeadingID,
			sources:       []string{"# One\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#one">One</a></li>
</ul>
</nav>
<h1 id="one">One</h1>
`,
		},
		{
			name:          "min and max depth",
			parserOptions: autoHeadingID,
			tocOptions:    []toc.ParserOption{toc.WithMinDepth(2), toc.WithMaxDepth(3)},
			sources:       []string{"# One\n\n## Two\n\n### Three\n\n#### Four\n\n## Two again\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#two">Two</a>
<ul>
<li><a href="#three">Three</a></li>
</ul>
</li>
<li><a href="#two-again">Two again</a></li>
</ul>
</nav>
<h1 id="one">One</h1>
<h2 id="two">Two</h2>
<h3 id="three">Three</h3>
<h4 id="four">Four</h4>
<h2 id="two-again">Two again</h2>
`,
		},
		{
			name:          "skipped and ascending levels",
			parserOptions: autoHeadingID,
			sources:       []string{"### Three\n\n# One\n\n### Three again\n\n## Two\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#three">Three</a></li>
<li><a href="#one">One</a>
<ul>
<li><a href="#three-again">Three again</a></li>
<li><a href="#two">Two</a></li>
</ul>
</li>
</ul>
</nav>
<h3 id="three">Three</h3>
<h1 id="one">One</h1>
<h3 id="three-again">Three again</h3>
<h2 id="two">Two</h2>
`,
		},
		{
			name:          "no headings",
			parserOptions: autoHeadingID,
			tocOptions:    []toc.ParserOption{toc.WithTitle("Contents")},
			sources:       []string{"just text\n"},
			want:          "<p>just text</p>\n",
		},
		{
			name:          "empty document",
			parserOptions: autoHeadingID,
			sources:       []string{""},
			want:          "",
		},
		{
			name:          "inline markup and escaping",
			parserOptions: autoHeadingID,
			tocOptions:    []toc.ParserOption{toc.WithTitle("A & <B>")},
			sources: []string{
				"# The `<code>` of *some* [link](https://example.com) &amp; \\*more\\*\n\nSetext\nheading\n-------\n",
			},
			want: `<nav class="toc" aria-labelledby="toc-title">
<h1 id="toc-title">A &amp; &lt;B&gt;</h1>
<ul id="toc-list">
<li><a href="#the-code-of-some-linkhttpsexamplecom-amp-more">The &lt;code&gt; of some link &amp; *more*</a>
<ul>
<li><a href="#heading">Setext heading</a></li>
</ul>
</li>
</ul>
</nav>
<h1 id="the-code-of-some-linkhttpsexamplecom-amp-more">The <code>&lt;code&gt;</code> of <em>some</em> <a href="https://example.com">link</a> &amp; *more*</h1>
<h2 id="heading">Setext
heading</h2>
`,
		},
		{
			name:    "headings without id get one",
			sources: []string{"# One\n\n## One\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#one">One</a>
<ul>
<li><a href="#one-1">One</a></li>
</ul>
</li>
</ul>
</nav>
<h1 id="one">One</h1>
<h2 id="one-1">One</h2>
`,
		},
		{
			name:          "custom heading ids",
			parserOptions: []parser.Option{parser.WithAutoHeadingID(), parser.WithAttribute()},
			sources:       []string{"# One {#first}\n\n## Two {#second .chapter}\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#first">One</a>
<ul>
<li><a href="#second">Two</a></li>
</ul>
</li>
</ul>
</nav>
<h1 id="first">One</h1>
<h2 id="second" class="chapter">Two</h2>
`,
		},
		{
			name:          "headings in containers and empty headings",
			parserOptions: autoHeadingID,
			sources:       []string{"> # Quoted\n\n- ## Listed\n\n#\n"},
			want: `<nav class="toc">
<ul id="toc-list">
<li><a href="#quoted">Quoted</a>
<ul>
<li><a href="#listed">Listed</a></li>
</ul>
</li>
</ul>
</nav>
<blockquote>
<h1 id="quoted">Quoted</h1>
</blockquote>
<ul>
<li>
<h2 id="listed">Listed</h2>
</li>
</ul>
<h1 id="heading"></h1>
`,
		},
		{
			name:          "multiple documents share one table of contents",
			parserOptions: autoHeadingID,
			tocOptions:    []toc.ParserOption{toc.WithTitle("Contents")},
			sources:       []string{"# Intro\n\n## Details\n", "### Deeper\n\n# Intro\n\n## TOC list\n"},
			want: `<nav class="toc" aria-labelledby="toc-title">
<h1 id="toc-title">Contents</h1>
<ul id="toc-list">
<li><a href="#intro">Intro</a>
<ul>
<li><a href="#details">Details</a>
<ul>
<li><a href="#deeper">Deeper</a></li>
</ul>
</li>
</ul>
</li>
<li><a href="#intro-1">Intro</a>
<ul>
<li><a href="#toc-list-1">TOC list</a></li>
</ul>
</li>
</ul>
</nav>
<h1 id="intro">Intro</h1>
<h2 id="details">Details</h2>
<h3 id="deeper">Deeper</h3>
<h1 id="intro-1">Intro</h1>
<h2 id="toc-list-1">TOC list</h2>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convert(t, tt.parserOptions, tt.tocOptions, tt.sources...)
			if got != tt.want {
				t.Errorf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}
