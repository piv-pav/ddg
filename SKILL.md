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
ddg search "query" [-l limit]
```

Returns: Title, description, and URL for each result (default: 10 results)

**Example:**
```bash
ddg search "golang error handling best practices" -l 5
```

### Fetch URL as Markdown

```bash
ddg read URL
```

Fetches page, cleans with go-readability, converts to markdown.

**Example:**
```bash
ddg read https://go.dev/doc/effective_go
```

## Output Format

**Search:** Plain text, one result per block:
```
Title
Description
URL

```

**Read:** Clean markdown content extracted from page.

## Notes

- Supports proxy via `HTTP_PROXY`/`HTTPS_PROXY` env vars
- Retry logic with exponential backoff (3 attempts)
- Search uses DuckDuckGo HTML interface (no API key needed)
