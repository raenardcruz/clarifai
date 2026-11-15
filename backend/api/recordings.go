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
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"created_at"`
	Duration    *float64     `json:"duration"`
	IsMedical   bool         `json:"is_medical"`
	Source      string       `json:"source"`
	ExternalURL *string      `json:"external_url,omitempty"`
	Tags        []TagSummary `json:"tags"`
}

func normalizeRecordingSource(source string) string {
	s := strings.TrimSpace(source)
	if s == "" {
		return services.SourceLocal
	}
	return s
}

func getSpeechmaticsKey(settings *models.Settings) string {
	if settings != nil && settings.SpeechmaticsAPIKey != nil && *settings.SpeechmaticsAPIKey != "" {
		return *settings.SpeechmaticsAPIKey
	}
	return os.Getenv("SPEECHMATICS_API_KEY")
}

func SetupRecordingsRoutes(router *gin.RouterGroup) {
	// Public routes (sharing & meeting invites) do not require auth
	router.GET("/share/:recording_id", getSharedRecording)
	router.GET("/invite/:token", getMeetingInvite)
	router.POST("/invite/:token/accept", acceptMeetingInvite)

	// Auth group
	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(), RequireApprovedUser())
	{
		authGroup.POST("", uploadAudio)
		authGroup.POST("/transcript", createRecordingFromTranscript)
		authGroup.POST("/upload/init", initChunkedUpload)
		authGroup.POST("/upload/chunk", uploadChunk)
		authGroup.POST("/upload/complete", completeChunkedUpload)
		authGroup.GET("", getRecordings)
		authGroup.GET("/recent", getRecentRecordings)
		authGroup.GET("/speechmatics-usage", getSpeechmaticsUsage)
		authGroup.POST("/:recording_id/invite", createMeetingInvite)
		authGroup.POST("/invite/:token/join", joinMeetingInvite)
		authGroup.GET("/:recording_id", getRecordingDetails)
		authGroup.PATCH("/:recording_id", updateRecording)
		authGroup.PUT("/:recording_id/transcript", replaceRecordingTranscript)
		authGroup.PUT("/:recording_id/tags", setRecordingTags)
		authGroup.POST("/bulk-tags", bulkUpdateRecordingTags)
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

func getUserUsedSeconds(db *gorm.DB, userID uint) float64 {
	var totalSeconds float64
	db.Model(&models.Recording{}).
		Where("user_id = ? AND duration IS NOT NULL AND (source IS NULL OR source = '' OR source = 'local') AND audio_path != ''", userID).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalSeconds)
	return totalSeconds
}

func isUserUsageLimitExceeded(db *gorm.DB, user *models.User) bool {
	// Usage limits apply only to non-admins.
	if user.Role == "admin" {
		return false
	}
	limitSeconds := user.TranscriptionLimitHours * 3600.0
	if limitSeconds <= 0 {
		return true
	}
	return getUserUsedSeconds(db, user.ID) >= limitSeconds
}

func canAccessRecording(db *gorm.DB, recordingID string, userID uint) bool {
	var count int64
	db.Model(&models.Recording{}).Where("id = ? AND user_id = ?", recordingID, userID).Count(&count)
	if count > 0 {
		return true
	}
	db.Model(&models.RecordingAccess{}).Where("recording_id = ? AND user_id = ?", recordingID, userID).Count(&count)
	return count > 0
}

func uploadAudio(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	if isUserUsageLimitExceeded(db, currentUser) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Transcription usage limit reached. Please contact an administrator."})
		return
	}

	var settings models.Settings
	db.First(&settings)

	if getSpeechmaticsKey(&settings) == "" {
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

	if tagIDsStr := strings.TrimSpace(c.PostForm("tag_ids")); tagIDsStr != "" {
		var tagIDs []uint
		if strings.HasPrefix(tagIDsStr, "[") {
			_ = json.Unmarshal([]byte(tagIDsStr), &tagIDs)
		} else {
			for _, part := range strings.Split(tagIDsStr, ",") {
				if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil {
					tagIDs = append(tagIDs, uint(id))
				}
			}
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			db.Where("id IN ? AND user_id = ?", tagIDs, currentUser.ID).Find(&tags)
			if len(tags) > 0 {
				_ = db.Model(&recording).Association("Tags").Replace(tags)
			}
		}
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
	TagIDs      []uint `json:"tag_ids"`
}


func initChunkedUpload(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	if isUserUsageLimitExceeded(db, currentUser) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Transcription usage limit reached. Please contact an administrator."})
		return
	}

	var settings models.Settings
	db.First(&settings)

	if getSpeechmaticsKey(&settings) == "" {
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

	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		db.Where("id IN ? AND user_id = ?", req.TagIDs, currentUser.ID).Find(&tags)
		if len(tags) > 0 {
			_ = db.Model(&recording).Association("Tags").Replace(tags)
		}
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
	sourceFilter := strings.TrimSpace(c.Query("source"))
	tagIDsParam := strings.TrimSpace(c.Query("tag_ids"))
	tagMode := c.DefaultQuery("tag_mode", "all")
	includeSubtags := c.DefaultQuery("include_subtags", "true") != "false"
	untaggedOnly := c.Query("untagged") == "true"

	db := models.GetDB()
	_, tagByID, childrenMap := loadUserTagGraph(db, currentUser.ID)

	query := db.Model(&models.Recording{}).Preload("Tags").Where(
		"recordings.user_id = ? OR EXISTS (SELECT 1 FROM recording_accesses ra WHERE ra.recording_id = recordings.id AND ra.user_id = ?)",
		currentUser.ID, currentUser.ID,
	)

	if isMedicalFilter == "true" {
		query = query.Where("recordings.is_medical = ?", true)
	} else if isMedicalFilter == "false" {
		query = query.Where("recordings.is_medical = ?", false)
	}

	if sourceFilter != "" && sourceFilter != "all" {
		if sourceFilter == "local" {
			query = query.Where("(recordings.source IS NULL OR recordings.source = '' OR recordings.source = ?)", "local")
		} else {
			query = query.Where("recordings.source = ?", sourceFilter)
		}
	}

	if untaggedOnly {
		query = query.Where("NOT EXISTS (SELECT 1 FROM recording_tags WHERE recording_tags.recording_id = recordings.id)")
	} else if tagIDsParam != "" {
		var requestedTagIDs []uint
		for _, part := range strings.Split(tagIDsParam, ",") {
			if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil {
				requestedTagIDs = append(requestedTagIDs, uint(id))
			}
		}

		if len(requestedTagIDs) > 0 {
			if tagMode == "any" {
				expandedSet := make(map[uint]bool)
				for _, tid := range requestedTagIDs {
					if includeSubtags {
						for _, dID := range getDescendantTagIDs(tid, childrenMap) {
							expandedSet[dID] = true
						}
					} else {
						expandedSet[tid] = true
					}
				}
				expandedIDs := make([]uint, 0, len(expandedSet))
				for id := range expandedSet {
					expandedIDs = append(expandedIDs, id)
				}
				query = query.Where("EXISTS (SELECT 1 FROM recording_tags rt WHERE rt.recording_id = recordings.id AND rt.tag_id IN ?)", expandedIDs)
			} else {
				// "all" mode: recording must match every requested tag (or its descendants if includeSubtags)
				for _, tid := range requestedTagIDs {
					var matchIDs []uint
					if includeSubtags {
						matchIDs = getDescendantTagIDs(tid, childrenMap)
					} else {
						matchIDs = []uint{tid}
					}
					query = query.Where("EXISTS (SELECT 1 FROM recording_tags rt WHERE rt.recording_id = recordings.id AND rt.tag_id IN ?)", matchIDs)
				}
			}
		}
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN transcript_segments ON transcript_segments.recording_id = recordings.id").
			Where("recordings.title ILIKE ? OR transcript_segments.text ILIKE ? OR transcript_segments.speaker ILIKE ?", searchPattern, searchPattern, searchPattern).
			Distinct()
	}

	var recordings []models.Recording
	query.Order("recordings.created_at desc").Find(&recordings)

	response := []RecordingOut{}
	for _, r := range recordings {
		response = append(response, RecordingOut{
			ID:          r.ID,
			Title:       r.Title,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
			Duration:    r.Duration,
			IsMedical:   r.IsMedical,
			Source:      normalizeRecordingSource(r.Source),
			ExternalURL: r.ExternalURL,
			Tags:        buildTagSummaries(r.Tags, tagByID),
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
	_, tagByID, _ := loadUserTagGraph(db, currentUser.ID)

	accessClause := "recordings.user_id = ? OR EXISTS (SELECT 1 FROM recording_accesses ra WHERE ra.recording_id = recordings.id AND ra.user_id = ?)"

	var recordings []models.Recording
	db.Preload("Tags").Where(accessClause, currentUser.ID, currentUser.ID).Order("created_at desc").Limit(5).Find(&recordings)

	var totalRecs int64
	db.Model(&models.Recording{}).Where(accessClause, currentUser.ID, currentUser.ID).Count(&totalRecs)

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
			ID:          r.ID,
			Title:       r.Title,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
			Duration:    r.Duration,
			IsMedical:   r.IsMedical,
			Source:      normalizeRecordingSource(r.Source),
			ExternalURL: r.ExternalURL,
			Tags:        buildTagSummaries(r.Tags, tagByID),
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

	apiKey := getSpeechmaticsKey(&settings)
	isUnlimited := currentUser.Role == "admin"

	respondUsage := func(totalSeconds float64) {
		usedHours := math.Round((totalSeconds/3600.0)*100) / 100
		if isUnlimited {
			c.JSON(http.StatusOK, gin.H{
				"used_seconds":  totalSeconds,
				"limit_seconds": nil,
				"used_hours":    usedHours,
				"limit_hours":   nil,
				"percentage":    0,
				"is_unlimited":  true,
			})
			return
		}
		limitHours := currentUser.TranscriptionLimitHours
		limitSeconds := limitHours * 3600.0
		percentage := 0.0
		if limitSeconds > 0 {
			percentage = math.Min(math.Round((totalSeconds/limitSeconds)*1000)/10, 100.0)
		} else if totalSeconds > 0 {
			percentage = 100.0
		}
		c.JSON(http.StatusOK, gin.H{
			"used_seconds":  totalSeconds,
			"limit_seconds": limitSeconds,
			"used_hours":    usedHours,
			"limit_hours":   limitHours,
			"percentage":    percentage,
			"is_unlimited":  false,
		})
	}

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
							respondUsage(durationHrs * 3600.0)
							return
						}
					}
				}
			}
		}
	}

	// Fallback to local DB calculation
	respondUsage(getUserUsedSeconds(db, currentUser.ID))
}

func getRecordingDetails(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()
	_, tagByID, _ := loadUserTagGraph(db, currentUser.ID)

	if !canAccessRecording(db, recordingID, currentUser.ID) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	var recording models.Recording
	if err := db.Preload("Tags").Preload("Segments", func(db *gorm.DB) *gorm.DB {
		return db.Order("transcript_segments.start_time ASC, transcript_segments.id ASC")
	}).Where("id = ?", recordingID).First(&recording).Error; err != nil {
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
		"id":           recording.ID,
		"title":        recording.Title,
		"status":       recording.Status,
		"created_at":   recording.CreatedAt.Format(time.RFC3339),
		"duration":     recording.Duration,
		"summary_md":   recording.SummaryMD,
		"is_medical":   recording.IsMedical,
		"source":       normalizeRecordingSource(recording.Source),
		"external_url": recording.ExternalURL,
		"is_owner":     recording.UserID == currentUser.ID,
		"segments":     segments,
		"tags":         buildTagSummaries(recording.Tags, tagByID),
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
		"id":           recording.ID,
		"title":        recording.Title,
		"status":       recording.Status,
		"created_at":   recording.CreatedAt.Format(time.RFC3339),
		"duration":     recording.Duration,
		"summary_md":   recording.SummaryMD,
		"is_medical":   recording.IsMedical,
		"source":       normalizeRecordingSource(recording.Source),
		"external_url": recording.ExternalURL,
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

type SetRecordingTagsRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

func setRecordingTags(c *gin.Context) {
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

	var req SetRecordingTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		db.Where("id IN ? AND user_id = ?", req.TagIDs, currentUser.ID).Find(&tags)
	}

	if err := db.Model(&recording).Association("Tags").Replace(tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to update recording tags"})
		return
	}

	_, tagByID, _ := loadUserTagGraph(db, currentUser.ID)
	c.JSON(http.StatusOK, gin.H{
		"id":   recording.ID,
		"tags": buildTagSummaries(tags, tagByID),
	})
}

type BulkTagsRequest struct {
	RecordingIDs []string `json:"recording_ids" binding:"required"`
	TagIDs       []uint   `json:"tag_ids" binding:"required"`
	Action       string   `json:"action" binding:"required"` // "add" or "remove"
}

func bulkUpdateRecordingTags(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req BulkTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if req.Action != "add" && req.Action != "remove" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "action must be 'add' or 'remove'"})
		return
	}

	db := models.GetDB()
	var recordings []models.Recording
	db.Where("id IN ? AND user_id = ?", req.RecordingIDs, currentUser.ID).Find(&recordings)
	if len(recordings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"detail": "No matching recordings found"})
		return
	}

	var tags []models.Tag
	db.Where("id IN ? AND user_id = ?", req.TagIDs, currentUser.ID).Find(&tags)
	if len(tags) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"detail": "No matching tags found"})
		return
	}

	for i := range recordings {
		if req.Action == "add" {
			if err := db.Model(&recordings[i]).Association("Tags").Append(tags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to add tags"})
				return
			}
		} else {
			if err := db.Model(&recordings[i]).Association("Tags").Delete(tags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to remove tags"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Bulk tag update completed",
		"recordings_count":  len(recordings),
		"tags_affected":     len(tags),
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
	db.Where("recording_id = ?", recordingID).Order("start_time asc, id asc").Find(&segments)
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

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := services.MarkImportedMeetingDeleted(tx, recording.UserID, recording.Source, recording.ExternalID); err != nil {
			return err
		}
		if err := tx.Where("recording_id = ?", recordingID).Delete(&models.TranscriptSegment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&recording).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to delete recording"})
		return
	}

	if recording.AudioPath != "" {
		if _, err := os.Stat(recording.AudioPath); err == nil {
			os.Remove(recording.AudioPath)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recording deleted successfully"})
}

func createMeetingInvite(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	recordingID := c.Param("recording_id")
	db := models.GetDB()

	var recording models.Recording
	if err := db.Where("id = ?", recordingID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Recording not found"})
		return
	}

	if recording.UserID != currentUser.ID && currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Only the recording owner can create an invite link"})
		return
	}

	var existing models.MeetingInvite
	if err := db.Where("recording_id = ? AND created_by = ?", recordingID, currentUser.ID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"token":        existing.Token,
			"recording_id": existing.RecordingID,
		})
		return
	}

	invite := models.MeetingInvite{
		Token:       uuid.New().String(),
		RecordingID: recordingID,
		CreatedBy:   currentUser.ID,
	}
	if err := db.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create invite link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        invite.Token,
		"recording_id": invite.RecordingID,
	})
}

func getMeetingInvite(c *gin.Context) {
	token := c.Param("token")
	db := models.GetDB()

	var invite models.MeetingInvite
	if err := db.Where("token = ?", token).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Invalid or expired meeting invite link"})
		return
	}

	var recording models.Recording
	if err := db.Where("id = ?", invite.RecordingID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Meeting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        invite.Token,
		"recording_id": recording.ID,
		"title":        recording.Title,
		"created_at":   recording.CreatedAt.Format(time.RFC3339),
		"duration":     recording.Duration,
		"is_medical":   recording.IsMedical,
		"source":       normalizeRecordingSource(recording.Source),
	})
}

type AcceptInviteRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

func acceptMeetingInvite(c *gin.Context) {
	token := c.Param("token")
	var input AcceptInviteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Username and password are required"})
		return
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		username = strings.TrimSpace(input.Email)
	}
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Username is required"})
		return
	}
	if len(input.Password) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Password must be at least 4 characters"})
		return
	}

	db := models.GetDB()
	var invite models.MeetingInvite
	if err := db.Where("token = ?", token).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Invalid or expired meeting invite link"})
		return
	}

	var existing models.User
	if err := db.Where("email = ?", username).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Username already taken. Please choose another or sign in."})
		return
	}

	hashedPassword, err := GetPasswordHash(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to hash password"})
		return
	}

	newUser := models.User{
		Email:          username,
		HashedPassword: hashedPassword,
		Role:           "user",
		IsApproved:     true,
	}
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	access := models.RecordingAccess{
		RecordingID: invite.RecordingID,
		UserID:      newUser.ID,
	}
	db.FirstOrCreate(&access, models.RecordingAccess{RecordingID: invite.RecordingID, UserID: newUser.ID})

	jwtToken, err := CreateAccessToken(newUser.Email, newUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": jwtToken,
		"token_type":   "bearer",
		"recording_id": invite.RecordingID,
		"user": UserOut{
			ID:                      newUser.ID,
			Email:                   newUser.Email,
			Role:                    newUser.Role,
			IsApproved:              newUser.IsApproved,
			TranscriptionLimitHours: newUser.TranscriptionLimitHours,
		},
	})
}

func joinMeetingInvite(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	token := c.Param("token")
	db := models.GetDB()

	var invite models.MeetingInvite
	if err := db.Where("token = ?", token).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Invalid or expired meeting invite link"})
		return
	}

	var recording models.Recording
	if err := db.Where("id = ?", invite.RecordingID).First(&recording).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Meeting not found"})
		return
	}

	if recording.UserID != currentUser.ID {
		db.FirstOrCreate(&models.RecordingAccess{}, models.RecordingAccess{
			RecordingID: invite.RecordingID,
			UserID:      currentUser.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recording_id": invite.RecordingID,
	})
}

type TranscriptSegmentInput struct {
	StartTime *float64 `json:"start_time"`
	EndTime   *float64 `json:"end_time"`
	Speaker   string   `json:"speaker"`
	Text      string   `json:"text"`
}

type CreateTranscriptRecordingRequest struct {
	Title          string                   `json:"title"`
	Transcript     string                   `json:"transcript"`
	Segments       []TranscriptSegmentInput `json:"segments"`
	DefaultSpeaker string                   `json:"default_speaker"`
	IsMedical      bool                     `json:"is_medical"`
	TagIDs         []uint                   `json:"tag_ids"`
	AutoSummarize  *bool                    `json:"auto_summarize"`
}

type UpdateTranscriptRequest struct {
	Transcript     string                   `json:"transcript"`
	Segments       []TranscriptSegmentInput `json:"segments"`
	DefaultSpeaker string                   `json:"default_speaker"`
	Mode           string                   `json:"mode"` // "replace" (default) or "append"
}

func buildTranscriptSegments(recordingID string, rawTranscript string, inputSegments []TranscriptSegmentInput, defaultSpeaker string, timeOffset float64) ([]models.TranscriptSegment, *float64) {
	defSpeaker := strings.TrimSpace(defaultSpeaker)
	if defSpeaker == "" {
		defSpeaker = "Speaker 1"
	}

	var parsed []services.ImportedSegment
	if len(inputSegments) > 0 {
		lastStart := -0.001
		for _, in := range inputSegments {
			txt := strings.TrimSpace(in.Text)
			if txt == "" {
				continue
			}
			spk := strings.TrimSpace(in.Speaker)
			if spk == "" {
				spk = defSpeaker
			}
			var start float64
			if in.StartTime != nil && *in.StartTime >= 0 {
				start = *in.StartTime
			} else {
				start = math.Round(math.Max(0, lastStart+0.001)*1000) / 1000
			}
			end := start
			if in.EndTime != nil && *in.EndTime >= start {
				end = *in.EndTime
			}
			lastStart = start
			parsed = append(parsed, services.ImportedSegment{
				StartTime: start,
				EndTime:   end,
				Speaker:   spk,
				Text:      txt,
			})
		}
	} else {
		parsed = services.ParseTranscriptText(rawTranscript, defSpeaker)
	}

	if len(parsed) == 0 {
		return nil, nil
	}

	dbSegs := make([]models.TranscriptSegment, 0, len(parsed))
	var maxEnd float64
	for _, seg := range parsed {
		start := seg.StartTime + timeOffset
		end := seg.EndTime + timeOffset
		if end < start {
			end = start
		}
		if end > maxEnd {
			maxEnd = end
		}
		dbSegs = append(dbSegs, models.TranscriptSegment{
			RecordingID: recordingID,
			StartTime:   start,
			EndTime:     end,
			Speaker:     seg.Speaker,
			Text:        seg.Text,
		})
	}

	var durationPtr *float64
	if maxEnd >= 1.0 {
		d := math.Round(maxEnd*100) / 100
		durationPtr = &d
	}

	return dbSegs, durationPtr
}

func createRecordingFromTranscript(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateTranscriptRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request payload"})
		return
	}

	jobID := uuid.New().String()
	segments, durationPtr := buildTranscriptSegments(jobID, req.Transcript, req.Segments, req.DefaultSpeaker, 0)
	if len(segments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Transcript text is required and must contain non-empty content"})
		return
	}

	db := models.GetDB()
	var settings models.Settings
	db.First(&settings)

	autoSummarize := settings.AISummarizationMode == "auto"
	if req.AutoSummarize != nil {
		autoSummarize = *req.AutoSummarize
	}

	customTitle := strings.TrimSpace(req.Title)
	preserveTitle := customTitle != ""
	title := customTitle
	if title == "" {
		if req.IsMedical {
			title = "Clinical Consultation"
		} else {
			title = "Transcript Note"
		}
	}

	initialStatus := "completed"
	if autoSummarize {
		initialStatus = "summarizing"
	}

	recording := models.Recording{
		ID:        jobID,
		UserID:    currentUser.ID,
		Title:     title,
		Filename:  "transcript.txt",
		AudioPath: "",
		Status:    initialStatus,
		Duration:  durationPtr,
		IsMedical: req.IsMedical,
		Source:    services.SourceLocal,
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&recording).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&segments, 200).Error; err != nil {
			return err
		}
		if len(req.TagIDs) > 0 {
			var tags []models.Tag
			tx.Where("id IN ? AND user_id = ?", req.TagIDs, currentUser.ID).Find(&tags)
			if len(tags) > 0 {
				if err := tx.Model(&recording).Association("Tags").Replace(tags); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not save transcript recording"})
		return
	}

	if autoSummarize {
		models.GlobalJobsDB.Set(jobID, models.JobStatus{
			ID:       jobID,
			Filename: recording.Filename,
			Status:   "summarizing",
			Progress: 85,
		})
		go services.GenerateSummaryWithOptions(jobID, "", preserveTitle)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Transcript saved successfully",
		"job_id":        jobID,
		"id":            jobID,
		"title":         recording.Title,
		"status":        recording.Status,
		"segment_count": len(segments),
	})
}

func replaceRecordingTranscript(c *gin.Context) {
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

	if recording.Status == "pending" || recording.Status == "transcribing" || recording.Status == "diarizing" || recording.Status == "summarizing" {
		c.JSON(http.StatusConflict, gin.H{"detail": "Cannot modify transcript while recording is still processing"})
		return
	}

	var req UpdateTranscriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid request payload"})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "replace"
	}
	if mode != "replace" && mode != "append" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "mode must be 'replace' or 'append'"})
		return
	}

	timeOffset := 0.0
	if mode == "append" {
		var existing []models.TranscriptSegment
		db.Where("recording_id = ?", recordingID).Order("start_time desc, id desc").Limit(1).Find(&existing)
		if len(existing) > 0 {
			maxExisting := existing[0].EndTime
			if existing[0].StartTime > maxExisting {
				maxExisting = existing[0].StartTime
			}
			timeOffset = math.Round((maxExisting+0.001)*1000) / 1000
		}
	}

	segments, durationPtr := buildTranscriptSegments(recordingID, req.Transcript, req.Segments, req.DefaultSpeaker, timeOffset)
	if len(segments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Transcript text is required and must contain non-empty content"})
		return
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		if mode == "replace" {
			if err := tx.Where("recording_id = ?", recordingID).Delete(&models.TranscriptSegment{}).Error; err != nil {
				return err
			}
		}
		if err := tx.CreateInBatches(&segments, 200).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{}
		if recording.Status == "error" {
			updates["status"] = "completed"
			updates["error_message"] = nil
		}
		if recording.Duration == nil && durationPtr != nil {
			updates["duration"] = *durationPtr
		}
		if len(updates) > 0 {
			if err := tx.Model(&recording).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to save transcript segments"})
		return
	}

	var allSegments []models.TranscriptSegment
	db.Where("recording_id = ?", recordingID).Order("start_time asc, id asc").Find(&allSegments)

	segOut := make([]gin.H, 0, len(allSegments))
	for _, s := range allSegments {
		segOut = append(segOut, gin.H{
			"id":      s.ID,
			"start":   s.StartTime,
			"end":     s.EndTime,
			"speaker": s.Speaker,
			"text":    s.Text,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Transcript updated successfully",
		"segment_count": len(allSegments),
		"segments":      segOut,
	})
}
