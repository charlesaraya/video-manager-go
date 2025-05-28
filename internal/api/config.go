package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	DB               *database.Queries
	Platform         string
	TokenSecret      string
	Port             string
	AssetsDirPath    string
	AssetsBrowserURL string
	AppDirPath       string
	S3BucketName     string
	S3BucketRegion   string
	S3Client         *s3.Client
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
	assetsBrowserURL := os.Getenv("ASSETS_BROWSER_URL")
	if tokenSecret == "" {
		return nil, fmt.Errorf("failed to set ASSETS_DIR_PATH environment variable")
	}
	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	if s3BucketName == "" {
		return nil, fmt.Errorf("failed to set S3_BUCKET environment variable")
	}
	s3BucketRegion := os.Getenv("S3_BUCKET_REGION")
	if s3BucketRegion == "" {
		return nil, fmt.Errorf("failed to set S3_BUCKET environment variable")
	}
	awsSDKConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws default config")
	}
	return &Config{
		DB:               database.New(db),
		Platform:         platform,
		TokenSecret:      tokenSecret,
		Port:             port,
		AppDirPath:       appDirPath,
		AssetsBrowserURL: assetsBrowserURL,
		AssetsDirPath:    assetsDirPath,
		S3BucketName:     s3BucketName,
		S3BucketRegion:   s3BucketRegion,
		S3Client:         s3.NewFromConfig(awsSDKConfig),
	}, nil
}
