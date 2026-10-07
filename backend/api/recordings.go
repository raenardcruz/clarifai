package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"note-taker/backend/models"
	"note-taker/backend/services"

	"gorm.io/gorm"
)

const DataDir = "data"

func init() {
	os.MkdirAll(DataDir, os.ModePerm)
	go cleanupOrphanedTempChunks(2 * time.Hour)
}

func cleanupOrphanedTempChunks(maxAge time.Duration) {
	tempBase := filepath.Join(DataDir, "temp_chunks")
	entries, err := os.ReadDir(tempBase)
	if err != nil {
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(tempBase, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if now.Sub(info.ModTime()) > maxAge {
				os.RemoveAll(dirPath)
			}
		}
	}
}

type SegmentUpdate struct {
	Speaker string `json:"speaker" binding:"required"`
	Text    string `json:"text" binding:"required"`
}

type RecordingOut struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	Duration  *float64 `json:"duration"`
	IsMedical bool     `json:"is_medical"`
}

func SetupRecordingsRoutes(router *gin.RouterGroup) {
	// Public route (sharing) does not require auth
	router.GET("/share/:recording_id", getSharedRecording)

	// Auth group
	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(), RequireApprovedUser())
	{
		authGroup.POST("", uploadAudio)
		authGroup.POST("/upload/init", initChunkedUpload)
		authGroup.POST("/upload/chunk", uploadChunk)
		authGroup.POST("/upload/complete", completeChunkedUpload)
		authGroup.GET("", getRecordings)
		authGroup.GET("/recent", getRecentRecordings)
		authGroup.GET("/speechmatics-usage", getSpeechmaticsUsage)
		authGroup.GET("/:recording_id", getRecordingDetails)
		authGroup.PATCH("/:recording_id", updateRecording)
		authGroup.PUT("/:recording_id/segments/:segment_id", updateSegment)
		authGroup.POST("/:recording_id/summarize", triggerSummary)
		authGroup.POST("/:recording_id/detect-speakers", detectRecordingSpeakers)
		authGroup.DELETE("/:recording_id", deleteRecording)
	}
}


func getContextUser(c *gin.Context) (*models.User, error) {
	userVal, exists := c.Get("current_user")
	if !exists {
		return nil, errors.New("user not found in context")
	}
	return userVal.(*models.User), nil
}

func uploadAudio(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	var settings models.Settings
	db.First(&settings)

	if settings.SpeechmaticsAPIKey == nil || *settings.SpeechmaticsAPIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Speechmatics API key is not configured. Please set it in Settings."})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "File is required"})
		return
	}

	jobID := uuid.New().String()
	filePath := filepath.Join(DataDir, fmt.Sprintf("%s_%s", jobID, file.Filename))

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not save file"})
		return
	}

	isMedical := c.PostForm("is_medical") == "true" || c.PostForm("is_medical") == "1"

	recording := models.Recording{
		ID:        jobID,
		UserID:    currentUser.ID,
		Title:     file.Filename,
		Filename:  file.Filename,
		AudioPath: filePath,
		Status:    "pending",
		IsMedical: isMedical,
	}

	if err := db.Create(&recording).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not save recording"})
		return
	}

	go services.ProcessSpeechmaticsJob(jobID, filePath)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "job_id": jobID})
}

type InitUploadRequest struct {
	Filename    string `json:"filename" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
	FileSize    int64  `json:"file_size"`
}

type CompleteUploadRequest struct {
	UploadID    string `json:"upload_id" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
	IsMedical   bool   `json:"is_medical"`
}


func initChunkedUpload(c *gin.Context) {
	_, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	var settings models.Settings
	db.First(&settings)

	if settings.SpeechmaticsAPIKey == nil || *settings.SpeechmaticsAPIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Speechmatics API key is not configured. Please set it in Settings."})
		return
	}

	var req InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request parameters"})
		return
	}

	if req.TotalChunks <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Total chunks must be greater than 0"})
		return
	}

	uploadID := uuid.New().String()
	tempDir := filepath.Join(DataDir, "temp_chunks", uploadID)
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not create chunk session"})
		return
	}

	go cleanupOrphanedTempChunks(2 * time.Hour)

	c.JSON(http.StatusOK, gin.H{
		"upload_id": uploadID,
	})
}

func uploadChunk(c *gin.Context) {
	_, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uploadID := c.PostForm("upload_id")
	chunkIndexStr := c.PostForm("chunk_index")
	if uploadID == "" || chunkIndexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "upload_id and chunk_index are required"})
		return
	}

	if _, err := uuid.Parse(uploadID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid upload_id format"})
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid chunk_index"})
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Chunk file is required"})
		return
	}

	tempDir := filepath.Join(DataDir, "temp_chunks", uploadID)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Upload session not found or expired"})
		return
	}

	chunkFilePath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := c.SaveUploadedFile(file, chunkFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not save chunk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Chunk uploaded successfully",
		"chunk_index": chunkIndex,
	})
}

func completeChunkedUpload(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request parameters"})
		return
	}

	if _, err := uuid.Parse(req.UploadID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid upload_id format"})
		return
	}

	tempDir := filepath.Join(DataDir, "temp_chunks", req.UploadID)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Upload session not found"})
		return
	}

	// Ensure all chunk files exist
	for i := 0; i < req.TotalChunks; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": fmt.Sprintf("Missing chunk %d", i)})
			return
		}
	}

	// Stitch chunks together
	jobID := uuid.New().String()
	cleanFilename := filepath.Base(req.Filename)
	if cleanFilename == "." || cleanFilename == "/" {
		cleanFilename = "audio_recording"
	}
	filePath := filepath.Join(DataDir, fmt.Sprintf("%s_%s", jobID, cleanFilename))

	destFile, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not create merged file"})
		return
	}

	for i := 0; i < req.TotalChunks; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			destFile.Close()
			os.Remove(filePath)
			c.JSON(http.StatusInternalServerError, gin.H{"detail": fmt.Sprintf("Could not read chunk %d", i)})
			return
		}
		_, err = io.Copy(destFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			destFile.Close()
			os.Remove(filePath)
			c.JSON(http.StatusInternalServerError, gin.H{"detail": fmt.Sprintf("Could not append chunk %d", i)})
			return
		}
	}
	destFile.Close()

	// Clean up temp directory
	os.RemoveAll(tempDir)

	db := models.GetDB()
	recording := models.Recording{
		ID:        jobID,
		UserID:    currentUser.ID,
		Title:     cleanFilename,
		Filename:  cleanFilename,
		AudioPath: filePath,
		Status:    "pending",
		IsMedical: req.IsMedical,
	}

	if err := db.Create(&recording).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not save recording record"})
		return
	}

	go services.ProcessSpeechmaticsJob(jobID, filePath)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded and stitched successfully", "job_id": jobID})
}

func getRecordings(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	isMedicalFilter := c.Query("is_medical")
	db := models.GetDB()
	query := db.Model(&models.Recording{}).Where("user_id = ?", currentUser.ID)

	if isMedicalFilter == "true" {
		query = query.Where("recordings.is_medical = ?", true)
	} else if isMedicalFilter == "false" {
		query = query.Where("recordings.is_medical = ?", false)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN transcript_segments ON transcript_segments.recording_id = recordings.id").
			Where("recordings.title ILIKE ? OR transcript_segments.text ILIKE ? OR transcript_segments.speaker ILIKE ?", searchPattern, searchPattern, searchPattern).
			Distinct()
	}

	var recordings []models.Recording
	query.Order("created_at desc").Find(&recordings)

	response := []RecordingOut{}
	for _, r := range recordings {
		response = append(response, RecordingOut{
			ID:        r.ID,
			Title:     r.Title,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			Duration:  r.Duration,
			IsMedical: r.IsMedical,
		})
	}

	c.JSON(http.StatusOK, response)
}

func getRecentRecordings(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()

	var recordings []models.Recording
	db.Where("user_id = ?", currentUser.ID).Order("created_at desc").Limit(5).Find(&recordings)

	var totalRecs int64
	db.Model(&models.Recording{}).Where("user_id = ?", currentUser.ID).Count(&totalRecs)

	var allRecordings []models.Recording
	db.Where("user_id = ?", currentUser.ID).Find(&allRecordings)

	var totalHours float64
	for _, r := range allRecordings {
		if r.Duration != nil {
			totalHours += *r.Duration
		}
	}
	totalHours = totalHours / 3600.0

	recResponse := []RecordingOut{}
	for _, r := range recordings {
		recResponse = append(recResponse, RecordingOut{
			ID:        r.ID,
			Title:     r.Title,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			Duration:  r.Duration,
			IsMedical: r.IsMedical,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recResponse,
		"statistics": gin.H{
			"total_recordings":   totalRecs,
			"total_hours":        math.Round(totalHours*100) / 100,
			"insights_generated": totalRecs * 3, // mock
		},
	})

}

func getSpeechmaticsUsage(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	var settings models.Settings
	db.First(&settings)

	apiKey := ""
	if settings.SpeechmaticsAPIKey != nil {
		apiKey = *settings.SpeechmaticsAPIKey
	}

	limitHours := currentUser.TranscriptionLimitHours
	limitSeconds := limitHours * 3600.0

	if apiKey != "" {
		now := time.Now()
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

		req, err := http.NewRequest("GET", "https://asr.api.speechmatics.com/v2/usage", nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			q := req.URL.Query()
			q.Add("since", firstOfMonth)
			req.URL.RawQuery = q.Encode()

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)

				var data map[string]interface{}
				if err := json.Unmarshal(body, &data); err == nil {
					if summary, ok := data["summary"].(map[string]interface{}); ok {
						if durationHrs, ok := summary["duration_hrs"].(float64); ok {
							usedHours := durationHrs
							totalSeconds := usedHours * 3600.0
							percentage := math.Min(math.Round((totalSeconds/limitSeconds)*1000)/10, 100.0)
							c.JSON(http.StatusOK, gin.H{
								"used_seconds":  totalSeconds,
								"limit_seconds": limitSeconds,
								"used_hours":    math.Round(usedHours*100) / 100,
								"limit_hours":   limitHours,
								"percentage":    percentage,
							})
							return
						}
					}
				}
			}
		}
	}

	// Fallback to local DB calculation
	var allRecordings []models.Recording
	db.Where("user_id = ?", currentUser.ID).Find(&allRecordings)
	var totalSeconds float64
	for _, r := range allRecordings {
		if r.Duration != nil {
			totalSeconds += *r.Duration
		}
	}

	percentage := math.Min(math.Round((totalSeconds/limitSeconds)*1000)/10, 100.0)
	c.JSON(http.StatusOK, gin.H{
		"used_seconds":  totalSeconds,
		"limit_seconds": limitSeconds,
		"used_hours":    math.Round((totalSeconds/3600.0)*100) / 100,
		"limit_hours":   limitHours,
		"percentage":    percentage,
	})
}

func getRecordingDetails(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
		return db.Order("transcript_segments.start_time ASC")
	}).Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	segments := []gin.H{}
	for _, s := range recording.Segments {
		segments = append(segments, gin.H{
			"id":      s.ID,
			"start":   s.StartTime,
			"end":     s.EndTime,
			"speaker": s.Speaker,
			"text":    s.Text,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         recording.ID,
		"title":      recording.Title,
		"status":     recording.Status,
		"created_at": recording.CreatedAt.Format(time.RFC3339),
		"duration":   recording.Duration,
		"summary_md": recording.SummaryMD,
		"is_medical": recording.IsMedical,
		"segments":   segments,
	})
}

func getSharedRecording(c *gin.Context) {
	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ?", recordingID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         recording.ID,
		"title":      recording.Title,
		"status":     recording.Status,
		"created_at": recording.CreatedAt.Format(time.RFC3339),
		"duration":   recording.Duration,
		"summary_md": recording.SummaryMD,
		"is_medical": recording.IsMedical,
	})
}

type UpdateRecordingRequest struct {
	Title *string `json:"title"`
}

func updateRecording(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	var req UpdateRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		recording.Title = strings.TrimSpace(*req.Title)
	}

	db.Save(&recording)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Recording updated successfully",
		"id":         recording.ID,
		"is_medical": recording.IsMedical,
		"title":      recording.Title,
	})
}

func updateSegment(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	segmentIDStr := c.Param("segment_id")
	segmentID, err := strconv.Atoi(segmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid segment ID"})
		return
	}

	var input SegmentUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := models.GetDB()
	var recording models.Recording
	if err := db.Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	var segment models.TranscriptSegment
	if err := db.Where("id = ? AND recording_id = ?", segmentID, recordingID).First(&segment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Segment not found"})
		return
	}

	segment.Speaker = input.Speaker
	segment.Text = input.Text
	db.Save(&segment)

	c.JSON(http.StatusOK, gin.H{"message": "Segment updated successfully"})
}

type SummarizeRequest struct {
	Instruction string `json:"instruction"`
}

func triggerSummary(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	if recording.Status != "completed" && recording.Status != "summarized" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Transcription must be completed first"})
		return
	}

	var req SummarizeRequest
	_ = c.ShouldBindJSON(&req)

	go services.GenerateSummary(recordingID, req.Instruction)

	c.JSON(http.StatusOK, gin.H{"message": "Summarization triggered"})
}


func detectRecordingSpeakers(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	var segments []models.TranscriptSegment
	db.Where("recording_id = ?", recordingID).Order("start_time asc").Find(&segments)
	if len(segments) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No segments to analyze", "mapping": gin.H{}})
		return
	}

	speakerSet := make(map[string]bool)
	for _, s := range segments {
		if s.Speaker != "" {
			speakerSet[s.Speaker] = true
		}
	}
	speakerLabels := []string{}
	for s := range speakerSet {
		speakerLabels = append(speakerLabels, s)
	}

	var fullTextBuilder strings.Builder
	for _, s := range segments {
		fullTextBuilder.WriteString(fmt.Sprintf("%s: %s\n\n", s.Speaker, s.Text))
	}
	fullText := fullTextBuilder.String()

	mapping := services.DetectSpeakers(fullText, speakerLabels)

	updated := make(map[string]string)
	for oldName, newName := range mapping {
		oldNameClean := strings.TrimSpace(oldName)
		newNameClean := strings.TrimSpace(newName)
		if oldNameClean != "" && newNameClean != "" && oldNameClean != newNameClean {
			db.Model(&models.TranscriptSegment{}).
				Where("recording_id = ? AND speaker = ?", recordingID, oldNameClean).
				Update("speaker", newNameClean)
			updated[oldNameClean] = newNameClean
		}
	}

	c.JSON(http.StatusOK, gin.H{"mapping": updated})
}

func deleteRecording(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ? AND user_id = ?", recordingID, currentUser.ID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	db.Where("recording_id = ?", recordingID).Delete(&models.TranscriptSegment{})

	if recording.AudioPath != "" {
		if _, err := os.Stat(recording.AudioPath); err == nil {
			os.Remove(recording.AudioPath)
		}
	}

	db.Delete(&recording)

	c.JSON(http.StatusOK, gin.H{"message": "Recording deleted successfully"})
}
