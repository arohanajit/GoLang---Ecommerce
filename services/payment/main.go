package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := initDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	consulClient, err := initConsul()
	if err != nil {
		log.Fatal("Consul connection failed:", err)
	}

	if err := registerService(consulClient); err != nil {
		log.Fatal("Service registration failed:", err)
	}

	router := gin.Default()
	paymentClient := &DummyClient{SuccessRate: 1.0}

	v1 := router.Group("/api/v1/payments")
	{
		v1.POST("", CreatePaymentHandler(db, paymentClient))
		v1.GET("/:id", GetPaymentHandler(db))
	}

	router.GET("/health", healthCheck)

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
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
		return nil, err
	}

	if err := db.AutoMigrate(&Payment{}); err != nil {
		return nil, err
	}

	return db, nil
}

func initConsul() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if config.Address == "" {
		config.Address = "http://localhost:8500"
	}
	return api.NewClient(config)
}

func registerService(client *api.Client) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	registration := &api.AgentServiceRegistration{
		ID:      "payment-service",
		Name:    "payment-service",
		Port:    8004,
		Address: "payment-service",
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://payment-service:%s/health", port),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"payment", "api"},
	}

	return client.Agent().ServiceRegister(registration)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}
