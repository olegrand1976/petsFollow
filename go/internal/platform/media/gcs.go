package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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
	if err := ValidateSizeLimit(size, AbsoluteMaxBytes); err != nil {
		return "", err
	}
	w := s.client.Bucket(s.bucket).Object(objectKey).NewWriter(ctx)
	w.ContentType = contentType
	sensitive := IsSensitiveObjectKey(objectKey)
	if sensitive {
		// PHI: no public URL returned; rely on auth Open + UUID keys (GCP forbids IAM conditions on allUsers).
		w.CacheControl = "private, no-store"
	} else {
		w.CacheControl = "public, max-age=86400"
	}
	limited := io.LimitReader(r, AbsoluteMaxBytes+1)
	n, err := io.Copy(w, limited)
	if err != nil {
		_ = w.Close()
		return "", err
	}
	if n > AbsoluteMaxBytes {
		_ = w.Close()
		return "", ErrTooLarge
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	if sensitive {
		// No public URL — serve via authenticated API Open only.
		return "", nil
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, objectKey), nil
}

func (s *gcsStore) Delete(ctx context.Context, objectKey string) error {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return nil
	}
	err := s.client.Bucket(s.bucket).Object(objectKey).Delete(ctx)
	if err != nil && errors.Is(err, storage.ErrObjectNotExist) {
		return nil
	}
	return err
}

func (s *gcsStore) Open(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return nil, "", storage.ErrObjectNotExist
	}
	obj := s.client.Bucket(s.bucket).Object(objectKey)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, "", err
	}
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, "", err
	}
	ct := attrs.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	return r, ct, nil
}

func newObjectID() string {
	return uuid.NewString()
}
