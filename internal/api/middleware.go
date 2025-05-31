package api

import (
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/auth"
	"github.com/google/uuid"
)

func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=3600")
		next.ServeHTTP(w, r)
	})
}

func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(cfg *Config, handler func(*Config, uuid.UUID) http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		jwt, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		// Call the original handler with injected userUUID
		handler(cfg, userUUID).ServeHTTP(res, req)
	}
}
