package vetnews

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const (
	cliniciansBriefURL  = "https://www.cliniciansbrief.com/articles"
	cliniciansBriefBase = "https://www.cliniciansbrief.com"
)

// Next.js SPA often ships an empty shell; we still try HTML + JSON-LD patterns.
var (
	cbJSONLDArticleRe = regexp.MustCompile(`(?is)"@type"\s*:\s*"NewsArticle"[^}]*?"headline"\s*:\s*"([^"]+)"[^}]*?"url"\s*:\s*"([^"]+)"`)
	cbAnchorRe        = regexp.MustCompile(`(?is)<a[^>]+href="((?:https://www\.cliniciansbrief\.com)?/articles?/[^"]+)"[^>]*>([\s\S]*?)</a>`)
)

type cliniciansBriefProvider struct {
	http *HTTPClient
}

func NewCliniciansBriefProvider(http *HTTPClient) Provider {
	return &cliniciansBriefProvider{http: http}
}

func (p *cliniciansBriefProvider) ID() string   { return "clinicians_brief" }
func (p *cliniciansBriefProvider) Name() string { return "Clinician's Brief" }
func (p *cliniciansBriefProvider) DefaultTags() []string {
	return []string{"clinique", "canin-felin", "en"}
}

func (p *cliniciansBriefProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	body, err := p.http.Get(ctx, cliniciansBriefURL, "text/html")
	if err != nil {
		return nil, err
	}
	items := ParseCliniciansBriefHTML(string(body), 20)
	if len(items) == 0 {
		return nil, fmt.Errorf("clinicians_brief_empty_spa")
	}
	return items, nil
}

// ParseCliniciansBriefHTML best-effort extract from SSR / JSON-LD when present.
func ParseCliniciansBriefHTML(raw string, limit int) []RawItem {
	if limit <= 0 {
		limit = 20
	}
	out := make([]RawItem, 0, limit)
	seen := make(map[string]struct{})

	for _, m := range cbJSONLDArticleRe.FindAllStringSubmatch(raw, -1) {
		title := StripHTML(m[1])
		url := NormalizeURL(ResolveURL(cliniciansBriefBase, m[2]))
		if title == "" || url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		out = append(out, RawItem{Title: title, SourceURL: url})
		if len(out) >= limit {
			return out
		}
	}

	for _, m := range cbAnchorRe.FindAllStringSubmatch(raw, -1) {
		url := NormalizeURL(ResolveURL(cliniciansBriefBase, m[1]))
		title := StripHTML(m[2])
		if title == "" || url == "" || strings.HasSuffix(url, "/articles") {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		out = append(out, RawItem{Title: title, SourceURL: url})
		if len(out) >= limit {
			break
		}
	}
	return out
}
