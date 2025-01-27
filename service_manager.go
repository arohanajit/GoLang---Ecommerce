package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Service struct {
	Name    string
	Path    string
	Process *exec.Cmd
}

func main() {
	// Define services
	services := []Service{
		{Name: "consul", Path: "consul"},
		{Name: "gateway", Path: "gateway"},
		{Name: "user-service", Path: "services/user"},
		{Name: "product-service", Path: "services/product"},
		{Name: "order-service", Path: "services/order"},
		{Name: "inventory-service", Path: "services/inventory"},
		{Name: "payment-service", Path: "services/payment"},
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create WaitGroup for services
	var wg sync.WaitGroup

	// Start services
	for i := range services {
		wg.Add(1)
		go func(s *Service) {
			defer wg.Done()
			startService(ctx, s)
		}(&services[i])
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down services...")
	cancel()

	// Wait for all services to shut down with timeout
	shutdownChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownChan)
	}()

	select {
	case <-shutdownChan:
		log.Println("All services shut down gracefully")
	case <-time.After(30 * time.Second):
		log.Println("Some services did not shut down gracefully")
	}
}

func startService(ctx context.Context, s *Service) {
	for {
		select {
		case <-ctx.Done():
			if s.Process != nil && s.Process.Process != nil {
				log.Printf("Stopping %s...\n", s.Name)
				if err := s.Process.Process.Signal(syscall.SIGTERM); err != nil {
					log.Printf("Error stopping %s: %v\n", s.Name, err)
					s.Process.Process.Kill()
				}
			}
			return
		default:
			log.Printf("Starting %s...\n", s.Name)

			// Special handling for consul
			if s.Name == "consul" {
				s.Process = exec.Command("consul", "agent", "-dev")
			} else {
				s.Process = exec.Command("go", "run", ".")
			}
			s.Process.Dir = s.Path

			// Set up environment variables
			s.Process.Env = os.Environ()

			// Redirect output
			s.Process.Stdout = os.Stdout
			s.Process.Stderr = os.Stderr

			if err := s.Process.Start(); err != nil {
				log.Printf("Error starting %s: %v\n", s.Name, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Wait for process to exit
			if err := s.Process.Wait(); err != nil {
				if ctx.Err() == nil {
					log.Printf("%s exited with error: %v\n", s.Name, err)
					time.Sleep(5 * time.Second)
				}
			}

			if ctx.Err() != nil {
				return
			}
		}
	}
}
