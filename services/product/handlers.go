package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ListProducts handles GET /api/products
func ListProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []Product
		if err := db.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

// CreateProduct handles POST /api/products
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Product
		if err := c.ShouldBindJSON(&product); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure images is initialized
		if product.Images == nil {
			product.Images = pq.StringArray{}
		}

		log.Printf("Attempting to create product: %+v", product)

		// Use a transaction to create the product
		tx := db.Begin()
		if err := tx.Create(&product).Error; err != nil {
			tx.Rollback()
			log.Printf("Database error creating product: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update images separately if needed
		if len(product.Images) > 0 {
			if err := tx.Model(&product).Update("images", product.Images).Error; err != nil {
				tx.Rollback()
				log.Printf("Error updating images: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		tx.Commit()
		log.Printf("Successfully created product with ID: %v", product.ID)
		c.JSON(http.StatusCreated, product)
	}
}

// GetProduct handles GET /api/products/:id
func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid product ID format",
				"details": "Product ID must be in UUID format",
				"code":    "INVALID_PRODUCT_ID",
			})
			return
		}

		var product Product
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Product not found",
					"details": fmt.Sprintf("Product with ID %s not found", id),
					"code":    "PRODUCT_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch product",
				"details": "An internal error occurred",
				"code":    "FETCH_PRODUCT_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

// UpdateProduct handles PUT /api/products/:id
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid product ID format",
				"details": "Product ID must be in UUID format",
				"code":    "INVALID_PRODUCT_ID",
			})
			return
		}

		var product Product
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Product not found",
					"details": fmt.Sprintf("Product with ID %s not found", id),
					"code":    "PRODUCT_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch product",
				"details": "An internal error occurred",
				"code":    "FETCH_PRODUCT_FAILED",
			})
			return
		}

		var updatedProduct Product
		if err := c.ShouldBindJSON(&updatedProduct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Only update specified fields
		if err := db.Model(&product).Updates(updatedProduct).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

// DeleteProduct handles DELETE /api/products/:id
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid product ID format",
				"details": "Product ID must be in UUID format",
				"code":    "INVALID_PRODUCT_ID",
			})
			return
		}

		if err := db.Delete(&Product{}, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Product not found",
					"details": fmt.Sprintf("Product with ID %s not found", id),
					"code":    "PRODUCT_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete product",
				"details": "An internal error occurred",
				"code":    "DELETE_PRODUCT_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
	}
}
