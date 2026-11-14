package plantuml_test

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/lehrcode/martok/plantuml"
	"github.com/yuin/goldmark-diagram"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"

// decode decodes a diagram source encoded using the PlantUML text encoding.
func decode(t *testing.T, encoded string) string {
	t.Helper()
	compressed, err := base64.NewEncoding(alphabet).WithPadding(base64.NoPadding).DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	src, err := io.ReadAll(flate.NewReader(bytes.NewReader(compressed)))
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// render converts the Markdown source with the PlantUML renderer.
func render(t *testing.T, source string, opts ...plantuml.Option) string {
	t.Helper()
	var (
		mdParser   = parser.New()
		mdRenderer = html.New(html.WithExtensions(diagram.NewHTMLRenderer(
			diagram.WithRenderer(diagram.LanguagePlantUML, plantuml.NewRenderer(opts...)),
		)))
		rendered strings.Builder
	)
	if err := mdRenderer.Render(&rendered, []byte(source), mdParser.Parse([]byte(source))); err != nil {
		t.Fatal(err)
	}
	return rendered.String()
}

func TestRenderer(t *testing.T) {
	tests := []struct {
		name       string
		opts       []plantuml.Option
		diagram    string
		wantPrefix string
	}{
		{
			name:       "default URL prefix",
			diagram:    "@startuml\nBob -> Alice : hello\n@enduml",
			wantPrefix: plantuml.DefaultURLPrefix,
		},
		{
			name:       "custom URL prefix",
			opts:       []plantuml.Option{plantuml.WithURLPrefix("https://example.com/plantuml/png/")},
			diagram:    "@startuml\nBob -> Alice : hello\n@enduml",
			wantPrefix: "https://example.com/plantuml/png/",
		},
		{
			name:       "URL prefix without a trailing slash",
			opts:       []plantuml.Option{plantuml.WithURLPrefix("https://example.com/plantuml/png")},
			diagram:    "@startuml\nBob -> Alice : hello\n@enduml",
			wantPrefix: "https://example.com/plantuml/png/",
		},
		{
			name:       "empty URL prefix",
			opts:       []plantuml.Option{plantuml.WithURLPrefix("")},
			diagram:    "@startuml\nBob -> Alice : hello\n@enduml",
			wantPrefix: "/",
		},
		{
			name:       "URL prefix is escaped",
			opts:       []plantuml.Option{plantuml.WithURLPrefix(`https://example.com/a&b/"c"/`)},
			diagram:    "@startuml\nBob -> Alice : hello\n@enduml",
			wantPrefix: "https://example.com/a&amp;b/&quot;c&quot;/",
		},
		{
			name:       "non-ASCII and HTML characters",
			diagram:    "@startuml\nBob -> Alice : Grüße & \"<b>\"\n@enduml",
			wantPrefix: plantuml.DefaultURLPrefix,
		},
		{
			name:       "empty diagram",
			diagram:    "",
			wantPrefix: plantuml.DefaultURLPrefix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render(t, "```plantuml\n"+tt.diagram+"\n```\n", tt.opts...)

			encoded, ok := strings.CutPrefix(got, `<img src="`+tt.wantPrefix)
			if !ok {
				t.Fatalf("got %q, want an image with the URL prefix %q", got, tt.wantPrefix)
			}
			if encoded, ok = strings.CutSuffix(encoded, "\">\n"); !ok {
				t.Fatalf("got %q, want an image", got)
			}
			if len(encoded)%4 != 0 {
				t.Errorf("encoded diagram %q is not padded to a multiple of 4 characters", encoded)
			}
			if strings.Trim(encoded, alphabet) != "" {
				t.Errorf("encoded diagram %q contains characters that are not URL-safe", encoded)
			}
			if src := decode(t, encoded); src != tt.diagram {
				t.Errorf("decoded diagram = %q, want %q", src, tt.diagram)
			}
		})
	}
}

// TestDecodeReference checks the decoding used by the tests against a diagram
// encoded by the PlantUML server itself.
func TestDecodeReference(t *testing.T) {
	const want = "@startuml\nBob -> Alice : hello\n@enduml"
	if got := decode(t, "SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiX8pSd9vt98pKi1IW80"); got != want {
		t.Errorf("decoded diagram = %q, want %q", got, want)
	}
}

func TestOtherLanguages(t *testing.T) {
	const want = "<pre><code class=\"language-go\">package main\n</code></pre>\n"
	if got := render(t, "```go\npackage main\n```\n"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
