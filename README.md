# ddg

DuckDuckGo search and web fetch CLI tool.

## Features

- Search DuckDuckGo with customizable result limits
- Fetch and convert web pages to clean markdown (uses go-readability for content extraction)

## Installation

```bash
task build
task install
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
task build

# Install to ~/.local/bin
task install

# Clean build artifacts
task clean
```

## Dependencies

- github.com/spf13/cobra - CLI framework
- github.com/PuerkitoBio/goquery - HTML parsing for search results
- github.com/go-shiori/go-readability - Article extraction and cleaning
- github.com/JohannesKaufmann/html-to-markdown/v2 - HTML to Markdown conversion
