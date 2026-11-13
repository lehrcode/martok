package toc

import (
	"io"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/util"
)

type htmlRendererExtension struct {
}

// NewHTMLRenderer returns a new html.Extension that renders [TableOfContents]
// nodes as a nav element with nested lists of links to the headings:
//
//	<nav class="toc" aria-labelledby="toc-title">
//	<h1 id="toc-title">Contents</h1>
//	<ul id="toc-list">
//	<li><a href="#heading">Heading</a>
//	<ul>
//	<li><a href="#subheading">Subheading</a></li>
//	</ul>
//	</li>
//	</ul>
//	</nav>
//
// A table of contents without items is not rendered.
func NewHTMLRenderer() html.Extension {
	return &htmlRendererExtension{}
}

// RendererOptions implements html.Extension.
func (e *htmlRendererExtension) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[gast.NodeKind]html.NodeRenderer{
			KindTableOfContents: html.NodeRendererFunc(e.renderTableOfContents),
			KindItem:            html.NodeRendererFunc(e.renderItem),
		}),
	}
}

func (e *htmlRendererExtension) renderTableOfContents(
	writer io.Writer, _ []byte, node gast.Node, entering bool, rc renderer.Context,
) (gast.WalkStatus, error) {
	n := node.(*TableOfContents)
	if !n.HasChildren() {
		return gast.WalkSkipChildren, nil
	}
	w := writer.(util.BufWriter)
	if !entering {
		_, _ = w.WriteString("</ul>\n</nav>\n")
		return gast.WalkContinue, nil
	}
	tw := html.ContextTextWriter(rc)
	_, _ = w.WriteString(`<nav class="toc"`)
	if n.Title != "" {
		_, _ = w.WriteString(` aria-labelledby="`)
		_, _ = tw.WriteString(n.TitleID)
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString(">\n")
	if n.Title != "" {
		_, _ = w.WriteString(`<h1 id="`)
		_, _ = tw.WriteString(n.TitleID)
		_, _ = w.WriteString(`">`)
		_, _ = tw.WriteString(n.Title)
		_, _ = w.WriteString("</h1>\n")
	}
	_, _ = w.WriteString(`<ul id="`)
	_, _ = tw.WriteString(n.ListID)
	_, _ = w.WriteString("\">\n")
	return gast.WalkContinue, nil
}

func (e *htmlRendererExtension) renderItem(
	writer io.Writer, _ []byte, node gast.Node, entering bool, rc renderer.Context,
) (gast.WalkStatus, error) {
	n := node.(*Item)
	w := writer.(util.BufWriter)
	if !entering {
		if n.HasChildren() {
			_, _ = w.WriteString("</ul>\n")
		}
		_, _ = w.WriteString("</li>\n")
		return gast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<li><a href="#`)
	_, _ = html.ContextLinkURLWriter(rc).WriteString(n.ID)
	_, _ = w.WriteString(`">`)
	_, _ = html.ContextTextWriter(rc).WriteString(n.Title)
	_, _ = w.WriteString("</a>")
	if n.HasChildren() {
		_, _ = w.WriteString("\n<ul>\n")
	}
	return gast.WalkContinue, nil
}
