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

const MaxUploadBytes = 2 << 20 // 2 MiB

var (
	ErrTooLarge      = errors.New("file too large")
	ErrInvalidType   = errors.New("invalid content type")
	ErrEmptyFile     = errors.New("empty file")
	ErrNotConfigured = errors.New("media store not configured")
)

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// Store uploads avatar / pet photo binaries and returns a publicly reachable URL.
type Store interface {
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (publicURL string, err error)
}

type Bundle struct {
	Store         Store
	LocalHandler  http.Handler // nil when GCS is used
	LocalMount    string       // e.g. "/media/"
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
	ext, ok := allowedTypes[ct]
	if !ok {
		return "", ErrInvalidType
	}
	return ext, nil
}

func NormalizeContentType(contentType, filename string) (string, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct == "" || ct == "application/octet-stream" {
		switch strings.ToLower(path.Ext(filename)) {
		case ".jpg", ".jpeg":
			ct = "image/jpeg"
		case ".png":
			ct = "image/png"
		case ".webp":
			ct = "image/webp"
		}
	}
	if _, err := ExtForContentType(ct); err != nil {
		return "", err
	}
	return ct, nil
}

func ValidateSize(size int64) error {
	if size <= 0 {
		return ErrEmptyFile
	}
	if size > MaxUploadBytes {
		return ErrTooLarge
	}
	return nil
}

func ObjectKey(kind, entityID, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", kind, entityID, newObjectID(), ext)
}
