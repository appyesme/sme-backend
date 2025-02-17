package main

import (
	"fmt"
	"log"
	"net/http"
	"sme-backend/src/core/config"
	"sme-backend/src/core/database"
	"sme-backend/src/core/router"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func main() {
	// Set up the router
	app := chi.NewRouter()
	app.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}))

	// Set up environment and database
	config.SetupEnv()
	database.ConnectDB()

	// Set up routes
	router.SetupRoutes(app)

	// Create the HTTP server
	server := &http.Server{Addr: fmt.Sprintf(":%s", config.Config("SERVER_PORT")), Handler: app}
	fmt.Println("Server started at port:", config.Config("SERVER_PORT"))

	// Configure graceful shutdown
	database.ConfigureGracefulShutdown(server)
	// Start the HTTP server (blocking call)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
