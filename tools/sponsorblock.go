package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// sponsorBlockURL is the SponsorBlock skipSegments endpoint.
const sponsorBlockURL = "https://sponsor.ajay.app/api/skipSegments"

// sponsorBlockCategories are all skip categories that add no value to a video.
var sponsorBlockCategories = []string{
	"sponsor",
	"selfpromo",
	"interaction",
	"intro",
	"outro",
	"preview",
	"music_offtopic",
}

// skipSegment is a SponsorBlock skip segment. Segment is [start, end] seconds.
type skipSegment struct {
	Segment  []float64 `json:"segment"`
	Category string    `json:"category"`
}

// fetchSponsorSegments returns skip segments for a video. A 404 (no segments)
// or any network error is returned as an error; the caller treats it as "no
// filtering".
func fetchSponsorSegments(ctx context.Context, videoID string) ([]skipSegment, error) {
	cats, err := json.Marshal(sponsorBlockCategories)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s?videoID=%s&categories=%s",
		sponsorBlockURL, url.QueryEscape(videoID), url.QueryEscape(string(cats)))

	body, err := doGetOK(ctx, u)
	if err != nil {
		return nil, err
	}

	var segments []skipSegment
	if err := json.Unmarshal(body, &segments); err != nil {
		return nil, err
	}

	return segments, nil
}

// renderTranscript joins cues into flowing text, removing cues that overlap
// SponsorBlock skip segments and emitting a [removed: <category>] marker once
// per skipped segment.
func renderTranscript(cues []transcriptCue, segments []skipSegment) string {
	var lines []string
	var cur strings.Builder
	marked := make(map[int]bool, len(segments))

	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			lines = append(lines, s)
			cur.Reset()
		}
	}

	for _, cue := range cues {
		if idx := overlappingSegmentIndex(cue, segments); idx >= 0 {
			flush()
			if !marked[idx] {
				lines = append(lines, "[removed: "+segments[idx].Category+"]")
				marked[idx] = true
			}
			continue
		}
		if cur.Len() > 0 {
			cur.WriteString(" ")
		}
		cur.WriteString(cue.Text)
	}
	flush()

	return strings.Join(lines, "\n")
}

// overlappingSegmentIndex returns the index of the first skip segment that
// overlaps the cue, or -1 if none.
func overlappingSegmentIndex(cue transcriptCue, segments []skipSegment) int {
	cueEnd := cue.Start + cue.Duration
	for i, seg := range segments {
		if len(seg.Segment) < 2 {
			continue
		}
		segStart, segEnd := seg.Segment[0], seg.Segment[1]
		if cue.Start < segEnd && cueEnd > segStart {
			return i
		}
	}
	return -1
}
