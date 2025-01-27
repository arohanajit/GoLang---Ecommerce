package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order represents the order data model
type Order struct {
	gorm.Model
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID        uuid.UUID   `gorm:"type:uuid;not null"`
	OrderItems    []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
	TotalAmount   float64     `json:"total_amount" gorm:"not null"`
	Status        string      `json:"status" gorm:"default:'pending'"`
	PaymentStatus string      `json:"payment_status" gorm:"default:'unpaid'"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	gorm.Model
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"`
}
