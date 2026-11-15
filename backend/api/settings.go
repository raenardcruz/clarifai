package api

import (
	"net/http"
	"time"

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

	FathomEnabled    bool    `json:"fathom_enabled"`
	FathomMethod     string  `json:"fathom_method"`
	FathomAPIKey     *string `json:"fathom_api_key,omitempty"`
	FathomMCPURL     string  `json:"fathom_mcp_url"`

	FirefliesEnabled bool    `json:"fireflies_enabled"`
	FirefliesMethod  string  `json:"fireflies_method"`
	FirefliesAPIKey  *string `json:"fireflies_api_key,omitempty"`
	FirefliesMCPURL  string  `json:"fireflies_mcp_url"`

	OtterEnabled     bool    `json:"otter_enabled"`
	OtterMethod      string  `json:"otter_method"`
	OtterAPIKey      *string `json:"otter_api_key,omitempty"`
	OtterAPIBaseURL  string  `json:"otter_api_base_url"`
	OtterMCPURL      string  `json:"otter_mcp_url"`
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

	FathomEnabled        bool    `json:"fathom_enabled"`
	FathomMethod         string  `json:"fathom_method"`
	FathomAPIKey         *string `json:"fathom_api_key,omitempty"`
	FathomMCPURL         string  `json:"fathom_mcp_url"`
	FathomLastSyncAt     *string `json:"fathom_last_sync_at,omitempty"`
	FathomLastSyncStatus string  `json:"fathom_last_sync_status,omitempty"`

	FirefliesEnabled        bool    `json:"fireflies_enabled"`
	FirefliesMethod         string  `json:"fireflies_method"`
	FirefliesAPIKey         *string `json:"fireflies_api_key,omitempty"`
	FirefliesMCPURL         string  `json:"fireflies_mcp_url"`
	FirefliesLastSyncAt     *string `json:"fireflies_last_sync_at,omitempty"`
	FirefliesLastSyncStatus string  `json:"fireflies_last_sync_status,omitempty"`

	OtterEnabled        bool    `json:"otter_enabled"`
	OtterMethod         string  `json:"otter_method"`
	OtterAPIKey         *string `json:"otter_api_key,omitempty"`
	OtterAPIBaseURL     string  `json:"otter_api_base_url"`
	OtterMCPURL         string  `json:"otter_mcp_url"`
	OtterLastSyncAt     *string `json:"otter_last_sync_at,omitempty"`
	OtterLastSyncStatus string  `json:"otter_last_sync_status,omitempty"`
}

func maskSecret(val *string) *string {
	if val != nil && *val != "" {
		masked := "••••••••"
		return &masked
	}
	return nil
}

func applyMaskedSecretUpdate(target **string, incoming *string) {
	if incoming == nil || *incoming == "••••••••" {
		return
	}
	if *incoming == "" {
		*target = nil
	} else {
		val := *incoming
		*target = &val
	}
}

func formatOptionalTime(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func defaultIfEmpty(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func buildSettingsOut(settings *models.Settings) SettingsOut {
	return SettingsOut{
		AISummarizationMode:     settings.AISummarizationMode,
		ExecutiveSummaryPrompt:  settings.ExecutiveSummaryPrompt,
		MedicalSummaryPrompt:    settings.MedicalSummaryPrompt,
		SpeechmaticsAPIKey:      maskSecret(settings.SpeechmaticsAPIKey),
		OllamaURL:               defaultIfEmpty(settings.OllamaURL, "http://localhost:11434"),
		OllamaModel:             defaultIfEmpty(settings.OllamaModel, "gemma4:12b-mlx"),
		AIProvider:              defaultIfEmpty(settings.AIProvider, "ollama"),
		GeminiModel:             defaultIfEmpty(settings.GeminiModel, "gemini-3.8-flash"),
		GeminiAPIKey:            maskSecret(settings.GeminiAPIKey),
		GeminiAPIBase:           settings.GeminiAPIBase,
		FathomEnabled:           settings.FathomEnabled,
		FathomMethod:            defaultIfEmpty(settings.FathomMethod, "api"),
		FathomAPIKey:            maskSecret(settings.FathomAPIKey),
		FathomMCPURL:            defaultIfEmpty(settings.FathomMCPURL, services.DefaultFathomMCPURL),
		FathomLastSyncAt:        formatOptionalTime(settings.FathomLastSyncAt),
		FathomLastSyncStatus:    settings.FathomLastSyncStatus,
		FirefliesEnabled:        settings.FirefliesEnabled,
		FirefliesMethod:         defaultIfEmpty(settings.FirefliesMethod, "api"),
		FirefliesAPIKey:         maskSecret(settings.FirefliesAPIKey),
		FirefliesMCPURL:         defaultIfEmpty(settings.FirefliesMCPURL, services.DefaultFirefliesMCPURL),
		FirefliesLastSyncAt:     formatOptionalTime(settings.FirefliesLastSyncAt),
		FirefliesLastSyncStatus: settings.FirefliesLastSyncStatus,
		OtterEnabled:            settings.OtterEnabled,
		OtterMethod:             defaultIfEmpty(settings.OtterMethod, "api"),
		OtterAPIKey:             maskSecret(settings.OtterAPIKey),
		OtterAPIBaseURL:         defaultIfEmpty(settings.OtterAPIBaseURL, services.DefaultOtterAPIBase),
		OtterMCPURL:             settings.OtterMCPURL,
		OtterLastSyncAt:         formatOptionalTime(settings.OtterLastSyncAt),
		OtterLastSyncStatus:     settings.OtterLastSyncStatus,
	}
}

func loadOrCreateSettings() models.Settings {
	db := models.GetDB()
	var settings models.Settings
	if err := db.First(&settings).Error; err != nil {
		settings = models.Settings{
			AIProvider:      "ollama",
			OllamaURL:       "http://localhost:11434",
			OllamaModel:     "gemma4:12b-mlx",
			GeminiModel:     "gemini-3.8-flash",
			FathomMethod:    "api",
			FathomMCPURL:    services.DefaultFathomMCPURL,
			FirefliesMethod: "api",
			FirefliesMCPURL: services.DefaultFirefliesMCPURL,
			OtterMethod:     "api",
			OtterAPIBaseURL: services.DefaultOtterAPIBase,
		}
		db.Create(&settings)
	}
	return settings
}

func SetupSettingsRoutes(router *gin.RouterGroup) {
	// GET route is publicly authenticated (for regular users)
	router.GET("", AuthMiddleware(), RequireApprovedUser(), func(c *gin.Context) {
		settings := loadOrCreateSettings()
		c.JSON(http.StatusOK, buildSettingsOut(&settings))
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
		settings := loadOrCreateSettings()

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

		applyMaskedSecretUpdate(&settings.SpeechmaticsAPIKey, input.SpeechmaticsAPIKey)
		applyMaskedSecretUpdate(&settings.GeminiAPIKey, input.GeminiAPIKey)

		// Third-party integration settings
		settings.FathomEnabled = input.FathomEnabled
		settings.FathomMethod = defaultIfEmpty(input.FathomMethod, "api")
		settings.FathomMCPURL = defaultIfEmpty(input.FathomMCPURL, services.DefaultFathomMCPURL)
		applyMaskedSecretUpdate(&settings.FathomAPIKey, input.FathomAPIKey)

		settings.FirefliesEnabled = input.FirefliesEnabled
		settings.FirefliesMethod = defaultIfEmpty(input.FirefliesMethod, "api")
		settings.FirefliesMCPURL = defaultIfEmpty(input.FirefliesMCPURL, services.DefaultFirefliesMCPURL)
		applyMaskedSecretUpdate(&settings.FirefliesAPIKey, input.FirefliesAPIKey)

		settings.OtterEnabled = input.OtterEnabled
		settings.OtterMethod = defaultIfEmpty(input.OtterMethod, "api")
		settings.OtterAPIBaseURL = defaultIfEmpty(input.OtterAPIBaseURL, services.DefaultOtterAPIBase)
		settings.OtterMCPURL = input.OtterMCPURL
		applyMaskedSecretUpdate(&settings.OtterAPIKey, input.OtterAPIKey)

		db.Save(&settings)

		c.JSON(http.StatusOK, buildSettingsOut(&settings))
	})
}
