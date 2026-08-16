# Changelog

## [0.5.2] - 2026-08-16

### Added
- MCP server startup logging: `--mcp` and `--mcp-http` print the binary name, version, and transport (plus the listen address for HTTP) to stderr

## [0.5.1] - 2026-08-16

### Added
- `ddg upgrade` command: checks the latest GitHub tag and self-upgrades via `go install`
- Consistent `v`-prefixed version string (`v<version>-dev` for local builds)

## [0.5.0] - 2026-08-16

### Added
- YouTube transcript support: `ddg read <youtube-url>` returns channel, title, duration, date, and a native-language transcript
- SponsorBlock filtering: sponsor/selfpromo/interaction/intro/outro/preview/music_offtopic segments removed and marked inline as `[removed: <category>]`
- Native-language detection via the auto-generated (asr) caption track (correct even for multi-audio dubbed videos)
- URL shapes: `watch?v=`, `youtu.be`, `/shorts/`, `/live/`, `/embed/`, `m.youtube.com`, `music.youtube.com`

## [0.4.0] - 2026-08-16

### Added
- MCP server support: `--mcp` (stdio) and `--mcp-http <port>` (StreamableHTTP)
- `web_search` tool (structured results with explicit title/url/info)
- `web_read` tool (markdown conversion, reuses the read path)

## [0.3.2] - 2026-06-02

### Changed
- `just install` now installs to `~/go/bin` instead of `~/.local/bin`

## [0.3.1] - 2026-06-02

### Changed
- Module renamed from `ddg-cli` to `ddg` for simpler `go install`
- Version detected from Go build info when installed via `go install`
- Repository renamed on Codeberg from `ddg-cli` to `ddg`

## [0.3.0] - 2026-06-02

### Changed
- Search results now output as proper Markdown with clickable title links
- Titles formatted as `# [Title](URL)` for better readability
- Horizontal rules (`---`) separate search results

## [0.2.2] - 2026-05-28

### Added
- `--json` flag for search and read commands (machine-readable output)
- Shared HTTP client with retry logic (`tools/http.go`)
- URL validation helper

### Changed
- Refactored HTTP client setup into shared module
- Both commands now use `cmd.Context()` for proper signal handling

## [0.2.0] - 2026-05-01

### Changed
- Switched from go-task to just (replaced Taskfile.yml with justfile)

## [0.1.1] - 2026-04-30

### Fixed
- Improved fallback for JavaScript-heavy sites in `read` command
- Now uses extraction ratio check (HTML size >10KB + <2% content extracted) instead of absolute length
- Prevents false positives on small valid responses (e.g., ifconfig.me)

## [0.1.0] - 2026-04-30

### Added
- Initial release
- DuckDuckGo search command with limit support
- URL fetch command with go-readability cleaning and markdown conversion
- Retry logic with exponential backoff
- Proxy support via environment variables
