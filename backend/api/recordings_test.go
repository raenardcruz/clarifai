package api

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupRecordingsRoutesNoConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRecordingsRoutes(r.Group("/api/recordings"))
}

func TestBuildTranscriptSegmentsFromRawAndStructured(t *testing.T) {
	segs, dur := buildTranscriptSegments("rec-1", "Alice: Hello there\nBob: Hi Alice", nil, "", 0)
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
	if segs[0].RecordingID != "rec-1" || segs[0].Speaker != "Alice" || segs[0].Text != "Hello there" {
		t.Errorf("unexpected seg 0: %+v", segs[0])
	}
	if segs[1].Speaker != "Bob" || segs[1].Text != "Hi Alice" {
		t.Errorf("unexpected seg 1: %+v", segs[1])
	}
	if dur != nil {
		t.Errorf("expected nil duration when no timestamps supplied, got %v", *dur)
	}

	start0 := 10.0
	end0 := 15.5
	structured := []TranscriptSegmentInput{
		{StartTime: &start0, EndTime: &end0, Speaker: "", Text: "First"},
		{Speaker: "Carol", Text: "Second"},
	}
	segs2, dur2 := buildTranscriptSegments("rec-2", "", structured, "Host", 5.0)
	if len(segs2) != 2 {
		t.Fatalf("expected 2 structured segments, got %d", len(segs2))
	}
	if segs2[0].Speaker != "Host" || segs2[0].StartTime != 15.0 || segs2[0].EndTime != 20.5 {
		t.Errorf("unexpected structured seg 0: %+v", segs2[0])
	}
	if segs2[1].Speaker != "Carol" || segs2[1].StartTime <= 15.0 {
		t.Errorf("unexpected structured seg 1: %+v", segs2[1])
	}
	if dur2 == nil || *dur2 != 20.5 {
		t.Errorf("expected duration 20.5, got %v", dur2)
	}

	emptySegs, emptyDur := buildTranscriptSegments("rec-3", "   \n  ", nil, "", 0)
	if len(emptySegs) != 0 || emptyDur != nil {
		t.Errorf("expected empty result for blank transcript, got %d segs", len(emptySegs))
	}
}
