package tools

import "testing"

func TestParseYouTubeID(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want string
		ok   bool
	}{
		{"watch", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"watch extra params", "https://www.youtube.com/watch?v=abc123&list=PLxyz&t=10", "abc123", true},
		{"youtu.be", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"shorts", "https://www.youtube.com/shorts/abc123", "abc123", true},
		{"live", "https://www.youtube.com/live/abc123", "abc123", true},
		{"embed", "https://www.youtube.com/embed/abc123", "abc123", true},
		{"mobile", "https://m.youtube.com/watch?v=abc123", "abc123", true},
		{"music", "https://music.youtube.com/watch?v=abc123", "abc123", true},
		{"bare host", "https://youtube.com/watch?v=abc123", "abc123", true},
		{"non-youtube", "https://example.com/article", "", false},
		{"channel", "https://www.youtube.com/@somechannel", "", false},
		{"playlist", "https://www.youtube.com/playlist?list=PLxyz", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseYouTubeID(tc.url)
			if ok != tc.ok || got != tc.want {
				t.Errorf("parseYouTubeID(%q) = (%q, %v), want (%q, %v)", tc.url, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"213", "3:33"},
		{"60", "1:00"},
		{"59", "0:59"},
		{"3600", "1:00:00"},
		{"3661", "1:01:01"},
		{"", ""},
		{"abc", ""},
	}
	for _, tc := range cases {
		if got := formatDuration(tc.in); got != tc.want {
			t.Errorf("formatDuration(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"2009-10-24T23:57:33-07:00", "2009-10-24"},
		{"2024-01-02", "2024-01-02"},
		{"short", "short"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := formatDate(tc.in); got != tc.want {
			t.Errorf("formatDate(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSelectNativeTrack(t *testing.T) {
	tracks := []captionTrack{
		{LanguageCode: "en", Kind: ""},
		{LanguageCode: "en", Kind: "asr"},
		{LanguageCode: "de-DE", Kind: ""},
	}
	got := selectNativeTrack(tracks)
	if got == nil || got.LanguageCode != "en" || got.Kind != "" {
		t.Errorf("selectNativeTrack = %+v, want manual en track", got)
	}

	// only auto-generated native track
	asrOnly := []captionTrack{
		{LanguageCode: "en", Kind: "asr"},
		{LanguageCode: "ja", Kind: "asr"},
	}
	got = selectNativeTrack(asrOnly)
	if got == nil || got.LanguageCode != "en" || got.Kind != "asr" {
		t.Errorf("selectNativeTrack(asrOnly) = %+v, want en asr track", got)
	}

	// dubbed video: manual tracks listed before the original, but the asr track
	// reveals the original (English) language.
	dubbed := []captionTrack{
		{LanguageCode: "ar", Kind: ""},
		{LanguageCode: "bn", Kind: ""},
		{LanguageCode: "zh-Hans", Kind: ""},
		{LanguageCode: "en", Kind: ""},
		{LanguageCode: "en", Kind: "asr"},
		{LanguageCode: "fr", Kind: ""},
	}
	got = selectNativeTrack(dubbed)
	if got == nil || got.LanguageCode != "en" || got.Kind != "" {
		t.Errorf("selectNativeTrack(dubbed) = %+v, want manual en track", got)
	}

	if selectNativeTrack(nil) != nil {
		t.Error("selectNativeTrack(nil) should be nil")
	}
}

func TestRenderTranscript(t *testing.T) {
	cues := []transcriptCue{
		{Text: "hello", Start: 0, Duration: 1},
		{Text: "world", Start: 1, Duration: 1},
		{Text: "sponsor", Start: 5, Duration: 2},
		{Text: "sponsor2", Start: 6, Duration: 2},
		{Text: "goodbye", Start: 10, Duration: 1},
	}

	segments := []skipSegment{
		{Segment: []float64{5, 8}, Category: "sponsor"},
	}

	got := renderTranscript(cues, segments)
	want := "hello world\n[removed: sponsor]\ngoodbye"
	if got != want {
		t.Errorf("renderTranscript =\n%q\nwant\n%q", got, want)
	}
}

func TestRenderTranscriptNoSegments(t *testing.T) {
	cues := []transcriptCue{
		{Text: "a", Start: 0, Duration: 1},
		{Text: "b", Start: 1, Duration: 1},
	}
	got := renderTranscript(cues, nil)
	if got != "a b" {
		t.Errorf("renderTranscript = %q, want %q", got, "a b")
	}
}

func TestOverlappingSegmentIndex(t *testing.T) {
	segments := []skipSegment{
		{Segment: []float64{5, 8}, Category: "sponsor"},
	}

	cases := []struct {
		cue  transcriptCue
		want int
	}{
		{transcriptCue{Start: 0, Duration: 1}, -1},
		{transcriptCue{Start: 5, Duration: 1}, 0},
		{transcriptCue{Start: 7, Duration: 2}, 0},
		{transcriptCue{Start: 4.5, Duration: 1}, 0}, // overlaps at boundary
		{transcriptCue{Start: 8, Duration: 1}, -1},  // starts right at end
	}
	for _, tc := range cases {
		if got := overlappingSegmentIndex(tc.cue, segments); got != tc.want {
			t.Errorf("overlappingSegmentIndex(%+v) = %d, want %d", tc.cue, got, tc.want)
		}
	}
}

func TestParseTranscriptXML(t *testing.T) {
	xmlData := `<transcript><text start="0" dur="1">Hello &amp; hi</text><text start="1" dur="1">world</text></transcript>`
	cues, err := parseTranscriptXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 {
		t.Fatalf("got %d cues, want 2", len(cues))
	}
	if cues[0].Text != "Hello & hi" {
		t.Errorf("cues[0].Text = %q, want %q", cues[0].Text, "Hello & hi")
	}
	if cues[1].Start != 1 || cues[1].Duration != 1 {
		t.Errorf("cues[1] = %+v", cues[1])
	}
}
