# Martok

Markdown to HTML and PDF converter

## Installation

```bash
go install github.com/lehrcode/martok@latest
```

## Usage

```bash
martok -o output.html -t "Title" -a "Christian Winkler" -e lang=de input1.md input2.md
```

Multiple Markdown files are converted into one document.

| Flag                    | Description                                                               | Default                                               |
|-------------------------|---------------------------------------------------------------------------|-------------------------------------------------------|
| `-o`, `--output`        | filename of the document                                                  | name of the first input file with `.html` or `.pdf`   |
| `-f`, `--format`        | output format, `html` or `pdf`                                            | `pdf` for an output filename with `.pdf`, else `html` |
| `--pdf-margin`          | page margins of the PDF in inches                                         | `0.4`                                                 |
| `--pdf-outline`         | generate PDF with outline                                                 | off                                                   |
| `-t`, `--title`         | title of the document                                                     | first heading of the Markdown files                   |
| `-a`, `--author`        | author of the document                                                    | full names or usernames of the input files' owners    |
| `--ai-generated`        | mark the document as AI generated                                         | off                                                   |
| `-s`, `--stylesheet`    | URI of a stylesheet used instead of the embedded styles (repeatable)      | embedded styles                                       |
| `-j`, `--script-module` | URI of a JavaScript module loaded at the end of the document (repeatable) | none                                                  |
| `-e`, `--variable`      | extra template variable as `name=value` (repeatable)                      | none                                                  |
| `-m`, `--template`      | filename of a Go HTML template used instead of the embedded template      | embedded template                                     |

The embedded template uses the variables `lang` (language of the document,
`en` if not set) and `icon` (URI of the favicon).

A template also gets the author as `.Author`, the most recent modification
time of the input files as `.LastModified`, the page margins of the PDF in
inches as `.PDFMargin` and `--ai-generated` as `.AIGenerated`. The function
`aiGeneratedImage` returns an "AI generated" image as a data URI.

### PDF

The document is converted to PDF if the output filename ends with `.pdf` or
with `--format pdf`. The PDF is printed by a headless Chrome, so Google Chrome,
Chromium or Microsoft Edge has to be installed.

```bash
martok -o output.pdf --toc --pdf-margin 0.8 input1.md input2.md
martok --format pdf input.md
```

A single HTML file given instead of Markdown files is printed to PDF as it is:

```bash
martok page.html
```

While printing, the document is served by a temporary web server on a free port
that is only reachable from the local machine. It also serves the files of the
directories of the input files, so that relative URIs of embedded files like
images work. With input files from several directories a file is taken from the
first directory that contains it.

`--pdf-outline` embeds the headings of the document as the outline (bookmarks)
of the PDF. `--pdf-margin` sets all four page margins. The paper size is A4
unless a stylesheet sets another one, for example `@page { size: letter }`.

The header and the footer of the pages are the templates `page-header` and
`page-footer` of the HTML template. The embedded template has an empty header
and a footer with the author on the left, the page number and the number of
pages in the middle and the date of the last modification of the input files on
the right. With `--ai-generated` an "AI generated" image is printed in front of
the author. A template given with `--template` gets a header or footer by
defining it, without it the PDF has none:

```gotemplate
{{ define "page-footer" -}}
    <div style="width: 100%; font-size: 8pt; text-align: center">
        {{ .Author }} – <span class="pageNumber"></span> – {{ .LastModified.Format "2006-01-02" }}
    </div>
{{- end }}
```

Chrome renders the header and the footer separately from the document: the
stylesheets of the document do not apply to them, so they need their own styles
including a font size, and images have to be data URIs. Elements with the
classes `pageNumber`, `totalPages`, `title`, `date` and `url` are filled in by
Chrome. They are printed in the top and bottom margin over the full width of
the page and are cut off if `--pdf-margin` is too small.

### Heading ids

Headings get GitHub-style ids that are restricted to ASCII characters: the text
is lowercased, spaces become hyphens and everything except `a-z`, `0-9`, `_`
and `-` is removed. German umlauts are transliterated (`## Über uns` gets
`id="ueber-uns"`, `ß` becomes `ss`). The ids are unique across all Markdown
files of a document: repeated headings get a numeric suffix (`ueber-uns-1`).

### Table of contents

`--toc` puts a table of contents with links to all headings at the beginning of
the HTML document. With multiple Markdown files it covers the headings of all
files.

```bash
martok --toc --toc-title "Inhalt" --toc-min-depth 2 --toc-max-depth 3 input1.md input2.md
```

| Flag              | Description                                         | Default    |
|-------------------|-----------------------------------------------------|------------|
| `--toc`           | generate the table of contents and the PDF outline  | off        |
| `--toc-title`     | title of the table of contents (`""` for no title)  | `Contents` |
| `--toc-min-depth` | minimum heading level included (`0` = no limit)     | `0`        |
| `--toc-max-depth` | maximum heading level included (`0` = no limit)     | `0`        |

The table of contents is rendered as `<nav class="toc">` containing the title
(`<h1 id="toc-title">`) and nested lists (`<ul id="toc-list">`).

## License

[MIT](LICENSE)
