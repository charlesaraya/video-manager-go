package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/api"
)

func main() {
	cfg, err := api.Load()
	if err != nil {
		log.Fatal(fmt.Errorf("error loading api config: %w", err))
	}
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + cfg.Port,
	}
	// 2. Set up handlers
	mux.Handle("/", api.AppHandler(cfg))

	mux.HandleFunc("POST /api/users", api.CreateUserHandler(cfg))
	mux.HandleFunc("POST /api/login", api.LoginHandler(cfg))
	mux.HandleFunc("POST /api/refresh", api.RefreshTokenHandler(cfg))
	mux.HandleFunc("POST /api/revoke", api.RevokeTokenHandler(cfg))

	// 3. Start server
	log.Printf("Serving: http://localhost:%s/\n", cfg.Port)
	server.ListenAndServe()
}
