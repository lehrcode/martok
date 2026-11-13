package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v3/formatters/html"
	"github.com/alecthomas/kingpin/v2"
	"github.com/lehrcode/martok/slugger"
	"github.com/lehrcode/martok/toc"
	"github.com/yuin/goldmark-diagram"
	"github.com/yuin/goldmark-highlighting/v3"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type Document struct {
	Title         string
	Variables     map[string]string
	Stylesheets   []string
	ScriptModules []string
	Body          template.HTML
}

const (
	formatHTML = "html"
	formatPDF  = "pdf"
)

var (
	inputFiles     []string
	outputFilename string
	outputFormat   string
	pdfMargin      float64
	document       = Document{Variables: make(map[string]string)}

	//go:embed template.gohtml
	outputTemplateSource   string
	outputTemplateFilename string
	outputTemplate         *template.Template
	generateTOC            bool
	titleOfTOC             string
	minDepthOfTOC          int
	maxDepthOfTOC          int
)

func main() {
	kingpin.Flag("variable", "extra template variable").Short('e').StringMapVar(&document.Variables)
	kingpin.Flag("stylesheet", "stylesheet URI").Short('s').StringsVar(&document.Stylesheets)
	kingpin.Flag("script-module", "JavaScript module URI").Short('j').StringsVar(&document.ScriptModules)
	kingpin.Flag("output", "output filename").Short('o').StringVar(&outputFilename)
	kingpin.Flag("format", "output format (default: pdf if the output filename ends with .pdf, otherwise html)").Short('f').EnumVar(&outputFormat, formatHTML, formatPDF)
	kingpin.Flag("margin", "page margins of the PDF in inches").Default("0.4").Float64Var(&pdfMargin)
	kingpin.Flag("title", "document title (default: first heading)").Short('t').StringVar(&document.Title)
	kingpin.Flag("template", "HTML template filename").Short('m').StringVar(&outputTemplateFilename)
	kingpin.Flag("toc", "generate table-of-contents (and the outline of a PDF)").BoolVar(&generateTOC)
	kingpin.Flag("toc-title", "table-of-contents title").Default("Contents").StringVar(&titleOfTOC)
	kingpin.Flag("toc-min-depth", "minimum headline depth included in table-of-contents").IntVar(&minDepthOfTOC)
	kingpin.Flag("toc-max-depth", "maximum headline depth included in table-of-contents").IntVar(&maxDepthOfTOC)
	kingpin.Arg("input", "Markdown filenames or a single HTML filename to convert to PDF").Required().ExistingFilesVar(&inputFiles)
	kingpin.Parse()

	// A single HTML file is not converted from Markdown but printed as it is.
	htmlInput := len(inputFiles) == 1 && isHTMLFilename(inputFiles[0])
	format, err := resolveFormat(outputFormat, outputFilename, htmlInput)
	if err != nil {
		log.Fatal(err)
	}
	if pdfMargin < 0 {
		log.Fatalf("invalid margin %v: must not be negative", pdfMargin)
	}
	if outputFilename == "" {
		outputFilename = strings.TrimSuffix(filepath.Base(inputFiles[0]), filepath.Ext(inputFiles[0])) + "." + format
	}

	if outputTemplateFilename != "" {
		log.Printf("Output template filename: %s\n", outputTemplateFilename)
		if t, err := template.ParseFiles(outputTemplateFilename); err != nil {
			log.Fatal(err)
		} else {
			outputTemplate = t
		}
	} else {
		outputTemplate = template.Must(template.New("template.gohtml").Parse(outputTemplateSource))
	}

	var (
		output  []byte
		options = pdfOptions{Margin: pdfMargin, Outline: generateTOC}
	)
	switch {
	case format == formatHTML:
		output = renderHTML()
	case htmlInput:
		options.Footer = renderPageFooter()
		output, err = printToPDF(filesHandler(inputDirs()), "/"+url.PathEscape(filepath.Base(inputFiles[0])), options)
	default:
		// The footer is rendered after the document, because rendering the
		// document sets the default document title.
		html := renderHTML()
		options.Footer = renderPageFooter()
		output, err = printToPDF(documentHandler(html, filesHandler(inputDirs())), "/", options)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(outputFilename, output, 0644); err != nil {
		log.Fatal(err)
	}
}

// renderPageFooter returns the footer template for the pages of a PDF. It is
// the template "page-footer" of the output template. If the output template
// does not define it, there is no footer and the result is an empty string.
func renderPageFooter() string {
	if outputTemplate.Lookup("page-footer") == nil {
		return ""
	}
	var footer strings.Builder
	if err := outputTemplate.ExecuteTemplate(&footer, "page-footer", document); err != nil {
		log.Fatal(err)
	}
	return footer.String()
}

// isHTMLFilename reports whether the filename has the extension of an HTML file.
func isHTMLFilename(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".html" || ext == ".htm"
}

// resolveFormat returns the output format. Without an explicit format it is
// derived from the extension of the output filename. An HTML input file can
// only be converted to PDF.
func resolveFormat(format, outputFilename string, htmlInput bool) (string, error) {
	if format == "" {
		switch {
		case strings.EqualFold(filepath.Ext(outputFilename), ".pdf"):
			format = formatPDF
		case htmlInput && !isHTMLFilename(outputFilename):
			format = formatPDF
		default:
			format = formatHTML
		}
	}
	if htmlInput && format != formatPDF {
		return "", fmt.Errorf("an HTML file can only be converted to %s", formatPDF)
	}
	return format, nil
}

// inputDirs returns the directories of the input files without duplicates.
// They are served to Chrome, so that the relative URIs of embedded files work.
func inputDirs() []string {
	var dirs []string
	for _, inputFilename := range inputFiles {
		dir, err := filepath.Abs(filepath.Dir(inputFilename))
		if err != nil {
			log.Fatal(err)
		}
		if !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// renderHTML converts the Markdown input files into one HTML document.
func renderHTML() []byte {
	var parserExtensions = []parser.Extension{
		extension.NewGFMParser(),
		slugger.NewParser(),
	}

	var rendererExtensions = []html.Extension{
		extension.NewGFMHTMLRenderer(),
		diagram.NewHTMLRenderer(),
		highlighting.NewHTMLRenderer(
			highlighting.WithFormatterOptions(chromahtml.WithClasses(true)),
			highlighting.WithExcludeLanguages("mermaid", "plantuml"),
		),
	}

	if generateTOC {
		parserExtensions = append(parserExtensions, toc.NewParser(
			toc.WithTitle(titleOfTOC),
			toc.WithMinDepth(minDepthOfTOC),
			toc.WithMaxDepth(maxDepthOfTOC),
		))
		rendererExtensions = append(rendererExtensions, toc.NewHTMLRenderer())
	}

	var (
		mdParser = parser.New(parser.WithExtensions(parserExtensions...), parser.WithAutoHeadingID())
		// All input files are parsed with the same context: it makes the heading
		// ids unique across the files and lets the table-of-contents collect the
		// headings of all of them.
		mdParserContext = parser.WithContext(parser.NewContext())
		mdHtmlRenderer  = html.New(html.WithExtensions(rendererExtensions...), html.WithUnsafe())
	)

	// All input files have to be parsed before the first one is rendered,
	// because the table-of-contents at the beginning of the first document
	// also lists the headings of the following ones.
	var (
		sources   = make([][]byte, 0, len(inputFiles))
		documents = make([]ast.Node, 0, len(inputFiles))
	)
	for _, inputFilename := range inputFiles {
		inputBytes, err := os.ReadFile(inputFilename)
		if err != nil {
			log.Fatal(err)
		}
		sources = append(sources, inputBytes)
		documents = append(documents, mdParser.Parse(inputBytes, mdParserContext))
	}

	var convertedHtml strings.Builder
	for i, doc := range documents {
		if err := mdHtmlRenderer.Render(&convertedHtml, sources[i], doc); err != nil {
			log.Fatal(err)
		}
	}

	if document.Title == "" {
		document.Title = firstHeadingText(documents, sources)
	}

	var finalHtml bytes.Buffer
	document.Body = template.HTML(convertedHtml.String())
	if err := outputTemplate.Execute(&finalHtml, document); err != nil {
		log.Fatal(err)
	}
	return finalHtml.Bytes()
}

// firstHeadingText returns the plain text of the first heading of the given
// documents that has any text. It returns an empty string if there is none.
func firstHeadingText(documents []ast.Node, sources [][]byte) string {
	var headingText string
	for i, doc := range documents {
		_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if heading, ok := node.(*ast.Heading); ok && entering {
				if headingText = toc.PlainText(heading, sources[i]); headingText != "" {
					return ast.WalkStop, nil
				}
				return ast.WalkSkipChildren, nil
			}
			return ast.WalkContinue, nil
		})
		if headingText != "" {
			break
		}
	}
	return headingText
}
