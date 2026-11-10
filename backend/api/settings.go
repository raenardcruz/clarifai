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
	MedicalSummaryPrompt   string  `json:"medical_summary_prompt"`
	SpeechmaticsAPIKey     *string `json:"speechmatics_api_key,omitempty"`
	OllamaURL              string  `json:"ollama_url"`
	OllamaModel            string  `json:"ollama_model"`
	AIProvider             string  `json:"ai_provider"`
	GeminiModel            string  `json:"gemini_model"`
	GeminiAPIKey           *string `json:"gemini_api_key,omitempty"`
	GeminiAPIBase          string  `json:"gemini_api_base,omitempty"`
}

type SettingsOut struct {
	AISummarizationMode    string  `json:"ai_summarization_mode"`
	ExecutiveSummaryPrompt string  `json:"executive_summary_prompt"`
	MedicalSummaryPrompt   string  `json:"medical_summary_prompt"`
	SpeechmaticsAPIKey     *string `json:"speechmatics_api_key,omitempty"`
	OllamaURL              string  `json:"ollama_url"`
	OllamaModel            string  `json:"ollama_model"`
	AIProvider             string  `json:"ai_provider"`
	GeminiModel            string  `json:"gemini_model"`
	GeminiAPIKey           *string `json:"gemini_api_key,omitempty"`
	GeminiAPIBase          string  `json:"gemini_api_base,omitempty"`
}

func SetupSettingsRoutes(router *gin.RouterGroup) {
	// GET route is publicly authenticated (for regular users)
	router.GET("", AuthMiddleware(), RequireApprovedUser(), func(c *gin.Context) {
		db := models.GetDB()
		var settings models.Settings
		if err := db.First(&settings).Error; err != nil {
			// Create default settings if not exists
			settings = models.Settings{
				AIProvider:  "ollama",
				OllamaURL:   "http://localhost:11434",
				OllamaModel: "gemma4:12b-mlx",
				GeminiModel: "gemini-3.8-flash",
			}
			db.Create(&settings)
		}

		var maskedSpeechmaticsKey *string
		if settings.SpeechmaticsAPIKey != nil && *settings.SpeechmaticsAPIKey != "" {
			masked := "••••••••"
			maskedSpeechmaticsKey = &masked
		}

		var maskedGeminiKey *string
		if settings.GeminiAPIKey != nil && *settings.GeminiAPIKey != "" {
			masked := "••••••••"
			maskedGeminiKey = &masked
		}

		provider := settings.AIProvider
		if provider == "" {
			provider = "ollama"
		}
		geminiModel := settings.GeminiModel
		if geminiModel == "" {
			geminiModel = "gemini-3.8-flash"
		}
		ollamaModel := settings.OllamaModel
		if ollamaModel == "" {
			ollamaModel = "gemma4:12b-mlx"
		}
		ollamaURL := settings.OllamaURL
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}

		c.JSON(http.StatusOK, SettingsOut{
			AISummarizationMode:    settings.AISummarizationMode,
			ExecutiveSummaryPrompt: settings.ExecutiveSummaryPrompt,
			MedicalSummaryPrompt:   settings.MedicalSummaryPrompt,
			SpeechmaticsAPIKey:     maskedSpeechmaticsKey,
			OllamaURL:              ollamaURL,
			OllamaModel:            ollamaModel,
			AIProvider:             provider,
			GeminiModel:            geminiModel,
			GeminiAPIKey:           maskedGeminiKey,
			GeminiAPIBase:          settings.GeminiAPIBase,
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

	// GET route to retrieve available Gemini models
	router.GET("/gemini-models", AuthMiddleware(), RequireApprovedUser(), func(c *gin.Context) {
		c.JSON(http.StatusOK, services.GetGeminiModels())
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
			settings = models.Settings{
				AIProvider:  "ollama",
				OllamaURL:   "http://localhost:11434",
				OllamaModel: "gemma4:12b-mlx",
				GeminiModel: "gemini-3.8-flash",
			}
			db.Create(&settings)
		}

		settings.AISummarizationMode = input.AISummarizationMode
		settings.ExecutiveSummaryPrompt = input.ExecutiveSummaryPrompt
		if input.MedicalSummaryPrompt != "" {
			settings.MedicalSummaryPrompt = input.MedicalSummaryPrompt
		}
		if input.OllamaURL != "" {
			settings.OllamaURL = input.OllamaURL
		}
		if input.OllamaModel != "" {
			settings.OllamaModel = input.OllamaModel
		}
		if input.AIProvider != "" {
			settings.AIProvider = input.AIProvider
		}
		if input.GeminiModel != "" {
			settings.GeminiModel = input.GeminiModel
		}
		settings.GeminiAPIBase = input.GeminiAPIBase

		if input.SpeechmaticsAPIKey != nil {
			if *input.SpeechmaticsAPIKey != "••••••••" {
				if *input.SpeechmaticsAPIKey == "" {
					settings.SpeechmaticsAPIKey = nil
				} else {
					settings.SpeechmaticsAPIKey = input.SpeechmaticsAPIKey
				}
			}
		}

		if input.GeminiAPIKey != nil {
			if *input.GeminiAPIKey != "••••••••" {
				if *input.GeminiAPIKey == "" {
					settings.GeminiAPIKey = nil
				} else {
					settings.GeminiAPIKey = input.GeminiAPIKey
				}
			}
		}

		db.Save(&settings)

		var maskedSpeechmaticsKey *string
		if settings.SpeechmaticsAPIKey != nil && *settings.SpeechmaticsAPIKey != "" {
			masked := "••••••••"
			maskedSpeechmaticsKey = &masked
		}

		var maskedGeminiKey *string
		if settings.GeminiAPIKey != nil && *settings.GeminiAPIKey != "" {
			masked := "••••••••"
			maskedGeminiKey = &masked
		}

		provider := settings.AIProvider
		if provider == "" {
			provider = "ollama"
		}
		geminiModel := settings.GeminiModel
		if geminiModel == "" {
			geminiModel = "gemini-3.8-flash"
		}
		ollamaModel := settings.OllamaModel
		if ollamaModel == "" {
			ollamaModel = "gemma4:12b-mlx"
		}
		ollamaURL := settings.OllamaURL
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}

		c.JSON(http.StatusOK, SettingsOut{
			AISummarizationMode:    settings.AISummarizationMode,
			ExecutiveSummaryPrompt: settings.ExecutiveSummaryPrompt,
			MedicalSummaryPrompt:   settings.MedicalSummaryPrompt,
			SpeechmaticsAPIKey:     maskedSpeechmaticsKey,
			OllamaURL:              ollamaURL,
			OllamaModel:            ollamaModel,
			AIProvider:             provider,
			GeminiModel:            geminiModel,
			GeminiAPIKey:           maskedGeminiKey,
			GeminiAPIBase:          settings.GeminiAPIBase,
		})
	})
}
