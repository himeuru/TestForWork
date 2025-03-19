package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	_ "testForWork/docs"
	"testForWork/internal/api"
	"testForWork/internal/config"
	"testForWork/internal/database"
	"time"
)

// @title Music API
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	log.Println("starting server...")
	log.Println("loading .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
		return
	}
	log.Println("Successfully loaded .env file")

	cfg := config.LoadConfig()

	log.Println("Loading database...")
	db, err := database.ConnectAndSetup(cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("Successfully connected to database")

	defer db.Close()
	defer log.Println("Database disconnected")

	handler := api.NewHandler(db, cfg.MusicAPI)
	
	server := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler.Routes(),
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
