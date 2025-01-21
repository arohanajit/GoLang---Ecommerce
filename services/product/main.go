package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

	log.Printf("Attempting to connect to database with DSN: %s", dsn)

	// Connect to the database with debug logging
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Add SQL query logging
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Successfully connected to database")

	// Add debug logging for migrations
	log.Println("Running database migrations...")

	// Drop existing table to ensure clean state
	if err := db.Migrator().DropTable(&Product{}); err != nil {
		log.Fatal("Failed to drop table:", err)
	}

	// Create the table
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create the array column properly
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

	log.Println("Database migrations completed successfully")

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
		ID:      "product-service-" + os.Getenv("HOST_IP"),
		Name:    "product-service",
		Port:    productPort,
		Address: os.Getenv("HOST_IP"),
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", os.Getenv("HOST_IP"), productPort),
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
	router := gin.Default()

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

	// Setup routes
	v1 := router.Group("/api/v1/products")
	{
		v1.GET("", func(c *gin.Context) {
			log.Printf("Handling GET request for products")
			ListProducts(db)(c)
		})
		v1.POST("", func(c *gin.Context) {
			log.Printf("Handling POST request for products")
			CreateProduct(db)(c)
		})
		v1.GET("/:id", GetProduct(db))
		v1.PUT("/:id", UpdateProduct(db))
		v1.DELETE("/:id", DeleteProduct(db))
	}

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router.Run(":" + port)
}
