package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/api"
)

func main() {
	cfg, err := api.LoadConfig()
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

	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir(cfg.AssetsDirPath)))
	mux.Handle(cfg.AssetsBrowserURL, api.CacheMiddleware(assetsHandler))

	mux.HandleFunc("POST /api/users", api.CreateUserHandler(cfg))
	mux.HandleFunc("POST /api/login", api.LoginHandler(cfg))
	mux.HandleFunc("POST /api/refresh", api.RefreshTokenHandler(cfg))
	mux.HandleFunc("POST /api/revoke", api.RevokeTokenHandler(cfg))

	mux.HandleFunc("GET /api/videos", api.AuthMiddleware(cfg, api.GetAllVideosHandler))
	mux.HandleFunc("GET /api/videos/{videoID}", api.GetVideoHandler(cfg))
	mux.HandleFunc("POST /api/videos", api.AuthMiddleware(cfg, api.AddVideoHandler))
	mux.HandleFunc("DELETE /api/videos/{videoID}", api.AuthMiddleware(cfg, api.DeleteVideoHandler))
	mux.HandleFunc("UPDATE /api/videos/{videoID}", api.AuthMiddleware(cfg, api.UploadThumbnailHandler))
	mux.HandleFunc("POST /api/video_upload/{videoID}", api.AuthMiddleware(cfg, api.UploadVideosHandler))

	mux.HandleFunc("POST /admin/reset", api.ResetHandler(cfg))

	// 3. Start server
	log.Printf("Serving: http://localhost:%s/\n", cfg.Port)
	server.ListenAndServe()
}
