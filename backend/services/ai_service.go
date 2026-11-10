package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"note-taker/backend/models"
)

const OllamaModel = "gemma4:12b-mlx"
const DefaultGeminiModel = "gemini-3.8-flash"

var GeminiModels = []string{
	"gemini-4-argon",
	"gemini-3.8-flash",
	"gemini-3.5-flash-lite",
}

func GetGeminiModels() []string {
	return GeminiModels
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiCandidate struct {
	Content struct {
		Parts []GeminiPart `json:"parts"`
		Role  string       `json:"role"`
	} `json:"content"`
	FinishReason string `json:"finishReason"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func getAISettings() (provider, ollamaURL, ollamaModel, geminiModel, geminiKey, geminiBase string) {
	provider = "ollama"
	ollamaURL = "http://localhost:11434"
	ollamaModel = OllamaModel
	geminiModel = DefaultGeminiModel
	geminiKey = os.Getenv("GEMINI_API_KEY")
	geminiBase = os.Getenv("GEMINI_API_BASE")

	if envOllama := os.Getenv("OLLAMA_URL"); envOllama != "" {
		ollamaURL = envOllama
	} else if envOllamaHost := os.Getenv("OLLAMA_HOST"); envOllamaHost != "" {
		ollamaURL = envOllamaHost
	}

	db := models.GetDB()
	if db != nil {
		var settings models.Settings
		if err := db.First(&settings).Error; err == nil {
			if settings.AIProvider != "" {
				provider = settings.AIProvider
			}
			if settings.OllamaURL != "" {
				ollamaURL = settings.OllamaURL
			}
			if settings.OllamaModel != "" {
				ollamaModel = settings.OllamaModel
			}
			if settings.GeminiModel != "" {
				geminiModel = settings.GeminiModel
			}
			if settings.GeminiAPIKey != nil && *settings.GeminiAPIKey != "" {
				geminiKey = *settings.GeminiAPIKey
			}
			if settings.GeminiAPIBase != "" {
				geminiBase = settings.GeminiAPIBase
			}
		}
	}
	return
}

func getOllamaURL(path string) string {
	_, ollamaURL, _, _, _, _ := getAISettings()
	url := strings.TrimSuffix(ollamaURL, "/")
	if strings.Contains(url, "/api/") {
		if strings.HasSuffix(url, "/api/generate") && path == "/api/tags" {
			return strings.Replace(url, "/api/generate", "/api/tags", 1)
		}
		if strings.HasSuffix(url, "/api/tags") && path == "/api/generate" {
			return strings.Replace(url, "/api/tags", "/api/generate", 1)
		}
		return url
	}
	return url + path
}

func getModelName() string {
	provider, _, ollamaModel, geminiModel, _, _ := getAISettings()
	if strings.EqualFold(provider, "gemini") {
		return geminiModel
	}
	return ollamaModel
}

func GetOllamaModels() ([]string, error) {
	resp, err := http.Get(getOllamaURL("/api/tags"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama returned status code %d", resp.StatusCode)
	}

	var tagsResp OllamaTagsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &tagsResp); err != nil {
		return nil, err
	}

	var modelsList []string
	for _, m := range tagsResp.Models {
		modelsList = append(modelsList, m.Name)
	}

	// If no models found, return a default/fallback list
	if len(modelsList) == 0 {
		modelsList = []string{OllamaModel}
	}

	return modelsList, nil
}

func callGemini(model string, apiKey string, customBase string, prompt string) (string, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("Gemini API key is not configured. Please enter your Gemini API key in Settings")
	}
	if model == "" {
		model = DefaultGeminiModel
	}

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	if customBase != "" {
		apiURL = fmt.Sprintf("%s/models/%s:generateContent", strings.TrimSuffix(customBase, "/"), model)
	}

	reqPayload := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("could not marshal Gemini request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not create Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not connect to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp GeminiResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error != nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("Gemini API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if geminiResp.Error != nil && geminiResp.Error.Message != "" {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("Gemini returned no response candidates")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
			return "", fmt.Errorf("Gemini generation stopped (%s)", candidate.FinishReason)
		}
		return "", fmt.Errorf("Gemini returned empty candidate content")
	}

	var sb strings.Builder
	for _, part := range candidate.Content.Parts {
		sb.WriteString(part.Text)
	}

	res := strings.TrimSpace(sb.String())
	if res == "" {
		return "", fmt.Errorf("Gemini returned empty response text")
	}

	return res, nil
}

func callOllama(model string, prompt string) (string, error) {
	if model == "" {
		model = OllamaModel
	}
	reqPayload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("could not marshal Ollama request: %w", err)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Post(getOllamaURL("/api/generate"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not connect to Ollama. Please check if Ollama is running.\n\nError details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Ollama response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %v", err)
	}

	if strings.TrimSpace(ollamaResp.Response) == "" {
		return "", fmt.Errorf("Ollama returned empty response")
	}

	return ollamaResp.Response, nil
}

func generateText(prompt string) (string, error) {
	provider, _, ollamaModel, geminiModel, geminiKey, geminiBase := getAISettings()
	if strings.EqualFold(provider, "gemini") {
		return callGemini(geminiModel, geminiKey, geminiBase, prompt)
	}
	return callOllama(ollamaModel, prompt)
}

func GenerateCustomSummary(transcript string, prompt string) string {
	systemPrompt := fmt.Sprintf("%s\n\nFormat the output in clean Markdown.", prompt)
	fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript)

	resp, err := generateText(fullPrompt)
	if err != nil {
		provider, _, _, _, _, _ := getAISettings()
		providerName := "Ollama"
		if strings.EqualFold(provider, "gemini") {
			providerName = "Google Gemini"
		}
		return fmt.Sprintf("> [!ERROR]\n> Could not generate summary via %s.\n\nError details: %v", providerName, err)
	}

	if strings.TrimSpace(resp) == "" {
		return "Error: No response generated."
	}

	return resp
}

func GenerateSummaryAndTitle(transcript string) string {
	return GenerateCustomSummary(transcript, "Provide a concise executive summary and a list of action items based on this transcript.")
}

func GenerateTitle(transcript string) string {
	return GenerateTitleForRecording(transcript, false)
}

func GenerateTitleForRecording(transcript string, isMedical bool) string {
	systemPrompt := "Based on the following transcript, generate a concise, professional, and descriptive title for this meeting or conversation. The title should be 3 to 7 words long. Do not include quotes, markdown formatting, or the word 'Title:' in your response. Return ONLY the title text."
	fallback := "Meeting Note"
	if isMedical {
		systemPrompt = "Based on the following medical conversation transcript, generate a concise, professional, and descriptive clinical title for this consultation or medical encounter (e.g. 'Cardiology Follow-up: Treatment Review', 'Pediatric Well-Child Consultation', 'Clinical Evaluation: Acute Bronchitis'). The title should be 3 to 7 words long. Do not include quotes, markdown formatting, or the word 'Title:' in your response. Return ONLY the title text."
		fallback = "Clinical Consultation"
	}
	fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript)

	resp, err := generateText(fullPrompt)
	if err != nil {
		return fallback
	}

	title := strings.TrimSpace(resp)
	title = strings.ReplaceAll(title, "\"", "")
	title = strings.ReplaceAll(title, "'", "")
	title = strings.TrimSpace(title)

	if strings.HasSuffix(title, ".") {
		title = title[:len(title)-1]
	}

	if len(title) > 80 {
		title = title[:80] + "..."
	}

	if title == "" {
		return fallback
	}

	return title
}

func DetectSpeakers(transcript string, speakerLabels []string) map[string]string {
	labelsStr := strings.Join(speakerLabels, ", ")
	systemPrompt := fmt.Sprintf(
		"You are an assistant that analyzes transcripts and identifies the names of speakers based on the context of the conversation.\n"+
			"Here are the speaker labels currently in the transcript: [%s].\n"+
			"For each label, identify the speaker's actual name using clues from the dialog (e.g. self-identification, someone addressing them, context of the meeting).\n"+
			"Return a clean JSON object mapping the original speaker label to the identified name (e.g. {\"Speaker 1\": \"Alice Smith\"}).\n"+
			"Ensure the returned JSON maps only the given labels: [%s]. If you cannot identify a speaker, map them to a more descriptive name or keep the original label.\n"+
			"Return ONLY the JSON object. Do not include any formatting, markdown backticks, explanations, or other text.",
		labelsStr, labelsStr,
	)

	result := make(map[string]string)
	fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript)

	resp, err := generateText(fullPrompt)
	if err != nil {
		return result
	}

	text := strings.TrimSpace(resp)

	// Strip markdown code block if present
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			if strings.HasPrefix(lines[0], "```json") || strings.HasPrefix(lines[0], "```") {
				text = strings.Join(lines[1:len(lines)-1], "\n")
			}
		}
	}
	text = strings.TrimSpace(text)

	_ = json.Unmarshal([]byte(text), &result)
	return result
}
