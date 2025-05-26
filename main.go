package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/api"
)

func main() {
	apiCfg, err := api.Load()
	if err != nil {
		log.Fatal(fmt.Errorf("error loading api config: %w", err))
	}
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + apiCfg.Port,
	}
	// 2. Set up handlers
	mux.Handle("/", api.AppHandler(apiCfg))
	mux.HandleFunc("POST /api/users", api.CreateUserHandler(apiCfg))
	mux.HandleFunc("POST /api/login", api.LoginHandler(apiCfg))
	mux.HandleFunc("POST /api/refresh", api.RefreshTokenHandler(apiCfg))
	mux.HandleFunc("POST /api/revoke", api.RevokeTokenHandler(apiCfg))

	// 3. Start server
	log.Printf("Serving: http://localhost:%s/\n", apiCfg.Port)
	server.ListenAndServe()
}
