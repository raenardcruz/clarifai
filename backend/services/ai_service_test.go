package services

import (
	"os"
	"testing"
)

func TestGenerateTitleForRecordingFallback(t *testing.T) {
	// Point OLLAMA_URL to an unreachable local port to test offline fallback
	origURL := os.Getenv("OLLAMA_URL")
	defer os.Setenv("OLLAMA_URL", origURL)
	os.Setenv("OLLAMA_URL", "http://127.0.0.1:59999")

	generalTitle := GenerateTitleForRecording("Sample meeting transcript", false)
	if generalTitle != "Meeting Note" {
		t.Errorf("expected 'Meeting Note', got '%s'", generalTitle)
	}

	medicalTitle := GenerateTitleForRecording("Sample patient clinical consultation", true)
	if medicalTitle != "Clinical Consultation" {
		t.Errorf("expected 'Clinical Consultation', got '%s'", medicalTitle)
	}
}
