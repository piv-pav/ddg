# ddg

DuckDuckGo search and web fetch CLI tool.

## Features

- Search DuckDuckGo with customizable result limits
- Fetch and convert web pages to clean markdown (uses go-readability for content extraction)
- Fetch YouTube video transcripts with SponsorBlock filtering (channel, title, duration, date, clean transcript)
- Fetch Reddit posts and comment threads (title, subreddit, author, score, body, threaded comments) — no auth, no stored cookies
- Run as an MCP server (`--mcp` stdio / `--mcp-http <port>` StreamableHTTP) exposing `web_search` and `web_read` tools

## Installation

```bash
go install github.com/piv-pav/ddg@latest
```

Or build from source:

```bash
git clone https://github.com/piv-pav/ddg.git
cd ddg
just build
just install
```

## Usage

### Search DuckDuckGo

```bash
ddg search "golang cobra" --limit 5
# or short
ddg search "golang cobra" -l 5

# JSON output
ddg search "golang cobra" --json
```

### Fetch and convert URL to markdown

```bash
ddg read https://example.com/article

# JSON output
ddg read https://example.com/article --json
```

### Fetch a YouTube transcript

```bash
ddg read https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

Returns the channel, title, duration, publish date, and a clean native-language
transcript. Sponsor/intro/outro/self-promo segments are removed via SponsorBlock
and marked inline as `[removed: <category>]`.

### Fetch a Reddit post

```bash
ddg read https://www.reddit.com/r/golang/comments/<id>/<slug>/
```

Returns the post (title, subreddit, author, score, body) and its threaded
comments as markdown; add `--json` for a structured object. Reddit serves an
interstitial JavaScript challenge to non-browser clients; ddg solves it
in-process. Each read is atomic — no authentication, no stored cookies, no
session state between reads. Recognized URL shapes: `.../comments/<id>/<slug>/`,
`/comments/<id>/`, `redd.it/<id>`, and `www`/`old`/`np`/`new` host variants.

### Run as an MCP server

```bash
# stdio transport
ddg --mcp

# StreamableHTTP transport on a port
ddg --mcp-http 8080
```

Exposes two tools:
- `web_search(query, limit)` — structured DuckDuckGo results
- `web_read(url)` — fetch and convert a URL to markdown

On startup, the server logs its name, version, and transport (plus the listen
address for HTTP) to stderr.

### Upgrade

```bash
ddg upgrade
```

Checks the latest GitHub release and self-upgrades via `go install` when a newer version is available.

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
- github.com/modelcontextprotocol/go-sdk - MCP server and transports
