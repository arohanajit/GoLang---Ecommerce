// handlers.go
package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateInventoryItemRequest struct {
	ProductID       uuid.UUID `json:"product_id" binding:"required"`
	Quantity        int       `json:"quantity" binding:"required,min=0"`
	ReorderPoint    int       `json:"reorder_point" binding:"required,min=0"`
	ReorderQuantity int       `json:"reorder_quantity" binding:"required,min=1"`
	Location        string    `json:"location" binding:"required"`
	BatchNumber     string    `json:"batch_number"`
	ExpiryDate      string    `json:"expiry_date"`
	Notes           string    `json:"notes"`
}

func CreateInventoryItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateInventoryItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var expiryDate *time.Time
		if req.ExpiryDate != "" {
			parsedTime, err := time.Parse("2006-01-02", req.ExpiryDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expiry date format"})
				return
			}
			expiryDate = &parsedTime
		}

		item := InventoryItem{
			ProductID:       req.ProductID,
			Quantity:        req.Quantity,
			ReorderPoint:    req.ReorderPoint,
			ReorderQuantity: req.ReorderQuantity,
			Location:        req.Location,
			BatchNumber:     req.BatchNumber,
			ExpiryDate:      expiryDate,
			LastStockCheck:  time.Now(),
			Notes:           req.Notes,
		}

		if err := db.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inventory item"})
			return
		}

		// Check if we need to create a low stock alert
		if item.Quantity <= item.ReorderPoint {
			alert := StockAlert{
				ItemID:  item.ID,
				Type:    "low_stock",
				Message: "Stock level is below reorder point",
				Status:  "pending",
			}
			db.Create(&alert)
		}

		c.JSON(http.StatusCreated, item)
	}
}

type UpdateStockRequest struct {
	Quantity  int    `json:"quantity" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=received shipped adjusted damaged"`
	Reference string `json:"reference"`
	Notes     string `json:"notes"`
}

func UpdateStock(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateStockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate quantity
		if req.Type == "shipped" && req.Quantity < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid quantity: shipped quantity cannot be negative",
				"code":  "INVALID_QUANTITY",
			})
			return
		}

		id := c.Param("id")
		itemID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		var item InventoryItem
		if err := db.First(&item, "id = ?", itemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Calculate new stock level
		var newStock int
		switch req.Type {
		case "received":
			newStock = item.Quantity + req.Quantity
		case "shipped":
			newStock = item.Quantity - req.Quantity
			if newStock < 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Insufficient stock",
					"code":  "INSUFFICIENT_STOCK",
				})
				return
			}
		case "adjusted":
			newStock = req.Quantity
		case "damaged":
			newStock = item.Quantity - req.Quantity
			if newStock < 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid quantity: cannot damage more items than available",
					"code":  "INVALID_QUANTITY",
				})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction type"})
			return
		}

		// Start a transaction
		tx := db.Begin()

		previousStock := item.Quantity
		newStock = newStock

		// Create transaction record
		transaction := InventoryTransaction{
			ItemID:        item.ID,
			Type:          req.Type,
			Quantity:      req.Quantity,
			PreviousStock: previousStock,
			NewStock:      newStock,
			Reference:     req.Reference,
			Notes:         req.Notes,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
			return
		}

		// Update item stock
		if err := tx.Model(&item).Update("quantity", newStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}

		// Check if we need to create a low stock alert
		if newStock <= item.ReorderPoint {
			alert := StockAlert{
				ItemID:  item.ID,
				Type:    "low_stock",
				Message: "Stock level is below reorder point",
				Status:  "pending",
			}
			if err := tx.Create(&alert).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
				return
			}
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"previous_stock": previousStock,
			"new_stock":      newStock,
			"transaction":    transaction,
		})
	}
}

func GetInventoryItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var item InventoryItem
		if err := db.First(&item, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory item"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func ListInventory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []InventoryItem
		if err := db.Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory"})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func GetTransactionHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemID := c.Param("id")
		var transactions []InventoryTransaction
		if err := db.Where("item_id = ?", itemID).Find(&transactions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}
		c.JSON(http.StatusOK, transactions)
	}
}
