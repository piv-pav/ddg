---
name: web
description: DuckDuckGo search and web fetch. Use for searching current info and fetching web content as clean markdown.
---

# ddg - DuckDuckGo Search and Web Fetch

Search DuckDuckGo and fetch web content as clean markdown.

## When to Use

- Search for current information, documentation, tutorials, libraries
- Fetch and read web articles, documentation pages, blog posts
- Get clean, readable markdown from any URL

## Commands

### Search DuckDuckGo

```bash
ddg search "query" [-l limit] [--json]
```

Returns: Title, description, and URL for each result (default: 10 results)

**Example:**
```bash
ddg search "golang error handling best practices" -l 5
ddg search "query" --json  # machine-readable output
```

### Fetch URL as Markdown

```bash
ddg read URL [--json]
```

Fetches page, cleans with go-readability, converts to markdown.

**Example:**
```bash
ddg read https://go.dev/doc/effective_go
ddg read https://example.com --json  # returns {url, content}

## Output Format

**Search (plain):** One result per block:
```
Title
Description
URL

```

**Search (--json):** Array of `{title, info, url}` objects.

**Read (plain):** Clean markdown content.

**Read (--json):** `{url, content}` object.

## Notes

- Supports proxy via `HTTP_PROXY`/`HTTPS_PROXY` env vars
- Retry logic with exponential backoff (3 attempts)
- Search uses DuckDuckGo HTML interface (no API key needed)
