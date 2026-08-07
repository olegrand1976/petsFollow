package vetnews

import (
	"context"
	"log"
	"time"
)

// ArticleStore persists classified articles (implemented by store.Store).
type ArticleStore interface {
	UpsertVetNewsArticle(ctx context.Context, a Article) (inserted bool, err error)
	ListVetNewsArticles(ctx context.Context, filter ListFilter) ([]Article, error)
	PurgeOldVetNewsArticles(ctx context.Context, olderThan time.Time) (int64, error)
}

// ListFilter for API reads.
type ListFilter struct {
	SourceID   string
	Importance string
	Limit      int
}

// Ingester runs all providers and upserts into the store.
type Ingester struct {
	Registry *Registry
	Store    ArticleStore
}

func NewIngester(store ArticleStore, providers ...Provider) *Ingester {
	if len(providers) == 0 {
		providers = DefaultProviders(NewHTTPClient())
	}
	return &Ingester{
		Registry: NewRegistry(providers...),
		Store:    store,
	}
}

// Run fetches every source; one failure does not abort the others.
func (in *Ingester) Run(ctx context.Context) RunResult {
	started := time.Now().UTC()
	res := RunResult{StartedAt: started, Sources: make([]SourceResult, 0)}
	if in == nil || in.Store == nil {
		return res
	}
	for _, p := range in.Registry.All() {
		sr := in.ingestOne(ctx, p)
		LogSourceResult("vetnews", sr)
		res.Sources = append(res.Sources, sr)
		res.Inserted += sr.Inserted
		res.Updated += sr.Updated
		res.Skipped += sr.Skipped
	}
	if n, err := in.Store.PurgeOldVetNewsArticles(ctx, time.Now().UTC().AddDate(0, 0, -90)); err != nil {
		log.Printf("vetnews purge_err=%v", err)
	} else if n > 0 {
		log.Printf("vetnews purged_old=%d", n)
	}
	res.DurationMs = time.Since(started).Milliseconds()
	return res
}

func (in *Ingester) ingestOne(ctx context.Context, p Provider) SourceResult {
	sr := SourceResult{SourceID: p.ID()}
	raw, err := p.Fetch(ctx)
	if err != nil {
		sr.Error = err.Error()
		return sr
	}
	sr.Fetched = len(raw)
	now := time.Now().UTC()
	for _, item := range raw {
		item.Title = StripHTML(item.Title)
		item.Summary = truncateRunes(StripHTML(item.Summary), 500)
		item.SourceURL = NormalizeURL(item.SourceURL)
		if item.Title == "" || item.SourceURL == "" {
			sr.Skipped++
			continue
		}
		cat, imp, tags := Classify(item, p.ID(), p.DefaultTags())
		art := Article{
			SourceID:    p.ID(),
			SourceName:  p.Name(),
			SourceURL:   item.SourceURL,
			Title:       item.Title,
			Summary:     item.Summary,
			ImageURL:    item.ImageURL,
			PublishedAt: item.PublishedAt,
			Category:    cat,
			Importance:  imp,
			Tags:        tags,
			FetchedAt:   now,
		}
		inserted, err := in.Store.UpsertVetNewsArticle(ctx, art)
		if err != nil {
			log.Printf("vetnews upsert source=%s url=%s err=%v", p.ID(), art.SourceURL, err)
			sr.Skipped++
			continue
		}
		if inserted {
			sr.Inserted++
		} else {
			sr.Updated++
		}
	}
	return sr
}
