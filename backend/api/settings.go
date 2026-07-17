package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"note-taker/backend/models"
	"note-taker/backend/services"
)

type SettingsUpdate struct {
	AISummarizationMode    string  `json:"ai_summarization_mode" binding:"required"`
	ExecutiveSummaryPrompt string  `json:"executive_summary_prompt" binding:"required"`
	SpeechmaticsAPIKey     *string `json:"speechmatics_api_key,omitempty"`
	OllamaModel            string  `json:"ollama_model"`
}

type SettingsOut struct {
	AISummarizationMode    string  `json:"ai_summarization_mode"`
	ExecutiveSummaryPrompt string  `json:"executive_summary_prompt"`
	SpeechmaticsAPIKey     *string `json:"speechmatics_api_key,omitempty"`
	OllamaModel            string  `json:"ollama_model"`
}

func SetupSettingsRoutes(router *gin.RouterGroup) {
	// GET route is publicly authenticated (for regular users)
	router.GET("", AuthMiddleware(), RequireApprovedUser(), func(c *gin.Context) {
		db := models.GetDB()
		var settings models.Settings
		if err := db.First(&settings).Error; err != nil {
			// Create default settings if not exists
			settings = models.Settings{}
			db.Create(&settings)
		}

		var maskedKey *string
		if settings.SpeechmaticsAPIKey != nil && *settings.SpeechmaticsAPIKey != "" {
			masked := "••••••••"
			maskedKey = &masked
		}

		c.JSON(http.StatusOK, SettingsOut{
			AISummarizationMode:    settings.AISummarizationMode,
			ExecutiveSummaryPrompt: settings.ExecutiveSummaryPrompt,
			SpeechmaticsAPIKey:     maskedKey,
			OllamaModel:            settings.OllamaModel,
		})
	})

	// GET route to retrieve available Ollama models
	router.GET("/ollama-models", AuthMiddleware(), RequireApprovedUser(), func(c *gin.Context) {
		modelsList, err := services.GetOllamaModels()
		if err != nil {
			// Return a slice with default model if Ollama request fails
			c.JSON(http.StatusOK, []string{"gemma4:12b-mlx"})
			return
		}
		c.JSON(http.StatusOK, modelsList)
	})

	// PUT route requires Admin privileges
	router.PUT("", AuthMiddleware(), RequireAdminUser(), func(c *gin.Context) {
		var input SettingsUpdate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		db := models.GetDB()
		var settings models.Settings
		if err := db.First(&settings).Error; err != nil {
			settings = models.Settings{}
			db.Create(&settings)
		}

		settings.AISummarizationMode = input.AISummarizationMode
		settings.ExecutiveSummaryPrompt = input.ExecutiveSummaryPrompt
		if input.OllamaModel != "" {
			settings.OllamaModel = input.OllamaModel
		}

		if input.SpeechmaticsAPIKey != nil && *input.SpeechmaticsAPIKey != "••••••••" {
			settings.SpeechmaticsAPIKey = input.SpeechmaticsAPIKey
		}

		db.Save(&settings)

		var maskedKey *string
		if settings.SpeechmaticsAPIKey != nil && *settings.SpeechmaticsAPIKey != "" {
			masked := "••••••••"
			maskedKey = &masked
		}

		c.JSON(http.StatusOK, SettingsOut{
			AISummarizationMode:    settings.AISummarizationMode,
			ExecutiveSummaryPrompt: settings.ExecutiveSummaryPrompt,
			SpeechmaticsAPIKey:     maskedKey,
			OllamaModel:            settings.OllamaModel,
		})
	})
}
