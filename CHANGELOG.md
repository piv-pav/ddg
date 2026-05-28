# Changelog

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
