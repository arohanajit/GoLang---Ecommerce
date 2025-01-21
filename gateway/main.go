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
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
)

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
		HTTP:     fmt.Sprintf("http://%s:%d/health", registration.Address, registration.Port),
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
			products.Any("", proxyToService(consulClient, "product-service", "/api/v1/products"))
			products.Any("/*path", proxyToService(consulClient, "product-service", "/api/v1/products"))
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			orders.Any("", proxyToService(consulClient, "order-service", "/api/v1/orders"))
			orders.Any("/*path", proxyToService(consulClient, "order-service", "/api/v1/orders"))
		}

		// User routes (placeholder)
		users := v1.Group("/users")
		{
			users.POST("/register", placeholderHandler("User Service - Register"))
			users.POST("/login", placeholderHandler("User Service - Login"))
			users.GET("/me", placeholderHandler("User Service - Profile"))
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

		// Discover service instances from Consul
		services, _, err := consulClient.Health().Service(serviceName, "", true, nil)
		if err != nil {
			log.Printf("Error discovering service: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover service"})
			return
		}

		if len(services) == 0 {
			log.Printf("No healthy instances found for service: %s", serviceName)
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Select the first healthy service instance (implement load balancing logic here if needed)
		service := services[0].Service

		// Create a new URL for the target service
		targetURL := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", service.Address, service.Port),
		}

		// Log the forwarding details
		log.Printf("Forwarding to service: %s at %s", serviceName, targetURL.String())

		// Create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Modify the request director to adjust the request URL
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = c.Request.URL.Path

			// Retain query parameters
			req.URL.RawQuery = c.Request.URL.RawQuery

			// Set X-Forwarded-Host header
			req.Header.Set("X-Forwarded-Host", c.Request.Host)
		}

		// Error handling for the reverse proxy
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Error proxying to service %s: %v", serviceName, err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy to service"})
		}

		// Serve the request using the reverse proxy
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
