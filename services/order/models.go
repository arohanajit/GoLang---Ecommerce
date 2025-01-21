package main

import (
	"time"

	"gorm.io/gorm"
)

// Order represents the order data model
type Order struct {
	gorm.Model
	ID            string      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID        string      `json:"user_id" gorm:"type:uuid;not null"`
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
	OrderID   string  `json:"order_id" gorm:"type:uuid;not null"`
	ProductID string  `json:"product_id" gorm:"type:uuid;not null"`
	Quantity  int     `json:"quantity" gorm:"not null"`
	Price     float64 `json:"price" gorm:"not null"`
}
