# Martok

Markdown to HTML converter

## Installation

```bash
go install github.com/lehrcode/martok@latest
```

## Usage

```bash
martok -o output.html -t "Title" -e author="Christian Winkler" -e lang=de input1.md input2.md
```

Multiple Markdown files are converted into one HTML document.

| Flag                    | Description                                                               | Default                                      |
|-------------------------|---------------------------------------------------------------------------|----------------------------------------------|
| `-o`, `--output`        | filename of the HTML document                                             | name of the first Markdown file with `.html` |
| `-t`, `--title`         | title of the HTML document                                                | first heading of the Markdown files          |
| `-s`, `--stylesheet`    | URI of a stylesheet used instead of the embedded styles (repeatable)      | embedded styles                              |
| `-j`, `--script-module` | URI of a JavaScript module loaded at the end of the document (repeatable) | none                                         |
| `-e`, `--variable`      | extra template variable as `name=value` (repeatable)                      | none                                         |
| `-m`, `--template`      | filename of a Go HTML template used instead of the embedded template      | embedded template                            |

The embedded template uses the variables `lang` (language of the document,
`en` if not set), `author` and `icon` (URI of the favicon).

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
| `--toc`           | generate the table of contents                      | off        |
| `--toc-title`     | title of the table of contents (`""` for no title)  | `Contents` |
| `--toc-min-depth` | minimum heading level included (`0` = no limit)     | `0`        |
| `--toc-max-depth` | maximum heading level included (`0` = no limit)     | `0`        |

The table of contents is rendered as `<nav class="toc">` containing the title
(`<h1 id="toc-title">`) and nested lists (`<ul id="toc-list">`).

## License

[MIT](LICENSE)
