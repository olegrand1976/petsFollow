package media

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type gcsStore struct {
	client *storage.Client
	bucket string
}

func newGCS(bucket string) (*gcsStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client, err := storage.NewClient(ctx, option.WithScopes(storage.ScopeReadWrite))
	if err != nil {
		return nil, fmt.Errorf("gcs client: %w", err)
	}
	return &gcsStore{client: client, bucket: bucket}, nil
}

func (s *gcsStore) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (string, error) {
	if err := ValidateSize(size); err != nil {
		return "", err
	}
	if _, err := ExtForContentType(contentType); err != nil {
		return "", err
	}
	w := s.client.Bucket(s.bucket).Object(objectKey).NewWriter(ctx)
	w.ContentType = contentType
	w.CacheControl = "public, max-age=86400"
	limited := io.LimitReader(r, MaxUploadBytes+1)
	n, err := io.Copy(w, limited)
	if err != nil {
		_ = w.Close()
		return "", err
	}
	if n > MaxUploadBytes {
		_ = w.Close()
		return "", ErrTooLarge
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, objectKey), nil
}

func newObjectID() string {
	return uuid.NewString()
}
