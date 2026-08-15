package tools

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// youtubeConsentCookies bypass YouTube's EU consent wall so the watch page
// returns the full player response instead of a consent redirect.
var youtubeConsentCookies = []*http.Cookie{
	{Name: "CONSENT", Value: "YES+cb.20210328-17-p0.en+FX+700"},
	{Name: "SOCS", Value: "CAI"},
}

// innertubePlayerURL is the ANDROID innertube endpoint whose captionTracks
// carry working baseUrls (the WEB client's baseUrls return empty responses).
const innertubePlayerURL = "https://www.youtube.com/youtubei/v1/player?key=%s"

// youtubeInfo carries everything needed to render a YouTube read result.
type youtubeInfo struct {
	Title      string
	Channel    string
	Duration   string
	Date       string
	Transcript string
}

// youtubeJSON is the JSON output shape for a YouTube read.
type youtubeJSON struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Channel    string `json:"channel"`
	Duration   string `json:"duration"`
	Date       string `json:"date"`
	Transcript string `json:"transcript"`
}

// captionTrack is one entry of the innertube captionTracks list.
type captionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind"`
}

// transcriptCue is a single subtitle line with timing, used for filtering.
type transcriptCue struct {
	Text     string
	Start    float64
	Duration float64
}

// parseYouTubeID extracts the video ID from a YouTube URL. It returns false
// for non-video URLs (channels, playlists, or non-YouTube hosts).
func parseYouTubeID(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(u.Hostname())

	if host == "youtu.be" {
		id := strings.Trim(u.Path, "/")
		return id, id != ""
	}

	if host != "youtube.com" && !strings.HasSuffix(host, ".youtube.com") {
		return "", false
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 {
		switch parts[0] {
		case "shorts", "live", "embed":
			if parts[1] != "" {
				return parts[1], true
			}
		}
	}

	if id := u.Query().Get("v"); id != "" {
		return id, true
	}

	return "", false
}

// fetchYouTube retrieves metadata, transcript, and SponsorBlock-filtered text
// for a video ID.
func fetchYouTube(ctx context.Context, videoID string) (*youtubeInfo, error) {
	watchURL := "https://www.youtube.com/watch?v=" + videoID

	body, err := doGetOKWithCookies(ctx, watchURL, youtubeConsentCookies)
	if err != nil {
		return nil, err
	}
	watchHTML := string(body)

	apiKey := extractInnertubeAPIKey(watchHTML)
	if apiKey == "" {
		return nil, fmt.Errorf("innertube API key not found")
	}

	watchPR, err := extractWatchPlayerResponse(watchHTML)
	if err != nil {
		return nil, err
	}

	player, err := fetchInnertubePlayer(ctx, apiKey, videoID)
	if err != nil {
		return nil, err
	}

	track := selectNativeTrack(player.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks)
	if track == nil {
		return nil, fmt.Errorf("no caption tracks")
	}

	cues, err := fetchTranscriptCues(ctx, track.BaseURL)
	if err != nil {
		return nil, err
	}

	segments, _ := fetchSponsorSegments(ctx, videoID) // nil on failure → no filtering

	return &youtubeInfo{
		Title:      player.VideoDetails.Title,
		Channel:    player.VideoDetails.Author,
		Duration:   formatDuration(player.VideoDetails.LengthSeconds),
		Date:       formatDate(watchPR.Microformat.PlayerMicroformatRenderer.PublishDate),
		Transcript: renderTranscript(cues, segments),
	}, nil
}

// innertubePlayerResponse captures the fields we need from the innertube player.
type innertubePlayerResponse struct {
	VideoDetails struct {
		Title         string `json:"title"`
		Author        string `json:"author"`
		LengthSeconds string `json:"lengthSeconds"`
	} `json:"videoDetails"`
	Captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []captionTrack `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
}

// fetchInnertubePlayer calls the ANDROID innertube player endpoint.
func fetchInnertubePlayer(ctx context.Context, apiKey, videoID string) (*innertubePlayerResponse, error) {
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "ANDROID",
				"clientVersion": "20.10.38",
			},
		},
		"videoId": videoID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := doPostJSON(ctx, fmt.Sprintf(innertubePlayerURL, apiKey), body)
	if err != nil {
		return nil, err
	}

	var pr innertubePlayerResponse
	if err := json.Unmarshal(resp, &pr); err != nil {
		return nil, fmt.Errorf("parse innertube player response: %w", err)
	}

	return &pr, nil
}

var innertubeAPIKeyRe = regexp.MustCompile(`"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`)

func extractInnertubeAPIKey(htmlContent string) string {
	if m := innertubeAPIKeyRe.FindStringSubmatch(htmlContent); len(m) == 2 {
		return m[1]
	}
	return ""
}

// watchPlayerResponse carries the publish date from ytInitialPlayerResponse.
type watchPlayerResponse struct {
	Microformat struct {
		PlayerMicroformatRenderer struct {
			PublishDate string `json:"publishDate"`
		} `json:"playerMicroformatRenderer"`
	} `json:"microformat"`
}

// extractWatchPlayerResponse locates the ytInitialPlayerResponse object in the
// watch page HTML and unmarshals it, using brace matching (the object is not
// delimited by a simple regex).
func extractWatchPlayerResponse(htmlContent string) (*watchPlayerResponse, error) {
	const marker = "var ytInitialPlayerResponse = "

	idx := strings.Index(htmlContent, marker)
	if idx < 0 {
		return nil, fmt.Errorf("ytInitialPlayerResponse not found")
	}

	start := idx + len(marker)
	if start >= len(htmlContent) || htmlContent[start] != '{' {
		return nil, fmt.Errorf("malformed ytInitialPlayerResponse")
	}

	depth := 0
	end := -1
	inString := false
	escaped := false
	for i := start; i < len(htmlContent); i++ {
		c := htmlContent[i]
		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
				goto done
			}
		}
	}
done:
	if end < 0 {
		return nil, fmt.Errorf("unbalanced ytInitialPlayerResponse")
	}

	var pr watchPlayerResponse
	if err := json.Unmarshal([]byte(htmlContent[start:end+1]), &pr); err != nil {
		return nil, fmt.Errorf("parse ytInitialPlayerResponse: %w", err)
	}

	return &pr, nil
}

// selectNativeTrack picks the original-language caption track. The
// auto-generated (asr) track is always in the original audio language (dubs
// are manually captioned), so that language is used as the source of truth.
// A manual track in that language is preferred for quality; fall back to the
// asr track, then to the first track.
func selectNativeTrack(tracks []captionTrack) *captionTrack {
	if len(tracks) == 0 {
		return nil
	}

	lang := ""
	for i := range tracks {
		if tracks[i].Kind == "asr" {
			lang = tracks[i].LanguageCode
			break
		}
	}
	if lang == "" {
		lang = tracks[0].LanguageCode
	}

	for i := range tracks {
		if tracks[i].LanguageCode == lang && tracks[i].Kind != "asr" {
			return &tracks[i]
		}
	}

	for i := range tracks {
		if tracks[i].LanguageCode == lang {
			return &tracks[i]
		}
	}

	return &tracks[0]
}

// fetchTranscriptCues fetches the caption track URL and parses the XML
// transcript into cues.
func fetchTranscriptCues(ctx context.Context, baseURL string) ([]transcriptCue, error) {
	trackURL := strings.Replace(baseURL, "&fmt=srv3", "", -1)

	body, err := doGetOK(ctx, trackURL)
	if err != nil {
		return nil, err
	}

	return parseTranscriptXML(string(body))
}

// parseTranscriptXML parses the timedtext XML into cues with timing.
func parseTranscriptXML(xmlData string) ([]transcriptCue, error) {
	type xmlTranscript struct {
		Texts []struct {
			Text     string  `xml:",chardata"`
			Start    float64 `xml:"start,attr"`
			Duration float64 `xml:"dur,attr"`
		} `xml:"text"`
	}

	var t xmlTranscript
	if err := xml.Unmarshal([]byte(xmlData), &t); err != nil {
		return nil, fmt.Errorf("parse transcript XML: %w", err)
	}

	cues := make([]transcriptCue, 0, len(t.Texts))
	for _, e := range t.Texts {
		cues = append(cues, transcriptCue{
			Text:     html.UnescapeString(strings.TrimSpace(e.Text)),
			Start:    e.Start,
			Duration: e.Duration,
		})
	}

	return cues, nil
}

// formatDuration renders seconds as H:MM:SS or M:SS.
func formatDuration(seconds string) string {
	s, err := strconv.Atoi(seconds)
	if err != nil {
		return ""
	}
	h := s / 3600
	m := (s % 3600) / 60
	sec := s % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%d:%02d", m, sec)
}

// formatDate returns the YYYY-MM-DD portion of an ISO timestamp.
func formatDate(iso string) string {
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

// formatYouTubeMarkdown renders the plain markdown output.
func formatYouTubeMarkdown(info *youtubeInfo) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", info.Title)
	fmt.Fprintf(&sb, "Channel: %s\n", info.Channel)
	fmt.Fprintf(&sb, "Duration: %s\n", info.Duration)
	fmt.Fprintf(&sb, "Published: %s\n\n", info.Date)
	sb.WriteString("## Transcript\n\n")
	sb.WriteString(info.Transcript)
	return sb.String()
}
