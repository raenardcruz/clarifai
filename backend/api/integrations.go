package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"note-taker/backend/models"
	"note-taker/backend/services"
)

type SyncIntegrationsRequest struct {
	Provider string `json:"provider"` // "all", "fathom", "fireflies", "otter"
}

type TestIntegrationRequest struct {
	Method      string `json:"method"`
	APIKey      string `json:"api_key"`
	EndpointURL string `json:"endpoint_url"`
}

func SetupIntegrationsRoutes(router *gin.RouterGroup) {
	router.POST("/sync", AuthMiddleware(), RequireAdminUser(), syncIntegrationsHandler)
	router.POST("/:provider/test", AuthMiddleware(), RequireAdminUser(), testIntegrationHandler)
}

func syncIntegrationsHandler(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SyncIntegrationsRequest
	_ = c.ShouldBindJSON(&req)

	db := models.GetDB()
	results, err := services.SyncAllEnabledIntegrations(c.Request.Context(), db, currentUser.ID, req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	settings := loadOrCreateSettings()
	c.JSON(http.StatusOK, gin.H{
		"results":  results,
		"settings": buildSettingsOut(&settings),
	})
}

func testIntegrationHandler(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if provider != services.ProviderFathom && provider != services.ProviderFireflies && provider != services.ProviderOtter {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Unsupported integration provider"})
		return
	}

	var req TestIntegrationRequest
	_ = c.ShouldBindJSON(&req)

	settings := loadOrCreateSettings()

	method := strings.TrimSpace(req.Method)
	apiKey := strings.TrimSpace(req.APIKey)
	endpointURL := strings.TrimSpace(req.EndpointURL)

	switch provider {
	case services.ProviderFathom:
		if method == "" {
			method = defaultIfEmpty(settings.FathomMethod, "api")
		}
		if apiKey == "" || apiKey == "••••••••" {
			if settings.FathomAPIKey != nil {
				apiKey = *settings.FathomAPIKey
			} else {
				apiKey = ""
			}
		}
		if endpointURL == "" && method == "mcp" {
			endpointURL = defaultIfEmpty(settings.FathomMCPURL, services.DefaultFathomMCPURL)
		}
	case services.ProviderFireflies:
		if method == "" {
			method = defaultIfEmpty(settings.FirefliesMethod, "api")
		}
		if apiKey == "" || apiKey == "••••••••" {
			if settings.FirefliesAPIKey != nil {
				apiKey = *settings.FirefliesAPIKey
			} else {
				apiKey = ""
			}
		}
		if endpointURL == "" && method == "mcp" {
			endpointURL = defaultIfEmpty(settings.FirefliesMCPURL, services.DefaultFirefliesMCPURL)
		}
	case services.ProviderOtter:
		if method == "" {
			method = defaultIfEmpty(settings.OtterMethod, "api")
		}
		if apiKey == "" || apiKey == "••••••••" {
			if settings.OtterAPIKey != nil {
				apiKey = *settings.OtterAPIKey
			} else {
				apiKey = ""
			}
		}
		if endpointURL == "" {
			if method == "mcp" {
				endpointURL = settings.OtterMCPURL
			} else {
				endpointURL = defaultIfEmpty(settings.OtterAPIBaseURL, services.DefaultOtterAPIBase)
			}
		}
	}

	res := services.TestIntegrationConnection(c.Request.Context(), provider, method, apiKey, endpointURL)
	c.JSON(http.StatusOK, res)
}
