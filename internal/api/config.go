package api

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/charlesaraya/video-manager-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const (
	AllowedPlatform   string = "dev"
	MimeTypeImagePNG  string = "image/png"
	MimeTypeImageJPEG string = "image/png"
	MimeTypeVideo     string = "video/mp4"
	MimeTypeAudio     string = "audio/mp3"
	MimeTypeText      string = "text/html"
)

type Config struct {
	DB            *database.Queries
	Platform      string
	TokenSecret   string
	Port          string
	AssetsDirPath string
	AppDirPath    string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("failed to set DB_PATH environment variable")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.New("error opening the database")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		return nil, fmt.Errorf("failed to set PLATFORM environment variable")
	}
	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		return nil, fmt.Errorf("failed to set TOKEN_SECRET environment variable")
	}
	port := os.Getenv("PORT")
	if tokenSecret == "" {
		return nil, fmt.Errorf("failed to set PORT environment variable")
	}
	appDirPath := os.Getenv("APP_DIR_PATH")
	if tokenSecret == "" {
		return nil, fmt.Errorf("failed to set APP_DIR_PATH environment variable")
	}
	assetsDirPath := os.Getenv("ASSETS_DIR_PATH")
	if tokenSecret == "" {
		return nil, fmt.Errorf("failed to set ASSETS_DIR_PATH environment variable")
	}
	return &Config{
		DB:            database.New(db),
		Platform:      platform,
		TokenSecret:   tokenSecret,
		Port:          port,
		AppDirPath:    appDirPath,
		AssetsDirPath: assetsDirPath,
	}, nil
}
