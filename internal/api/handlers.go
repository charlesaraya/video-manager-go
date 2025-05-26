package api

import "net/http"

func AppHandler(cfg *Config) http.Handler {
	return http.FileServer(http.Dir(cfg.AppDirPath))
}
