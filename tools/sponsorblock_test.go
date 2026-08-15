package tools

import (
	"encoding/json"
	"testing"
)

func TestSkipSegmentUnmarshal(t *testing.T) {
	// Realistic SponsorBlock response shape.
	data := `[
		{"category":"sponsor","actionType":"skip","segment":[1026.702,1043.767],"videoDuration":1842.221,"locked":1,"votes":1,"description":""},
		{"category":"intro","actionType":"skip","segment":[0,15.5],"videoDuration":1842.221,"locked":0,"votes":0,"description":""}
	]`

	var segments []skipSegment
	if err := json.Unmarshal([]byte(data), &segments); err != nil {
		t.Fatal(err)
	}
	if len(segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(segments))
	}
	if segments[0].Category != "sponsor" {
		t.Errorf("segments[0].Category = %q", segments[0].Category)
	}
	if len(segments[0].Segment) != 2 || segments[0].Segment[0] != 1026.702 || segments[0].Segment[1] != 1043.767 {
		t.Errorf("segments[0].Segment = %v", segments[0].Segment)
	}
	if segments[1].Category != "intro" {
		t.Errorf("segments[1].Category = %q", segments[1].Category)
	}
}
