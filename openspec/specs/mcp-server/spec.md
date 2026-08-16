## Purpose

Serve ddg's search and read functionality — including the YouTube transcript path — as MCP tools over stdio or StreamableHTTP.

## Requirements

### Requirement: MCP server flags
The `ddg` root command SHALL accept `--mcp` and `--mcp-http <port>` flags that start an MCP server instead of running a subcommand. `--mcp` SHALL serve over stdio; `--mcp-http <port>` SHALL serve StreamableHTTP on the given port.

#### Scenario: stdio server
- **WHEN** `ddg --mcp` is invoked
- **THEN** an MCP server starts over stdin/stdout and serves until the client disconnects

#### Scenario: StreamableHTTP server
- **WHEN** `ddg --mcp-http 8080` is invoked
- **THEN** an MCP server listens on `:8080` over StreamableHTTP

### Requirement: web_search tool
The MCP server SHALL expose a `web_search` tool that searches DuckDuckGo. It SHALL accept a required `query` string and an optional `limit` (default 10). It SHALL return **structured** results with `title`, `url`, and `info` fields.

#### Scenario: Structured search results
- **WHEN** `web_search` is called with a query
- **THEN** it returns a structured list of results, each with an explicit `title`, `url`, and `info`

#### Scenario: Limit applied
- **WHEN** `web_search` is called with `limit=5`
- **THEN** at most 5 results are returned

### Requirement: web_read tool
The MCP server SHALL expose a `web_read` tool that fetches a URL and returns markdown text. For YouTube URLs it SHALL return the transcript markdown (title, channel, duration, date, transcript) using the same logic as `ddg read`.

#### Scenario: Fetch markdown
- **WHEN** `web_read` is called with a non-YouTube URL
- **THEN** it returns the readability-converted markdown as text

#### Scenario: YouTube transcript
- **WHEN** `web_read` is called with a YouTube URL
- **THEN** it returns the transcript markdown (same as `ddg read`), including SponsorBlock filtering

### Requirement: Shared logic, no duplication
The MCP tools SHALL call the same underlying functions as the CLI subcommands (`searchDuckDuckGo` and a shared `readURL`), so behavior SHALL be identical between CLI and MCP.

#### Scenario: Single entry point
- **WHEN** the read logic changes (e.g., new YouTube handling)
- **THEN** both `ddg read` and MCP `web_read` reflect the change without separate updates

### Requirement: Startup logging
When the MCP server starts, `ddg` SHALL write one line to stderr containing the binary name, version, and active transport. The line SHALL be written before the server begins serving. Nothing SHALL be written to stdout for startup logging.

#### Scenario: stdio startup log
- **WHEN** `ddg --mcp` is invoked
- **THEN** stderr receives a single line naming `ddg`, the running version, and the `stdio` transport

#### Scenario: HTTP startup log with port
- **WHEN** `ddg --mcp-http 8080` is invoked
- **THEN** stderr receives a single line naming `ddg`, the running version, the `http` transport, and the resolved listen address `:8080`

#### Scenario: stdout stays clean
- **WHEN** `ddg --mcp` is invoked
- **THEN** stdout carries no startup log output, preserving the MCP JSON-RPC channel
