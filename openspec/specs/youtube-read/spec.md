## Purpose

Detect YouTube video URLs passed to `read` and return channel, title, duration, date, and a native-language transcript with SponsorBlock segments removed, as markdown.

## Requirements

### Requirement: YouTube URL detection
The `read` command SHALL detect when a target URL identifies a YouTube video and route it to the transcript path instead of generic readability conversion. Detection SHALL recognize the following URL shapes and extract the video ID:

- `youtube.com/watch?v=<id>`
- `youtu.be/<id>`
- `youtube.com/shorts/<id>`
- `youtube.com/live/<id>`
- `youtube.com/embed/<id>`
- `m.youtube.com/watch?v=<id>`
- `music.youtube.com/watch?v=<id>`

Non-YouTube URLs SHALL continue through the existing readability conversion path unchanged.

#### Scenario: Standard watch URL
- **WHEN** `read` is invoked with `https://www.youtube.com/watch?v=dQw4w9WgXcQ`
- **THEN** the URL is detected as YouTube and the video ID `dQw4w9WgXcQ` is extracted

#### Scenario: Short youtu.be URL
- **WHEN** `read` is invoked with `https://youtu.be/dQw4w9WgXcQ`
- **THEN** the video ID `dQw4w9WgXcQ` is extracted and the transcript path is used

#### Scenario: Shorts / live / embed / mobile / music hosts
- **WHEN** `read` is invoked with a URL using `youtube.com/shorts/<id>`, `youtube.com/live/<id>`, `youtube.com/embed/<id>`, `m.youtube.com/watch?v=<id>`, or `music.youtube.com/watch?v=<id>`
- **THEN** the video ID is extracted and the transcript path is used

#### Scenario: Non-YouTube URL
- **WHEN** `read` is invoked with `https://example.com/article`
- **THEN** the existing readability conversion path is used and no YouTube processing occurs

### Requirement: Metadata extraction
For a detected YouTube video, the system SHALL retrieve the video title, channel name, duration, and publish date.

#### Scenario: Metadata present
- **WHEN** a YouTube video has available title, author, length, and publish date
- **THEN** all four fields are extracted and included in the output

### Requirement: Native-language transcript
The system SHALL return the transcript in the video's native (spoken) language without requiring a language flag. The native language SHALL be determined from the auto-generated (`asr`) caption track, which is always in the original audio language; a manual track in that language is preferred for quality. Transcript text SHALL be rendered as plain text with no timestamps.

#### Scenario: Dubbed video with multiple tracks
- **WHEN** the video has caption tracks in multiple languages (including a dubbed language listed first)
- **THEN** the original language is used (from the `asr` track), not the first-listed dubbed language

### Requirement: Flowing transcript text
The transcript SHALL be rendered as concatenated plain text: individual caption cues joined with spaces into continuous text. The system SHALL NOT insert line breaks based on gaps, pauses, or punctuation.

#### Scenario: Continuous text
- **WHEN** a transcript is rendered
- **THEN** caption cue texts are joined into flowing text without gap- or punctuation-based line breaks

### Requirement: SponsorBlock filtering
The system SHALL fetch SponsorBlock skip segments for the video and remove the overlapping transcript content. Segments SHALL be requested for all skip categories: `sponsor`, `selfpromo`, `interaction`, `intro`, `outro`, `preview`, `music_offtopic`. Each removal SHALL be marked inline as `[removed: <category>]`.

#### Scenario: Sponsored segment removed
- **WHEN** a video has a sponsor segment and the transcript contains text within that time range
- **THEN** the overlapping transcript text is removed and a `[removed: sponsor]` marker is emitted at that position

#### Scenario: No SponsorBlock segments
- **WHEN** SponsorBlock returns no segments (404) or is unreachable
- **THEN** the transcript is returned unmodified without error and without markers

### Requirement: Markdown output format
For a detected YouTube video, `read` SHALL emit plain markdown with the title, channel, duration, date, and transcript:

```markdown
# <title>

Channel: <channel>
Duration: <duration>
Published: <date>

<transcript heading>

<transcript text>
```

The transcript section SHALL be introduced by a `Transcript` heading. Duration SHALL be formatted as `H:MM:SS` or `M:SS`.

#### Scenario: Plain output
- **WHEN** `read` is invoked on a YouTube URL without `--json`
- **THEN** the markdown document above is printed with the transcript under a `Transcript` heading

#### Scenario: JSON output
- **WHEN** `read` is invoked on a YouTube URL with `--json`
- **THEN** a JSON object containing `url`, `title`, `channel`, `duration`, `date`, and `transcript` is printed

### Requirement: Clean error on transcript failure
When a YouTube video has no captions available or transcript retrieval fails, the system SHALL return a clean error (to stderr) instead of falling back to the readability conversion of the watch page. No page chrome (footer links, navigation) SHALL appear in the output for a YouTube URL.

#### Scenario: Transcripts disabled
- **WHEN** a YouTube video has captions disabled or unavailable
- **THEN** `read` exits non-zero with a clean error message and no page-chrome output
