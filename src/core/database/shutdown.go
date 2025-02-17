package database

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ConfigureGracefulShutdown(server *http.Server) {
	// Signal channel to listen for termination signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for a signal
	go func() {
		<-sig
		log.Println("Signal received. Initiating graceful shutdown...")

		// Context with timeout for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Shutdown HTTP server
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		} else {
			log.Println("HTTP server shut down gracefully.")
		}

		// Close database connection
		if CONTEXT_API_DB != nil {
			api_db, err := CONTEXT_API_DB.DB()
			if err != nil {
				log.Printf("Error retrieving database connection: %v", err)
			} else if err := api_db.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			} else {
				log.Println("Database connection closed successfully.")
			}
		}

		log.Println("Shutdown process completed. Exiting...")
		os.Exit(0)
	}()
}
