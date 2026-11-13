# mdtohtml

Markdown to html coverter

## Installation

```bash
go install github.com/lehrcode/mdtohtml@latest
```

## Usage

```bash
mdtohtml -o output.html -t "Title" -e "author=Christian+Winkler&lang=de" input1.md input2.md
```

### Heading ids

Headings get GitHub-style ids that are restricted to ASCII characters: the text
is lowercased, spaces become hyphens and everything except `a-z`, `0-9`, `_`
and `-` is removed. German umlauts are transliterated (`## Über uns` gets
`id="ueber-uns"`, `ß` becomes `ss`). The ids are unique across all markdown
files of a document: repeated headings get a numeric suffix (`ueber-uns-1`).

### Table of contents

`--toc` puts a table of contents with links to all headings at the beginning of
the HTML document. With multiple markdown files it covers the headings of all
files.

```bash
mdtohtml --toc --toc-title "Inhalt" --toc-min-depth 2 --toc-max-depth 3 input1.md input2.md
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
