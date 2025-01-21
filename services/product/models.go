package main

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model                 // This embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string         `json:"name" gorm:"not null" binding:"required"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null" binding:"required"`
	Stock       int64          `json:"stock" gorm:"not null" binding:"required"`
	CategoryID  string         `json:"category_id"`
	Images      pq.StringArray `json:"images" gorm:"type:text[];default:'{}'::text[]"`
}
