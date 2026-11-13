package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"os"
	"path/filepath"
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

var (
	inputFiles     []string
	outputFilename string
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
	kingpin.Flag("title", "document title (default: first heading)").Short('t').StringVar(&document.Title)
	kingpin.Flag("template", "HTML template filename").Short('m').StringVar(&outputTemplateFilename)
	kingpin.Flag("toc", "generate table-of-contents").BoolVar(&generateTOC)
	kingpin.Flag("toc-title", "table-of-contents title").Default("Contents").StringVar(&titleOfTOC)
	kingpin.Flag("toc-min-depth", "minimum headline depth included in table-of-contents").IntVar(&minDepthOfTOC)
	kingpin.Flag("toc-max-depth", "maximum headline depth included in table-of-contents").IntVar(&maxDepthOfTOC)
	kingpin.Arg("md", "Markdown filename").Required().ExistingFilesVar(&inputFiles)
	kingpin.Parse()

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
	if outputFilename == "" {
		outputFilename = strings.TrimSuffix(filepath.Base(inputFiles[0]), filepath.Ext(inputFiles[0])) + ".html"
	}

	if err := os.WriteFile(outputFilename, finalHtml.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
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
