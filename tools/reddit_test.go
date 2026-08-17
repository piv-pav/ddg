package tools

import (
	"strings"
	"testing"
)

func TestParseRedditURL(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		wantID string
		wantOK bool
	}{
		{"comments with slug", "https://www.reddit.com/r/golang/comments/abc123/some_title/", "abc123", true},
		{"comments no slug", "https://www.reddit.com/r/golang/comments/abc123/", "abc123", true},
		{"bare comments", "https://www.reddit.com/comments/abc123/", "abc123", true},
		{"old host", "https://old.reddit.com/r/golang/comments/abc123/t/", "abc123", true},
		{"np host", "https://np.reddit.com/r/golang/comments/abc123/", "abc123", true},
		{"bare reddit.com", "https://reddit.com/r/golang/comments/abc123/t/", "abc123", true},
		{"redd.it short", "https://redd.it/abc123", "abc123", true},
		{"subreddit listing", "https://www.reddit.com/r/golang/", "", false},
		{"user profile", "https://www.reddit.com/user/someone/", "", false},
		{"reddit home", "https://www.reddit.com/", "", false},
		{"non-reddit", "https://example.com/article", "", false},
		{"redd.it nested", "https://redd.it/abc/def", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := parseRedditURL(tt.url)
			if ok != tt.wantOK || id != tt.wantID {
				t.Errorf("parseRedditURL(%q) = (%q, %v), want (%q, %v)", tt.url, id, ok, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestSolveRedditChallenge(t *testing.T) {
	// Compact form as served by Reddit (no spaces around => and +).
	page := `<html><body>
	<script>document.addEventListener("DOMContentLoaded",async function(){var n=await(async e=>e+e)("fc8de649b7b72b6a");e.elements.namedItem("solution").value=n},{once:!0});</script>
	<form hidden method="GET" action="/comments/abc123/">
	<input type="hidden" name="solution" />
	<input type="hidden" name="js_challenge" value="1"/>
	<input type="hidden" name="token" value="deadbeefcafe"/>
	</form>
	</body></html>`

	sol, tok, ok := solveRedditChallenge(page)
	if !ok {
		t.Fatal("expected challenge to solve")
	}
	if want := "fc8de649b7b72b6afc8de649b7b72b6a"; sol != want {
		t.Errorf("solution = %q, want %q", sol, want)
	}
	if tok != "deadbeefcafe" {
		t.Errorf("token = %q, want deadbeefcafe", tok)
	}
}

func TestSolveRedditChallengeSpacedForm(t *testing.T) {
	// Tolerant of the spaced arrow-fn form too.
	page := `<script>await(async e => e + e)("aabb")</script>
	<input name="token" value="tok1"/>`
	sol, tok, ok := solveRedditChallenge(page)
	if !ok || sol != "aabbaabb" || tok != "tok1" {
		t.Errorf("got (%q, %q, %v), want (aabbaabb, tok1, true)", sol, tok, ok)
	}
}

func TestSolveRedditChallengeNoChallenge(t *testing.T) {
	if _, _, ok := solveRedditChallenge("<html>no challenge here</html>"); ok {
		t.Error("expected no solution for non-challenge page")
	}
}

func TestFormatRedditMarkdown(t *testing.T) {
	info := &redditInfo{
		Title:     "Test Post",
		Author:    "alice",
		Subreddit: "golang",
		Score:     42,
		Body:      "This is the body.",
		Comments: []redditComment{
			{
				Author: "bob", Score: 10, Body: "Top comment.",
				Replies: []redditComment{
					{Author: "carol", Score: 3, Body: "A reply."},
				},
			},
			{Author: "dave", Score: 1, Body: "Another top comment."},
		},
	}

	out := formatRedditMarkdown(info)

	for _, want := range []string{
		"# Test Post",
		"Subreddit: r/golang",
		"Author: u/alice",
		"Score: 42",
		"This is the body.",
		"## Comments",
		"- [10] u/bob: Top comment.",
		"  - [3] u/carol: A reply.",
		"- [1] u/dave: Another top comment.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestFormatRedditMarkdownLinkPost(t *testing.T) {
	// Empty body (link post) must not emit a stray blank body block.
	info := &redditInfo{Title: "Link", Author: "eve", Subreddit: "pics", Score: 5}
	out := formatRedditMarkdown(info)
	if strings.Contains(out, "Score: 5\n\n\n") {
		t.Errorf("stray blank body block in link post output:\n%s", out)
	}
	if !strings.Contains(out, "Score: 5") {
		t.Errorf("missing score line:\n%s", out)
	}
}
