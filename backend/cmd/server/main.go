package main

import (
	"log"
	"net/http"
	"time"

	"bank-repayment-record/backend/internal/config"
	"bank-repayment-record/backend/internal/httpapi"
	"bank-repayment-record/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("load timezone: %v", err)
	}

	store, err := storage.OpenMySQL(cfg.MySQL, loc)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			log.Printf("close database: %v", closeErr)
		}
	}()

	server := httpapi.NewServer(store, loc, httpapi.ServerOptions{
		FrontendDistDir:    cfg.FrontendDistDir,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		Auth:               cfg.Auth,
		BeianText:          cfg.BeianText,
	})
	addr := ":" + cfg.Port
	log.Printf("backend listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
