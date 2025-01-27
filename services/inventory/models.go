// models.go
package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryItem struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null"`
	Quantity        int       `gorm:"not null"`
	ReorderPoint    int       `gorm:"not null"`
	ReorderQuantity int       `gorm:"not null"`
	Location        string    `gorm:"not null"`
	Status          string    `gorm:"not null;default:'available'"`
	BatchNumber     string    `gorm:"index"`
	ExpiryDate      *time.Time
	LastStockCheck  time.Time
	Notes           string
}

type InventoryTransaction struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ItemID        uuid.UUID `gorm:"type:uuid;not null"`
	Type          string    `gorm:"not null"` // "received", "shipped", "adjusted", "damaged"
	Quantity      int       `gorm:"not null"`
	PreviousStock int       `gorm:"not null"`
	NewStock      int       `gorm:"not null"`
	Reference     string    `gorm:"index"` // Order ID, Purchase Order ID, etc.
	Notes         string
	Item          InventoryItem `gorm:"foreignKey:ItemID"`
}

type StockAlert struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ItemID     uuid.UUID `gorm:"type:uuid;not null"`
	Type       string    `gorm:"not null"` // "low_stock", "expired", "damaged"
	Message    string    `gorm:"not null"`
	Status     string    `gorm:"not null;default:'pending'"`
	ResolvedAt *time.Time
	Item       InventoryItem `gorm:"foreignKey:ItemID"`
}
