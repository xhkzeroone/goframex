package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.io/xhkzeroone/goframex/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize usecase: %v", err)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	if err := app.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	if app.Infrastructure.DB != nil {
		if err := app.Infrastructure.DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	log.Println("Server stopped")
}
