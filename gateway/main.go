package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
)

// validateUUID middleware checks if the :id parameter is a valid UUID
func validateUUID(errorCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.Next()
			return
		}

		if _, err := uuid.Parse(id); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid ID format",
				"details": "ID must be in UUID format",
				"code":    errorCode,
			})
			return
		}
		c.Next()
	}
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Get the PORT from the environment, or default to 8081
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("Error: PORT must be a number:", err)
	}
	if port == 0 {
		port = 8081
	}

	// Initialize Gin router
	r := gin.Default()

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

	// Register API Gateway with Consul
	registration := new(api.AgentServiceRegistration)
	registration.ID = "api-gateway"              // Unique ID for this instance
	registration.Name = "api-gateway"            // Service name
	registration.Port = port                     // API Gateway port
	registration.Address = os.Getenv("HOST_IP")  // Host IP or network interface
	registration.Check = &api.AgentServiceCheck{ // Simple health check
		HTTP:     fmt.Sprintf("http://gateway:%d/health", registration.Port),
		Interval: "10s",
		Timeout:  "1s",
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// Deregister API Gateway on exit
	defer consulClient.Agent().ServiceDeregister(registration.ID)

	// Basic health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Routes
	v1 := r.Group("/api/v1")
	{
		// Product routes
		products := v1.Group("/products")
		{
			products.GET("", proxyToService(consulClient, "product-service", "/api/v1/products"))
			products.POST("", proxyToService(consulClient, "product-service", "/api/v1/products"))
			products.GET("/:id", validateUUID("INVALID_PRODUCT_ID"), proxyToService(consulClient, "product-service", "/api/v1/products/:id"))
			products.PUT("/:id", validateUUID("INVALID_PRODUCT_ID"), proxyToService(consulClient, "product-service", "/api/v1/products/:id"))
			products.DELETE("/:id", validateUUID("INVALID_PRODUCT_ID"), proxyToService(consulClient, "product-service", "/api/v1/products/:id"))
		}

		// Order routes with UUID validation
		orders := v1.Group("/orders")
		{
			// Base endpoints
			orders.GET("", proxyToService(consulClient, "order-service", "/api/v1/orders"))
			orders.POST("", proxyToService(consulClient, "order-service", "/api/v1/orders"))

			// Endpoints with ID parameter
			orders.GET("/:id", validateUUID("INVALID_ORDER_ID"), proxyToService(consulClient, "order-service", "/api/v1/orders"))
			orders.PUT("/:id", validateUUID("INVALID_ORDER_ID"), proxyToService(consulClient, "order-service", "/api/v1/orders"))
			orders.DELETE("/:id", validateUUID("INVALID_ORDER_ID"), proxyToService(consulClient, "order-service", "/api/v1/orders"))
		}

		// User routes
		users := v1.Group("/users")
		{
			// Public routes
			users.POST("/register", proxyToService(consulClient, "user-service", "/register"))
			users.POST("/login", proxyToService(consulClient, "user-service", "/login"))
			users.POST("/forgot-password", proxyToService(consulClient, "user-service", "/forgot-password"))
			users.POST("/reset-password", proxyToService(consulClient, "user-service", "/reset-password"))

			// Protected routes
			users.GET("/profile", proxyToService(consulClient, "user-service", "/profile"))
			users.PUT("/profile", proxyToService(consulClient, "user-service", "/profile"))
			users.PUT("/profile/change-password", proxyToService(consulClient, "user-service", "/profile/change-password")) // Updated path
			users.DELETE("/profile", proxyToService(consulClient, "user-service", "/profile"))

			// Address routes
			users.POST("/addresses", proxyToService(consulClient, "user-service", "/addresses"))
			users.GET("/addresses", proxyToService(consulClient, "user-service", "/addresses"))
			users.PUT("/addresses/:id", proxyToService(consulClient, "user-service", "/addresses/:id"))
			users.DELETE("/addresses/:id", proxyToService(consulClient, "user-service", "/addresses/:id"))
			users.PUT("/addresses/:id/default", proxyToService(consulClient, "user-service", "/addresses"))
		}

		// Inventory routes
		inventory := v1.Group("/inventory")
		{
			inventory.POST("/items", proxyToService(consulClient, "inventory-service", "/api/v1/inventory/items"))
			inventory.GET("/items", proxyToService(consulClient, "inventory-service", "/api/v1/inventory/items"))
			inventory.GET("/items/:id", validateUUID("INVALID_INVENTORY_ID"), proxyToService(consulClient, "inventory-service", "/api/v1/inventory/items/:id"))
			inventory.PUT("/items/:id/stock", validateUUID("INVALID_INVENTORY_ID"), proxyToService(consulClient, "inventory-service", "/api/v1/inventory/items/:id/stock"))
			inventory.GET("/items/:id/transactions", validateUUID("INVALID_INVENTORY_ID"), proxyToService(consulClient, "inventory-service", "/api/v1/inventory/items/:id/transactions"))
		}

		// Payment routes
		payments := v1.Group("/payments")
		{
			payments.POST("", proxyToService(consulClient, "payment-service", "/api/v1/payments"))
			payments.GET("/:id", validateUUID("INVALID_PAYMENT_ID"), proxyToService(consulClient, "payment-service", "/api/v1/payments/:id"))
		}
	}

	// Configure graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Handle shutdown gracefully
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// serviceName - consul service name
// targetPath - path which will be forwarded to a service along with path parameters
func proxyToService(consulClient *api.Client, serviceName, targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Incoming request to gateway: %s %s", c.Request.Method, c.Request.URL.Path)

		services, _, err := consulClient.Health().Service(serviceName, "", true, nil)
		if err != nil {
			log.Printf("Error discovering service: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to discover service",
				"code":  "SERVICE_DISCOVERY_ERROR",
			})
			return
		}

		if len(services) == 0 {
			log.Printf("No healthy instances found for service: %s", serviceName)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service unavailable",
				"code":  "SERVICE_UNAVAILABLE",
			})
			return
		}

		service := services[0].Service
		targetURL := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", service.Address, service.Port),
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host

			// Preserve the full path from the gateway
			req.URL.Path = targetPath
			// Replace :id with the actual ID value
			if strings.Contains(targetPath, ":id") {
				id := c.Param("id")
				req.URL.Path = strings.ReplaceAll(req.URL.Path, ":id", id)
			}

			// Add any additional path segments from the original request
			if c.Param("path") != "" {
				req.URL.Path += c.Param("path")
			}

			req.URL.RawQuery = c.Request.URL.RawQuery
			req.Header.Set("X-Forwarded-Host", c.Request.Host)

			log.Printf("Forwarding request to: %s %s", req.Method, req.URL.String())
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Failed to proxy request",
				"code":    "PROXY_ERROR",
				"details": err.Error(),
			})
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// placeholderHandler creates a simple handler that returns a message indicating which service it's for.
func placeholderHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Response from API Gateway - " + serviceName,
		})
	}
}
