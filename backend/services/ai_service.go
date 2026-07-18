package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"note-taker/backend/models"
)

const OllamaModel = "gemma4:12b-mlx"

func getOllamaURL(path string) string {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = os.Getenv("OLLAMA_HOST")
	}
	if url == "" {
		url = "http://localhost:11434"
	}
	url = strings.TrimSuffix(url, "/")
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

func getModelName() string {
	db := models.GetDB()
	if db != nil {
		var settings models.Settings
		if err := db.First(&settings).Error; err == nil && settings.OllamaModel != "" {
			return settings.OllamaModel
		}
	}
	return OllamaModel
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

func GenerateCustomSummary(transcript string, prompt string) string {
	systemPrompt := fmt.Sprintf("%s\n\nFormat the output in clean Markdown.", prompt)
	fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript)

	reqPayload := OllamaRequest{
		Model:  getModelName(),
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Sprintf("> [!ERROR]\n> Could not marshal request. Error: %v", err)
	}

	resp, err := http.Post(getOllamaURL("/api/generate"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("> [!ERROR]\n> Could not generate summary. Please check if Ollama is running.\n\nError details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("> [!ERROR]\n> Ollama returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("> [!ERROR]\n> Failed to read Ollama response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return fmt.Sprintf("> [!ERROR]\n> Failed to parse Ollama response: %v", err)
	}

	if ollamaResp.Response == "" {
		return "Error: No response generated."
	}

	return ollamaResp.Response
}

func GenerateSummaryAndTitle(transcript string) string {
	return GenerateCustomSummary(transcript, "Provide a concise executive summary and a list of action items based on this transcript.")
}

func GenerateTitle(transcript string) string {
	systemPrompt := "Based on the following transcript, generate a concise, professional, and descriptive title for this meeting or conversation. The title should be 3 to 7 words long. Do not include quotes, markdown formatting, or the word 'Title:' in your response. Return ONLY the title text."
	fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript)

	reqPayload := OllamaRequest{
		Model:  getModelName(),
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "Meeting Note"
	}

	resp, err := http.Post(getOllamaURL("/api/generate"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "Meeting Note"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "Meeting Note"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Meeting Note"
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "Meeting Note"
	}

	title := strings.TrimSpace(ollamaResp.Response)
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
		return "Meeting Note"
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

	reqPayload := OllamaRequest{
		Model:  getModelName(),
		Prompt: fmt.Sprintf("%s\n\nTranscript:\n%s", systemPrompt, transcript),
		Stream: false,
	}

	result := make(map[string]string)

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return result
	}

	resp, err := http.Post(getOllamaURL("/api/generate"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return result
	}

	text := strings.TrimSpace(ollamaResp.Response)
	
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
