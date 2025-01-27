package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreatePaymentRequest struct {
	OrderID       uuid.UUID `json:"order_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
}

func CreatePaymentHandler(db *gorm.DB, client *DummyClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid payment request",
				"details": err.Error(),
				"code":    "INVALID_PAYMENT_REQUEST",
			})
			return
		}

		payment := Payment{
			OrderID:       req.OrderID,
			Amount:        req.Amount,
			PaymentMethod: req.PaymentMethod,
			Status:        "pending",
		}

		// Process with dummy client
		if err := client.Process(&payment); err != nil {
			payment.Status = "failed"
			db.Create(&payment)
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":   "Payment processing failed",
				"details": err.Error(),
				"code":    "PAYMENT_FAILED",
			})
			return
		}

		payment.Status = "completed"
		db.Create(&payment)

		c.JSON(http.StatusCreated, payment)
	}
}

func (d *DummyClient) Process(p *Payment) error {
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() > d.SuccessRate {
		return errors.New("dummy payment processing failed")
	}
	p.TransactionID = uuid.New().String()
	return nil
}

func GetPaymentHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid payment ID format",
				"details": "Payment ID must be in UUID format",
				"code":    "INVALID_PAYMENT_ID",
			})
			return
		}

		var payment Payment
		if err := db.First(&payment, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Payment not found",
					"details": fmt.Sprintf("Payment with ID %s not found", id),
					"code":    "PAYMENT_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch payment",
				"details": "An internal error occurred",
				"code":    "FETCH_PAYMENT_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, payment)
	}
}
