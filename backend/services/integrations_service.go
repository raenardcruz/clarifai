package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"note-taker/backend/models"
)

const (
	SourceLocal     = "local"
	ProviderFathom    = "fathom"
	ProviderFireflies = "fireflies"
	ProviderOtter     = "otter"

	DefaultFathomAPIBase      = "https://api.fathom.ai/external/v1"
	DefaultFathomMCPURL       = "https://api.fathom.ai/mcp"
	DefaultFirefliesGraphQL   = "https://api.fireflies.ai/graphql"
	DefaultFirefliesMCPURL    = "https://api.fireflies.ai/mcp"
	DefaultOtterAPIBase       = "https://api.otter.ai"
)

// Package-level endpoints allow unit tests to inject httptest server URLs cleanly.
var (
	FathomAPIBaseOverride    string
	FirefliesGraphQLOverride string
)

// ImportedSegment represents a single speaker-attributed transcript utterance from a 3rd-party note-taker.
type ImportedSegment struct {
	StartTime float64
	EndTime   float64
	Speaker   string
	Text      string
}

// ImportedMeeting is the normalized representation of a meeting fetched from Fathom, Fireflies.ai, or Otter.ai.
type ImportedMeeting struct {
	ExternalID  string
	ExternalURL string
	Title       string
	CreatedAt   time.Time
	Duration    *float64
	SummaryMD   string
	Segments    []ImportedSegment
}

// ProviderSyncResult summarizes the outcome of syncing a single integration provider.
type ProviderSyncResult struct {
	Provider string `json:"provider"`
	Method   string `json:"method"`
	Fetched  int    `json:"fetched"`
	Imported int    `json:"imported"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
	Error    string `json:"error,omitempty"`
	SyncedAt string `json:"synced_at"`
}

// IntegrationTestResult reports connection diagnostics for a provider.
type IntegrationTestResult struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message"`
	Tools   []string `json:"tools,omitempty"`
}

var integrationHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

func resolveFathomAPIBase() string {
	if FathomAPIBaseOverride != "" {
		return strings.TrimRight(FathomAPIBaseOverride, "/")
	}
	return DefaultFathomAPIBase
}

func resolveFirefliesGraphQLURL() string {
	if FirefliesGraphQLOverride != "" {
		return FirefliesGraphQLOverride
	}
	return DefaultFirefliesGraphQL
}

// SyncAllEnabledIntegrations runs sync for every enabled integration (or a single named provider).
func SyncAllEnabledIntegrations(ctx context.Context, db *gorm.DB, userID uint, targetProvider string) ([]ProviderSyncResult, error) {
	var settings models.Settings
	if err := db.First(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	target := strings.ToLower(strings.TrimSpace(targetProvider))
	var providers []string
	switch target {
	case "", "all":
		if settings.FathomEnabled {
			providers = append(providers, ProviderFathom)
		}
		if settings.FirefliesEnabled {
			providers = append(providers, ProviderFireflies)
		}
		if settings.OtterEnabled {
			providers = append(providers, ProviderOtter)
		}
	case ProviderFathom, ProviderFireflies, ProviderOtter:
		providers = []string{target}
	default:
		return nil, fmt.Errorf("unknown integration provider: %s", targetProvider)
	}

	results := make([]ProviderSyncResult, 0, len(providers))
	for _, p := range providers {
		res := syncSingleProvider(ctx, db, userID, &settings, p)
		results = append(results, res)
	}

	_ = db.Save(&settings).Error
	return results, nil
}

func syncSingleProvider(ctx context.Context, db *gorm.DB, userID uint, settings *models.Settings, provider string) ProviderSyncResult {
	now := time.Now().UTC()
	res := ProviderSyncResult{
		Provider: provider,
		SyncedAt: now.Format(time.RFC3339),
	}

	var meetings []ImportedMeeting
	var err error

	switch provider {
	case ProviderFathom:
		method := normalizeMethod(settings.FathomMethod)
		res.Method = method
		apiKey := derefString(settings.FathomAPIKey)
		if method == "mcp" {
			mcpURL := strings.TrimSpace(settings.FathomMCPURL)
			if mcpURL == "" {
				mcpURL = DefaultFathomMCPURL
			}
			meetings, err = fetchViaMCP(ctx, provider, mcpURL, apiKey)
		} else {
			if apiKey == "" {
				err = fmt.Errorf("Fathom API key is not configured")
			} else {
				meetings, err = fetchFathomAPI(ctx, resolveFathomAPIBase(), apiKey)
			}
		}
		settings.FathomLastSyncAt = &now
		if err != nil {
			res.Error = err.Error()
			settings.FathomLastSyncStatus = "Error: " + truncateStatus(err.Error())
		}

	case ProviderFireflies:
		method := normalizeMethod(settings.FirefliesMethod)
		res.Method = method
		apiKey := derefString(settings.FirefliesAPIKey)
		if method == "mcp" {
			mcpURL := strings.TrimSpace(settings.FirefliesMCPURL)
			if mcpURL == "" {
				mcpURL = DefaultFirefliesMCPURL
			}
			meetings, err = fetchViaMCP(ctx, provider, mcpURL, apiKey)
		} else {
			if apiKey == "" {
				err = fmt.Errorf("Fireflies.ai API key is not configured")
			} else {
				meetings, err = fetchFirefliesAPI(ctx, resolveFirefliesGraphQLURL(), apiKey)
			}
		}
		settings.FirefliesLastSyncAt = &now
		if err != nil {
			res.Error = err.Error()
			settings.FirefliesLastSyncStatus = "Error: " + truncateStatus(err.Error())
		}

	case ProviderOtter:
		method := normalizeMethod(settings.OtterMethod)
		res.Method = method
		apiKey := derefString(settings.OtterAPIKey)
		if method == "mcp" {
			mcpURL := strings.TrimSpace(settings.OtterMCPURL)
			if mcpURL == "" {
				err = fmt.Errorf("Otter.ai MCP Server URL is not configured")
			} else {
				meetings, err = fetchViaMCP(ctx, provider, mcpURL, apiKey)
			}
		} else {
			if apiKey == "" {
				err = fmt.Errorf("Otter.ai API key is not configured")
			} else {
				baseURL := strings.TrimSpace(settings.OtterAPIBaseURL)
				if baseURL == "" {
					baseURL = DefaultOtterAPIBase
				}
				meetings, err = fetchOtterAPI(ctx, strings.TrimRight(baseURL, "/"), apiKey)
			}
		}
		settings.OtterLastSyncAt = &now
		if err != nil {
			res.Error = err.Error()
			settings.OtterLastSyncStatus = "Error: " + truncateStatus(err.Error())
		}
	}

	if err != nil {
		return res
	}

	res.Fetched = len(meetings)
	imported, updated, skipped, dbErr := upsertImportedMeetings(db, userID, provider, meetings)
	if dbErr != nil {
		res.Error = dbErr.Error()
		setLastSyncStatus(settings, provider, "Error: "+truncateStatus(dbErr.Error()))
		return res
	}

	res.Imported = imported
	res.Updated = updated
	res.Skipped = skipped

	statusSummary := fmt.Sprintf("Synced %d fetched (%d new, %d updated, %d skipped)", res.Fetched, res.Imported, res.Updated, res.Skipped)
	setLastSyncStatus(settings, provider, statusSummary)
	return res
}

func setLastSyncStatus(settings *models.Settings, provider, status string) {
	switch provider {
	case ProviderFathom:
		settings.FathomLastSyncStatus = status
	case ProviderFireflies:
		settings.FirefliesLastSyncStatus = status
	case ProviderOtter:
		settings.OtterLastSyncStatus = status
	}
}

func truncateStatus(s string) string {
	if len(s) > 220 {
		return s[:220] + "..."
	}
	return s
}

func normalizeMethod(m string) string {
	if strings.EqualFold(strings.TrimSpace(m), "mcp") {
		return "mcp"
	}
	return "api"
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

// TestIntegrationConnection validates credentials / MCP endpoint reachability for a provider.
func TestIntegrationConnection(ctx context.Context, provider, method, apiKey, endpointURL string) IntegrationTestResult {
	provider = strings.ToLower(strings.TrimSpace(provider))
	method = normalizeMethod(method)
	apiKey = strings.TrimSpace(apiKey)
	endpointURL = strings.TrimSpace(endpointURL)

	if method == "mcp" {
		if endpointURL == "" {
			switch provider {
			case ProviderFathom:
				endpointURL = DefaultFathomMCPURL
			case ProviderFireflies:
				endpointURL = DefaultFirefliesMCPURL
			default:
				return IntegrationTestResult{OK: false, Message: "MCP Server URL is required"}
			}
		}
		client := buildProviderMCPClient(provider, endpointURL, apiKey)
		if err := client.Initialize(ctx); err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		tools, err := client.ListTools(ctx)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: fmt.Sprintf("Connected to MCP server, but tools/list failed: %v", err)}
		}
		toolNames := make([]string, 0, len(tools))
		for _, t := range tools {
			toolNames = append(toolNames, t.Name)
		}
		return IntegrationTestResult{
			OK:      true,
			Message: fmt.Sprintf("Connected via MCP (%d tool(s) discovered)", len(toolNames)),
			Tools:   toolNames,
		}
	}

	if apiKey == "" {
		return IntegrationTestResult{OK: false, Message: "API key is required"}
	}

	switch provider {
	case ProviderFathom:
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolveFathomAPIBase()+"/meetings?limit=1", nil)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		req.Header.Set("X-Api-Key", apiKey)
		resp, err := integrationHTTPClient.Do(req)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return IntegrationTestResult{OK: false, Message: "Invalid Fathom API key (unauthorized)"}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return IntegrationTestResult{OK: false, Message: fmt.Sprintf("Fathom API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
		}
		return IntegrationTestResult{OK: true, Message: "Successfully connected to Fathom API"}

	case ProviderFireflies:
		payload := map[string]interface{}{
			"query": `query { transcripts(limit: 1) { id title } }`,
		}
		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, resolveFirefliesGraphQLURL(), bytes.NewReader(bodyBytes))
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := integrationHTTPClient.Do(req)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		defer resp.Body.Close()
		rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return IntegrationTestResult{OK: false, Message: "Invalid Fireflies.ai API key (unauthorized)"}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return IntegrationTestResult{OK: false, Message: fmt.Sprintf("Fireflies API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))}
		}
		var gqlResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(rawBody, &gqlResp); err == nil && len(gqlResp.Errors) > 0 {
			return IntegrationTestResult{OK: false, Message: "Fireflies GraphQL error: " + gqlResp.Errors[0].Message}
		}
		return IntegrationTestResult{OK: true, Message: "Successfully connected to Fireflies.ai GraphQL API"}

	case ProviderOtter:
		base := strings.TrimRight(endpointURL, "/")
		if base == "" {
			base = DefaultOtterAPIBase
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/conversations?limit=1", nil)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := integrationHTTPClient.Do(req)
		if err != nil {
			return IntegrationTestResult{OK: false, Message: err.Error()}
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return IntegrationTestResult{OK: false, Message: "Invalid Otter.ai Bearer token (unauthorized)"}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return IntegrationTestResult{OK: false, Message: fmt.Sprintf("Otter API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
		}
		return IntegrationTestResult{OK: true, Message: "Successfully connected to Otter.ai Public API"}
	}

	return IntegrationTestResult{OK: false, Message: "Unknown integration provider"}
}

// MarkImportedMeetingDeleted records a deleted imported meeting's external ID so future syncs skip it.
func MarkImportedMeetingDeleted(db *gorm.DB, userID uint, source string, externalID *string) error {
	if db == nil || externalID == nil {
		return nil
	}
	extID := strings.TrimSpace(*externalID)
	src := strings.ToLower(strings.TrimSpace(source))
	if extID == "" || src == "" || src == SourceLocal {
		return nil
	}
	entry := models.DeletedImport{
		UserID:     userID,
		Source:     src,
		ExternalID: extID,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "source"}, {Name: "external_id"}},
		DoNothing: true,
	}).Create(&entry).Error
}

func loadDeletedImportIDs(db *gorm.DB, userID uint, source string, meetings []ImportedMeeting) (map[string]bool, error) {
	deleted := make(map[string]bool)
	if db == nil || len(meetings) == 0 {
		return deleted, nil
	}
	src := strings.ToLower(strings.TrimSpace(source))
	if src == "" || src == SourceLocal {
		return deleted, nil
	}

	extIDs := make([]string, 0, len(meetings))
	seen := make(map[string]bool, len(meetings))
	for _, m := range meetings {
		extID := strings.TrimSpace(m.ExternalID)
		if extID == "" || seen[extID] {
			continue
		}
		seen[extID] = true
		extIDs = append(extIDs, extID)
	}
	if len(extIDs) == 0 {
		return deleted, nil
	}

	var dismissed []string
	if err := db.Model(&models.DeletedImport{}).
		Where("user_id = ? AND source = ? AND external_id IN ?", userID, src, extIDs).
		Pluck("external_id", &dismissed).Error; err != nil {
		return nil, err
	}
	for _, id := range dismissed {
		deleted[id] = true
	}
	return deleted, nil
}

func filterOutDeletedMeetings(meetings []ImportedMeeting, deletedIDs map[string]bool) ([]ImportedMeeting, int) {
	if len(deletedIDs) == 0 {
		return meetings, 0
	}
	kept := make([]ImportedMeeting, 0, len(meetings))
	skipped := 0
	for _, m := range meetings {
		extID := strings.TrimSpace(m.ExternalID)
		if extID != "" && deletedIDs[extID] {
			skipped++
			continue
		}
		kept = append(kept, m)
	}
	return kept, skipped
}

func upsertImportedMeetings(db *gorm.DB, userID uint, source string, meetings []ImportedMeeting) (imported, updated, skipped int, err error) {
	deletedIDs, err := loadDeletedImportIDs(db, userID, source, meetings)
	if err != nil {
		return 0, 0, 0, err
	}
	meetings, skipped = filterOutDeletedMeetings(meetings, deletedIDs)

	for _, m := range meetings {
		extID := strings.TrimSpace(m.ExternalID)
		if extID == "" {
			continue
		}

		var existing models.Recording
		findErr := db.Where("user_id = ? AND source = ? AND external_id = ?", userID, source, extID).First(&existing).Error
		if findErr == gorm.ErrRecordNotFound {
			title := strings.TrimSpace(m.Title)
			if title == "" {
				title = fmt.Sprintf("%s Meeting", providerDisplayName(source))
			}

			status := "completed"
			if strings.TrimSpace(m.SummaryMD) != "" {
				status = "summarized"
			}

			var extURLPtr *string
			if u := strings.TrimSpace(m.ExternalURL); u != "" {
				extURLPtr = &u
			}

			createdAt := m.CreatedAt
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}

			rec := models.Recording{
				ID:          uuid.New().String(),
				UserID:      userID,
				Title:       title,
				Filename:    fmt.Sprintf("%s_%s", source, extID),
				AudioPath:   "",
				Status:      status,
				SummaryMD:   strings.TrimSpace(m.SummaryMD),
				Duration:    m.Duration,
				Source:      source,
				ExternalID:  &extID,
				ExternalURL: extURLPtr,
				CreatedAt:   createdAt,
			}

			txErr := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&rec).Error; err != nil {
					return err
				}
				if len(m.Segments) > 0 {
					segs := make([]models.TranscriptSegment, 0, len(m.Segments))
					for _, s := range m.Segments {
						txt := strings.TrimSpace(s.Text)
						if txt == "" {
							continue
						}
						spk := strings.TrimSpace(s.Speaker)
						if spk == "" {
							spk = "Speaker"
						}
						segs = append(segs, models.TranscriptSegment{
							RecordingID: rec.ID,
							StartTime:   s.StartTime,
							EndTime:     s.EndTime,
							Speaker:     spk,
							Text:        txt,
						})
					}
					if len(segs) > 0 {
						if err := tx.CreateInBatches(&segs, 200).Error; err != nil {
							return err
						}
					}
				}
				return nil
			})
			if txErr != nil {
				return imported, updated, skipped, txErr
			}
			imported++
			continue
		} else if findErr != nil {
			return imported, updated, skipped, findErr
		}

		// Record already exists — backfill missing summary, URL, duration, or segments if absent.
		changed := false
		if strings.TrimSpace(existing.SummaryMD) == "" && strings.TrimSpace(m.SummaryMD) != "" {
			existing.SummaryMD = strings.TrimSpace(m.SummaryMD)
			existing.Status = "summarized"
			changed = true
		}
		if (existing.ExternalURL == nil || *existing.ExternalURL == "") && strings.TrimSpace(m.ExternalURL) != "" {
			u := strings.TrimSpace(m.ExternalURL)
			existing.ExternalURL = &u
			changed = true
		}
		if existing.Duration == nil && m.Duration != nil {
			existing.Duration = m.Duration
			changed = true
		}

		var segCount int64
		db.Model(&models.TranscriptSegment{}).Where("recording_id = ?", existing.ID).Count(&segCount)
		if segCount == 0 && len(m.Segments) > 0 {
			segs := make([]models.TranscriptSegment, 0, len(m.Segments))
			for _, s := range m.Segments {
				txt := strings.TrimSpace(s.Text)
				if txt == "" {
					continue
				}
				spk := strings.TrimSpace(s.Speaker)
				if spk == "" {
					spk = "Speaker"
				}
				segs = append(segs, models.TranscriptSegment{
					RecordingID: existing.ID,
					StartTime:   s.StartTime,
					EndTime:     s.EndTime,
					Speaker:     spk,
					Text:        txt,
				})
			}
			if len(segs) > 0 {
				if err := db.CreateInBatches(&segs, 200).Error; err != nil {
					return imported, updated, skipped, err
				}
				changed = true
			}
		}

		if changed {
			if err := db.Save(&existing).Error; err != nil {
				return imported, updated, skipped, err
			}
			updated++
		} else {
			skipped++
		}
	}
	return imported, updated, skipped, nil
}

func providerDisplayName(source string) string {
	switch source {
	case ProviderFathom:
		return "Fathom"
	case ProviderFireflies:
		return "Fireflies.ai"
	case ProviderOtter:
		return "Otter.ai"
	default:
		return "ClarifAi"
	}
}

// ============================================================================
// Fathom REST API Client (https://api.fathom.ai/external/v1)
// ============================================================================

type fathomMeetingsResponse struct {
	Limit      int                 `json:"limit"`
	NextCursor string              `json:"next_cursor"`
	Items      []fathomMeetingItem `json:"items"`
}

type fathomMeetingItem struct {
	RecordingID        json.RawMessage `json:"recording_id"`
	ID                 json.RawMessage `json:"id"`
	Title              string          `json:"title"`
	MeetingTitle       string          `json:"meeting_title"`
	URL                string          `json:"url"`
	ShareURL           string          `json:"share_url"`
	CreatedAt          string          `json:"created_at"`
	ScheduledStartTime string          `json:"scheduled_start_time"`
	ScheduledEndTime   string          `json:"scheduled_end_time"`
	RecordingStartTime string          `json:"recording_start_time"`
	RecordingEndTime   string          `json:"recording_end_time"`
	DefaultSummary     *struct {
		TemplateName      string `json:"template_name"`
		MarkdownFormatted string `json:"markdown_formatted"`
	} `json:"default_summary"`
	ActionItems []fathomActionItem     `json:"action_items"`
	Transcript  []fathomTranscriptTurn `json:"transcript"`
}

type fathomActionItem struct {
	Description string `json:"description"`
	Text        string `json:"text"`
	Completed   bool   `json:"completed"`
	Timestamp   string `json:"timestamp"`
	Assignee    *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"assignee"`
}

type fathomTranscriptTurn struct {
	Speaker *struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
		Email       string `json:"matched_calendar_invitee_email"`
	} `json:"speaker"`
	SpeakerName string      `json:"speaker_name"`
	Text        string      `json:"text"`
	Timestamp   interface{} `json:"timestamp"`
	StartTime   interface{} `json:"start_time"`
	EndTime     interface{} `json:"end_time"`
}

type fathomRecordingSummaryResponse struct {
	Summary           string             `json:"summary"`
	MarkdownFormatted string             `json:"markdown_formatted"`
	ActionItems       []fathomActionItem `json:"action_items"`
	DefaultSummary    *struct {
		MarkdownFormatted string `json:"markdown_formatted"`
	} `json:"default_summary"`
}

type fathomRecordingTranscriptResponse struct {
	Segments   []fathomTranscriptTurn `json:"segments"`
	Transcript []fathomTranscriptTurn `json:"transcript"`
}

func fetchFathomAPI(ctx context.Context, baseURL, apiKey string) ([]ImportedMeeting, error) {
	var allMeetings []ImportedMeeting
	cursor := ""
	const maxPages = 5

	for page := 0; page < maxPages; page++ {
		u, err := url.Parse(baseURL + "/meetings")
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("include_summary", "true")
		q.Set("include_transcript", "true")
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		u.RawQuery = q.Encode()

		var respData fathomMeetingsResponse
		if err := doJSONGet(ctx, u.String(), map[string]string{"X-Api-Key": apiKey}, &respData); err != nil {
			return nil, fmt.Errorf("Fathom API error: %w", err)
		}

		for _, item := range respData.Items {
			recID := parseRawID(item.RecordingID)
			if recID == "" {
				recID = parseRawID(item.ID)
			}
			if recID == "" {
				continue
			}

			title := strings.TrimSpace(item.MeetingTitle)
			if title == "" {
				title = strings.TrimSpace(item.Title)
			}
			if title == "" {
				title = "Fathom Meeting"
			}

			extURL := strings.TrimSpace(item.URL)
			if extURL == "" {
				extURL = strings.TrimSpace(item.ShareURL)
			}

			createdAt := parseAnyTimestamp(item.RecordingStartTime, item.CreatedAt, item.ScheduledStartTime)

			summaryMD := ""
			if item.DefaultSummary != nil {
				summaryMD = strings.TrimSpace(item.DefaultSummary.MarkdownFormatted)
			}
			actionItems := item.ActionItems

			// Fallback to /recordings/{id}/summary if inline summary was not populated
			if summaryMD == "" {
				var sumResp fathomRecordingSummaryResponse
				sumURL := fmt.Sprintf("%s/recordings/%s/summary", baseURL, url.PathEscape(recID))
				if err := doJSONGet(ctx, sumURL, map[string]string{"X-Api-Key": apiKey}, &sumResp); err == nil {
					if sumResp.MarkdownFormatted != "" {
						summaryMD = strings.TrimSpace(sumResp.MarkdownFormatted)
					} else if sumResp.DefaultSummary != nil && sumResp.DefaultSummary.MarkdownFormatted != "" {
						summaryMD = strings.TrimSpace(sumResp.DefaultSummary.MarkdownFormatted)
					} else if sumResp.Summary != "" {
						summaryMD = strings.TrimSpace(sumResp.Summary)
					}
					if len(actionItems) == 0 && len(sumResp.ActionItems) > 0 {
						actionItems = sumResp.ActionItems
					}
				}
			}

			if len(actionItems) > 0 && !strings.Contains(strings.ToLower(summaryMD), "action item") {
				summaryMD = appendFathomActionItems(summaryMD, actionItems)
			}

			turns := item.Transcript
			if len(turns) == 0 {
				var trResp fathomRecordingTranscriptResponse
				trURL := fmt.Sprintf("%s/recordings/%s/transcript", baseURL, url.PathEscape(recID))
				if err := doJSONGet(ctx, trURL, map[string]string{"X-Api-Key": apiKey}, &trResp); err == nil {
					if len(trResp.Segments) > 0 {
						turns = trResp.Segments
					} else if len(trResp.Transcript) > 0 {
						turns = trResp.Transcript
					}
				}
			}

			segments := normalizeFathomTurns(turns)
			dur := computeFathomDuration(item.RecordingStartTime, item.RecordingEndTime, item.ScheduledStartTime, item.ScheduledEndTime, segments)

			allMeetings = append(allMeetings, ImportedMeeting{
				ExternalID:  recID,
				ExternalURL: extURL,
				Title:       title,
				CreatedAt:   createdAt,
				Duration:    dur,
				SummaryMD:   summaryMD,
				Segments:    segments,
			})
		}

		if strings.TrimSpace(respData.NextCursor) == "" || len(respData.Items) == 0 {
			break
		}
		cursor = respData.NextCursor
	}

	return allMeetings, nil
}

func appendFathomActionItems(summaryMD string, items []fathomActionItem) string {
	var b strings.Builder
	if summaryMD != "" {
		b.WriteString(summaryMD)
		b.WriteString("\n\n")
	}
	b.WriteString("### Action Items\n")
	for _, ai := range items {
		desc := strings.TrimSpace(ai.Description)
		if desc == "" {
			desc = strings.TrimSpace(ai.Text)
		}
		if desc == "" {
			continue
		}
		check := "[ ]"
		if ai.Completed {
			check = "[x]"
		}
		assignee := ""
		if ai.Assignee != nil && strings.TrimSpace(ai.Assignee.Name) != "" {
			assignee = fmt.Sprintf(" (*%s*)", strings.TrimSpace(ai.Assignee.Name))
		}
		b.WriteString(fmt.Sprintf("- %s %s%s\n", check, desc, assignee))
	}
	return strings.TrimSpace(b.String())
}

func normalizeFathomTurns(turns []fathomTranscriptTurn) []ImportedSegment {
	segments := make([]ImportedSegment, 0, len(turns))
	for _, t := range turns {
		txt := strings.TrimSpace(t.Text)
		if txt == "" {
			continue
		}
		spk := strings.TrimSpace(t.SpeakerName)
		if spk == "" && t.Speaker != nil {
			if strings.TrimSpace(t.Speaker.DisplayName) != "" {
				spk = strings.TrimSpace(t.Speaker.DisplayName)
			} else if strings.TrimSpace(t.Speaker.Name) != "" {
				spk = strings.TrimSpace(t.Speaker.Name)
			}
		}
		if spk == "" {
			spk = "Speaker"
		}

		startSec := parseFlexibleSeconds(t.StartTime)
		if startSec == 0 && t.Timestamp != nil {
			startSec = parseFlexibleSeconds(t.Timestamp)
		}
		endSec := parseFlexibleSeconds(t.EndTime)

		segments = append(segments, ImportedSegment{
			StartTime: startSec,
			EndTime:   endSec,
			Speaker:   spk,
			Text:      txt,
		})
	}
	fillMissingEndTimes(segments)
	return segments
}

func computeFathomDuration(recStart, recEnd, schedStart, schedEnd string, segments []ImportedSegment) *float64 {
	if s := parseAnyTimestamp(recStart); !s.IsZero() {
		if e := parseAnyTimestamp(recEnd); !e.IsZero() && e.After(s) {
			d := e.Sub(s).Seconds()
			return &d
		}
	}
	if s := parseAnyTimestamp(schedStart); !s.IsZero() {
		if e := parseAnyTimestamp(schedEnd); !e.IsZero() && e.After(s) {
			d := e.Sub(s).Seconds()
			return &d
		}
	}
	return durationFromSegments(segments)
}

// ============================================================================
// Fireflies.ai GraphQL API Client (https://api.fireflies.ai/graphql)
// ============================================================================

type firefliesGraphQLResponse struct {
	Data struct {
		Transcripts []firefliesTranscript `json:"transcripts"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type firefliesTranscript struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Date          json.RawMessage `json:"date"`
	Duration      float64         `json:"duration"`
	TranscriptURL string          `json:"transcript_url"`
	Summary       *struct {
		Overview        string   `json:"overview"`
		ShortSummary    string   `json:"short_summary"`
		Gist            string   `json:"gist"`
		BulletGist      string   `json:"bullet_gist"`
		ActionItems     string   `json:"action_items"`
		Outline         string   `json:"outline"`
		Keywords        []string `json:"keywords"`
		TopicsDiscussed []string `json:"topics_discussed"`
	} `json:"summary"`
	Sentences []struct {
		Index       int     `json:"index"`
		Text        string  `json:"text"`
		RawText     string  `json:"raw_text"`
		StartTime   float64 `json:"start_time"`
		EndTime     float64 `json:"end_time"`
		SpeakerID   json.RawMessage `json:"speaker_id"`
		SpeakerName string  `json:"speaker_name"`
	} `json:"sentences"`
	Speakers []struct {
		ID   json.RawMessage `json:"id"`
		Name string          `json:"name"`
	} `json:"speakers"`
}

func fetchFirefliesAPI(ctx context.Context, graphqlURL, apiKey string) ([]ImportedMeeting, error) {
	const query = `query ListTranscripts($limit: Int, $skip: Int) {
  transcripts(limit: $limit, skip: $skip) {
    id
    title
    date
    duration
    transcript_url
    summary {
      overview
      short_summary
      gist
      bullet_gist
      action_items
      outline
      keywords
      topics_discussed
    }
    sentences {
      index
      text
      raw_text
      start_time
      end_time
      speaker_id
      speaker_name
    }
    speakers {
      id
      name
    }
  }
}`

	var allMeetings []ImportedMeeting
	const pageSize = 25
	const maxPages = 2

	for page := 0; page < maxPages; page++ {
		reqBody := map[string]interface{}{
			"query": query,
			"variables": map[string]interface{}{
				"limit": pageSize,
				"skip":  page * pageSize,
			},
		}
		payload, _ := json.Marshal(reqBody)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlURL, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := integrationHTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Fireflies API request failed: %w", err)
		}
		rawBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024))
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("Fireflies authentication failed (HTTP %d): invalid API key", resp.StatusCode)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("Fireflies API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
		}

		var gqlResp firefliesGraphQLResponse
		if err := json.Unmarshal(rawBody, &gqlResp); err != nil {
			return nil, fmt.Errorf("failed to parse Fireflies GraphQL response: %w", err)
		}
		if len(gqlResp.Errors) > 0 {
			return nil, fmt.Errorf("Fireflies GraphQL error: %s", gqlResp.Errors[0].Message)
		}

		items := gqlResp.Data.Transcripts
		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" {
				continue
			}

			speakerMap := make(map[string]string, len(item.Speakers))
			for _, sp := range item.Speakers {
				sid := parseRawID(sp.ID)
				if sid != "" && strings.TrimSpace(sp.Name) != "" {
					speakerMap[sid] = strings.TrimSpace(sp.Name)
				}
			}

			segments := make([]ImportedSegment, 0, len(item.Sentences))
			maxEnd := 0.0
			for _, s := range item.Sentences {
				txt := strings.TrimSpace(s.Text)
				if txt == "" {
					txt = strings.TrimSpace(s.RawText)
				}
				if txt == "" {
					continue
				}
				spk := strings.TrimSpace(s.SpeakerName)
				if spk == "" {
					sid := parseRawID(s.SpeakerID)
					if mapped, ok := speakerMap[sid]; ok {
						spk = mapped
					} else if sid != "" {
						spk = "Speaker " + sid
					} else {
						spk = "Speaker"
					}
				}
				if s.EndTime > maxEnd {
					maxEnd = s.EndTime
				}
				segments = append(segments, ImportedSegment{
					StartTime: s.StartTime,
					EndTime:   s.EndTime,
					Speaker:   spk,
					Text:      txt,
				})
			}
			fillMissingEndTimes(segments)

			var durPtr *float64
			if item.Duration > 0 {
				d := item.Duration
				// If duration appears to be in minutes compared to sentence end_time in seconds, convert to seconds
				if maxEnd > 0 && math.Abs(d*60-maxEnd) < math.Abs(d-maxEnd) {
					d = d * 60
				}
				durPtr = &d
			} else {
				durPtr = durationFromSegments(segments)
			}

			summaryMD := formatFirefliesSummary(item.Summary)
			createdAt := parseFirefliesDate(item.Date)

			allMeetings = append(allMeetings, ImportedMeeting{
				ExternalID:  strings.TrimSpace(item.ID),
				ExternalURL: strings.TrimSpace(item.TranscriptURL),
				Title:       strings.TrimSpace(item.Title),
				CreatedAt:   createdAt,
				Duration:    durPtr,
				SummaryMD:   summaryMD,
				Segments:    segments,
			})
		}

		if len(items) < pageSize {
			break
		}
	}

	return allMeetings, nil
}

func formatFirefliesSummary(s *struct {
	Overview        string   `json:"overview"`
	ShortSummary    string   `json:"short_summary"`
	Gist            string   `json:"gist"`
	BulletGist      string   `json:"bullet_gist"`
	ActionItems     string   `json:"action_items"`
	Outline         string   `json:"outline"`
	Keywords        []string `json:"keywords"`
	TopicsDiscussed []string `json:"topics_discussed"`
}) string {
	if s == nil {
		return ""
	}
	var sections []string

	overview := strings.TrimSpace(s.Overview)
	if overview == "" {
		overview = strings.TrimSpace(s.ShortSummary)
	}
	if overview == "" {
		overview = strings.TrimSpace(s.Gist)
	}
	if overview != "" {
		sections = append(sections, "## Overview\n"+overview)
	}
	if bg := strings.TrimSpace(s.BulletGist); bg != "" {
		sections = append(sections, "## Key Takeaways\n"+bg)
	}
	if out := strings.TrimSpace(s.Outline); out != "" {
		sections = append(sections, "## Outline\n"+out)
	}
	if ai := strings.TrimSpace(s.ActionItems); ai != "" {
		sections = append(sections, "## Action Items\n"+ai)
	}
	if len(s.TopicsDiscussed) > 0 {
		var tb strings.Builder
		tb.WriteString("## Topics Discussed\n")
		for _, t := range s.TopicsDiscussed {
			if strings.TrimSpace(t) != "" {
				tb.WriteString("- " + strings.TrimSpace(t) + "\n")
			}
		}
		sections = append(sections, strings.TrimSpace(tb.String()))
	}
	if len(s.Keywords) > 0 {
		sections = append(sections, "**Keywords:** "+strings.Join(s.Keywords, ", "))
	}

	return strings.TrimSpace(strings.Join(sections, "\n\n"))
}

// ============================================================================
// Otter.ai Public REST API Client (https://api.otter.ai)
// ============================================================================

type otterConversation struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	URL             string  `json:"url"`
	CreatedAt       string  `json:"created_at"`
	Duration        float64 `json:"duration"`
	AbstractSummary string  `json:"abstract_summary"`
	Summary         string  `json:"summary"`
	Outline         []struct {
		Topic     string `json:"topic"`
		Title     string `json:"title"`
		Timestamp string `json:"timestamp"`
	} `json:"outline"`
	ActionItems []struct {
		Task     string `json:"task"`
		Text     string `json:"text"`
		Assignee string `json:"assignee"`
	} `json:"action_items"`
	Transcript []struct {
		Speaker *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"speaker"`
		SpeakerName string      `json:"speaker_name"`
		StartTime   interface{} `json:"start_time"`
		EndTime     interface{} `json:"end_time"`
		Text        string      `json:"text"`
	} `json:"transcript"`
}

func fetchOtterAPI(ctx context.Context, baseURL, apiKey string) ([]ImportedMeeting, error) {
	var allMeetings []ImportedMeeting
	cursor := ""
	const maxPages = 5

	for page := 0; page < maxPages; page++ {
		u, err := url.Parse(baseURL + "/conversations")
		if err != nil {
			return nil, err
		}
		if cursor != "" {
			q := u.Query()
			q.Set("cursor", cursor)
			u.RawQuery = q.Encode()
		}

		var rawEnvelope struct {
			Meta *struct {
				HasMore    bool   `json:"has_more"`
				NextCursor string `json:"next_cursor"`
			} `json:"meta"`
			NextCursor    string              `json:"next_cursor"`
			Data          []otterConversation `json:"data"`
			Results       []otterConversation `json:"results"`
			Conversations []otterConversation `json:"conversations"`
			Speeches      []otterConversation `json:"speeches"`
		}

		if err := doJSONGet(ctx, u.String(), map[string]string{"Authorization": "Bearer " + apiKey}, &rawEnvelope); err != nil {
			return nil, fmt.Errorf("Otter.ai API error: %w", err)
		}

		items := rawEnvelope.Data
		if len(items) == 0 {
			items = rawEnvelope.Conversations
		}
		if len(items) == 0 {
			items = rawEnvelope.Results
		}
		if len(items) == 0 {
			items = rawEnvelope.Speeches
		}

		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" {
				continue
			}

			// Fetch full conversation detail if transcript or summary is absent
			if len(item.Transcript) == 0 || (strings.TrimSpace(item.AbstractSummary) == "" && strings.TrimSpace(item.Summary) == "") {
				detailURL := fmt.Sprintf("%s/conversations/%s", baseURL, url.PathEscape(item.ID))
				var detailRaw json.RawMessage
				if err := doJSONGet(ctx, detailURL, map[string]string{"Authorization": "Bearer " + apiKey}, &detailRaw); err == nil {
					var wrapped struct {
						Data *otterConversation `json:"data"`
					}
					if err := json.Unmarshal(detailRaw, &wrapped); err == nil && wrapped.Data != nil && wrapped.Data.ID != "" {
						item = *wrapped.Data
					} else {
						var direct otterConversation
						if err := json.Unmarshal(detailRaw, &direct); err == nil && direct.ID != "" {
							item = direct
						}
					}
				}
			}

			segments := make([]ImportedSegment, 0, len(item.Transcript))
			for _, seg := range item.Transcript {
				txt := strings.TrimSpace(seg.Text)
				if txt == "" {
					continue
				}
				spk := strings.TrimSpace(seg.SpeakerName)
				if spk == "" && seg.Speaker != nil {
					spk = strings.TrimSpace(seg.Speaker.Name)
				}
				if spk == "" {
					spk = "Speaker"
				}
				segments = append(segments, ImportedSegment{
					StartTime: parseFlexibleSeconds(seg.StartTime),
					EndTime:   parseFlexibleSeconds(seg.EndTime),
					Speaker:   spk,
					Text:      txt,
				})
			}
			fillMissingEndTimes(segments)

			var durPtr *float64
			if item.Duration > 0 {
				d := item.Duration
				durPtr = &d
			} else {
				durPtr = durationFromSegments(segments)
			}

			allMeetings = append(allMeetings, ImportedMeeting{
				ExternalID:  strings.TrimSpace(item.ID),
				ExternalURL: strings.TrimSpace(item.URL),
				Title:       strings.TrimSpace(item.Title),
				CreatedAt:   parseAnyTimestamp(item.CreatedAt),
				Duration:    durPtr,
				SummaryMD:   formatOtterSummary(item),
				Segments:    segments,
			})
		}

		next := ""
		if rawEnvelope.Meta != nil && rawEnvelope.Meta.NextCursor != "" {
			next = rawEnvelope.Meta.NextCursor
		} else if rawEnvelope.NextCursor != "" {
			next = rawEnvelope.NextCursor
		}
		if next == "" || len(items) == 0 {
			break
		}
		cursor = next
	}

	return allMeetings, nil
}

func formatOtterSummary(c otterConversation) string {
	var sections []string

	overview := strings.TrimSpace(c.AbstractSummary)
	if overview == "" {
		overview = strings.TrimSpace(c.Summary)
	}
	if overview != "" {
		sections = append(sections, "## Overview\n"+overview)
	}

	if len(c.Outline) > 0 {
		var ob strings.Builder
		ob.WriteString("## Outline\n")
		for _, o := range c.Outline {
			topic := strings.TrimSpace(o.Topic)
			if topic == "" {
				topic = strings.TrimSpace(o.Title)
			}
			if topic == "" {
				continue
			}
			if o.Timestamp != "" {
				ob.WriteString(fmt.Sprintf("- **%s** — %s\n", o.Timestamp, topic))
			} else {
				ob.WriteString(fmt.Sprintf("- %s\n", topic))
			}
		}
		sections = append(sections, strings.TrimSpace(ob.String()))
	}

	if len(c.ActionItems) > 0 {
		var ab strings.Builder
		ab.WriteString("## Action Items\n")
		for _, a := range c.ActionItems {
			task := strings.TrimSpace(a.Task)
			if task == "" {
				task = strings.TrimSpace(a.Text)
			}
			if task == "" {
				continue
			}
			if strings.TrimSpace(a.Assignee) != "" {
				ab.WriteString(fmt.Sprintf("- [ ] %s (*%s*)\n", task, strings.TrimSpace(a.Assignee)))
			} else {
				ab.WriteString(fmt.Sprintf("- [ ] %s\n", task))
			}
		}
		sections = append(sections, strings.TrimSpace(ab.String()))
	}

	return strings.TrimSpace(strings.Join(sections, "\n\n"))
}

// ============================================================================
// MCP Integration Fetcher (Fathom MCP, Fireflies MCP, Otter MCP)
// ============================================================================

func buildProviderMCPClient(provider, mcpURL, token string) *MCPClient {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
		if provider == ProviderFathom {
			headers["X-Api-Key"] = token
		}
	}
	return NewMCPClient(mcpURL, headers)
}

func fetchViaMCP(ctx context.Context, provider, mcpURL, token string) ([]ImportedMeeting, error) {
	client := buildProviderMCPClient(provider, mcpURL, token)
	if err := client.Initialize(ctx); err != nil {
		return nil, err
	}
	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("MCP tools/list failed: %w", err)
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("MCP server exposed no tools")
	}

	listTool := pickMCPTool(tools,
		[]string{
			"list_meetings", "get_meetings", "search_meetings",
			"fireflies_get_transcripts", "get_transcripts", "list_transcripts", "search_transcripts",
			"list_conversations", "get_conversations", "list_speeches", "get_speeches",
		},
		[]string{"meeting", "transcript", "conversation", "speech"},
		[]string{"list", "get", "search", "fetch"},
	)
	if listTool == nil {
		return nil, fmt.Errorf("no meeting/transcript listing tool found on MCP server")
	}

	listArgs := buildMCPListArgs(listTool)
	rawListText, err := client.CallTool(ctx, listTool.Name, listArgs)
	if err != nil {
		return nil, err
	}

	meetings := parseMCPMeetingsPayload(rawListText)

	transcriptTool := pickMCPTool(tools,
		[]string{
			"get_transcript", "get_meeting_transcript", "fireflies_get_transcript",
			"get_conversation", "get_speech", "fetch_transcript",
		},
		[]string{"transcript", "conversation", "speech"},
		[]string{"get", "fetch", "read"},
	)
	summaryTool := pickMCPTool(tools,
		[]string{
			"get_summary", "get_meeting_summary", "fireflies_get_summary", "fetch_summary",
		},
		[]string{"summary"},
		[]string{"get", "fetch", "read"},
	)

	for i := range meetings {
		m := &meetings[i]
		if len(m.Segments) == 0 && transcriptTool != nil && transcriptTool.Name != listTool.Name {
			args := buildMCPIDArgs(transcriptTool, m.ExternalID)
			if detailText, err := client.CallTool(ctx, transcriptTool.Name, args); err == nil {
				parsed := parseMCPMeetingsPayload(detailText)
				if len(parsed) > 0 {
					if len(parsed[0].Segments) > 0 {
						m.Segments = parsed[0].Segments
					}
					if m.SummaryMD == "" && parsed[0].SummaryMD != "" {
						m.SummaryMD = parsed[0].SummaryMD
					}
					if m.Duration == nil && parsed[0].Duration != nil {
						m.Duration = parsed[0].Duration
					}
				} else if segs := parseMCPTranscriptSegmentsText(detailText); len(segs) > 0 {
					m.Segments = segs
				}
			}
		}

		if m.SummaryMD == "" && summaryTool != nil && summaryTool.Name != listTool.Name {
			args := buildMCPIDArgs(summaryTool, m.ExternalID)
			if sumText, err := client.CallTool(ctx, summaryTool.Name, args); err == nil {
				m.SummaryMD = extractMCPSummaryText(sumText)
			}
		}

		if m.Duration == nil {
			m.Duration = durationFromSegments(m.Segments)
		}
	}

	return meetings, nil
}

func pickMCPTool(tools []MCPTool, exactNames, nounHints, verbHints []string) *MCPTool {
	byLower := make(map[string]*MCPTool, len(tools))
	for i := range tools {
		byLower[strings.ToLower(tools[i].Name)] = &tools[i]
	}
	for _, exact := range exactNames {
		if t, ok := byLower[strings.ToLower(exact)]; ok {
			return t
		}
	}
	for i := range tools {
		name := strings.ToLower(tools[i].Name)
		hasNoun := false
		for _, n := range nounHints {
			if strings.Contains(name, n) {
				hasNoun = true
				break
			}
		}
		hasVerb := false
		for _, v := range verbHints {
			if strings.Contains(name, v) {
				hasVerb = true
				break
			}
		}
		if hasNoun && hasVerb {
			return &tools[i]
		}
	}
	return nil
}

func buildMCPListArgs(tool *MCPTool) map[string]interface{} {
	args := map[string]interface{}{}
	props, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		return args
	}
	if _, has := props["limit"]; has {
		args["limit"] = 25
	}
	if _, has := props["include_summary"]; has {
		args["include_summary"] = true
	}
	if _, has := props["include_transcript"]; has {
		args["include_transcript"] = true
	}
	return args
}

func buildMCPIDArgs(tool *MCPTool, externalID string) map[string]interface{} {
	args := map[string]interface{}{}
	if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
		for _, candidate := range []string{"id", "transcript_id", "transcriptId", "recording_id", "recordingId", "meeting_id", "meetingId", "conversation_id", "conversationId", "speech_id"} {
			if _, has := props[candidate]; has {
				args[candidate] = externalID
				return args
			}
		}
	}
	args["id"] = externalID
	return args
}

func parseMCPMeetingsPayload(rawText string) []ImportedMeeting {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return nil
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}

	var rawItems []map[string]interface{}
	switch v := parsed.(type) {
	case []interface{}:
		for _, elem := range v {
			if m, ok := elem.(map[string]interface{}); ok {
				rawItems = append(rawItems, m)
			}
		}
	case map[string]interface{}:
		for _, key := range []string{"items", "meetings", "transcripts", "conversations", "speeches", "data", "results"} {
			if arr, ok := v[key].([]interface{}); ok {
				for _, elem := range arr {
					if m, ok := elem.(map[string]interface{}); ok {
						rawItems = append(rawItems, m)
					}
				}
				break
			}
		}
		if len(rawItems) == 0 {
			// Single meeting object
			if _, hasID := v["id"]; hasID || v["recording_id"] != nil {
				rawItems = append(rawItems, v)
			} else if d, ok := v["data"].(map[string]interface{}); ok {
				rawItems = append(rawItems, d)
			}
		}
	}

	meetings := make([]ImportedMeeting, 0, len(rawItems))
	for _, item := range rawItems {
		extID := firstNonEmptyString(item, "recording_id", "id", "transcript_id", "conversation_id")
		if extID == "" {
			continue
		}
		title := firstNonEmptyString(item, "meeting_title", "title", "name")
		extURL := firstNonEmptyString(item, "url", "share_url", "transcript_url")
		createdAt := parseAnyTimestamp(
			firstNonEmptyString(item, "created_at", "recording_start_time", "scheduled_start_time", "date"),
		)

		summaryMD := extractMCPSummaryFromMap(item)
		segments := extractMCPSegmentsFromMap(item)
		fillMissingEndTimes(segments)

		var durPtr *float64
		if dVal, ok := item["duration"]; ok {
			d := parseFlexibleSeconds(dVal)
			if d > 0 {
				durPtr = &d
			}
		}
		if durPtr == nil {
			durPtr = durationFromSegments(segments)
		}

		meetings = append(meetings, ImportedMeeting{
			ExternalID:  extID,
			ExternalURL: extURL,
			Title:       title,
			CreatedAt:   createdAt,
			Duration:    durPtr,
			SummaryMD:   summaryMD,
			Segments:    segments,
		})
	}

	return meetings
}

func extractMCPSummaryFromMap(item map[string]interface{}) string {
	for _, key := range []string{"markdown_formatted", "abstract_summary"} {
		if s, ok := item[key].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	for _, key := range []string{"default_summary", "summary"} {
		switch v := item[key].(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case map[string]interface{}:
			var parts []string
			for _, subKey := range []string{"markdown_formatted", "overview", "short_summary", "gist", "bullet_gist", "outline", "action_items"} {
				if sub, ok := v[subKey].(string); ok && strings.TrimSpace(sub) != "" {
					parts = append(parts, strings.TrimSpace(sub))
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, "\n\n")
			}
		}
	}
	return ""
}

func extractMCPSummaryText(rawText string) string {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return ""
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
		if s := extractMCPSummaryFromMap(obj); s != "" {
			return s
		}
	}
	return trimmed
}

func extractMCPSegmentsFromMap(item map[string]interface{}) []ImportedSegment {
	for _, key := range []string{"transcript", "sentences", "segments"} {
		if arr, ok := item[key].([]interface{}); ok {
			return parseSegmentSlice(arr)
		}
	}
	return nil
}

func parseMCPTranscriptSegmentsText(rawText string) []ImportedSegment {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}
	switch v := parsed.(type) {
	case []interface{}:
		return parseSegmentSlice(v)
	case map[string]interface{}:
		return extractMCPSegmentsFromMap(v)
	}
	return nil
}

func parseSegmentSlice(arr []interface{}) []ImportedSegment {
	segs := make([]ImportedSegment, 0, len(arr))
	for _, elem := range arr {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		txt := firstNonEmptyString(m, "text", "raw_text", "content")
		if txt == "" {
			continue
		}
		spk := firstNonEmptyString(m, "speaker_name", "speakerName")
		if spk == "" {
			if spObj, ok := m["speaker"].(map[string]interface{}); ok {
				spk = firstNonEmptyString(spObj, "display_name", "name")
			} else if spStr, ok := m["speaker"].(string); ok {
				spk = strings.TrimSpace(spStr)
			}
		}
		if spk == "" {
			spk = "Speaker"
		}
		startSec := parseFlexibleSeconds(m["start_time"])
		if startSec == 0 && m["timestamp"] != nil {
			startSec = parseFlexibleSeconds(m["timestamp"])
		}
		endSec := parseFlexibleSeconds(m["end_time"])

		segs = append(segs, ImportedSegment{
			StartTime: startSec,
			EndTime:   endSec,
			Speaker:   spk,
			Text:      txt,
		})
	}
	fillMissingEndTimes(segs)
	return segs
}

// ============================================================================
// Shared HTTP & Parsing Utilities
// ============================================================================

func doJSONGet(ctx context.Context, targetURL string, headers map[string]string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := integrationHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024))
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed (HTTP %d): invalid or expired API key", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
	}

	return json.Unmarshal(body, out)
}

func parseRawID(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return strings.Trim(trimmed, `"`)
}

func firstNonEmptyString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		val, exists := m[k]
		if !exists || val == nil {
			continue
		}
		switch v := val.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			if v == float64(int64(v)) {
				return strconv.FormatInt(int64(v), 10)
			}
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int64:
			return strconv.FormatInt(v, 10)
		case int:
			return strconv.Itoa(v)
		}
	}
	return ""
}

func parseAnyTimestamp(candidates ...string) time.Time {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if ms, err := strconv.ParseInt(c, 10, 64); err == nil {
			if ms > 1e12 {
				return time.UnixMilli(ms).UTC()
			}
			return time.Unix(ms, 0).UTC()
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, c); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}

func parseFirefliesDate(raw json.RawMessage) time.Time {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return time.Time{}
	}
	var num float64
	if err := json.Unmarshal(raw, &num); err == nil && num > 0 {
		if num > 1e12 {
			return time.UnixMilli(int64(num)).UTC()
		}
		return time.Unix(int64(num), 0).UTC()
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return parseAnyTimestamp(s)
	}
	return time.Time{}
}

func parseFlexibleSeconds(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			h, err1 := strconv.ParseFloat(parts[0], 64)
			m, err2 := strconv.ParseFloat(parts[1], 64)
			sec, err3 := strconv.ParseFloat(parts[2], 64)
			if err1 == nil && err2 == nil && err3 == nil {
				return h*3600 + m*60 + sec
			}
		} else if len(parts) == 2 {
			m, err1 := strconv.ParseFloat(parts[0], 64)
			sec, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return m*60 + sec
			}
		}
	}
	return 0
}

func fillMissingEndTimes(segments []ImportedSegment) {
	for i := range segments {
		if segments[i].EndTime <= segments[i].StartTime {
			if i+1 < len(segments) && segments[i+1].StartTime > segments[i].StartTime {
				segments[i].EndTime = segments[i+1].StartTime
			} else {
				segments[i].EndTime = segments[i].StartTime + 4.0
			}
		}
	}
}

func durationFromSegments(segments []ImportedSegment) *float64 {
	if len(segments) == 0 {
		return nil
	}
	maxEnd := 0.0
	for _, s := range segments {
		if s.EndTime > maxEnd {
			maxEnd = s.EndTime
		} else if s.StartTime > maxEnd {
			maxEnd = s.StartTime
		}
	}
	if maxEnd <= 0 {
		return nil
	}
	return &maxEnd
}
