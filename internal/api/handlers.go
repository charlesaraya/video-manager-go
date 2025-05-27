package api

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func Error(res http.ResponseWriter, msg string, code int) {
	errorPayload := errorResponse{
		Error: msg,
	}
	data, err := json.Marshal(errorPayload)
	if err != nil {
		http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
		return
	}
	res.WriteHeader(code)
	res.Header().Set("Content-Type", "application/json")
	res.Write(data)
}

func AppHandler(cfg *Config) http.Handler {
	return http.FileServer(http.Dir(cfg.AppDirPath))
}

func AssetsHandler(cfg *Config) http.Handler {
	return http.FileServer(http.Dir(cfg.AssetsDirPath))
}

func ResetHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if cfg.Platform != AllowedPlatform {
			Error(res, "reset is only allowed in dev environment.", http.StatusForbidden)
			return
		}
		if err := cfg.DB.DeleteAllUsers(req.Context()); err != nil {
			Error(res, "failed to reset 'users' table", http.StatusInternalServerError)
			return
		}
		if err := cfg.DB.DeleteAllRefreshTokens(req.Context()); err != nil {
			Error(res, "failed to reset 'referesh_tokens' table", http.StatusInternalServerError)
			return
		}
		if err := cfg.DB.DeleteAllVideos(req.Context()); err != nil {
			Error(res, "failed to reset 'videos' table", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
