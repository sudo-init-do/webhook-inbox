package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/you/webhook-inbox/internal/config"
	httpapi "github.com/you/webhook-inbox/internal/http"
	"github.com/you/webhook-inbox/internal/storage"
)

func main() {
	cfg := config.Load()

	store, err := storage.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer store.Close()

	// auto-migrate
	if err := store.Migrate(context.Background(), "migrations/001_init.sql"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := &httpapi.Handlers{
		Store:  store,
		Config: cfg,
	}
	r := httpapi.NewRouter(h)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("shut down cleanly")
}
