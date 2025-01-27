package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	UserID string             `json:"user_id" binding:"required"`
	Items  []OrderItemRequest `json:"items" binding:"required"` // Use a DTO for items
}

type OrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

func CreateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": "Please check your request body format",
				"code":    "INVALID_REQUEST",
			})
			return
		}

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID format",
				"details": "User ID must be in UUID format",
				"code":    "INVALID_USER_ID",
			})
			return
		}

		var totalAmount float64
		var orderItems []OrderItem

		// Get product details and calculate total amount
		for _, item := range req.Items {
			productURL := fmt.Sprintf("%s/products/%s", os.Getenv("PRODUCT_SERVICE_URL"), item.ProductID)
			resp, err := http.Get(productURL)
			if err != nil || resp.StatusCode != http.StatusOK {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid product ID",
					"details": fmt.Sprintf("Product with ID %s not found", item.ProductID),
					"code":    "INVALID_PRODUCT_ID",
				})
				return
			}
			defer resp.Body.Close()

			var productDetails struct {
				Price float64 `json:"price"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&productDetails); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to decode product details",
					"details": err.Error(),
					"code":    "PRODUCT_DECODE_ERROR",
				})
				return
			}

			orderItems = append(orderItems, OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     productDetails.Price,
			})
			totalAmount += productDetails.Price * float64(item.Quantity)
		}

		newOrder := Order{
			UserID:      userID,
			TotalAmount: totalAmount,
			Status:      "pending",
			OrderItems:  orderItems,
		}

		if err := db.Create(&newOrder).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create order",
				"details": "An internal error occurred while creating the order",
				"code":    "ORDER_CREATION_FAILED",
			})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch orders",
				"details": "An internal error occurred while fetching orders",
				"code":    "FETCH_ORDERS_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

// GetOrder handles GET /api/v1/orders/:id
func GetOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		// Parse UUID
		if _, err := uuid.Parse(orderID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order ID format",
				"details": "Order ID must be in UUID format",
				"code":    "INVALID_ORDER_ID",
			})
			return
		}

		var order Order
		if err := db.Preload("OrderItems").First(&order, "id = ?", orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Order not found",
					"details": fmt.Sprintf("Order with ID %s not found", orderID),
					"code":    "ORDER_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch order",
				"details": "An internal error occurred",
				"code":    "FETCH_ORDER_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// UpdateOrder handles PUT /api/v1/orders/:id
func UpdateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order ID format",
				"details": "Order ID must be in UUID format",
				"code":    "INVALID_ORDER_ID",
			})
			return
		}

		// Find the existing order
		var order Order
		if err := db.Preload("OrderItems").First(&order, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Order not found",
					"details": "The requested order does not exist",
					"code":    "ORDER_NOT_FOUND",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch order",
					"details": "An internal error occurred while fetching the order",
					"code":    "FETCH_ORDER_FAILED",
				})
			}
			return
		}

		// Bind the updated order data
		var updatedOrder Order
		if err := c.ShouldBindJSON(&updatedOrder); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order data",
				"details": "Please check your order data format",
				"code":    "INVALID_ORDER_DATA",
			})
			return
		}

		// Delete existing order items
		if err := db.Delete(order.OrderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update order",
				"details": "An error occurred while updating order items",
				"code":    "UPDATE_ORDER_FAILED",
			})
			return
		}

		// Update order fields
		order.Status = updatedOrder.Status
		order.PaymentStatus = updatedOrder.PaymentStatus
		order.TotalAmount = updatedOrder.TotalAmount

		// Recreate order items
		for _, item := range updatedOrder.OrderItems {
			order.OrderItems = append(order.OrderItems, OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
		}

		// Save the updated order
		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update order",
				"details": "An internal error occurred while saving the order",
				"code":    "UPDATE_ORDER_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// DeleteOrder handles DELETE /api/v1/orders/:id
func DeleteOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order ID format",
				"details": "Order ID must be in UUID format",
				"code":    "INVALID_ORDER_ID",
			})
			return
		}

		// Find the order with its items
		var order Order
		if err := db.Preload("OrderItems").First(&order, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Order not found",
					"details": "The requested order does not exist",
					"code":    "ORDER_NOT_FOUND",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch order",
					"details": "An internal error occurred while fetching the order",
					"code":    "FETCH_ORDER_FAILED",
				})
			}
			return
		}

		// Delete order items
		if err := db.Delete(order.OrderItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete order",
				"details": "An error occurred while deleting order items",
				"code":    "DELETE_ORDER_FAILED",
			})
			return
		}

		// Delete the order
		if err := db.Delete(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete order",
				"details": "An internal error occurred while deleting the order",
				"code":    "DELETE_ORDER_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order and associated items deleted successfully",
			"code":    "ORDER_DELETED",
		})
	}
}
