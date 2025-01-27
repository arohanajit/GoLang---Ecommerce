package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// connectWithRetry attempts to connect to the database with retries
func connectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)", i+1, maxRetries)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			log.Println("Successfully connected to database")
			return db, nil
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(5 * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
}

// setupDatabase handles database migrations and schema updates
func setupDatabase(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Add this before AutoMigrate
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	// Check if the products table exists
	if !db.Migrator().HasTable(&Product{}) {
		log.Println("Products table does not exist, creating it...")
		if err := db.AutoMigrate(&Product{}); err != nil {
			return fmt.Errorf("failed to create products table: %v", err)
		}

		// Create the array column properly for new table
		if err := db.Exec(`
			DO $$ 
			BEGIN
				ALTER TABLE products 
				ALTER COLUMN images SET DEFAULT '{}',
				ALTER COLUMN images TYPE text[] USING CASE 
					WHEN images IS NULL THEN '{}'::text[]
					ELSE images::text[]
				END;
			EXCEPTION WHEN others THEN
				NULL;
			END $$;
		`).Error; err != nil {
			log.Printf("Warning: Failed to alter images column: %v", err)
		}
	} else {
		// Table exists, just run migrations for any schema updates
		if err := db.AutoMigrate(&Product{}); err != nil {
			return fmt.Errorf("failed to migrate database: %v", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/products", ListProducts(db))
		v1.POST("/products", CreateProduct(db))
		v1.GET("/products/:id", GetProduct(db))
		v1.PUT("/products/:id", UpdateProduct(db))
		v1.DELETE("/products/:id", DeleteProduct(db))
	}

	return router
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database connection details
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Connect to the database with retries
	db, err := connectWithRetry(dsn, 5)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup database with migrations
	if err := setupDatabase(db); err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// Consul client config
	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if consulConfig.Address == "" {
		consulConfig.Address = "http://localhost:8500" // Default Consul address
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// Register Product Service with Consul
	productPort, _ := strconv.Atoi(os.Getenv("PORT"))
	registration := &api.AgentServiceRegistration{
		ID:      "product-service",
		Name:    "product-service",
		Port:    productPort,
		Address: "127.0.0.1",
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://127.0.0.1:%d/health", productPort),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"product", "api"},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Failed to register product service with Consul: %v", err)
	}

	// Deregister Product Service on exit
	defer consulClient.Agent().ServiceDeregister(registration.ID)

	// Initialize Gin router
	router := setupRouter(db)

	// Add this before setting up routes
	router.Use(gin.Logger())
	router.Use(func(c *gin.Context) {
		log.Printf("Product Service received request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Basic health check endpoint for the Product Service
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router.Run("0.0.0.0:" + port)
}
