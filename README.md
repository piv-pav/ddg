# ddg-cli

DuckDuckGo search and web fetch CLI tool.

## Features

- Search DuckDuckGo with customizable result limits
- Fetch and convert web pages to clean markdown (uses go-readability for content extraction)

## Installation

```bash
go install codeberg.org/pivpav/ddg-cli@latest
```

Or build from source:

```bash
git clone https://codeberg.org/pivpav/ddg-cli.git
cd ddg-cli
just build
just install
```

## Usage

### Search DuckDuckGo

```bash
ddg search "golang cobra" --limit 5
# or short
ddg search "golang cobra" -l 5
```

### Fetch and convert URL to markdown

```bash
ddg read https://example.com/article
```

## Build

```bash
# Build binary
just build

# Install to ~/.local/bin
just install

# Clean build artifacts
just clean
```

## Dependencies

- github.com/spf13/cobra - CLI framework
- github.com/PuerkitoBio/goquery - HTML parsing for search results
- codeberg.org/readeck/go-readability/v2 - Article extraction and cleaning
- github.com/JohannesKaufmann/html-to-markdown/v2 - HTML to Markdown conversion
