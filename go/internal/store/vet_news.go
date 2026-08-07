package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/vetnews"
)

// UpsertVetNewsArticle inserts or refreshes by unique source_url.
// Returns inserted=true when a new row was created.
func (s *Store) UpsertVetNewsArticle(ctx context.Context, a vetnews.Article) (inserted bool, err error) {
	if s == nil || s.pool == nil {
		return false, errors.New("store_unavailable")
	}
	url := strings.TrimSpace(a.SourceURL)
	if url == "" || strings.TrimSpace(a.Title) == "" {
		return false, errors.New("invalid_article")
	}
	hash := vetnews.ContentHash(a.Title, url)
	tags := a.Tags
	if tags == nil {
		tags = []string{}
	}
	var id string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO ops.vet_news_articles (
			source_id, source_name, source_url, title, summary, image_url,
			published_at, category, importance, tags, content_hash, fetched_at
		) VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10,$11,$12)
		ON CONFLICT (source_url) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			source_name = EXCLUDED.source_name,
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			image_url = EXCLUDED.image_url,
			published_at = COALESCE(EXCLUDED.published_at, ops.vet_news_articles.published_at),
			category = EXCLUDED.category,
			importance = EXCLUDED.importance,
			tags = EXCLUDED.tags,
			content_hash = EXCLUDED.content_hash,
			fetched_at = EXCLUDED.fetched_at,
			updated_at = NOW()
		RETURNING id, (xmax = 0) AS inserted`,
		a.SourceID, a.SourceName, url, a.Title, a.Summary, a.ImageURL,
		a.PublishedAt, a.Category, a.Importance, tags, hash, a.FetchedAt,
	).Scan(&id, &inserted)
	return inserted, err
}

// ListVetNewsArticles returns recent articles (newest first).
func (s *Store) ListVetNewsArticles(ctx context.Context, filter vetnews.ListFilter) ([]vetnews.Article, error) {
	if s == nil || s.pool == nil {
		return nil, errors.New("store_unavailable")
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, source_id, source_name, source_url, title, summary,
			COALESCE(image_url, ''), published_at, category, importance, tags, fetched_at
		FROM ops.vet_news_articles
		WHERE ($1 = '' OR source_id = $1)
		  AND ($2 = '' OR importance = $2)
		ORDER BY published_at DESC NULLS LAST, fetched_at DESC
		LIMIT $3`,
		strings.TrimSpace(filter.SourceID),
		strings.TrimSpace(filter.Importance),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]vetnews.Article, 0, limit)
	for rows.Next() {
		var a vetnews.Article
		var publishedAt *time.Time
		if err := rows.Scan(
			&a.ID, &a.SourceID, &a.SourceName, &a.SourceURL, &a.Title, &a.Summary,
			&a.ImageURL, &publishedAt, &a.Category, &a.Importance, &a.Tags, &a.FetchedAt,
		); err != nil {
			return nil, err
		}
		a.PublishedAt = publishedAt
		if a.Tags == nil {
			a.Tags = []string{}
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// PurgeOldVetNewsArticles deletes rows older than cutoff (by published_at, else fetched_at).
func (s *Store) PurgeOldVetNewsArticles(ctx context.Context, olderThan time.Time) (int64, error) {
	if s == nil || s.pool == nil {
		return 0, errors.New("store_unavailable")
	}
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM ops.vet_news_articles
		WHERE COALESCE(published_at, fetched_at) < $1`, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
