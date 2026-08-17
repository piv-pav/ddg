package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

// redditMaxChallengeAttempts bounds the challenge solve/resubmit loop.
const redditMaxChallengeAttempts = 5

// redditCommentLimit caps how many comments the .json listing returns.
const redditCommentLimit = 100

// redditInfo carries everything needed to render a Reddit read result.
type redditInfo struct {
	Title     string
	Author    string
	Subreddit string
	Score     int
	Body      string
	Comments  []redditComment
}

// redditComment is one comment plus its nested replies.
type redditComment struct {
	Author  string          `json:"author"`
	Score   int             `json:"score"`
	Body    string          `json:"body"`
	Replies []redditComment `json:"replies,omitempty"`
}

// redditJSON is the JSON output shape for a Reddit read.
type redditJSON struct {
	URL       string          `json:"url"`
	Title     string          `json:"title"`
	Subreddit string          `json:"subreddit"`
	Author    string          `json:"author"`
	Score     int             `json:"score"`
	Body      string          `json:"body"`
	Comments  []redditComment `json:"comments"`
}

var redditPostPathRe = regexp.MustCompile(`(?:^|/)comments/([a-z0-9]+)`)

// parseRedditURL extracts the post ID from a Reddit post (comment thread) URL.
// It returns false for non-post Reddit URLs (subreddit listings, user
// profiles) and for non-Reddit hosts.
func parseRedditURL(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "old.")
	host = strings.TrimPrefix(host, "np.")
	host = strings.TrimPrefix(host, "new.")

	switch {
	case host == "redd.it":
		id := strings.Trim(u.Path, "/")
		// Reject nested paths; a share ID is a single segment.
		if id == "" || strings.Contains(id, "/") {
			return "", false
		}
		return id, true
	case host == "reddit.com":
		if m := redditPostPathRe.FindStringSubmatch(u.Path); len(m) == 2 {
			return m[1], true
		}
		return "", false
	default:
		return "", false
	}
}

var redditChallengeRe = regexp.MustCompile(`await\(async e\s*=>\s*e\s*\+\s*e\)\("([^"]+)"\)`)
var redditTokenRe = regexp.MustCompile(`name="token" value="([^"]+)"`)

// solveRedditChallenge extracts the challenge input and token from a challenge
// page and computes the solution. Reddit's challenge is the input string
// concatenated with itself.
func solveRedditChallenge(htmlContent string) (solution, token string, ok bool) {
	m := redditChallengeRe.FindStringSubmatch(htmlContent)
	if len(m) != 2 {
		return "", "", false
	}
	input := m[1]

	t := redditTokenRe.FindStringSubmatch(htmlContent)
	if len(t) != 2 {
		return "", "", false
	}

	return input + input, t[1], true
}

// isRedditChallenge reports whether the body is the JS challenge interstitial.
// It matches the hidden challenge form input specifically, not any occurrence
// of the string "js_challenge" (which also appears in scripts on real pages).
func isRedditChallenge(body string) bool {
	return strings.Contains(body, `name="js_challenge" value="1"`)
}

// newRedditClient builds a proxy-aware HTTP client with a fresh in-memory
// cookie jar. The jar is discarded when the client goes out of scope, so no
// Reddit session state persists between reads.
func newRedditClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	return &http.Client{
		Timeout: timeout,
		Jar:     jar,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}, nil
}

// redditGet performs a GET with browser-like headers using the given client
// (whose cookie jar collects the challenge session cookies).
func redditGet(ctx context.Context, client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", acceptHTML)
	req.Header.Set("Accept-Language", acceptLang)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	return string(body), nil
}

// clearRedditChallenge solves the interstitial challenge for a post, priming
// the client's cookie jar with the session cookies needed to fetch content.
func clearRedditChallenge(ctx context.Context, client *http.Client, postID string) error {
	postURL := "https://www.reddit.com/comments/" + postID + "/"

	body, err := redditGet(ctx, client, postURL)
	if err != nil {
		return err
	}

	for attempt := 0; attempt < redditMaxChallengeAttempts; attempt++ {
		if !isRedditChallenge(body) {
			return nil
		}
		solution, token, ok := solveRedditChallenge(body)
		if !ok {
			return fmt.Errorf("could not solve Reddit challenge")
		}
		solvedURL := fmt.Sprintf(
			"%s?solution=%s&js_challenge=1&token=%s&jsc_orig_r=",
			postURL, url.QueryEscape(solution), url.QueryEscape(token),
		)
		body, err = redditGet(ctx, client, solvedURL)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("could not clear Reddit challenge after %d attempts", redditMaxChallengeAttempts)
}

// redditListing mirrors the two-element array Reddit returns for a post's
// .json endpoint: [post listing, comments listing].
type redditListing struct {
	Data struct {
		Children []redditChild `json:"children"`
	} `json:"data"`
}

// redditChild is one entry in a listing's children array.
type redditChild struct {
	Kind string         `json:"kind"`
	Data redditNodeData `json:"data"`
}

// redditNodeData holds the fields we read from posts and comments. Replies are
// decoded lazily because Reddit uses either an object (a nested listing) or the
// empty string "" when there are none.
type redditNodeData struct {
	Title     string          `json:"title"`
	Author    string          `json:"author"`
	Subreddit string          `json:"subreddit"`
	Score     int             `json:"score"`
	Selftext  string          `json:"selftext"`
	Body      string          `json:"body"`
	Replies   json.RawMessage `json:"replies"`
}

// fetchReddit solves the challenge and retrieves the post plus threaded
// comments from Reddit's public JSON endpoint. Each call is atomic: it uses a
// fresh cookie jar that is discarded on return, requires no authentication,
// and stores nothing.
func fetchReddit(ctx context.Context, postID string) (*redditInfo, error) {
	client, err := newRedditClient()
	if err != nil {
		return nil, err
	}

	if err := clearRedditChallenge(ctx, client, postID); err != nil {
		return nil, err
	}

	jsonURL := fmt.Sprintf(
		"https://www.reddit.com/comments/%s/.json?raw_json=1&limit=%d",
		postID, redditCommentLimit,
	)
	body, err := redditGet(ctx, client, jsonURL)
	if err != nil {
		return nil, err
	}
	if isRedditChallenge(body) {
		return nil, fmt.Errorf("Reddit challenge not cleared for post %s", postID)
	}

	var listings []redditListing
	if err := json.Unmarshal([]byte(body), &listings); err != nil {
		return nil, fmt.Errorf("post %s not found or not a post", postID)
	}
	if len(listings) < 2 || len(listings[0].Data.Children) == 0 {
		return nil, fmt.Errorf("post %s not found", postID)
	}

	post := listings[0].Data.Children[0].Data
	info := &redditInfo{
		Title:     post.Title,
		Author:    post.Author,
		Subreddit: post.Subreddit,
		Score:     post.Score,
		Body:      post.Selftext,
		Comments:  parseRedditComments(listings[1].Data.Children),
	}
	return info, nil
}

// parseRedditComments recursively walks comment children, skipping non-t1
// kinds (e.g. "more" nodes).
func parseRedditComments(children []redditChild) []redditComment {
	var out []redditComment
	for _, c := range children {
		if c.Kind != "t1" {
			continue
		}
		comment := redditComment{
			Author: c.Data.Author,
			Score:  c.Data.Score,
			Body:   c.Data.Body,
		}
		if len(c.Data.Replies) > 0 && c.Data.Replies[0] == '{' {
			var repliesListing redditListing
			if err := json.Unmarshal(c.Data.Replies, &repliesListing); err == nil {
				comment.Replies = parseRedditComments(repliesListing.Data.Children)
			}
		}
		out = append(out, comment)
	}
	return out
}

// formatRedditMarkdown renders the plain markdown output.
func formatRedditMarkdown(info *redditInfo) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", info.Title)
	fmt.Fprintf(&sb, "Subreddit: r/%s\n", info.Subreddit)
	fmt.Fprintf(&sb, "Author: u/%s\n", info.Author)
	fmt.Fprintf(&sb, "Score: %d\n", info.Score)
	if body := strings.TrimSpace(info.Body); body != "" {
		fmt.Fprintf(&sb, "\n%s\n", body)
	}
	if len(info.Comments) > 0 {
		sb.WriteString("\n## Comments\n\n")
		writeRedditComments(&sb, info.Comments, 0)
	}
	return sb.String()
}

// writeRedditComments renders comments as an indented bullet list reflecting
// reply depth.
func writeRedditComments(sb *strings.Builder, comments []redditComment, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, c := range comments {
		body := strings.ReplaceAll(strings.TrimSpace(c.Body), "\n", " ")
		fmt.Fprintf(sb, "%s- [%d] u/%s: %s\n", indent, c.Score, c.Author, body)
		if len(c.Replies) > 0 {
			writeRedditComments(sb, c.Replies, depth+1)
		}
	}
}
