package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charlesaraya/video-manager-go/internal/auth"
	"github.com/charlesaraya/video-manager-go/internal/database"
	"github.com/google/uuid"
)

const (
	ErrDecodeRequestBody    string        = "failed to decode request body"
	ErrMarshalPayload       string        = "failed to marshal payload"
	MaxSessionDuration      time.Duration = time.Hour * 24
	MaxRefreshTokenDuration time.Duration = time.Hour * 24 * 60
)

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userPayload struct {
	User         database.User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func CreateUserHandler(cfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		params := loginPayload{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrDecodeRequestBody, http.StatusBadRequest)
			return
		}
		if params.Password == "" || params.Email == "" {
			http.Error(res, "Email and password are required", http.StatusBadRequest)
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := auth.CheckPasswordHash(hashedPassword, params.Password); err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		userUUID := uuid.New()
		userParams := database.CreateUserParams{
			ID:       userUUID.String(),
			Email:    params.Email,
			Password: hashedPassword,
		}
		user, err := cfg.DB.CreateUser(req.Context(), userParams)
		if err != nil {
			http.Error(res, "failed to create user", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func LoginHandler(cfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		params := loginPayload{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrDecodeRequestBody, http.StatusInternalServerError)
			return
		}
		user, err := cfg.DB.GetUserByEmail(req.Context(), params.Email)
		if err != nil {
			http.Error(res, "Incorrect email or password", http.StatusUnauthorized)
			return
		}
		if err := auth.CheckPasswordHash(user.Password, params.Password); err != nil {
			http.Error(res, "Incorrect email or password", http.StatusUnauthorized)
			return
		}
		userUUID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(res, "failed to parse uuid", http.StatusInternalServerError)
			return
		}
		jwt, err := auth.MakeJWT(userUUID, cfg.TokenSecret, MaxSessionDuration)
		if err != nil {
			http.Error(res, "failed to create access JWT", http.StatusInternalServerError)
			return
		}
		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			http.Error(res, "failed to create refresh token", http.StatusInternalServerError)
			return
		}
		refereshTokensParams := database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(MaxRefreshTokenDuration),
		}
		_, err = cfg.DB.CreateRefreshToken(req.Context(), refereshTokensParams)
		if err != nil {
			http.Error(res, "failed to create refresh token", http.StatusInternalServerError)
			return
		}
		payload := userPayload{
			User:         user,
			Token:        jwt,
			RefreshToken: refreshToken,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}
