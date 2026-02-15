package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	router := handlers.NewRouter(cassandraDB, wsHub, cfg.JWTSecret)

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
}
