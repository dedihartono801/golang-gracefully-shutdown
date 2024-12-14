package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dating-service/database"
	"dating-service/pkg/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Setup environment variables and initialize database connection
	config := config.SetupEnvFile()
	db := database.InitPostgres(config)

	app := fiber.New() // Initialize Fiber web framework

	// Start the server in a separate goroutine
	go func() {
		if err := app.Listen(":5004"); err != nil {
			log.Fatalf("listen: %s", err) // Log an error and exit if the server fails to start
		}
	}()

	// Channel to listen for OS signals (e.g., SIGINT, SIGTERM, SIGHUP)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Block until a termination signal is received
	<-quit
	log.Println("Shutdown signal received, shutting down server...")

	// Set a timeout for the shutdown process
	timeoutFunc := time.AfterFunc(10*time.Second, func() {
		log.Printf("timeout %d s has been elapsed, force exit", 10*time.Second) // If the timeout is reached, log a message and forcefully exit
		os.Exit(0)                                                              // Immediately terminate the process
	})
	defer timeoutFunc.Stop() // Ensure the timeout timer is stopped if shutdown completes in time

	// Attempt to gracefully shut down the Fiber server
	if err := app.Shutdown(); err != nil {
		log.Fatal(err) // Log and exit if the server fails to shut down
	}
	log.Println("Server shutdown completed")

	// Close the database connection
	if err := db.Close(); err != nil {
		log.Fatal(err) // Log and exit if the database fails to close
	}
	log.Println("Database connection closed")

	// Indicate that the server has exited gracefully
	log.Println("Server exited gracefully")
}
