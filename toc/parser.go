package toc

import (
	"strings"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type parserConfig struct {
	Title    string
	MinDepth int
	MaxDepth int
}

// A ParserOption is a functional option for the parser extension.
type ParserOption func(*parserConfig)

// WithTitle is a functional option that sets the title of the table of
// contents. Without a title only the list of headings is generated.
func WithTitle(title string) ParserOption {
	return func(c *parserConfig) {
		c.Title = title
	}
}

// WithMinDepth is a functional option that sets the minimum level of the
// headings that are included in the table of contents. Values less than 1
// include headings from level 1.
func WithMinDepth(depth int) ParserOption {
	return func(c *parserConfig) {
		c.MinDepth = depth
	}
}

// WithMaxDepth is a functional option that sets the maximum level of the
// headings that are included in the table of contents. Values less than 1
// include headings up to level 6.
func WithMaxDepth(depth int) ParserOption {
	return func(c *parserConfig) {
		c.MaxDepth = depth
	}
}

type parserExtension struct {
	config parserConfig
}

// NewParser returns a new parser.Extension that inserts a [TableOfContents]
// node at the beginning of the document. The node has an [Item] for every
// heading of the document.
//
// If multiple sources are parsed with the same parser.Context, they are
// handled as parts of one document: only the first parsed document gets a
// TableOfContents node, which then also collects the headings of the following
// documents. Such documents must be rendered after all of them have been
// parsed.
func NewParser(opts ...ParserOption) parser.Extension {
	e := &parserExtension{}
	for _, opt := range opts {
		opt(&e.config)
	}
	return e
}

// ParserOptions implements parser.Extension.
func (e *parserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithASTTransformers(
			util.Prioritized[parser.ASTTransformer](&astTransformer{config: e.config}, 1000),
		),
	}
}

var contextKey = parser.NewContextKey()

type astTransformer struct {
	config parserConfig
}

// Transform implements parser.ASTTransformer.
func (t *astTransformer) Transform(doc *gast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	toc, ok := pc.Get(contextKey).(*TableOfContents)
	if !ok {
		toc = NewTableOfContents(t.config.Title)
		toc.TitleID = string(pc.IDs().Generate([]byte("toc-title"), KindTableOfContents))
		toc.ListID = string(pc.IDs().Generate([]byte("toc-list"), KindTableOfContents))
		doc.InsertBefore(doc.FirstChild(), toc)
		pc.Set(contextKey, toc)
	}

	_ = gast.Walk(doc, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *TableOfContents:
			return gast.WalkSkipChildren, nil
		case *gast.Heading:
			if t.includes(n.Level) {
				if title := headingTitle(n, source); title != "" {
					toc.AddItem(NewItem(n.Level, headingID(n, title, source, pc), title))
				}
			}
			return gast.WalkSkipChildren, nil
		}
		return gast.WalkContinue, nil
	})
}

func (t *astTransformer) includes(level int) bool {
	if t.config.MinDepth > 0 && level < t.config.MinDepth {
		return false
	}
	if t.config.MaxDepth > 0 && level > t.config.MaxDepth {
		return false
	}
	return true
}

// headingID returns the element id of the given heading. Headings without an
// id (the parser has been created without parser.WithAutoHeadingID) get a
// generated one, because an item can not link to a heading without an id.
func headingID(heading *gast.Heading, title string, source []byte, pc parser.Context) string {
	if id, ok := heading.Attribute("id"); ok {
		return strings.Clone(id.Value(source))
	}
	id := string(pc.IDs().Generate([]byte(title), gast.KindHeading))
	heading.SetAttribute("id", text.NewMultiLineValueFromString(id, text.IdentityDecoder))
	return id
}

// headingTitle returns the plain text of the given heading without any inline
// markup.
func headingTitle(heading *gast.Heading, source []byte) string {
	var sb strings.Builder
	writePlainText(&sb, heading, source)
	return strings.TrimSpace(sb.String())
}

func writePlainText(sb *strings.Builder, parent gast.Node, source []byte) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *gast.Text:
			sb.WriteString(n.Value.Value(source))
			if n.SoftLineBreak() || n.HardLineBreak() {
				sb.WriteByte(' ')
			}
		case *gast.CodeSpan:
			sb.WriteString(n.Value.Value(source))
		case *gast.AutoLink:
			sb.WriteString(n.Label.Value(source))
		default:
			writePlainText(sb, n, source)
		}
	}
}
