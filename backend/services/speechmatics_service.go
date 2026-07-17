package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SpeechmaticsJobResponse struct {
	ID string `json:"id"`
}

type SpeechmaticsStatusResponse struct {
	Job struct {
		Status string `json:"status"`
	} `json:"job"`
}

func TranscribeWithSpeechmatics(filePath string, apiKey string, updateProgress func(string, int)) (string, error) {
	url := "https://asr.api.speechmatics.com/v2/jobs"

	// 1. Prepare configuration
	config := map[string]interface{}{
		"type": "transcription",
		"transcription_config": map[string]interface{}{
			"language":    "en",
			"diarization": "speaker",
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// 2. Prepare multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Config field
	configWriter, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="config"`},
		"Content-Type":        {`application/json`},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create config part: %w", err)
	}
	if _, err := configWriter.Write(configJSON); err != nil {
		return "", fmt.Errorf("failed to write config part: %w", err)
	}

	// File field
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileWriter, err := writer.CreateFormFile("data_file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create file part: %w", err)
	}
	if _, err := io.Copy(fileWriter, file); err != nil {
		return "", fmt.Errorf("failed to write file part: %w", err)
	}

	writer.Close()

	updateProgress("uploading to Speechmatics", 25)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to Speechmatics failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to submit job (status %d): %s", resp.StatusCode, string(respBody))
	}

	var jobResponse SpeechmaticsJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResponse); err != nil {
		return "", fmt.Errorf("failed to decode job response: %w", err)
	}

	jobID := jobResponse.ID
	fmt.Printf("Speechmatics job submitted successfully. Job ID: %s\n", jobID)

	// 3. Polling loop
	pollURL := fmt.Sprintf("%s/%s", url, jobID)
	maxRetries := 60
	retryCount := 0

	for retryCount < maxRetries {
		time.Sleep(10 * time.Second)
		retryCount++

		pollReq, err := http.NewRequest("GET", pollURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create poll request: %w", err)
		}
		pollReq.Header.Set("Authorization", "Bearer "+apiKey)

		pollResp, err := client.Do(pollReq)
		if err != nil {
			continue // retry on transport errors
		}
		defer pollResp.Body.Close()

		if pollResp.StatusCode != http.StatusOK {
			pollBody, _ := io.ReadAll(pollResp.Body)
			return "", fmt.Errorf("failed to fetch status (status %d): %s", pollResp.StatusCode, string(pollBody))
		}

		var statusResponse SpeechmaticsStatusResponse
		if err := json.NewDecoder(pollResp.Body).Decode(&statusResponse); err != nil {
			return "", fmt.Errorf("failed to decode status response: %w", err)
		}

		status := statusResponse.Job.Status
		pollProgress := 30 + int((float64(retryCount)/float64(maxRetries))*45.0)

		if status == "done" {
			break
		} else if status == "rejected" {
			return "", fmt.Errorf("Speechmatics job was rejected")
		} else {
			if updateProgress != nil {
				updateProgress(fmt.Sprintf("Speechmatics transcribing (%s)", status), minInt(pollProgress, 75))
			}
		}
	}

	if retryCount >= maxRetries {
		return "", fmt.Errorf("timed out waiting for Speechmatics job completion")
	}

	// 4. Fetch results
	if updateProgress != nil {
		updateProgress("fetching Speechmatics transcript", 75)
	}

	transcriptURL := fmt.Sprintf("%s/transcript?format=json-v2", pollURL)
	transcriptReq, err := http.NewRequest("GET", transcriptURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create transcript request: %w", err)
	}
	transcriptReq.Header.Set("Authorization", "Bearer "+apiKey)

	transcriptResp, err := client.Do(transcriptReq)
	if err != nil {
		return "", fmt.Errorf("failed to fetch transcript: %w", err)
	}
	defer transcriptResp.Body.Close()

	if transcriptResp.StatusCode != http.StatusOK {
		transcriptBody, _ := io.ReadAll(transcriptResp.Body)
		return "", fmt.Errorf("failed to retrieve transcript (status %d): %s", transcriptResp.StatusCode, string(transcriptBody))
	}

	var transcriptJSON map[string]interface{}
	if err := json.NewDecoder(transcriptResp.Body).Decode(&transcriptJSON); err != nil {
		return "", fmt.Errorf("failed to decode transcript JSON: %w", err)
	}

	if updateProgress != nil {
		updateProgress("formatting transcript", 78)
	}

	return FormatSpeechmaticsJSON(transcriptJSON), nil
}

func FormatSpeechmaticsJSON(data map[string]interface{}) string {
	resultsVal, ok := data["results"]
	if !ok {
		return "No transcription results found."
	}

	results, ok := resultsVal.([]interface{})
	if !ok || len(results) == 0 {
		return "No transcription results found."
	}

	var lines []string
	var currentSpeaker *string
	var currentText []string
	var segmentStart float64
	var segmentEnd float64

	for _, rawItem := range results {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}

		tokenType, _ := item["type"].(string)
		alternatives, _ := item["alternatives"].([]interface{})
		if len(alternatives) == 0 {
			continue
		}

		alt, ok := alternatives[0].(map[string]interface{})
		if !ok {
			continue
		}

		content, _ := alt["content"].(string)
		speaker, _ := alt["speaker"].(string)

		if content == "" {
			continue
		}

		startTime, _ := item["start_time"].(float64)
		endTimeVal, hasEndTime := item["end_time"]
		var endTime float64
		if hasEndTime {
			endTime, _ = endTimeVal.(float64)
		} else {
			duration, _ := item["duration"].(float64)
			endTime = startTime + duration
		}

		if tokenType == "word" {
			spk := "Unknown"
			if speaker != "" {
				spk = speaker
			}

			if currentSpeaker == nil {
				currentSpeaker = &spk
				segmentStart = startTime
				segmentEnd = endTime
				currentText = []string{content}
			} else if *currentSpeaker == spk {
				currentText = append(currentText, content)
				segmentEnd = endTime
			} else {
				// Flush segment
				textStr := strings.Join(currentText, " ")
				lines = append(lines, fmt.Sprintf("**Speaker %s** (%.2fs - %.2fs): %s", *currentSpeaker, segmentStart, segmentEnd, textStr))
				currentSpeaker = &spk
				segmentStart = startTime
				segmentEnd = endTime
				currentText = []string{content}
			}
		} else if tokenType == "punctuation" {
			if len(currentText) > 0 {
				if content == "." || content == "," || content == "!" || content == "?" || content == ";" || content == ":" {
					currentText[len(currentText)-1] = currentText[len(currentText)-1] + content
				} else {
					currentText = append(currentText, content)
				}
				segmentEnd = endTime
			} else {
				currentText = []string{content}
				segmentStart = startTime
				segmentEnd = endTime
			}
		}
	}

	if len(currentText) > 0 && currentSpeaker != nil {
		textStr := strings.Join(currentText, " ")
		lines = append(lines, fmt.Sprintf("**Speaker %s** (%.2fs - %.2fs): %s", *currentSpeaker, segmentStart, segmentEnd, textStr))
	}

	return strings.Join(lines, "\n\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
