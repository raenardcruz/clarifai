package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFetchFathomAPI_InlineAndFallbackAndPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "fathom-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/meetings":
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"limit":       10,
					"next_cursor": "page2",
					"items": []map[string]interface{}{
						{
							"recording_id":         101,
							"meeting_title":        "Q4 Roadmap Sync",
							"url":                  "https://fathom.video/calls/101",
							"recording_start_time": "2026-11-10T10:00:00Z",
							"recording_end_time":   "2026-11-10T10:30:00Z",
							"default_summary": map[string]interface{}{
								"template_name":      "general",
								"markdown_formatted": "## Summary\nAligned on Q4 milestones.",
							},
							"transcript": []map[string]interface{}{
								{
									"speaker":   map[string]interface{}{"display_name": "Alice"},
									"text":      "Let's review Q4 goals.",
									"timestamp": "00:00:05",
								},
								{
									"speaker":   map[string]interface{}{"display_name": "Bob"},
									"text":      "API launch is on track.",
									"timestamp": "00:00:15",
								},
							},
						},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"limit":       10,
				"next_cursor": "",
				"items": []map[string]interface{}{
					{
						"recording_id": "102",
						"title":        "Design Critique",
						"share_url":    "https://fathom.video/share/102",
						"created_at":   "2026-11-11T14:00:00Z",
					},
				},
			})
		case "/recordings/102/summary":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"summary": "Reviewed new navigation.",
				"action_items": []map[string]interface{}{
					{
						"description": "Finalize icons",
						"completed":   false,
						"assignee":    map[string]interface{}{"name": "Clara"},
					},
				},
			})
		case "/recordings/102/transcript":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"segments": []map[string]interface{}{
					{
						"start_time": "00:00:02",
						"end_time":   "00:00:09",
						"speaker":    map[string]interface{}{"name": "Clara"},
						"text":       "Here are the updated icons.",
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	meetings, err := fetchFathomAPI(context.Background(), srv.URL, "fathom-secret")
	if err != nil {
		t.Fatalf("fetchFathomAPI returned error: %v", err)
	}
	if len(meetings) != 2 {
		t.Fatalf("expected 2 meetings, got %d", len(meetings))
	}

	m1 := meetings[0]
	if m1.ExternalID != "101" || m1.Title != "Q4 Roadmap Sync" {
		t.Errorf("unexpected meeting 1: %+v", m1)
	}
	if m1.Duration == nil || *m1.Duration != 1800 {
		t.Errorf("expected duration 1800s, got %v", m1.Duration)
	}
	if len(m1.Segments) != 2 || m1.Segments[0].Speaker != "Alice" || m1.Segments[0].StartTime != 5 || m1.Segments[0].EndTime != 15 {
		t.Errorf("unexpected segments on meeting 1: %+v", m1.Segments)
	}

	m2 := meetings[1]
	if m2.ExternalID != "102" || m2.Title != "Design Critique" {
		t.Errorf("unexpected meeting 2: %+v", m2)
	}
	if !strings.Contains(m2.SummaryMD, "Reviewed new navigation.") || !strings.Contains(m2.SummaryMD, "Finalize icons") {
		t.Errorf("expected fallback summary + action item, got %q", m2.SummaryMD)
	}
	if len(m2.Segments) != 1 || m2.Segments[0].Speaker != "Clara" {
		t.Errorf("expected fallback transcript segments, got %+v", m2.Segments)
	}
}

func TestFetchFirefliesAPI_GraphQL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ff-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"transcripts": []map[string]interface{}{
					{
						"id":             "ff_abc",
						"title":          "Weekly Engineering Standup",
						"date":           1762855200000,
						"duration":       15.0, // minutes vs 900s end time
						"transcript_url": "https://app.fireflies.ai/view/ff_abc",
						"summary": map[string]interface{}{
							"overview":         "Covered backend sprint progress.",
							"bullet_gist":      "- Migration finished\n- Tests green",
							"action_items":     "Deploy staging build",
							"keywords":         []string{"backend", "sprint"},
							"topics_discussed": []string{"API", "Database"},
						},
						"speakers": []map[string]interface{}{
							{"id": 1, "name": "Dana"},
						},
						"sentences": []map[string]interface{}{
							{
								"index":        0,
								"text":         "Migration is complete.",
								"start_time":   2.0,
								"end_time":     6.0,
								"speaker_id":   1,
								"speaker_name": "",
							},
							{
								"index":        1,
								"text":         "Awesome, let's ship it.",
								"start_time":   880.0,
								"end_time":     900.0,
								"speaker_id":   2,
								"speaker_name": "Evan",
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	meetings, err := fetchFirefliesAPI(context.Background(), srv.URL, "ff-token")
	if err != nil {
		t.Fatalf("fetchFirefliesAPI returned error: %v", err)
	}
	if len(meetings) != 1 {
		t.Fatalf("expected 1 meeting, got %d", len(meetings))
	}
	m := meetings[0]
	if m.ExternalID != "ff_abc" || m.Title != "Weekly Engineering Standup" {
		t.Errorf("unexpected meeting metadata: %+v", m)
	}
	if m.Duration == nil || *m.Duration != 900 {
		t.Errorf("expected normalized duration 900s, got %v", m.Duration)
	}
	if len(m.Segments) != 2 || m.Segments[0].Speaker != "Dana" || m.Segments[1].Speaker != "Evan" {
		t.Errorf("unexpected segments: %+v", m.Segments)
	}
	if !strings.Contains(m.SummaryMD, "## Overview") || !strings.Contains(m.SummaryMD, "Deploy staging build") {
		t.Errorf("unexpected summary markdown: %q", m.SummaryMD)
	}
}

func TestFetchOtterAPI_ListAndDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer otter-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/conversations":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"meta": map[string]interface{}{"has_more": false, "next_cursor": ""},
				"data": []map[string]interface{}{
					{
						"id":         "conv_1",
						"title":      "Customer Discovery Call",
						"created_at": "2026-11-12T09:00:00Z",
					},
				},
			})
		case "/conversations/conv_1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":               "conv_1",
					"title":            "Customer Discovery Call",
					"url":              "https://otter.ai/u/conv_1",
					"created_at":       "2026-11-12T09:00:00Z",
					"abstract_summary": "Customer wants SSO and audit logs.",
					"outline": []map[string]interface{}{
						{"topic": "SSO Requirements", "timestamp": "00:01:00"},
					},
					"action_items": []map[string]interface{}{
						{"task": "Share SAML docs", "assignee": "Alex"},
					},
					"transcript": []map[string]interface{}{
						{
							"speaker":    map[string]interface{}{"id": "s1", "name": "Alex"},
							"start_time": 0.0,
							"end_time":   4.5,
							"text":       "Thanks for joining today.",
						},
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	meetings, err := fetchOtterAPI(context.Background(), srv.URL, "otter-key")
	if err != nil {
		t.Fatalf("fetchOtterAPI returned error: %v", err)
	}
	if len(meetings) != 1 {
		t.Fatalf("expected 1 meeting, got %d", len(meetings))
	}
	m := meetings[0]
	if m.ExternalID != "conv_1" || m.ExternalURL != "https://otter.ai/u/conv_1" {
		t.Errorf("unexpected Otter meeting: %+v", m)
	}
	if !strings.Contains(m.SummaryMD, "Customer wants SSO") || !strings.Contains(m.SummaryMD, "Share SAML docs") {
		t.Errorf("unexpected Otter summary: %q", m.SummaryMD)
	}
	if len(m.Segments) != 1 || m.Segments[0].Speaker != "Alex" {
		t.Errorf("unexpected Otter segments: %+v", m.Segments)
	}
}

func TestFetchViaMCP_StreamableHTTPAndSSE(t *testing.T) {
	var sawSession int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mcp-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-123")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"serverInfo":      map[string]string{"name": "mock-mcp"},
				},
			})
		case "notifications/initialized":
			if r.Header.Get("Mcp-Session-Id") == "sess-123" {
				atomic.StoreInt32(&sawSession, 1)
			}
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			// Test SSE response framing
			w.Header().Set("Content-Type", "text/event-stream")
			respJSON, _ := json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name": "fireflies_get_transcripts",
							"inputSchema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"limit": map[string]string{"type": "integer"},
								},
							},
						},
						{
							"name": "fireflies_get_transcript",
							"inputSchema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"transcriptId": map[string]string{"type": "string"},
								},
							},
						},
					},
				},
			})
			_, _ = w.Write([]byte("event: message\ndata: " + string(respJSON) + "\n\n"))
		case "tools/call":
			paramsBytes, _ := json.Marshal(req.Params)
			var callParams struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			_ = json.Unmarshal(paramsBytes, &callParams)
			w.Header().Set("Content-Type", "application/json")
			if callParams.Name == "fireflies_get_transcripts" {
				payloadText, _ := json.Marshal(map[string]interface{}{
					"transcripts": []map[string]interface{}{
						{
							"id":         "mcp-tr-1",
							"title":      "MCP Synced Call",
							"created_at": "2026-11-14T12:00:00Z",
						},
					},
				})
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]interface{}{
						"content": []map[string]string{{"type": "text", "text": string(payloadText)}},
					},
				})
				return
			}
			if callParams.Name == "fireflies_get_transcript" {
				if callParams.Arguments["transcriptId"] != "mcp-tr-1" {
					t.Errorf("expected transcriptId mcp-tr-1, got %v", callParams.Arguments)
				}
				detailText, _ := json.Marshal(map[string]interface{}{
					"id": "mcp-tr-1",
					"summary": map[string]interface{}{
						"overview": "Synced over MCP.",
					},
					"sentences": []map[string]interface{}{
						{"speaker_name": "Sam", "text": "Hello from MCP!", "start_time": 1.0, "end_time": 3.5},
					},
				})
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]interface{}{
						"content": []map[string]string{{"type": "text", "text": string(detailText)}},
					},
				})
			}
		}
	}))
	defer srv.Close()

	meetings, err := fetchViaMCP(context.Background(), ProviderFireflies, srv.URL, "mcp-token")
	if err != nil {
		t.Fatalf("fetchViaMCP returned error: %v", err)
	}
	if atomic.LoadInt32(&sawSession) != 1 {
		t.Errorf("expected Mcp-Session-Id header to be propagated")
	}
	if len(meetings) != 1 {
		t.Fatalf("expected 1 MCP meeting, got %d", len(meetings))
	}
	m := meetings[0]
	if m.ExternalID != "mcp-tr-1" || m.Title != "MCP Synced Call" {
		t.Errorf("unexpected MCP meeting: %+v", m)
	}
	if m.SummaryMD != "Synced over MCP." {
		t.Errorf("expected summary 'Synced over MCP.', got %q", m.SummaryMD)
	}
	if len(m.Segments) != 1 || m.Segments[0].Speaker != "Sam" || m.Segments[0].Text != "Hello from MCP!" {
		t.Errorf("unexpected MCP segments: %+v", m.Segments)
	}
}

func TestTestIntegrationConnection(t *testing.T) {
	fathomSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "good-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer fathomSrv.Close()

	origOverride := FathomAPIBaseOverride
	FathomAPIBaseOverride = fathomSrv.URL
	defer func() { FathomAPIBaseOverride = origOverride }()

	bad := TestIntegrationConnection(context.Background(), ProviderFathom, "api", "bad-key", "")
	if bad.OK {
		t.Errorf("expected bad key test to fail")
	}
	good := TestIntegrationConnection(context.Background(), ProviderFathom, "api", "good-key", "")
	if !good.OK {
		t.Errorf("expected good key test to succeed, got: %s", good.Message)
	}
}

func TestFilterOutDeletedMeetingsAndCollectIDs(t *testing.T) {
	meetings := []ImportedMeeting{
		{ExternalID: "101", Title: "Keep First"},
		{ExternalID: " 102 ", Title: "Deleted With Spaces"},
		{ExternalID: "103", Title: "Keep Second"},
		{ExternalID: "104", Title: "Deleted Exact"},
	}

	deletedIDs := map[string]bool{
		"102": true,
		"104": true,
	}

	kept, skipped := filterOutDeletedMeetings(meetings, deletedIDs)
	if skipped != 2 {
		t.Fatalf("expected 2 skipped meetings, got %d", skipped)
	}
	if len(kept) != 2 || kept[0].ExternalID != "101" || kept[1].ExternalID != "103" {
		t.Fatalf("unexpected kept meetings: %+v", kept)
	}

	// Empty deleted set passes meetings through untouched
	keptAll, skippedNone := filterOutDeletedMeetings(meetings, nil)
	if skippedNone != 0 || len(keptAll) != len(meetings) {
		t.Fatalf("expected all %d meetings with 0 skipped, got %d kept / %d skipped", len(meetings), len(keptAll), skippedNone)
	}
}

func TestMarkImportedMeetingDeletedNoOps(t *testing.T) {
	id := "101"
	empty := "   "
	// Nil DB, nil ExternalID, empty ExternalID, and local source must safely no-op without DB writes
	if err := MarkImportedMeetingDeleted(nil, 1, ProviderFathom, &id); err != nil {
		t.Errorf("expected nil db to no-op, got %v", err)
	}
	if err := MarkImportedMeetingDeleted(nil, 1, ProviderFathom, nil); err != nil {
		t.Errorf("expected nil externalID to no-op, got %v", err)
	}
	if err := MarkImportedMeetingDeleted(nil, 1, ProviderFathom, &empty); err != nil {
		t.Errorf("expected empty externalID to no-op, got %v", err)
	}
	if err := MarkImportedMeetingDeleted(nil, 1, SourceLocal, &id); err != nil {
		t.Errorf("expected local source to no-op, got %v", err)
	}
}
