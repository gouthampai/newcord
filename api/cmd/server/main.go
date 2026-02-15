package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"newcord/api/internal/db"
	"newcord/api/internal/handlers"
	"newcord/api/internal/websocket"
	"newcord/api/pkg/config"
)

func main() {
	cfg := config.Load()

	cassandraDB, err := db.NewCassandraDB(cfg.CassandraHosts, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer cassandraDB.Close()

	if err := cassandraDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()

	router := handlers.NewRouter(cassandraDB, wsHub, cfg.JWTSecret, cfg.AllowedOrigins)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	cassandraDB.Close()
	log.Println("Server exited")
}
