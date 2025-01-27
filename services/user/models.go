package main

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `sql:"index" json:"-"`
	Email               string     `gorm:"uniqueIndex;not null" json:"email"`
	Password            string     `gorm:"not null" json:"-"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	PhoneNumber         string     `json:"phone_number"`
	Role                string     `gorm:"default:'user'" json:"role"`
	DateOfBirth         *time.Time `json:"date_of_birth"`
	ProfilePicture      string     `json:"profile_picture"`
	Bio                 string     `json:"bio"`
	PreferredLanguage   string     `gorm:"default:'en'" json:"preferred_language"`
	Addresses           []Address  `gorm:"constraint:OnDelete:CASCADE;" json:"addresses"`
	PasswordResetToken  string     `gorm:"index" json:"-"`
	ResetTokenExpiresAt *time.Time `json:"-"`
}

type Address struct {
	gorm.Model
	Street     string    `json:"street"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Country    string    `json:"country"`
	PostalCode string    `json:"postal_code"`
	UserID     uuid.UUID `json:"user_id"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;"`
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
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}
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
	return u.PasswordResetToken == token && time.Now().Before(*u.ResetTokenExpiresAt)
}

// ClearResetToken clears the password reset token and expiration
func (u *User) ClearResetToken() {
	u.PasswordResetToken = ""
	u.ResetTokenExpiresAt = nil
}

// GetIDString returns the string representation of the user's UUID
func (u *User) GetIDString() string {
	return u.ID.String()
}
