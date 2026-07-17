package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"note-taker/backend/models"
)

func ProcessSpeechmaticsJob(jobID string, filePath string) {
	db := models.GetDB()
	var recording models.Recording
	if err := db.Where("id = ?", jobID).First(&recording).Error; err != nil {
		log.Printf("ProcessSpeechmaticsJob error: recording %s not found", jobID)
		cleanupFile(filePath)
		return
	}

	var settings models.Settings
	db.First(&settings)

	// Update job status in jobs DB
	models.GlobalJobsDB.Set(jobID, models.JobStatus{
		ID:       jobID,
		Filename: recording.Filename,
		Status:   "transcribing",
		Progress: 10,
	})

	recording.Status = "transcribing"
	db.Save(&recording)

	// Determine duration using ffprobe
	duration := 180.0
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		durStr := strings.TrimSpace(out.String())
		if dur, err := strconv.ParseFloat(durStr, 64); err == nil {
			duration = dur
		}
	}
	recording.Duration = &duration
	db.Save(&recording)

	apiKey := ""
	if settings.SpeechmaticsAPIKey != nil {
		apiKey = *settings.SpeechmaticsAPIKey
	}

	if apiKey != "" {
		// Call Speechmatics API directly to parse results and save them
		err := runSpeechmaticsPipeline(jobID, filePath, apiKey, &recording, db)
		if err != nil {
			failJob(jobID, &recording, db, err.Error())
			cleanupFile(filePath)
			return
		}
	} else {
		// Mock response for demonstration
		time.Sleep(5 * time.Second)
		mockSegments := []struct {
			Start   float64
			End     float64
			Speaker string
			Text    string
		}{
			{Start: 0.0, End: 5.5, Speaker: "Alex Rivera", Text: "Alright everyone, let's get started with the Q4 roadmap."},
			{Start: 6.0, End: 12.0, Speaker: "Sarah Jenkins", Text: "I reviewed the API integration requirements, we might need more time."},
			{Start: 12.5, End: 18.2, Speaker: "Mark Kim", Text: "I can confirm developer availability for the November Sprint."},
		}

		for _, seg := range mockSegments {
			dbSeg := models.TranscriptSegment{
				RecordingID: recording.ID,
				StartTime:   seg.Start,
				EndTime:     seg.End,
				Speaker:     seg.Speaker,
				Text:        seg.Text,
			}
			db.Create(&dbSeg)
		}
	}

	// Generate summary if auto
	if settings.AISummarizationMode == "auto" {
		models.GlobalJobsDB.Set(jobID, models.JobStatus{
			ID:       jobID,
			Filename: recording.Filename,
			Status:   "summarizing",
			Progress: 85,
		})

		recording.Status = "summarizing"
		db.Save(&recording)

		GenerateSummarySync(recording.ID, db, settings.ExecutiveSummaryPrompt)
	} else {
		recording.Status = "completed"
		db.Save(&recording)

		models.GlobalJobsDB.Set(jobID, models.JobStatus{
			ID:       jobID,
			Filename: recording.Filename,
			Status:   "completed",
			Progress: 100,
		})
	}

	cleanupFile(filePath)
}

func runSpeechmaticsPipeline(jobID, filePath, apiKey string, recording *models.Recording, db *gorm.DB) error {
	url := "https://asr.api.speechmatics.com/v2/jobs"

	config := map[string]interface{}{
		"type": "transcription",
		"transcription_config": map[string]interface{}{
			"language":    "en",
			"diarization": "speaker",
		},
	}

	configJSON, _ := json.Marshal(config)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	configWriter, _ := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="config"`},
		"Content-Type":        {`application/json`},
	})
	configWriter.Write(configJSON)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileWriter, _ := writer.CreateFormFile("data_file", filepath.Base(filePath))
	io.Copy(fileWriter, file)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit job (status %d): %s", resp.StatusCode, string(respBody))
	}

	var jobResponse SpeechmaticsJobResponse
	json.NewDecoder(resp.Body).Decode(&jobResponse)

	speechmaticsJobID := jobResponse.ID
	recording.SpeechmaticsJobID = &speechmaticsJobID
	db.Save(recording)

	pollURL := fmt.Sprintf("%s/%s", url, speechmaticsJobID)
	maxRetries := 60
	retryCount := 0

	for retryCount < maxRetries {
		time.Sleep(10 * time.Second)
		retryCount++

		pollReq, _ := http.NewRequest("GET", pollURL, nil)
		pollReq.Header.Set("Authorization", "Bearer "+apiKey)
		pollResp, err := client.Do(pollReq)
		if err != nil {
			continue
		}
		defer pollResp.Body.Close()

		if pollResp.StatusCode != http.StatusOK {
			continue
		}

		var statusResponse SpeechmaticsStatusResponse
		json.NewDecoder(pollResp.Body).Decode(&statusResponse)

		status := statusResponse.Job.Status
		progress := 30 + int((float64(retryCount)/float64(maxRetries))*45.0)

		models.GlobalJobsDB.Set(jobID, models.JobStatus{
			ID:       jobID,
			Filename: recording.Filename,
			Status:   "transcribing",
			Progress: progress,
		})

		if status == "done" {
			break
		} else if status == "rejected" {
			return fmt.Errorf("Speechmatics job rejected")
		}
	}

	if retryCount >= maxRetries {
		return fmt.Errorf("Speechmatics job polling timed out")
	}

	transcriptURL := fmt.Sprintf("%s/transcript?format=json-v2", pollURL)
	tReq, _ := http.NewRequest("GET", transcriptURL, nil)
	tReq.Header.Set("Authorization", "Bearer "+apiKey)
	tResp, err := client.Do(tReq)
	if err != nil {
		return err
	}
	defer tResp.Body.Close()

	var transcriptJSON map[string]interface{}
	json.NewDecoder(tResp.Body).Decode(&transcriptJSON)

	resultsVal, ok := transcriptJSON["results"]
	if ok {
		if results, ok := resultsVal.([]interface{}); ok {
			parseAndSaveSegments(results, recording, db)
		}
	}

	return nil
}

func parseAndSaveSegments(results []interface{}, recording *models.Recording, db *gorm.DB) {
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
				dbSeg := models.TranscriptSegment{
					RecordingID: recording.ID,
					StartTime:   segmentStart,
					EndTime:     segmentEnd,
					Speaker:     *currentSpeaker,
					Text:        strings.Join(currentText, " "),
				}
				db.Create(&dbSeg)

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
			}
		}
	}

	if len(currentText) > 0 && currentSpeaker != nil {
		dbSeg := models.TranscriptSegment{
			RecordingID: recording.ID,
			StartTime:   segmentStart,
			EndTime:     segmentEnd,
			Speaker:     *currentSpeaker,
			Text:        strings.Join(currentText, " "),
		}
		db.Create(&dbSeg)
	}
}

func GenerateSummary(recordingID string) {
	db := models.GetDB()
	var recording models.Recording
	if err := db.Where("id = ?", recordingID).First(&recording).Error; err != nil {
		return
	}

	var settings models.Settings
	db.First(&settings)
	prompt := settings.ExecutiveSummaryPrompt
	if prompt == "" {
		prompt = "Summarize this."
	}

	models.GlobalJobsDB.Set(recordingID, models.JobStatus{
		ID:       recordingID,
		Filename: recording.Filename,
		Status:   "summarizing",
		Progress: 85,
	})

	recording.Status = "summarizing"
	db.Save(&recording)

	GenerateSummarySync(recordingID, db, prompt)
}

func GenerateSummarySync(recordingID string, db *gorm.DB, prompt string) {
	var recording models.Recording
	if err := db.Where("id = ?", recordingID).First(&recording).Error; err != nil {
		return
	}

	var segments []models.TranscriptSegment
	db.Where("recording_id = ?", recordingID).Order("start_time asc").Find(&segments)

	var fullTextBuilder strings.Builder
	for _, s := range segments {
		fullTextBuilder.WriteString(fmt.Sprintf("**%s**: %s\n\n", s.Speaker, s.Text))
	}
	fullText := fullTextBuilder.String()

	aiTitle := GenerateTitle(fullText)
	if aiTitle != "" {
		recording.Title = aiTitle
	}

	summaryMD := GenerateCustomSummary(fullText, prompt)
	recording.SummaryMD = summaryMD
	recording.Status = "completed"
	db.Save(&recording)

	models.GlobalJobsDB.Set(recordingID, models.JobStatus{
		ID:       recordingID,
		Filename: recording.Filename,
		Status:   "completed",
		Progress: 100,
	})
}

func failJob(jobID string, recording *models.Recording, db *gorm.DB, errMsg string) {
	recording.Status = "error"
	recording.ErrorMessage = &errMsg
	db.Save(recording)

	models.GlobalJobsDB.Set(jobID, models.JobStatus{
		ID:           jobID,
		Filename:     recording.Filename,
		Status:       "error",
		Progress:     100,
		ErrorMessage: errMsg,
	})
}

func cleanupFile(filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove file %s: %v", filePath, err)
		}
	}
}
