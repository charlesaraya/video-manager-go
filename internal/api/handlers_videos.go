package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/charlesaraya/video-manager-go/internal/auth"
	"github.com/charlesaraya/video-manager-go/internal/database"
	"github.com/google/uuid"
)

func AddVideoHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		jwt, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			http.Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		videoParams := database.CreateVideoParams{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&videoParams); err != nil {
			http.Error(res, ErrDecodeRequestBody, http.StatusInternalServerError)
			return
		}
		videoParams.ID = uuid.New().String()
		videoParams.UserID = userUUID.String()
		video, err := cfg.DB.CreateVideo(context.Background(), videoParams)
		if err != nil {
			http.Error(res, "failed to get videos", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(video)
		if err != nil {
			http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func GetVideoHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		jwt, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			http.Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		videoUUID, err := uuid.Parse(req.PathValue("videoID"))
		if err != nil {
			http.Error(res, "failed to parse video ID", http.StatusInternalServerError)
			return
		}
		video, err := cfg.DB.GetVideo(context.Background(), videoUUID.String())
		if err != nil {
			http.Error(res, "failed to get video", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(video)
		if err != nil {
			http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func GetAllVideosHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		jwt, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			http.Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		videos, err := cfg.DB.GetVideosByUser(context.Background(), userUUID.String())
		if err != nil {
			http.Error(res, "failed to get videos", http.StatusInternalServerError)
			return
		}
		videosPayload := []database.Video{}
		if len(videos) != 0 {
			videosPayload = videos
		}
		data, err := json.Marshal(videosPayload)
		if err != nil {
			http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func DeleteVideoHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, "failed to extract token", http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			http.Error(res, "failed to authorize request", http.StatusUnauthorized)
			return
		}
		deleteParams := database.DeleteVideoParams{
			ID:     req.PathValue("videoID"),
			UserID: userUUID.String(),
		}
		if err := cfg.DB.DeleteVideo(context.Background(), deleteParams); err != nil {
			http.Error(res, "failed to delete video", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
