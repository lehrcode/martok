# Martok

Markdown to HTML and PDF converter

## Installation

```bash
go install github.com/lehrcode/martok@latest
```

## Usage

```bash
martok -o output.html -t "Title" -e author="Christian Winkler" -e lang=de input1.md input2.md
```

Multiple Markdown files are converted into one document.

| Flag                    | Description                                                               | Default                                               |
|-------------------------|---------------------------------------------------------------------------|-------------------------------------------------------|
| `-o`, `--output`        | filename of the document                                                  | name of the first input file with `.html` or `.pdf`   |
| `-f`, `--format`        | output format, `html` or `pdf`                                            | `pdf` for an output filename with `.pdf`, else `html` |
| `--pdf-margin`          | page margins of the PDF in inches                                         | `0.4`                                                 |
| `--pdf-outline`         | generate PDF with outline                                                 | off                                                   |
| `-t`, `--title`         | title of the document                                                     | first heading of the Markdown files                   |
| `-s`, `--stylesheet`    | URI of a stylesheet used instead of the embedded styles (repeatable)      | embedded styles                                       |
| `-j`, `--script-module` | URI of a JavaScript module loaded at the end of the document (repeatable) | none                                                  |
| `-e`, `--variable`      | extra template variable as `name=value` (repeatable)                      | none                                                  |
| `-m`, `--template`      | filename of a Go HTML template used instead of the embedded template      | embedded template                                     |

The embedded template uses the variables `lang` (language of the document,
`en` if not set), `author` and `icon` (URI of the favicon).

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

`--toc` also embeds the headings of the document as the outline (bookmarks) of
the PDF. `--margin` sets all four page margins. The paper size is A4
unless a stylesheet sets another one, for example `@page { size: letter }`.

The footer of the pages is the template `page-footer` of the HTML template. The
embedded template prints the page number and the number of pages. A template
given with `--template` gets a footer by defining it, without it the PDF has no
footer:

```gotemplate
{{ define "page-footer" -}}
    <div style="width: 100%; font-size: 8pt; text-align: center"><span class="pageNumber"></span></div>
{{- end }}
```

Chrome renders the footer separately from the document: the stylesheets of the
document do not apply to it, so it needs its own styles including a font size.
Elements with the classes `pageNumber`, `totalPages`, `title`, `date` and `url`
are filled in by Chrome. The footer is printed in the bottom margin and is cut
off if `--margin` is too small.

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
