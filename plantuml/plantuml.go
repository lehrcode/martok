// Package plantuml provides a [diagram.Renderer] that renders PlantUML
// diagrams client-side: the browser loads the diagram as an image from a
// PlantUML server, so no plantuml command is required.
package plantuml

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"strings"

	"github.com/yuin/goldmark-diagram"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/util"
)

// DefaultURLPrefix is the default URL prefix of the PlantUML server that
// renders the diagrams.
const DefaultURLPrefix = "https://www.plantuml.com/plantuml/svg/"

type config struct {
	urlPrefix string
}

// Option is a functional option for [NewRenderer].
type Option func(*config)

// WithURLPrefix sets the URL prefix of the PlantUML server the encoded diagram
// source is appended to (e.g. to use a self-hosted PlantUML server, or another
// output format such as "png"). A missing trailing slash is added.
func WithURLPrefix(prefix string) Option {
	return func(c *config) {
		c.urlPrefix = prefix
	}
}

// NewRenderer returns a [diagram.Renderer] that encodes the diagram source
// using the PlantUML text encoding, appends it to the URL prefix of a PlantUML
// server and writes an `<img>` element that references the resulting URL.
//
// Note that the diagram source is sent to the PlantUML server as part of the
// URL.
func NewRenderer(opts ...Option) diagram.Renderer {
	cfg := config{urlPrefix: DefaultURLPrefix}
	for _, opt := range opts {
		opt(&cfg)
	}
	if !strings.HasSuffix(cfg.urlPrefix, "/") {
		cfg.urlPrefix += "/"
	}
	return diagram.RendererFunc(func(w util.BufWriter, source []byte, n *ast.CodeBlock, _ renderer.Context) error {
		var buf bytes.Buffer
		_, _ = n.Value.WriteTo(&buf, source)
		encoded, err := encode(bytes.TrimSpace(buf.Bytes()))
		if err != nil {
			return err
		}
		_, _ = w.WriteString(`<img src="`)
		_, _ = w.Write(util.EscapeHTML([]byte(cfg.urlPrefix)))
		_, _ = w.WriteString(encoded)
		_, _ = w.WriteString("\">\n")
		return nil
	})
}

// encoding is the base64 variant used by the PlantUML text encoding. It maps
// 3 bytes to 4 characters like the standard base64 encoding, but uses a
// different alphabet.
var encoding = base64.NewEncoding(
	"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_").WithPadding(base64.NoPadding)

// encode encodes a diagram source using the PlantUML text encoding
// (https://plantuml.com/text-encoding): the UTF-8 encoded source is compressed
// using Deflate and then encoded using [encoding].
func encode(src []byte) (string, error) {
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(src); err != nil {
		return "", err
	}
	if err := fw.Close(); err != nil {
		return "", err
	}
	// PlantUML pads the compressed data with zero bytes to a multiple of 3
	// bytes instead of using a padding character.
	for buf.Len()%3 != 0 {
		buf.WriteByte(0)
	}
	return encoding.EncodeToString(buf.Bytes()), nil
}
