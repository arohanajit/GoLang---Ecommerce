package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres" // Changed from "postgre" to "postgres"
	"gorm.io/gorm"
)

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
		ID:      "order-service-" + os.Getenv("HOST_IP"),
		Name:    "order-service",
		Port:    orderPort,
		Address: os.Getenv("HOST_IP"),
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", os.Getenv("HOST_IP"), orderPort),
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
		v1.GET("/:id", GetOrder(db))
		v1.PUT("/:id", UpdateOrder(db))
		v1.DELETE("/:id", DeleteOrder(db))
	}

	// Configure server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
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
