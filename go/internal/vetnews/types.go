// Package vetnews ingests veterinary scientific / sanitary news from multiple sources.
package vetnews

import (
	"context"
	"time"
)

// Importance levels drive UI colour pastilles.
const (
	ImportanceCritical = "critical"
	ImportanceHigh     = "high"
	ImportanceMedium   = "medium"
	ImportanceLow      = "low"
)

// Category domain tags (orthogonal to importance).
const (
	CategoryEpidemio   = "epidemio"
	CategoryRegulatory = "regulatory"
	CategoryClinical   = "clinical"
	CategoryScience    = "science"
	CategoryPractice   = "practice"
	CategoryGeneral    = "general"
)

// RawItem is a provider-normalized article before classification / upsert.
type RawItem struct {
	Title       string
	Summary     string
	SourceURL   string
	ImageURL    string
	PublishedAt *time.Time
	Tags        []string
}

// Article is the persisted / API shape.
type Article struct {
	ID          string     `json:"id"`
	SourceID    string     `json:"sourceId"`
	SourceName  string     `json:"sourceName"`
	SourceURL   string     `json:"sourceUrl"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	ImageURL    string     `json:"imageUrl,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	Category    string     `json:"category"`
	Importance  string     `json:"importance"`
	Tags        []string   `json:"tags"`
	FetchedAt   time.Time  `json:"fetchedAt"`
}

// Provider fetches recent items from one external source.
type Provider interface {
	ID() string
	Name() string
	DefaultTags() []string
	Fetch(ctx context.Context) ([]RawItem, error)
}

// SourceResult is per-provider ingest stats (CRON logs).
type SourceResult struct {
	SourceID string `json:"sourceId"`
	Fetched  int    `json:"fetched"`
	Inserted int    `json:"inserted"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
	Error    string `json:"error,omitempty"`
}

// RunResult aggregates a full ingest pass.
type RunResult struct {
	Sources    []SourceResult `json:"sources"`
	Inserted   int            `json:"inserted"`
	Updated    int            `json:"updated"`
	Skipped    int            `json:"skipped"`
	StartedAt  time.Time      `json:"startedAt"`
	DurationMs int64          `json:"durationMs"`
}

// ImportanceLegend is exposed to the UI for pastille colour help.
func ImportanceLegend() []map[string]string {
	return []map[string]string{
		{"id": ImportanceCritical, "color": "critical"},
		{"id": ImportanceHigh, "color": "high"},
		{"id": ImportanceMedium, "color": "medium"},
		{"id": ImportanceLow, "color": "low"},
	}
}
