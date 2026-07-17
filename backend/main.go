package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"note-taker/backend/api"
	"note-taker/backend/models"
)

func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			os.Setenv(key, val)
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	// Load environment variables
	ex, err := os.Executable()
	var dir string
	if err == nil {
		dir = filepath.Dir(ex)
	} else {
		dir = "."
	}
	loadEnv(filepath.Join(dir, ".env"))
	loadEnv(".env") // fallback to current dir

	// Initialize DB
	models.InitDB()

	// Initialize Gin Engine
	r := gin.Default()

	// Apply CORS
	r.Use(CORSMiddleware())

	// Welcome Route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to Note-Taker API"})
	})

	// Register Router Groups
	apiGroup := r.Group("/api")
	{
		api.SetupAuthRoutes(apiGroup.Group("/auth"))
		api.SetupUsersRoutes(apiGroup.Group("/users"))
		api.SetupSettingsRoutes(apiGroup.Group("/settings"))
		api.SetupRecordingsRoutes(apiGroup.Group("/recordings"))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
