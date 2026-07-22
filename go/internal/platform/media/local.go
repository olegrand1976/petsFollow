package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type localStore struct {
	root      string
	publicBase string // e.g. http://localhost:8291/media
}

func newLocal(root, apiPublicURL string) (*localStore, http.Handler, error) {
	if root == "" {
		root = "./data/uploads"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, nil, err
	}
	base := strings.TrimRight(apiPublicURL, "/") + "/media"
	st := &localStore{root: root, publicBase: base}
	fs := http.StripPrefix("/media/", http.FileServer(http.Dir(root)))
	return st, fs, nil
}

func (s *localStore) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (string, error) {
	_ = ctx
	_ = contentType
	if err := ValidateSizeLimit(size, AbsoluteMaxBytes); err != nil {
		return "", err
	}
	full := filepath.Join(s.root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer f.Close()
	limited := io.LimitReader(r, AbsoluteMaxBytes+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		return "", err
	}
	if n > AbsoluteMaxBytes {
		_ = os.Remove(full)
		return "", ErrTooLarge
	}
	return fmt.Sprintf("%s/%s", s.publicBase, objectKey), nil
}

func (s *localStore) Delete(ctx context.Context, objectKey string) error {
	_ = ctx
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.Contains(objectKey, "..") {
		return nil
	}
	full := filepath.Join(s.root, filepath.FromSlash(objectKey))
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
