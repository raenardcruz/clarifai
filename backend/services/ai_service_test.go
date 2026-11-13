package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
	"note-taker/backend/models"
)

func TestSettingsSchema(t *testing.T) {
	s, err := schema.Parse(&models.Settings{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	f := s.FieldsByName["AIProvider"]
	if f == nil {
		t.Fatalf("field AIProvider not found")
	}
	t.Logf("AIProvider DB Name: %s", f.DBName)
	if f.DBName != "ai_provider" {
		t.Errorf("expected DBName 'ai_provider', got '%s'", f.DBName)
	}
}

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

func TestCleanHTMLResponse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain html",
			in:   "<!DOCTYPE html><html><body>Hi</body></html>",
			want: "<!DOCTYPE html><html><body>Hi</body></html>",
		},
		{
			name: "fenced html",
			in:   "```html\n<!DOCTYPE html><html><body>Hi</body></html>\n```",
			want: "<!DOCTYPE html><html><body>Hi</body></html>",
		},
		{
			name: "fenced generic",
			in:   "```\n<!DOCTYPE html><html><body>Hi</body></html>\n```",
			want: "<!DOCTYPE html><html><body>Hi</body></html>",
		},
	}

	for _, tc := range cases {
		got := cleanHTMLResponse(tc.in)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBuildSummaryPromptIncludesTemplateAndMetadata(t *testing.T) {
	meta := SummaryMetadata{
		Title:    "Architecture Review",
		Date:     "Nov 13, 2026",
		Duration: "45m 10s",
		Speakers: "Alice, Bob",
	}
	out := buildSummaryPrompt("Alice: Hello world.", "Focus on KPIs and charts.", meta)

	required := []string{
		"Focus on KPIs and charts.",
		"Architecture Review",
		"Nov 13, 2026",
		"45m 10s",
		"Alice, Bob",
		"<!DOCTYPE html>",
		"document-sheet",
		"kpi-grid",
		"Alice: Hello world.",
	}
	for _, needle := range required {
		if !strings.Contains(out, needle) {
			t.Errorf("expected prompt to contain %q", needle)
		}
	}
}
