package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateOrder handles POST /api/v1/orders
func CreateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newOrder Order
		if err := c.ShouldBindJSON(&newOrder); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate that each order item has a product ID
		for _, item := range newOrder.OrderItems {
			if item.ProductID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required for each order item"})
				return
			}
		}

		if err := db.Create(&newOrder).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, newOrder)
	}
}

// ListOrders handles GET /api/v1/orders
func ListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []Order
		if err := db.Preload("OrderItems").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

// GetOrder handles GET /api/v1/orders/:id
func GetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var order Order
		if err := db.Preload("OrderItems").Where("id = ?", id).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// UpdateOrder handles PUT /api/v1/orders/:id
func UpdateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the existing order
		var order Order
		if err := db.Preload("OrderItems").First(&order, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Bind the updated order data from the request
		var updatedOrder Order
		if err := c.ShouldBindJSON(&updatedOrder); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Delete existing order items
		if err := db.Delete(order.OrderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old order items"})
			return
		}

		// Update order fields
		order.Status = updatedOrder.Status
		order.PaymentStatus = updatedOrder.PaymentStatus
		order.TotalAmount = updatedOrder.TotalAmount

		// Recreate order items from the updated order
		for _, item := range updatedOrder.OrderItems {
			order.OrderItems = append(order.OrderItems, OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
		}

		// Save the updated order and new order items
		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// DeleteOrder handles DELETE /api/v1/orders/:id
func DeleteOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Find the order with its items
		var order Order
		if err := db.Preload("OrderItems").First(&order, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Delete each order item associated with the order
		for _, item := range order.OrderItems {
			if err := db.Delete(&item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order items"})
				return
			}
		}

		// Delete the order
		if err := db.Delete(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order and associated items deleted"})
	}
}
