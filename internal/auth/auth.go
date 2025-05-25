package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	TokenTypeAccess      string = "video-manager-access"
	ErrMissingAuthHeader string = "missing authorization header"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash from password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("failed to compare hash and password: %w", err)
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    TokenTypeAccess,
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token secret: %w", err)
	}
	return signedToken, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse with claims: %w", err)
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get issuer from claims: %w", err)
	}
	if issuer != TokenTypeAccess {
		return uuid.Nil, errors.New("invalid issuer")
	}
	expirationTime, err := token.Claims.GetExpirationTime()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get expiration time from claims: %w", err)
	}
	if expirationTime.Time.Before(time.Now()) {
		return uuid.Nil, fmt.Errorf("failed to get expiration time from claims: %w", jwt.ErrTokenExpired)
	}
	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get subject from claims: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
	}
	return userUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeaderValue := headers.Get("Authorization")
	if authHeaderValue == "" {
		return "", errors.New(ErrMissingAuthHeader)
	}
	token := strings.TrimPrefix(authHeaderValue, "Bearer ")
	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key), nil
}

func GetApiKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", errors.New(ErrMissingAuthHeader)
	}
	apiKey = strings.TrimPrefix(apiKey, "ApiKey ")
	return apiKey, nil
}
