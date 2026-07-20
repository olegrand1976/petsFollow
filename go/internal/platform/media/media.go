package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

const (
	MaxUploadBytes       = 2 << 20  // 2 MiB — avatars / pet photos
	MaxMessageMediaBytes = 25 << 20 // 25 MiB — chat image / video
	MaxPitchAudioBytes   = 20 << 20 // 20 MiB — pitch simulation recordings
	AbsoluteMaxBytes     = MaxMessageMediaBytes
)

var (
	ErrTooLarge      = errors.New("file too large")
	ErrInvalidType   = errors.New("invalid content type")
	ErrEmptyFile     = errors.New("empty file")
	ErrNotConfigured = errors.New("media store not configured")
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var allowedMessageMediaTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/webm":      ".webm",
}

var allowedPitchAudioTypes = map[string]string{
	"audio/webm":       ".webm",
	"audio/ogg":        ".ogg",
	"audio/mpeg":       ".mp3",
	"audio/mp4":        ".m4a",
	"audio/wav":        ".wav",
	"audio/x-wav":      ".wav",
	"video/webm":       ".webm",
}

// Store uploads binaries and returns a publicly reachable URL.
type Store interface {
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (publicURL string, err error)
}

type Bundle struct {
	Store        Store
	LocalHandler http.Handler // nil when GCS is used
	LocalMount   string       // e.g. "/media/"
}

func New(cfg config.Config) (*Bundle, error) {
	if cfg.GCSMediaBucket != "" {
		st, err := newGCS(cfg.GCSMediaBucket)
		if err != nil {
			return nil, err
		}
		return &Bundle{Store: st}, nil
	}
	st, handler, err := newLocal(cfg.MediaLocalDir, cfg.APIPublicURL)
	if err != nil {
		return nil, err
	}
	return &Bundle{
		Store:        st,
		LocalHandler: handler,
		LocalMount:   "/media/",
	}, nil
}

func ExtForContentType(contentType string) (string, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	ext, ok := allowedImageTypes[ct]
	if !ok {
		return "", ErrInvalidType
	}
	return ext, nil
}

func ExtForMessageMedia(contentType string) (string, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	ext, ok := allowedMessageMediaTypes[ct]
	if !ok {
		return "", ErrInvalidType
	}
	return ext, nil
}

func ExtForPitchAudio(contentType string) (string, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	ext, ok := allowedPitchAudioTypes[ct]
	if !ok {
		return "", ErrInvalidType
	}
	return ext, nil
}

func NormalizePitchAudioType(contentType, filename string) (string, error) {
	ct := inferContentType(contentType, filename)
	if ct == "" || ct == "application/octet-stream" {
		switch strings.ToLower(path.Ext(filename)) {
		case ".webm":
			ct = "audio/webm"
		case ".ogg":
			ct = "audio/ogg"
		case ".mp3":
			ct = "audio/mpeg"
		case ".m4a":
			ct = "audio/mp4"
		case ".wav":
			ct = "audio/wav"
		}
	}
	if _, err := ExtForPitchAudio(ct); err != nil {
		return "", err
	}
	return ct, nil
}

func MediaKind(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if strings.HasPrefix(ct, "video/") {
		return "video"
	}
	return "image"
}

func NormalizeContentType(contentType, filename string) (string, error) {
	ct := inferContentType(contentType, filename)
	if _, err := ExtForContentType(ct); err != nil {
		return "", err
	}
	return ct, nil
}

func NormalizeMessageMediaType(contentType, filename string) (string, error) {
	ct := inferContentType(contentType, filename)
	if _, err := ExtForMessageMedia(ct); err != nil {
		return "", err
	}
	return ct, nil
}

func inferContentType(contentType, filename string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct == "" || ct == "application/octet-stream" {
		switch strings.ToLower(path.Ext(filename)) {
		case ".jpg", ".jpeg":
			ct = "image/jpeg"
		case ".png":
			ct = "image/png"
		case ".webp":
			ct = "image/webp"
		case ".mp4":
			ct = "video/mp4"
		case ".mov":
			ct = "video/quicktime"
		case ".webm":
			ct = "video/webm"
		}
	}
	return ct
}

func ValidateSize(size int64) error {
	return ValidateSizeLimit(size, MaxUploadBytes)
}

func ValidateSizeLimit(size, max int64) error {
	if size <= 0 {
		return ErrEmptyFile
	}
	if size > max {
		return ErrTooLarge
	}
	return nil
}

func ObjectKey(kind, entityID, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", kind, entityID, newObjectID(), ext)
}

// PublicURL builds a reachable URL for an object key (empty key → "").
func PublicURL(cfg config.Config, objectKey string) string {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return ""
	}
	if cfg.GCSMediaBucket != "" {
		return fmt.Sprintf("https://storage.googleapis.com/%s/%s", cfg.GCSMediaBucket, objectKey)
	}
	return strings.TrimRight(cfg.APIPublicURL, "/") + "/media/" + objectKey
}
