package services

import (
	"testing"
)

func TestParseTranscriptText_SpeakerDialogue(t *testing.T) {
	input := `Alice: Welcome everyone to the Q4 planning sync.
Today we need to finalize the API roadmap.

Bob: Thanks Alice! Engineering is ready for sprint 12.
Dr. Jane Doe: Let's also check compliance requirements.`

	segs := ParseTranscriptText(input, "")
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(segs), segs)
	}

	if segs[0].Speaker != "Alice" || segs[0].Text != "Welcome everyone to the Q4 planning sync.\nToday we need to finalize the API roadmap." {
		t.Errorf("unexpected seg 0: %+v", segs[0])
	}
	if segs[1].Speaker != "Bob" || segs[1].Text != "Thanks Alice! Engineering is ready for sprint 12." {
		t.Errorf("unexpected seg 1: %+v", segs[1])
	}
	if segs[2].Speaker != "Dr. Jane Doe" || segs[2].Text != "Let's also check compliance requirements." {
		t.Errorf("unexpected seg 2: %+v", segs[2])
	}
	if !(segs[0].StartTime < segs[1].StartTime && segs[1].StartTime < segs[2].StartTime) {
		t.Errorf("expected strictly increasing synthetic start times, got %v, %v, %v", segs[0].StartTime, segs[1].StartTime, segs[2].StartTime)
	}
}

func TestParseTranscriptText_TimestampedInline(t *testing.T) {
	input := `[00:12] Alice: First point
[01:05] Bob: Second point
(01:02:03) Carol: Third point`

	segs := ParseTranscriptText(input, "")
	if len(segs) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(segs), segs)
	}
	if segs[0].StartTime != 12 || segs[0].Speaker != "Alice" || segs[0].Text != "First point" {
		t.Errorf("unexpected seg 0: %+v", segs[0])
	}
	if segs[1].StartTime != 65 || segs[1].Speaker != "Bob" || segs[1].Text != "Second point" {
		t.Errorf("unexpected seg 1: %+v", segs[1])
	}
	if segs[2].StartTime != 3723 || segs[2].Speaker != "Carol" || segs[2].Text != "Third point" {
		t.Errorf("unexpected seg 2: %+v", segs[2])
	}
}

func TestParseTranscriptText_TwoLineHeaderFormat(t *testing.T) {
	input := `Alice Smith  0:15
Hello team, let's get started.

Bob Jones  1:02
Sounds great to me.`

	segs := ParseTranscriptText(input, "")
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d: %+v", len(segs), segs)
	}
	if segs[0].Speaker != "Alice Smith" || segs[0].StartTime != 15 || segs[0].Text != "Hello team, let's get started." {
		t.Errorf("unexpected seg 0: %+v", segs[0])
	}
	if segs[1].Speaker != "Bob Jones" || segs[1].StartTime != 62 || segs[1].Text != "Sounds great to me." {
		t.Errorf("unexpected seg 1: %+v", segs[1])
	}
}

func TestParseTranscriptText_WebVTTAndSRT(t *testing.T) {
	vtt := `WEBVTT

00:00:01.500 --> 00:00:04.000
<v Alice>Hello everyone.</v>

00:00:05.000 --> 00:00:08.250
Bob: Good morning!`

	segs := ParseTranscriptText(vtt, "")
	if len(segs) != 2 {
		t.Fatalf("expected 2 VTT segments, got %d: %+v", len(segs), segs)
	}
	if segs[0].Speaker != "Alice" || segs[0].StartTime != 1.5 || segs[0].EndTime != 4.0 || segs[0].Text != "Hello everyone." {
		t.Errorf("unexpected VTT seg 0: %+v", segs[0])
	}
	if segs[1].Speaker != "Bob" || segs[1].StartTime != 5.0 || segs[1].EndTime != 8.25 || segs[1].Text != "Good morning!" {
		t.Errorf("unexpected VTT seg 1: %+v", segs[1])
	}

	srt := `1
00:00:10,000 --> 00:00:14,500
Speaker 1: First SRT line.

2
00:00:15,000 --> 00:00:19,000
Second SRT line without speaker.`

	srtSegs := ParseTranscriptText(srt, "Host")
	if len(srtSegs) != 2 {
		t.Fatalf("expected 2 SRT segments, got %d: %+v", len(srtSegs), srtSegs)
	}
	if srtSegs[0].Speaker != "Speaker 1" || srtSegs[0].StartTime != 10.0 || srtSegs[0].EndTime != 14.5 {
		t.Errorf("unexpected SRT seg 0: %+v", srtSegs[0])
	}
	if srtSegs[1].Speaker != "Host" || srtSegs[1].StartTime != 15.0 || srtSegs[1].Text != "Second SRT line without speaker." {
		t.Errorf("unexpected SRT seg 1: %+v", srtSegs[1])
	}
}

func TestParseTranscriptText_PlainParagraphsAndURLs(t *testing.T) {
	plain := `We are meeting at 10:30 to check https://example.com/docs and review the sprint.
- Deadline: Friday at 5pm

Second paragraph continues the notes without any speaker labels.`

	segs := ParseTranscriptText(plain, "Narrator")
	if len(segs) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d: %+v", len(segs), segs)
	}
	if segs[0].Speaker != "Narrator" || segs[1].Speaker != "Narrator" {
		t.Errorf("expected default speaker 'Narrator', got %q and %q", segs[0].Speaker, segs[1].Speaker)
	}
	if segs[0].StartTime >= segs[1].StartTime {
		t.Errorf("expected increasing start times, got %v >= %v", segs[0].StartTime, segs[1].StartTime)
	}
}

func TestParseTranscriptText_JSON(t *testing.T) {
	jsonInput := `[
		{"speaker": "Alice", "text": "Hello", "start": 0, "end": 2.5},
		{"speaker": "Bob", "text": "Hi there", "start": 3.0, "end": 5.0}
	]`
	segs := ParseTranscriptText(jsonInput, "")
	if len(segs) != 2 {
		t.Fatalf("expected 2 JSON segments, got %d", len(segs))
	}
	if segs[0].Speaker != "Alice" || segs[0].EndTime != 2.5 {
		t.Errorf("unexpected JSON seg 0: %+v", segs[0])
	}
	if segs[1].Speaker != "Bob" || segs[1].StartTime != 3.0 {
		t.Errorf("unexpected JSON seg 1: %+v", segs[1])
	}
}
