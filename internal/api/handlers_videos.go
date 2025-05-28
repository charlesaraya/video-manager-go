package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charlesaraya/video-manager-go/internal/auth"
	"github.com/charlesaraya/video-manager-go/internal/database"
	"github.com/google/uuid"
)

func AddVideoHandler(cfg *Config) http.HandlerFunc {
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
		videoParams := database.CreateVideoParams{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&videoParams); err != nil {
			Error(res, ErrDecodeRequestBody, http.StatusInternalServerError)
			return
		}
		videoParams.ID = uuid.New().String()
		videoParams.UserID = userUUID.String()
		video, err := cfg.DB.CreateVideo(context.Background(), videoParams)
		if err != nil {
			Error(res, "failed to get videos", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(video)
		if err != nil {
			Error(res, ErrMarshalPayload, http.StatusInternalServerError)
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
			Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		videoUUID, err := uuid.Parse(req.PathValue("videoID"))
		if err != nil {
			Error(res, "failed to parse video ID", http.StatusInternalServerError)
			return
		}
		video, err := cfg.DB.GetVideo(context.Background(), videoUUID.String())
		if err != nil {
			Error(res, "failed to get video", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(video)
		if err != nil {
			Error(res, ErrMarshalPayload, http.StatusInternalServerError)
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
			Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(jwt, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to validate access jwt", http.StatusUnauthorized)
			return
		}
		videos, err := cfg.DB.GetVideosByUser(context.Background(), userUUID.String())
		if err != nil {
			Error(res, "failed to get videos", http.StatusInternalServerError)
			return
		}
		videosPayload := []database.Video{}
		if len(videos) != 0 {
			videosPayload = videos
		}
		data, err := json.Marshal(videosPayload)
		if err != nil {
			Error(res, ErrMarshalPayload, http.StatusInternalServerError)
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
			Error(res, "failed to extract token", http.StatusBadRequest)
			return
		}
		userUUID, err := auth.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to authorize request", http.StatusUnauthorized)
			return
		}
		deleteParams := database.DeleteVideoParams{
			ID:     req.PathValue("videoID"),
			UserID: userUUID.String(),
		}
		if err := cfg.DB.DeleteVideo(context.Background(), deleteParams); err != nil {
			Error(res, "failed to delete video", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}

func UploadThumbnailHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		videoUUID := req.PathValue("videoID")
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Error(res, "failed to extract token", http.StatusUnauthorized)
			return
		}
		_, err = auth.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to authorize request", http.StatusUnauthorized)
			return
		}
		const maxMemory = 10 << 20
		req.ParseMultipartForm(maxMemory)

		file, header, err := req.FormFile("thumbnail")
		if err != nil {
			Error(res, "failed to parse form file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
		if err != nil {
			Error(res, "failed to parse media", http.StatusInternalServerError)
			return
		}
		if mediaType != MimeTypeImageJPEG && mediaType != MimeTypeImagePNG {
			Error(res, "invalid media type", http.StatusInternalServerError)
			return
		}
		key := make([]byte, 32)
		rand.Read(key)
		fileTag := base64.RawURLEncoding.EncodeToString(key)
		mediaTypeSplit := strings.Split(mediaType, "/")
		fileName := fmt.Sprintf("%s.%s", fileTag, mediaTypeSplit[1])
		filePath := filepath.Join(cfg.AssetsDirPath, fileName)
		thumbnailFile, err := os.Create(filePath)
		if err != nil {
			Error(res, "failed to create thumbnail file", http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(thumbnailFile, file)
		if err != nil {
			Error(res, "failed to copy thumbnail file", http.StatusInternalServerError)
			return
		}
		videoParams := database.UpdateVideoThumbnailParams{
			ID:           videoUUID,
			ThumbnailUrl: cfg.AssetsBrowserURL + fileName,
		}
		video, err := cfg.DB.UpdateVideoThumbnail(context.Background(), videoParams)
		if err != nil {
			Error(res, "failed to update thumbnail file", http.StatusInternalServerError)
			return
		}
		payload, err := json.Marshal(video)
		if err != nil {
			Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(payload)
	}
}

func UploadVideosHandler(cfg *Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			Error(res, "failed to extract token", http.StatusUnauthorized)
			return
		}
		userUUID, err := auth.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			Error(res, "failed to authorize request", http.StatusUnauthorized)
			return
		}
		const uploadLimit = 1 << 30
		req.Body = http.MaxBytesReader(res, req.Body, uploadLimit)
		req.ParseMultipartForm(uploadLimit)

		videoUUID := req.PathValue("videoID")
		video, err := cfg.DB.GetVideo(context.Background(), videoUUID)
		if err != nil {
			Error(res, "failed to get video from DB", http.StatusInternalServerError)
			return
		}
		videoUserID, err := uuid.Parse(video.UserID)
		if err != nil {
			Error(res, "failed to parse the video user uuid", http.StatusInternalServerError)
			return
		}
		if videoUserID != userUUID {
			Error(res, "failed to authorize video owner", http.StatusUnauthorized)
			return
		}
		file, header, err := req.FormFile("video")
		if err != nil {
			Error(res, "failed to get the video from the request", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
		if err != nil {
			Error(res, "failed to parse media type", http.StatusInternalServerError)
			return
		}
		if mediaType != MimeTypeVideo {
			Error(res, "invalid media type", http.StatusInternalServerError)
			return
		}
		fileName := "tubely-upload.mp4"
		tempFile, err := os.CreateTemp("", fileName)
		if err != nil {
			Error(res, "failed to create temp file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(fileName)
		defer tempFile.Close()

		_, err = io.Copy(tempFile, file)
		if err != nil {
			Error(res, "failed to copy video file", http.StatusInternalServerError)
			return
		}
		_, err = tempFile.Seek(0, io.SeekStart)
		if err != nil {
			Error(res, "failed to reset read offset to beginning of file", http.StatusInternalServerError)
			return
		}
		key := make([]byte, 32)
		rand.Read(key)
		fileTag := base64.RawURLEncoding.EncodeToString(key)
		mediaTypeSplit := strings.Split(mediaType, "/")
		fileKeyName := fmt.Sprintf("%s.%s", fileTag, mediaTypeSplit[1])
		aspectRatio, err := getVideoAspectRatio(tempFile.Name())
		if err != nil {
			Error(res, "failed to get video aspect ratio", http.StatusInternalServerError)
			return
		}
		prefix := "other"
		switch aspectRatio {
		case "16:9":
			prefix = "landscape"
		case "9:16":
			prefix = "portrait"
		}
		tempFileName, err := processVideoForFastStart(tempFile.Name())
		if err != nil {
			Error(res, "failed to process video for fast start", http.StatusInternalServerError)
			return
		}
		tempFile, err = os.Open(tempFileName)
		if err != nil {
			Error(res, "failed to open preprocessed temp file", http.StatusInternalServerError)
			return
		}
		fileKeyName = filepath.Join(prefix, fileKeyName)
		putObjectInputParams := s3.PutObjectInput{
			Bucket:      &cfg.S3BucketName,
			Key:         &fileKeyName,
			Body:        tempFile,
			ContentType: &mediaType,
		}
		_, err = cfg.S3Client.PutObject(context.Background(), &putObjectInputParams)
		if err != nil {
			Error(res, "failed to put the object into s3", http.StatusInternalServerError)
			return
		}
		videoURL := "https://" + cfg.S3BucketName + ".s3." + cfg.S3BucketRegion + ".amazonaws.com/" + fileKeyName
		videoParams := database.UpdateVideoUrlParams{
			ID:       videoUUID,
			VideoUrl: videoURL,
		}
		video, err = cfg.DB.UpdateVideoUrl(context.Background(), videoParams)
		if err != nil {
			Error(res, "failed to upload video url", http.StatusInternalServerError)
			return
		}
		payload, err := json.Marshal(video)
		if err != nil {
			Error(res, ErrMarshalPayload, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(payload)
	}
}

func getVideoAspectRatio(filepath string) (string, error) {
	args := []string{"-v", "error", "-print_format", "json", "-show_streams", filepath}
	// Prep command
	cmd := exec.Command("ffprobe", args...)
	// Prep buffer to capture stdout
	buff := bytes.Buffer{}
	cmd.Stdout = &buff
	// Run command
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	// Unmarshal JSON
	var dim struct {
		Stream []struct {
			AspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}
	err = json.Unmarshal(buff.Bytes(), &dim)
	if err != nil {
		return "", err
	}
	return dim.Stream[0].AspectRatio, nil
}

func processVideoForFastStart(filepath string) (string, error) {
	output_filepath := filepath + ".preprocessing"
	args := []string{"-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", output_filepath}
	// Prep command
	cmd := exec.Command("ffmpeg", args...)
	// Prep buffer to capture stdout
	buff := bytes.Buffer{}
	cmd.Stdout = &buff
	// Run command
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return output_filepath, nil
}
