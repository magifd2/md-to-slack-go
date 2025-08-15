# md-to-slack-go

A CLI tool to convert Markdown to Slack Block Kit JSON.
This is a Go port of [github.com/magifd2/md-to-slack](https://github.com/magifd2/md-to-slack).

## Features

Converts the following Markdown elements to Slack Block Kit blocks:

- Headings (`# H1`, `## H2`) are converted to `header` blocks.
- Other headings (`### H3`, etc.) are converted to `section` blocks with bold text.
- Paragraphs are converted to `section` blocks.
- Blockquotes are converted to `section` blocks with quote formatting.
- Fenced code blocks are converted to `section` blocks with formatted code.
- Thematic breaks (horizontal rules) are converted to `divider` blocks.
- GFM Tables are converted to `table` blocks.
- Inline formatting (bold, italic, strikethrough, code, links) is preserved as `mrkdwn`.

## Installation

### From source

```bash
go install github.com/magifd2/md-to-slack-go/cmd/md-to-slack@latest
```

### From GitHub Releases

Alternatively, you can download a pre-compiled binary for your OS from the [GitHub Releases](https://github.com/magifd2/md-to-slack-go/releases) page.

### With `make`

If you have the repository cloned, you can build and install it using make:
```bash
make build
make install # Default prefix is /usr/local
```

## Usage

### From a file

```bash
md-to-slack <path/to/your/file.md>
```

### From standard input

```bash
cat file.md | md-to-slack
```

### Options

```
-version  Print version and exit
```

## Development

This project uses `make` for common development tasks.

- `make build`: Build the binary for your current OS and architecture.
- `make test`: Run tests.
- `make lint`: Run the linter.
- `make cross-compile`: Build for all target platforms (macOS, Linux, Windows).
- `make clean`: Clean up build artifacts.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.