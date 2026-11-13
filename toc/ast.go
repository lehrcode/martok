// Package toc provides goldmark extensions that generate a table of contents
// from the headings of a Markdown document.
//
// The package consists of a parser extension ([NewParser]) that collects the
// headings into a [TableOfContents] node at the beginning of the document and
// a renderer extension ([NewHTMLRenderer]) that renders this node as HTML.
package toc

import (
	gast "github.com/yuin/goldmark/v2/ast"
)

// A TableOfContents struct represents a table of contents of a Markdown
// document. Its children are [Item] nodes.
type TableOfContents struct {
	gast.BaseBlock

	// Title is the plain text title of the table of contents.
	// If Title is empty, no title is rendered.
	Title string

	// TitleID is the element id of the title.
	TitleID string

	// ListID is the element id of the top level list.
	ListID string
}

// AddItem adds the given item to the table of contents. The item becomes a
// child of the nearest preceding item with a lower level, so skipped heading
// levels do not produce empty intermediate entries.
func (n *TableOfContents) AddItem(item *Item) {
	var parent gast.Node = n
	for last := n.LastChild(); last != nil; last = last.LastChild() {
		lastItem, ok := last.(*Item)
		if !ok || lastItem.Level >= item.Level {
			break
		}
		parent = lastItem
	}
	parent.AppendChild(item)
}

// Dump implements Node.Dump.
func (n *TableOfContents) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Title":   n.Title,
		"TitleID": n.TitleID,
		"ListID":  n.ListID,
	})
}

// KindTableOfContents is a NodeKind of the TableOfContents node.
var KindTableOfContents = gast.NewNodeKind("TableOfContents")

// Kind implements Node.Kind.
func (n *TableOfContents) Kind() gast.NodeKind {
	return KindTableOfContents
}

// NewTableOfContents returns a new TableOfContents node.
func NewTableOfContents(title string) *TableOfContents {
	n := &TableOfContents{Title: title}
	n.Init(n)
	return n
}

// An Item struct represents an entry of a [TableOfContents] that links to a
// heading. Its children are the items of the subordinate headings.
type Item struct {
	gast.BaseBlock

	// Level is the level of the heading this item links to.
	Level int

	// ID is the element id of the heading this item links to.
	ID string

	// Title is the plain text of the heading this item links to.
	Title string
}

// Dump implements Node.Dump.
func (n *Item) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Level": n.Level,
		"ID":    n.ID,
		"Title": n.Title,
	})
}

// KindItem is a NodeKind of the Item node.
var KindItem = gast.NewNodeKind("TableOfContentsItem")

// Kind implements Node.Kind.
func (n *Item) Kind() gast.NodeKind {
	return KindItem
}

// NewItem returns a new Item node.
func NewItem(level int, id, title string) *Item {
	n := &Item{Level: level, ID: id, Title: title}
	n.Init(n)
	return n
}
