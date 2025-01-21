package main

import (
	"encoding/base64"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email                   string                  `gorm:"uniqueIndex;not null" json:"email"`
	Password                string                  `gorm:"not null" json:"-"`
	FirstName               string                  `json:"first_name"`
	LastName                string                  `json:"last_name"`
	PhoneNumber             string                  `json:"phone_number"`
	Role                    string                  `gorm:"default:'user'" json:"role"`
	DateOfBirth             *time.Time              `json:"date_of_birth"`
	ProfilePicture          string                  `json:"profile_picture"`
	Bio                     string                  `json:"bio"`
	PreferredLanguage       string                  `gorm:"default:'en'" json:"preferred_language"`
	NotificationPreferences NotificationPreferences `gorm:"embedded" json:"notification_preferences"`
	Addresses               []Address               `json:"addresses"`
	PasswordResetToken      string                  `gorm:"unique" json:"-"`
	ResetTokenExpiresAt     *time.Time              `json:"-"`
}

type NotificationPreferences struct {
	EmailNotifications bool `gorm:"default:true" json:"email_notifications"`
	OrderUpdates       bool `gorm:"default:true" json:"order_updates"`
	PromotionalEmails  bool `gorm:"default:true" json:"promotional_emails"`
	SecurityAlerts     bool `gorm:"default:true" json:"security_alerts"`
}

type Address struct {
	gorm.Model
	UserID     string `gorm:"type:uuid;not null" json:"user_id"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	IsDefault  bool   `gorm:"default:false" json:"is_default"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword checks if the provided password matches the hash
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// GeneratePasswordResetToken creates a new password reset token
func (u *User) GeneratePasswordResetToken() error {
	// Generate a random token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}

	// Set token and expiration
	u.PasswordResetToken = base64.URLEncoding.EncodeToString(token)
	expiresAt := time.Now().Add(15 * time.Minute)
	u.ResetTokenExpiresAt = &expiresAt

	return nil
}

// IsResetTokenValid checks if the reset token is valid and not expired
func (u *User) IsResetTokenValid(token string) bool {
	if u.PasswordResetToken == "" || u.ResetTokenExpiresAt == nil {
		return false
	}

	if u.PasswordResetToken != token {
		return false
	}

	return time.Now().Before(*u.ResetTokenExpiresAt)
}

// ClearResetToken clears the password reset token and expiration
func (u *User) ClearResetToken() {
	u.PasswordResetToken = ""
	u.ResetTokenExpiresAt = nil
}
