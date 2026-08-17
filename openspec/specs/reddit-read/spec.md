## Purpose

Detect Reddit post URLs passed to `read`, solve Reddit's interstitial JavaScript challenge in-process, and return the post (title, author, subreddit, score, body) plus its threaded comments as markdown, without user authentication or stored credentials.

## Requirements

### Requirement: Reddit URL detection
The `read` command SHALL detect when a target URL identifies a Reddit post (comment thread) and route it to the Reddit path instead of generic readability conversion. Detection SHALL recognize the following URL shapes and extract the post ID:

- `reddit.com/r/<sub>/comments/<id>/<slug>/`
- `reddit.com/r/<sub>/comments/<id>/`
- `reddit.com/comments/<id>/`
- `www.reddit.com`, `old.reddit.com`, and `np.reddit.com` host variants of the above
- `redd.it/<id>`

Non-Reddit URLs SHALL continue through the existing readability conversion path unchanged.

#### Scenario: Standard comments URL
- **WHEN** `read` is invoked with `https://www.reddit.com/r/golang/comments/abc123/some_title/`
- **THEN** the URL is detected as Reddit and the post ID `abc123` is extracted

#### Scenario: Short redd.it URL
- **WHEN** `read` is invoked with `https://redd.it/abc123`
- **THEN** the post ID `abc123` is extracted and the Reddit path is used

#### Scenario: Host variants
- **WHEN** `read` is invoked with an `old.reddit.com`, `np.reddit.com`, or bare `reddit.com` URL for a post
- **THEN** the post ID is extracted and the Reddit path is used

#### Scenario: Non-Reddit URL
- **WHEN** `read` is invoked with `https://example.com/article`
- **THEN** the existing readability conversion path is used and no Reddit processing occurs

#### Scenario: Reddit non-post URL
- **WHEN** `read` is invoked with a Reddit URL that is not a post (e.g. a subreddit listing or user profile) and no post ID can be extracted
- **THEN** the URL is not routed to the Reddit path

### Requirement: Challenge solving
Reddit serves an interstitial JavaScript challenge to non-browser clients. The system SHALL solve this challenge in-process, without a JavaScript engine or headless browser, by computing the solution the challenge script defines and resubmitting it with the challenge token, then retaining the session cookies Reddit sets in response.

#### Scenario: Challenge present
- **WHEN** a Reddit request returns the JavaScript challenge page containing a challenge token and input value
- **THEN** the system computes the solution, resubmits it with the token, and obtains the session cookies needed to fetch content

#### Scenario: Challenge solving retries
- **WHEN** Reddit returns a further challenge after a solution is submitted
- **THEN** the system solves and resubmits again up to a bounded number of attempts before giving up

### Requirement: Atomic, stateless, unauthenticated reads
Each Reddit read SHALL be self-contained: it SHALL use a fresh in-memory cookie jar per invocation and discard it when the read returns. The system SHALL NOT store cookies, tokens, or credentials on disk, SHALL NOT require or perform user authentication, and SHALL NOT carry Reddit session state between reads.

#### Scenario: No persisted state
- **WHEN** a Reddit read completes
- **THEN** no Reddit cookies, tokens, or credentials remain stored, and a subsequent read starts from a fresh cookie jar

#### Scenario: No user authentication
- **WHEN** a Reddit read is performed
- **THEN** it succeeds without any user login, OAuth user grant, or account credential

### Requirement: Post and comment retrieval
After solving the challenge, the system SHALL fetch the post and its comment thread from Reddit's public JSON endpoint for the post. It SHALL extract the post title, author, subreddit, score, and body text, and the threaded comments (author, score, body) preserving reply nesting.

#### Scenario: Post with comments
- **WHEN** a Reddit post with a body and comments is read
- **THEN** the title, author, subreddit, score, and body are extracted, along with the comment tree preserving reply depth

#### Scenario: Link post with no body
- **WHEN** a Reddit post has no self-text body (a link post)
- **THEN** the post metadata and comments are still returned and the empty body does not cause an error

### Requirement: Markdown output format
For a detected Reddit post, `read` SHALL emit markdown with the post metadata, body, and threaded comments:

```markdown
# <title>

Subreddit: r/<subreddit>
Author: u/<author>
Score: <score>

<body>

## Comments

- [<score>] u/<author>: <comment body>
  - [<score>] u/<author>: <nested reply body>
```

Nested replies SHALL be indented to reflect their depth in the thread.

#### Scenario: Plain output
- **WHEN** `read` is invoked on a Reddit post URL without `--json`
- **THEN** the markdown document above is printed with comments under a `Comments` heading and replies indented by depth

#### Scenario: JSON output
- **WHEN** `read` is invoked on a Reddit post URL with `--json`
- **THEN** a JSON object containing `url`, `title`, `subreddit`, `author`, `score`, `body`, and `comments` (with nested replies) is printed

### Requirement: Clean error on retrieval failure
When challenge solving fails, the post cannot be found, or content retrieval otherwise fails, the system SHALL return a clean error to stderr and exit non-zero, instead of falling back to readability conversion of the challenge page. No challenge-page chrome SHALL appear in the output for a Reddit URL.

#### Scenario: Challenge cannot be solved
- **WHEN** the challenge cannot be solved within the allowed attempts
- **THEN** `read` exits non-zero with a clean error message and no challenge-page output

#### Scenario: Post not found
- **WHEN** the extracted post ID does not resolve to a post
- **THEN** `read` exits non-zero with a clean error message
