package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres" // Changed from "postgre" to "postgres"
	"gorm.io/gorm"
)

func validateUUID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.Next()
			return
		}

		if _, err := uuid.Parse(id); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid order ID format",
				"details": "Order ID must be in UUID format",
				"code":    "INVALID_ORDER_ID",
			})
			return
		}
		c.Next()
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	redisClient, err := redis.NewClient(
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
	)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize cache
	cache := cache.NewServiceCache(redisClient) // Use appropriate constructor for each service

	// Add rate limiting
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	router.Use(rateLimiter.Middleware())

	// Consul client config
	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if consulConfig.Address == "" {
		consulConfig.Address = "http://localhost:8500"
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// Register Order Service with Consul
	orderPort, _ := strconv.Atoi(os.Getenv("PORT"))
	registration := &api.AgentServiceRegistration{
		ID:      "order-service",
		Name:    "order-service",
		Port:    orderPort,
		Address: "order-service",
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://order-service:%d/health", orderPort),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"order", "api"},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Failed to register order service with Consul: %v", err)
	}

	// Deregister Order Service on exit
	defer consulClient.Agent().ServiceDeregister(registration.ID)

	// Initialize Gin router
	router := gin.Default()
	router.Use(gin.Logger())

	// Basic health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Setup routes
	v1 := router.Group("/api/v1/orders")
	{
		v1.GET("", ListOrders(db))
		v1.POST("", CreateOrder(db))
		v1.GET("/:id", validateUUID(), GetOrder(db))
		v1.PUT("/:id", validateUUID(), UpdateOrder(db))
		v1.DELETE("/:id", validateUUID(), DeleteOrder(db))
	}

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	router.Run("0.0.0.0:" + port)
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Auto-migrate the Order and OrderItem models
	if err := db.AutoMigrate(&Order{}, &OrderItem{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
