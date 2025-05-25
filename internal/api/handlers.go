package api

import "net/http"

func AppHandler(apiCfg *ApiConfig) http.Handler {
	return http.FileServer(http.Dir(apiCfg.AppDirPath))
}
