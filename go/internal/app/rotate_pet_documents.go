package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

// IsRotatePetDocumentsCmd matches the one-shot key rotation sub-command.
func IsRotatePetDocumentsCmd(args []string) bool {
	return len(args) > 0 && strings.EqualFold(args[0], "rotate-pet-documents")
}

// RotatePetDocumentsOnly re-keys pet documents whose storage key leaked.
//
// Those objects were uploaded while documents/ was served publicly and their
// URL was persisted in file_url, so the key must be considered compromised.
// Each document is copied to a fresh random key, the old object is deleted and
// file_url is cleared — any previously shared link now 404s.
//
// A non-empty file_url is exactly what marks a legacy row, so re-running the
// command is a no-op.
func RotatePetDocumentsOnly(ctx context.Context, cfg config.Config, args []string) error {
	dryRun := false
	for _, a := range args[1:] {
		switch a {
		case "--dry-run":
			dryRun = true
		case "--help", "-h":
			fmt.Println("usage: petsfollow-api rotate-pet-documents [--dry-run]")
			return nil
		default:
			return fmt.Errorf("unknown flag %q (see --help)", a)
		}
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	bundle, err := media.New(cfg)
	if err != nil {
		return fmt.Errorf("media: %w", err)
	}

	type target struct{ id, objectKey string }
	var targets []target
	rows, err := pool.Query(ctx, `
		SELECT id::text, object_key
		FROM pets.documents
		WHERE COALESCE(object_key, '') <> '' AND COALESCE(file_url, '') <> ''
		ORDER BY created_at`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.objectKey); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("rotate-pet-documents: %d document(s) to rotate (dry-run=%v)", len(targets), dryRun)
	var rotated, failed int
	for _, t := range targets {
		newKey := media.ObjectKey("documents", petIDFromObjectKey(t.objectKey), path.Ext(t.objectKey))
		if dryRun {
			log.Printf("  %s: %s -> %s", t.id, t.objectKey, newKey)
			continue
		}
		if err := copyMediaObject(ctx, bundle.Store, t.objectKey, newKey); err != nil {
			log.Printf("  %s: rotation failed (%v) — left untouched", t.id, err)
			failed++
			continue
		}
		if _, err := pool.Exec(ctx, `
			UPDATE pets.documents SET object_key = $2, file_url = '' WHERE id = $1::uuid`,
			t.id, newKey); err != nil {
			// The copy is in place but unreferenced; the old object still serves the row.
			log.Printf("  %s: db update failed (%v)", t.id, err)
			failed++
			continue
		}
		if err := bundle.Store.Delete(ctx, t.objectKey); err != nil {
			log.Printf("  %s: old object not deleted (%v) — purge manually", t.id, err)
		}
		rotated++
	}
	log.Printf("rotate-pet-documents: %d rotated, %d failed", rotated, failed)
	if failed > 0 {
		return fmt.Errorf("%d document(s) could not be rotated", failed)
	}
	return nil
}

func copyMediaObject(ctx context.Context, st media.Store, oldKey, newKey string) error {
	rc, ct, err := st.Open(ctx, oldKey)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if _, err := st.Upload(ctx, newKey, bytes.NewReader(data), int64(len(data)), ct); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	return nil
}

// petIDFromObjectKey extracts the pet id segment of documents/{petID}/{object}.
func petIDFromObjectKey(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	if len(parts) >= 2 && parts[1] != "" {
		return parts[1]
	}
	return "unknown"
}
