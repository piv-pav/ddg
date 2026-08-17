## Purpose

Fetch a URL and return it as clean markdown from the `ddg read` command, routing YouTube and Reddit URLs to their dedicated handlers and converting everything else through readability, via a shared function reused by the MCP `web_read` tool.

## Requirements

### Requirement: Read command and alias
The `ddg read` command SHALL accept exactly one URL argument and output its content as markdown. The command SHALL also be invocable as `ddg fetch`.

#### Scenario: Read a URL
- **WHEN** `ddg read https://example.com/article` is invoked
- **THEN** the page content is fetched and printed as markdown

#### Scenario: Fetch alias
- **WHEN** `ddg fetch https://example.com/article` is invoked
- **THEN** it behaves identically to `ddg read` on the same URL

#### Scenario: Missing or extra argument
- **WHEN** `ddg read` is invoked with zero or more than one argument
- **THEN** the command reports a usage error and does not fetch anything

### Requirement: URL routing
The read logic SHALL inspect the target URL and dispatch it to the appropriate handler: YouTube URLs to the YouTube transcript path, Reddit post URLs to the Reddit path, and all other URLs to the generic readability conversion path. YouTube and Reddit detection and handling are defined by the `youtube-read` and `reddit-read` capabilities respectively; this capability owns only the dispatch decision and the generic path. The same routing SHALL apply whether invoked from the CLI or the MCP `web_read` tool.

#### Scenario: YouTube URL
- **WHEN** `read` is invoked with a YouTube video URL
- **THEN** the URL is routed to the YouTube transcript path as defined by the `youtube-read` capability

#### Scenario: Reddit post URL
- **WHEN** `read` is invoked with a Reddit post URL
- **THEN** the URL is routed to the Reddit path as defined by the `reddit-read` capability

#### Scenario: Generic URL
- **WHEN** `read` is invoked with a URL that is neither YouTube nor a Reddit post
- **THEN** the URL is routed to the generic readability conversion path

### Requirement: URL validation
Before fetching on the generic path, the system SHALL validate that the target URL is parseable and has both a scheme and a host. An invalid URL SHALL produce a clean error and SHALL NOT be fetched.

#### Scenario: Invalid URL
- **WHEN** `read` is invoked with a string that has no scheme or host
- **THEN** a clean `invalid URL` error is returned and no fetch is attempted

### Requirement: Readability conversion
On the generic path, the system SHALL fetch the page over HTTP, clean the HTML with a readability extractor to isolate the main article content, and convert the cleaned HTML to markdown. When readability extraction or HTML rendering fails, the system SHALL fall back to converting the raw fetched HTML to markdown.

#### Scenario: Article page
- **WHEN** a content-rich article page is read
- **THEN** the main content is extracted, converted to markdown, and returned

#### Scenario: Readability failure fallback
- **WHEN** readability extraction or rendering of the cleaned HTML fails
- **THEN** the raw fetched HTML is converted to markdown instead

### Requirement: Low-extraction fallback
When the fetched body is large (greater than 10,000 bytes) and readability extracts very little relative to the source (a cleaned-to-source ratio below 2%), the system SHALL discard the readability result and convert the raw fetched HTML to markdown instead, to avoid returning near-empty output for pages readability mishandles.

#### Scenario: Over-aggressive extraction
- **WHEN** a large page yields a cleaned-HTML-to-source ratio below 2%
- **THEN** the raw HTML is converted to markdown instead of the sparse readability result

#### Scenario: Normal extraction retained
- **WHEN** readability extracts a reasonable share of a large page (ratio at or above 2%)
- **THEN** the readability-based markdown is returned

### Requirement: Output formats
The `ddg read` command SHALL print markdown text by default. With `--json` on the generic path, it SHALL print a JSON object with `url` and `content` fields. YouTube and Reddit URLs under `--json` produce their own structured JSON as defined by their respective capabilities.

#### Scenario: Plain output
- **WHEN** `ddg read <url>` is invoked without `--json` on a generic URL
- **THEN** the markdown content is printed to stdout

#### Scenario: JSON output
- **WHEN** `ddg read <url> --json` is invoked on a generic URL
- **THEN** a JSON object with `url` and `content` fields is printed

### Requirement: Read error handling
When the generic-path fetch returns a non-`200` status or the body cannot be read or converted, the system SHALL return a clean error and the command SHALL exit non-zero. Transient network failures SHALL be retried with backoff before an error is reported.

#### Scenario: Non-success status
- **WHEN** the fetched URL responds with a status other than 200
- **THEN** `ddg read` exits non-zero with an error naming the status

#### Scenario: Retry on transient failure
- **WHEN** an initial fetch fails with a transient network error
- **THEN** the fetch is retried with backoff before any error is returned
