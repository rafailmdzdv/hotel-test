package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hotel-test/internal/config"
	"hotel-test/internal/handler"
	"hotel-test/internal/repository/postgres"
	"hotel-test/internal/service"
	"hotel-test/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Connecting to PostgreSQL...")
	dbPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbPool.Close()

	log.Println("Running database migrations...")
	if err := migrations.Run(ctx, dbPool); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	userRepo := postgres.NewUserRepository(dbPool)
	guestRepo := postgres.NewGuestRepository(dbPool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTDuration, cfg.BcryptCost)
	guestService := service.NewGuestService(guestRepo)

	router := handler.NewRouter(authService, guestService)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Hotel Guest Notebook server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-shutdownChan
	log.Println("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped successfully")
}
