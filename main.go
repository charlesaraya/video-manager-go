package main

import (
	"log"
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/api"
)

func main() {
	apiCfg, err := api.Load()
	if err != nil {
		log.Fatal("error loading api config")
	}
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + apiCfg.Port,
	}
	// 2. Set up handlers
	mux.Handle("/", api.AppHandler(apiCfg))

	// 3. Start server
	log.Printf("Serving: http://localhost:%s/\n", apiCfg.Port)
	server.ListenAndServe()
}
