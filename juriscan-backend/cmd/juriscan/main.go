package main

import (
	"log"

	"juriscan-backend/internal/app"
)

func main() {
	cfg := app.LoadConfigFromEnv()

	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("app init: %v", err)
	}

	log.Printf("juriscan backend listening on %s (env=%s)", cfg.HTTPAddr, cfg.AppEnv)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
