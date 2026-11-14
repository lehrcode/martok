# Martok

Command line tool that converts Markdown files into one HTML document or, by
printing it with a headless Chrome, into a PDF. The usage and all flags are
documented in `README.md`.

## Commands

```bash
go build ./...
go vet ./...
go test ./...
gofmt -l .                # must print nothing
GOOS=linux go vet .       # checks owner_unix.go, which is not built on Windows
```

`TestPrintToPDF` prints with a real browser and takes a few seconds. It is
skipped if no Chrome, Chromium or Edge is installed.

The development machine is Windows.

## Layout

- `main.go` – flags, the `Document` passed to the templates, Markdown to HTML
  (`renderHTML`), choice of the output format, author and modification date.
- `pdf.go` – temporary web server and printing with chromedp (`printToPDF`).
- `owner_windows.go`, `owner_unix.go` – `fileOwner`, the owner of a file. The
  standard library has no portable way to read it, hence the build-specific
  files and `golang.org/x/sys` on Windows.
- `template.gohtml` – embedded HTML template with the sub-templates
  `page-header` and `page-footer`.
- `assets/` – files embedded with `go:embed`.
- `toc/`, `slugger/` – Goldmark extensions for the table of contents and the
  heading ids.
- `plantuml/` – `diagram.Renderer` for goldmark-diagram that embeds a PlantUML
  diagram as an image loaded from a PlantUML server (`--plantuml-url-prefix`).
  MermaidJS diagrams use the client renderer of goldmark-diagram
  (`--mermaid-module-url`).

## How the PDF output works

- The format is `pdf` if `-o` ends with `.pdf` or with `--format pdf`. A single
  `.html`/`.htm` input file is not rendered from Markdown but printed as it is.
- The document is served at `/` by a web server on `127.0.0.1` with a port
  chosen by the operating system. All other paths are served from the
  directories of the input files (first directory that has the file wins), so
  that relative URIs of embedded files work.
- The page is printed after the lifecycle event `networkIdle` and not at the
  load event, because script modules such as Mermaid fetch further modules and
  render after the page has loaded.
- The paper size is `chromedp.PaperA4`. `PDFPreferCSSPageSize` lets a
  stylesheet override it with `@page`.

## Page header and footer

- `page-header` and `page-footer` are only rendered for PDF output
  (`renderPageTemplate`) and passed to Chrome as header and footer templates.
  They are never part of the HTML document. A template without them gets no
  header or footer.
- Chrome renders them in a separate document. The stylesheets of the page do
  not apply, the default font size is 0 and the box spans the full page width,
  so they carry inline styles, an explicit font size and the horizontal padding
  `{{ .PDFMargin }}in`.
- Images are only shown as data URIs, URLs are not loaded. That is why
  `assets/ai-generated.svg` is embedded and exposed by the template function
  `aiGeneratedImage`.
- If only one of the two templates is set, Chrome prints its own default for
  the other one. `printToPDF` therefore passes `<span></span>` for the missing
  one.
- They are rendered after `renderHTML`, which sets the default document title.

## chromedp

The project uses chromedp v0.23, whose API differs a lot from older versions
and from most examples on the web: actions are generic (`Action[T]`),
`chromedp.Run` returns a value, `chromedp.Do` runs actions without one, and
events are read with `chromedp.Events`. Read the package documentation in the
module cache (`go doc github.com/chromedp/chromedp`) before changing `pdf.go`.

- Use the helpers of chromedp (`PrintToPDF`, `PDFMargin`, `PDFPaper`,
  `PaperA4`, `PDFOutlineAndTagged`, …) instead of own constants or raw
  `page.PrintToPDFParams`.
- The browser lives as long as the context of the first `Run`/`Do`. Start it
  with `chromedp.Do(ctx)` first and derive the timeout context afterwards.

## Conventions

- Code comments, `README.md` and commit messages are in English. The maintainer
  writes in German.
- Keep solutions simple.
- Flags bind directly to fields of `Document` where a template needs the value.
  PDF-only flags have the prefix `pdf-`.
- Update the flag table and the related section of `README.md` with every
  change of a flag or of the template data.
- Tests are table-driven. Functions that use the globals `outputTemplate` and
  `document` are tested by setting and restoring them.
- Commit only when asked, on the current feature branch. Commit messages have a
  short English subject and a body that explains the change.
