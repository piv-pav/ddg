## Purpose

Search DuckDuckGo from the `ddg search` command and return the top results — each with title, description, and URL — via a shared function reused by the MCP `web_search` tool, so CLI and MCP search behave identically.

## Requirements

### Requirement: DuckDuckGo search
The `ddg search` command SHALL accept a required query argument and search DuckDuckGo, returning result entries each consisting of a title, an informational snippet, and an absolute result URL. Search SHALL be performed against DuckDuckGo's HTML endpoint with the query URL-encoded.

#### Scenario: Basic query
- **WHEN** `ddg search "golang generics"` is invoked
- **THEN** DuckDuckGo is queried and a list of results is returned, each with a title, an info snippet, and an absolute `http(s)` URL

#### Scenario: Missing query
- **WHEN** `ddg search` is invoked with no query argument
- **THEN** the command reports a usage error and does not perform a search

### Requirement: Result parsing
The system SHALL parse each web result from the DuckDuckGo HTML response, extracting the result title text, the snippet text, and the result link. A result SHALL be included only when both a non-empty title and an absolute `http://` or `https://` URL are present; results lacking either SHALL be skipped.

#### Scenario: Well-formed result
- **WHEN** a web result has a title and an absolute link
- **THEN** it is included with its title, snippet, and URL

#### Scenario: Result without absolute URL
- **WHEN** a web result has no link or a non-absolute link
- **THEN** that result is skipped and does not appear in the output

### Requirement: Result limit
The `ddg search` command SHALL accept a `--limit` (`-l`) flag, defaulting to 10, that caps the number of results returned. When more results are available than the limit, the excess SHALL be truncated; a non-positive limit SHALL NOT truncate.

#### Scenario: Limit smaller than available
- **WHEN** `ddg search "query" --limit 5` returns more than 5 parsed results
- **THEN** at most 5 results are output

#### Scenario: Default limit
- **WHEN** `ddg search "query"` is invoked without `--limit`
- **THEN** at most 10 results are output

### Requirement: Output formats
The `ddg search` command SHALL emit plain markdown by default and JSON when `--json` is passed. In markdown, each result SHALL be rendered as a linked heading with its snippet, and multiple results SHALL be separated by a `---` divider. In JSON, results SHALL be an array of objects with `title`, `info`, and `url` fields.

#### Scenario: Markdown output
- **WHEN** `ddg search "query"` is invoked without `--json`
- **THEN** each result is printed as `# [<title>](<url>)` followed by its snippet, with `---` separating consecutive results

#### Scenario: JSON output
- **WHEN** `ddg search "query" --json` is invoked
- **THEN** a JSON array is printed, each element having `title`, `info`, and `url`

### Requirement: Search error handling
When the DuckDuckGo request returns a status other than `200` or `202`, or the response body cannot be parsed, the system SHALL return a clean error and the command SHALL exit non-zero. Transient network failures SHALL be retried with backoff before an error is reported.

#### Scenario: Non-success status
- **WHEN** DuckDuckGo responds with a status other than 200 or 202
- **THEN** `ddg search` exits non-zero with an error naming the status

#### Scenario: Retry on transient failure
- **WHEN** an initial request fails with a transient network error
- **THEN** the request is retried with backoff before any error is returned
