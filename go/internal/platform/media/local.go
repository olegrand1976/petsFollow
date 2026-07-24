package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type localStore struct {
	root       string
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

func (s *localStore) resolvePath(objectKey string) (string, error) {
	objectKey = strings.TrimSpace(objectKey)
	objectKey = path.Clean("/" + objectKey)
	objectKey = strings.TrimPrefix(objectKey, "/")
	if objectKey == "" || objectKey == "." || strings.HasPrefix(objectKey, "..") {
		return "", os.ErrNotExist
	}
	full := filepath.Join(s.root, filepath.FromSlash(objectKey))
	rootAbs, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", os.ErrNotExist
	}
	return fullAbs, nil
}

func (s *localStore) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (string, error) {
	_ = ctx
	_ = contentType
	if err := ValidateSizeLimit(size, AbsoluteMaxBytes); err != nil {
		return "", err
	}
	full, err := s.resolvePath(objectKey)
	if err != nil {
		return "", err
	}
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
	if IsSensitiveObjectKey(objectKey) {
		return "", nil
	}
	clean := path.Clean("/" + strings.TrimSpace(objectKey))
	clean = strings.TrimPrefix(clean, "/")
	return fmt.Sprintf("%s/%s", s.publicBase, clean), nil
}

func (s *localStore) Delete(ctx context.Context, objectKey string) error {
	_ = ctx
	full, err := s.resolvePath(objectKey)
	if err != nil {
		return nil
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *localStore) Open(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	_ = ctx
	full, err := s.resolvePath(objectKey)
	if err != nil {
		return nil, "", err
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, "", err
	}
	ct := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(full)) {
	case ".mp3":
		ct = "audio/mpeg"
	case ".m4a", ".mp4":
		ct = "audio/mp4"
	case ".wav":
		ct = "audio/wav"
	case ".webm":
		ct = "audio/webm"
	case ".ogg", ".oga":
		ct = "audio/ogg"
	}
	return f, ct, nil
}
