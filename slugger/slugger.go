// Package slugger provides a goldmark parser extension that generates
// GitHub-style element ids that are restricted to ASCII characters.
package slugger

import (
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/util"
)

var _ parser.IDGenerator = (*IDGenerator)(nil)

// combiningDiaeresis follows a base letter in decomposed umlauts ("ä").
const combiningDiaeresis = '̈'

// transliterations maps non-ASCII letters to their ASCII replacements.
var transliterations = map[rune]string{
	'ä': "ae", 'Ä': "ae",
	'ö': "oe", 'Ö': "oe",
	'ü': "ue", 'Ü': "ue",
	'ß': "ss", 'ẞ': "ss",
}

// An IDGenerator generates GitHub-style element ids: the value is lowercased,
// spaces are replaced with hyphens and everything except ASCII letters, ASCII
// digits, underscores and hyphens is removed.
//
// German umlauts are not removed but transliterated, so "Größe" becomes
// "groesse" and "Über uns" becomes "ueber-uns".
type IDGenerator struct {
}

// Generate implements parser.IDGenerator.
func (g *IDGenerator) Generate(value []byte, kind ast.NodeKind) []byte {
	value = util.TrimRightSpace(util.TrimLeftSpace(value))

	var sb strings.Builder
	sb.Grow(len(value))
	for len(value) > 0 {
		r, size := utf8.DecodeRune(value)
		value = value[size:]
		switch {
		case r == ' ' || r == '-':
			sb.WriteByte('-')
		case r == '_' || '0' <= r && r <= '9' || 'a' <= r && r <= 'z':
			sb.WriteByte(byte(r))
		case 'A' <= r && r <= 'Z':
			sb.WriteByte(byte(r) + ('a' - 'A'))
		case r == combiningDiaeresis:
			if s := sb.String(); s != "" && strings.IndexByte("aou", s[len(s)-1]) >= 0 {
				sb.WriteByte('e')
			}
		default:
			// everything without a transliteration is removed
			sb.WriteString(transliterations[r])
		}
	}

	if sb.Len() == 0 {
		if kind == ast.KindHeading {
			return []byte("heading")
		}
		return []byte("id")
	}
	return []byte(sb.String())
}

type parserExtension struct {
}

// NewParser returns a new parser.Extension that makes the parser generate
// element ids with an [IDGenerator], for example for the headings if the
// parser has been created with parser.WithAutoHeadingID.
func NewParser() parser.Extension {
	return &parserExtension{}
}

// ParserOptions implements parser.Extension.
func (e *parserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithIDGenerator(&IDGenerator{}),
	}
}
