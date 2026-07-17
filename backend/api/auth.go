package api

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"note-taker/backend/models"
)

var SecretKey = []byte(os.Getenv("SECRET_KEY"))

func init() {
	if len(SecretKey) == 0 {
		SecretKey = []byte("supersecretkey")
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type UserCreate struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserOut struct {
	ID                      uint    `json:"id"`
	Email                   string  `json:"email"`
	Role                    string  `json:"role"`
	IsApproved              bool    `json:"is_approved"`
	TranscriptionLimitHours float64 `json:"transcription_limit_hours"`
}

func VerifyPassword(plain, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

func GetPasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CreateAccessToken(email string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString(SecretKey)
}

// AuthMiddleware intercepts requests and parses user details into context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := authHeader
		// Handle "Bearer <token>"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return SecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		email, _ := claims["sub"].(string)
		db := models.GetDB()
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("current_user", &user)
		c.Next()
	}
}

func RequireApprovedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		user := userVal.(*models.User)
		if !user.IsApproved {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "User not approved by admin"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdminUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		user := userVal.(*models.User)
		if !user.IsApproved {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "User not approved by admin"})
			c.Abort()
			return
		}
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"detail": "Not enough privileges"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func SetupAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", func(c *gin.Context) {
		// FastAPI's OAuth2PasswordRequestForm expects form data: username and password
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Fallback to JSON if form data not present (for flexibility)
		if username == "" && password == "" {
			var jsonInput struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&jsonInput); err == nil {
				username = jsonInput.Username
				password = jsonInput.Password
			}
		}

		db := models.GetDB()
		var user models.User
		if err := db.Where("email = ?", username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		if !VerifyPassword(password, user.HashedPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password"})
			return
		}

		if !user.IsApproved {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Account pending admin approval"})
			return
		}

		token, err := CreateAccessToken(user.Email, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not create token"})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{
			AccessToken: token,
			TokenType:   "bearer",
		})
	})

	router.POST("/register", func(c *gin.Context) {
		var input UserCreate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		db := models.GetDB()
		var existing models.User
		if err := db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Email already registered"})
			return
		}

		hashedPassword, err := GetPasswordHash(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to hash password"})
			return
		}

		newUser := models.User{
			Email:          input.Email,
			HashedPassword: hashedPassword,
			Role:           "user",
			IsApproved:     false,
		}

		if err := db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, UserOut{
			ID:                      newUser.ID,
			Email:                   newUser.Email,
			Role:                    newUser.Role,
			IsApproved:              newUser.IsApproved,
			TranscriptionLimitHours: newUser.TranscriptionLimitHours,
		})
	})

	// Authenticated routes
	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(), RequireApprovedUser())
	{
		authGroup.GET("/me", func(c *gin.Context) {
			userVal, _ := c.Get("current_user")
			user := userVal.(*models.User)
			c.JSON(http.StatusOK, UserOut{
				ID:                      user.ID,
				Email:                   user.Email,
				Role:                    user.Role,
				IsApproved:              user.IsApproved,
				TranscriptionLimitHours: user.TranscriptionLimitHours,
			})
		})
	}
}
