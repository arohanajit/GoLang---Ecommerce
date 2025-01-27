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
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Consul client
	consulClient, err := initConsulClient()
	if err != nil {
		log.Fatal("Failed to initialize Consul client:", err)
	}

	// Register service with Consul
	if err := registerService(consulClient); err != nil {
		log.Fatal("Failed to register service:", err)
	}

	// Initialize router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API routes
	v1 := r.Group("/api/v1/inventory")
	{
		v1.POST("/items", CreateInventoryItem(db))
		v1.GET("/items", ListInventory(db))
		v1.GET("/items/:id", GetInventoryItem(db))
		v1.PUT("/items/:id/stock", UpdateStock(db))
		v1.GET("/items/:id/transactions", GetTransactionHistory(db))
	}

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	r.Run("0.0.0.0:" + port)
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
		return nil, err
	}

	// Enable uuid-ossp extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	// Auto-migrate the schema
	if err := db.AutoMigrate(&InventoryItem{}, &InventoryTransaction{}, &StockAlert{}); err != nil {
		return nil, err
	}

	return db, nil
}

func initConsulClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if config.Address == "" {
		config.Address = "http://localhost:8500"
	}
	return api.NewClient(config)
}

func registerService(client *api.Client) error {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	registration := &api.AgentServiceRegistration{
		ID:      "inventory-service",
		Name:    "inventory-service",
		Port:    port,
		Address: "inventory-service",
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://inventory-service:%d/health", port),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"inventory", "api"},
	}
	return client.Agent().ServiceRegister(registration)
}
