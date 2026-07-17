package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"note-taker/backend/models"
)

type UserUpdate struct {
	Role                    string  `json:"role" binding:"required"`
	IsApproved              bool    `json:"is_approved" binding:"exists"`
	TranscriptionLimitHours float64 `json:"transcription_limit_hours" binding:"required"`
}

func SetupUsersRoutes(router *gin.RouterGroup) {
	// Apply Auth and Admin middlewares to all user routes
	router.Use(AuthMiddleware(), RequireAdminUser())

	router.GET("", func(c *gin.Context) {
		db := models.GetDB()
		var users []models.User
		if err := db.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		var response []UserOut
		for _, u := range users {
			response = append(response, UserOut{
				ID:                      u.ID,
				Email:                   u.Email,
				Role:                    u.Role,
				IsApproved:              u.IsApproved,
				TranscriptionLimitHours: u.TranscriptionLimitHours,
			})
		}

		c.JSON(http.StatusOK, response)
	})

	router.PUT("/:user_id", func(c *gin.Context) {
		userIDStr := c.Param("user_id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid user ID"})
			return
		}

		var input UserUpdate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		db := models.GetDB()
		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
			return
		}

		if user.Email == "admin" && input.Role != "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot downgrade primary admin"})
			return
		}

		user.Role = input.Role
		user.IsApproved = input.IsApproved
		user.TranscriptionLimitHours = input.TranscriptionLimitHours
		db.Save(&user)

		c.JSON(http.StatusOK, UserOut{
			ID:                      user.ID,
			Email:                   user.Email,
			Role:                    user.Role,
			IsApproved:              user.IsApproved,
			TranscriptionLimitHours: user.TranscriptionLimitHours,
		})
	})

	router.DELETE("/:user_id", func(c *gin.Context) {
		userIDStr := c.Param("user_id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid user ID"})
			return
		}

		db := models.GetDB()
		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
			return
		}

		if user.Email == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot delete primary admin"})
			return
		}

		db.Delete(&user)

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	})
}
