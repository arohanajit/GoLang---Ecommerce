package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null"`
	Amount        float64   `gorm:"not null"`
	Currency      string    `gorm:"not null;default:'USD'"`
	Status        string    `gorm:"not null;default:'pending'"`
	PaymentMethod string    `gorm:"not null"`
	TransactionID string
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type DummyClient struct {
	SuccessRate float64
}
