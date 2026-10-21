package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blindtag_api/internal/config"
	"blindtag_api/internal/database"
	"blindtag_api/internal/router"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	log.Printf("[BlindTag API] Starting server in %s mode on port %s...", cfg.Env, cfg.Port)

	// 2. Connect to Supabase PostgreSQL database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[BlindTag API] Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 3. Setup router & routes
	engine := router.SetupRouter(cfg, db)

	// 4. HTTP Server with timeout settings for Cloud Run
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start server in a goroutine
	go func() {
		log.Printf("[BlindTag API] 🚀 Server running and listening on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[BlindTag API] Server listen failed: %v", err)
		}
	}()

	// 6. Graceful shutdown on SIGINT / SIGTERM (required for Cloud Run)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[BlindTag API] 🛑 Shutdown signal received, shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[BlindTag API] Server forced to shutdown: %v", err)
	}

	log.Println("[BlindTag API] Server exited smoothly.")
}
