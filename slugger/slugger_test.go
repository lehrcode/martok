package slugger_test

import (
	"strings"
	"testing"

	"github.com/lehrcode/mdtohtml/slugger"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func TestIDGenerator(t *testing.T) {
	tests := []struct {
		title string
		slug  string
	}{
		{"Number 32", "number-32"},
		{"Question?", "question"},
		{"1+2=3", "123"},
		{"Special !\"#$%&(parens)=?´* chars", "special-parens-chars"},
		{"One-Hyphen", "one-hyphen"},
		{"Multiple--Hyphens", "multiple--hyphens"},
		{"Trailing hyphen-", "trailing-hyphen-"},
		{"Many   spaces  here", "many---spaces--here"},
		{"  surrounding spaces \t", "surrounding-spaces"},
		{"Forward/slash", "forwardslash"},
		{"Backward\\slash", "backwardslash"},
		{"Under_score", "under_score"},
		{"Nonbreaking Space", "nonbreakingspace"},
		{"Tab\tseparated", "tabseparated"},
		{"Über uns", "ueber-uns"},
		{"Größe & Maße", "groesse--masse"},
		{"ÄÖÜ äöü ẞß", "aeoeue-aeoeue-ssss"},
		{"Schön und Ärger", "schoen-und-aerger"},
		{"naïve", "naive"},
		{"Resumé", "resum"},
		{"Ærø Łódź", "r-d"},
		{"God is good: 神真美好", "god-is-good-"},
		{"Emoji 🚀 launch", "emoji--launch"},
		{"x² and ½", "x-and-"},
		{"Fullwidth １２３ 123", "fullwidth--123"},
		{"The `code` of *some*", "the-code-of-some"},
		{"invalid \xff utf8", "invalid--utf8"},
		{"", "heading"},
		{"?!", "heading"},
		{"神真美好", "heading"},
	}

	generator := &slugger.IDGenerator{}
	for _, tt := range tests {
		if got := string(generator.Generate([]byte(tt.title), ast.KindHeading)); got != tt.slug {
			t.Errorf("Generate(%q) = %q, want %q", tt.title, got, tt.slug)
		}
	}

	if got := string(generator.Generate([]byte("?"), ast.KindParagraph)); got != "id" {
		t.Errorf("Generate for a non-heading = %q, want %q", got, "id")
	}
}

func TestParser(t *testing.T) {
	var (
		mdParser   = parser.New(parser.WithExtensions(slugger.NewParser()), parser.WithAutoHeadingID())
		mdRenderer = html.New()
		source     = "# Über uns\n\n## Über uns\n\nSetext Überschrift\n---\n"
		want       = "<h1 id=\"ueber-uns\">Über uns</h1>\n" +
			"<h2 id=\"ueber-uns-1\">Über uns</h2>\n" +
			"<h2 id=\"setext-ueberschrift\">Setext Überschrift</h2>\n"
	)

	// with and without a context given by the caller
	for _, opts := range [][]parser.ParseOption{nil, {parser.WithContext(parser.NewContext())}} {
		var out strings.Builder
		if err := mdRenderer.RenderStringSource(&out, source, mdParser.ParseStringSource(source, opts...)); err != nil {
			t.Fatal(err)
		}
		if out.String() != want {
			t.Errorf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", out.String(), want)
		}
	}
}
