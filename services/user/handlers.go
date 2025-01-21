package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
}

type UpdateProfileRequest struct {
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	PhoneNumber       string     `json:"phone_number"`
	DateOfBirth       *time.Time `json:"date_of_birth"`
	ProfilePicture    string     `json:"profile_picture"`
	Bio               string     `json:"bio"`
	PreferredLanguage string     `json:"preferred_language"`
}

type UpdateNotificationPreferencesRequest struct {
	EmailNotifications bool `json:"email_notifications"`
	OrderUpdates       bool `json:"order_updates"`
	PromotionalEmails  bool `json:"promotional_emails"`
	SecurityAlerts     bool `json:"security_alerts"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func Register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user already exists
		var existingUser User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		user := User{
			Email:       req.Email,
			Password:    req.Password,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			PhoneNumber: req.PhoneNumber,
		}

		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"user_id": user.ID,
		})
	}
}

func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq LoginRequest
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var user User
		if err := db.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Create token with user ID as string
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": strconv.FormatUint(uint64(user.ID), 10),
			"role":    user.Role,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user":  user,
		})
	}
}

func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in token"})
			return
		}

		var user User
		if err := db.Preload("Addresses").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func UpdateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Update only provided fields
		updates := make(map[string]interface{})
		if req.FirstName != "" {
			updates["first_name"] = req.FirstName
		}
		if req.LastName != "" {
			updates["last_name"] = req.LastName
		}
		if req.PhoneNumber != "" {
			updates["phone_number"] = req.PhoneNumber
		}
		if req.DateOfBirth != nil {
			updates["date_of_birth"] = req.DateOfBirth
		}
		if req.ProfilePicture != "" {
			updates["profile_picture"] = req.ProfilePicture
		}
		if req.Bio != "" {
			updates["bio"] = req.Bio
		}
		if req.PreferredLanguage != "" {
			updates["preferred_language"] = req.PreferredLanguage
		}

		if err := db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func UpdateNotificationPreferences(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req UpdateNotificationPreferencesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		updates := NotificationPreferences{
			EmailNotifications: req.EmailNotifications,
			OrderUpdates:       req.OrderUpdates,
			PromotionalEmails:  req.PromotionalEmails,
			SecurityAlerts:     req.SecurityAlerts,
		}

		if err := db.Model(&user).Updates(map[string]interface{}{
			"notification_preferences": updates,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification preferences"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func AddAddress(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		var address Address
		if err := c.ShouldBindJSON(&address); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		address.UserID = userID
		if err := db.Create(&address).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
			return
		}

		c.JSON(http.StatusCreated, address)
	}
}

func ListAddresses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		var addresses []Address
		if err := db.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch addresses"})
			return
		}
		c.JSON(http.StatusOK, addresses)
	}
}

// RequestPasswordReset handles the password reset request
func RequestPasswordReset(db *gorm.DB, emailService *EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RequestPasswordResetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find user by email
		var user User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			// Don't reveal if email exists or not for security
			c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive a password reset link"})
			return
		}

		// Generate reset token
		if err := user.GeneratePasswordResetToken(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
			return
		}

		// Save token to database
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
			return
		}

		// Send reset email
		if err := emailService.SendPasswordResetEmail(user.Email, user.PasswordResetToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive a password reset link"})
	}
}

// ResetPassword handles the password reset
func ResetPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find user by reset token
		var user User
		if err := db.Where("password_reset_token = ?", req.Token).First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		// Validate token
		if !user.IsResetTokenValid(req.Token) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		// Update password
		user.Password = req.Password
		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Clear reset token
		user.ClearResetToken()

		// Save changes
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
	}
}
