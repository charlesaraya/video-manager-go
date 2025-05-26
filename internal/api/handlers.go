package api

import "net/http"

func AppHandler(cfg *Config) http.Handler {
	return http.FileServer(http.Dir(cfg.AppDirPath))
}

func ResetHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if cfg.Platform != AllowedPlatform {
			http.Error(res, "reset is only allowed in dev environment.", http.StatusForbidden)
			return
		}
		if err := cfg.DB.DeleteAllUsers(req.Context()); err != nil {
			http.Error(res, "failed to reset 'users' table", http.StatusInternalServerError)
			return
		}
		if err := cfg.DB.DeleteAllRefreshTokens(req.Context()); err != nil {
			http.Error(res, "failed to reset 'referesh_tokens' table", http.StatusInternalServerError)
			return
		}
		if err := cfg.DB.DeleteAllVideos(req.Context()); err != nil {
			http.Error(res, "failed to reset 'videos' table", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
