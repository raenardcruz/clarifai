package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestGetGeminiModels(t *testing.T) {
	expected := []string{
		"gemini-4-argon",
		"gemini-3.8-flash",
		"gemini-3.5-flash-lite",
	}
	actual := GetGeminiModels()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

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

func TestCallGeminiMissingKey(t *testing.T) {
	_, err := callGemini("gemini-3.8-flash", "", "", "test prompt")
	if err == nil {
		t.Errorf("expected error when API key is empty, got nil")
	}
}

func TestCallGeminiSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-gemini-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp := GeminiResponse{
			Candidates: []GeminiCandidate{
				{
					Content: struct {
						Parts []GeminiPart `json:"parts"`
						Role  string       `json:"role"`
					}{
						Parts: []GeminiPart{
							{Text: "Generated summary from Gemini."},
						},
						Role: "model",
					},
					FinishReason: "STOP",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	res, err := callGemini("gemini-3.8-flash", "test-gemini-key", server.URL, "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error calling gemini: %v", err)
	}

	if res != "Generated summary from Gemini." {
		t.Errorf("expected 'Generated summary from Gemini.', got '%s'", res)
	}
}
