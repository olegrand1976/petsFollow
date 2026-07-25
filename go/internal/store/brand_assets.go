package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	BrandAssetQRAndroid = "qr_android"
	BrandAssetQRIOS     = "qr_ios"
)

// BrandAsset is a platform-level media asset (store QR images, optional store URLs).
type BrandAsset struct {
	Key       string    `json:"key"`
	ObjectKey string    `json:"objectKey"`
	PublicURL string    `json:"publicUrl"`
	StoreURL  string    `json:"storeUrl"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Store) UpsertBrandAsset(ctx context.Context, key, objectKey, publicURL, storeURL string) (BrandAsset, error) {
	var a BrandAsset
	err := s.pool.QueryRow(ctx, `
		INSERT INTO platform.brand_assets (key, object_key, public_url, store_url, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (key) DO UPDATE SET
			object_key = EXCLUDED.object_key,
			public_url = EXCLUDED.public_url,
			store_url = CASE
				WHEN EXCLUDED.store_url <> '' THEN EXCLUDED.store_url
				ELSE platform.brand_assets.store_url
			END,
			updated_at = NOW()
		RETURNING key, object_key, public_url, store_url, updated_at`,
		key, objectKey, publicURL, storeURL,
	).Scan(&a.Key, &a.ObjectKey, &a.PublicURL, &a.StoreURL, &a.UpdatedAt)
	return a, err
}

func (s *Store) PatchBrandAssetStoreURL(ctx context.Context, key, storeURL string) (BrandAsset, error) {
	var a BrandAsset
	err := s.pool.QueryRow(ctx, `
		UPDATE platform.brand_assets SET store_url = $2, updated_at = NOW()
		WHERE key = $1
		RETURNING key, object_key, public_url, store_url, updated_at`,
		key, storeURL,
	).Scan(&a.Key, &a.ObjectKey, &a.PublicURL, &a.StoreURL, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return BrandAsset{}, ErrNotFound
	}
	return a, err
}

func (s *Store) GetBrandAsset(ctx context.Context, key string) (BrandAsset, error) {
	var a BrandAsset
	err := s.pool.QueryRow(ctx, `
		SELECT key, object_key, public_url, store_url, updated_at
		FROM platform.brand_assets WHERE key = $1`, key,
	).Scan(&a.Key, &a.ObjectKey, &a.PublicURL, &a.StoreURL, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return BrandAsset{}, ErrNotFound
	}
	return a, err
}

func (s *Store) ListBrandAssets(ctx context.Context) ([]BrandAsset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT key, object_key, public_url, store_url, updated_at
		FROM platform.brand_assets ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BrandAsset
	for rows.Next() {
		var a BrandAsset
		if err := rows.Scan(&a.Key, &a.ObjectKey, &a.PublicURL, &a.StoreURL, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []BrandAsset{}
	}
	return out, rows.Err()
}

// StoreQRAssets returns qr_android / qr_ios when present (empty PublicURL if missing).
func (s *Store) StoreQRAssets(ctx context.Context) (android, ios BrandAsset, err error) {
	android, aerr := s.GetBrandAsset(ctx, BrandAssetQRAndroid)
	if aerr != nil && !errors.Is(aerr, ErrNotFound) {
		return BrandAsset{}, BrandAsset{}, aerr
	}
	ios, ierr := s.GetBrandAsset(ctx, BrandAssetQRIOS)
	if ierr != nil && !errors.Is(ierr, ErrNotFound) {
		return BrandAsset{}, BrandAsset{}, ierr
	}
	return android, ios, nil
}
