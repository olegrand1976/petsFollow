package vetnews

import (
	"context"
	"regexp"
)

const (
	// User-facing diseases URL redirects; medias/actualites holds sanitary communiqués.
	woahNewsURL  = "https://www.woah.org/fr/medias/actualites/"
	woahNewsBase = "https://www.woah.org"
)

var (
	woahCardRe  = regexp.MustCompile(`(?is)<li class="cards__item cards__item--post[^"]*"[^>]*>(.*?)</li>`)
	woahTitleRe = regexp.MustCompile(`(?is)<h3 class="cards__title">\s*<a href="([^"]+)"[^>]*>([\s\S]*?)</a>`)
	woahDateRe  = regexp.MustCompile(`(?is)cards__infos-item--date[^>]*>\s*([^<]+)`)
	woahTypeRe  = regexp.MustCompile(`(?is)<p class="cards__type">\s*([^<]+)`)
	woahImgRe   = regexp.MustCompile(`(?is)<img[^>]+(?:data-src|src)="(https://www\.woah\.org/[^"]+)"`)
)

type woahProvider struct {
	http *HTTPClient
}

func NewWOAHProvider(http *HTTPClient) Provider {
	return &woahProvider{http: http}
}

func (p *woahProvider) ID() string   { return "woah" }
func (p *woahProvider) Name() string { return "WOAH / OMSA" }
func (p *woahProvider) DefaultTags() []string {
	return []string{"epidemio", "sante-publique", "fr"}
}

func (p *woahProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	body, err := p.http.Get(ctx, woahNewsURL, "text/html")
	if err != nil {
		return nil, err
	}
	return ParseWOAHHTML(string(body), 20), nil
}

// ParseWOAHHTML extracts news cards from OMSA medias/actualites.
func ParseWOAHHTML(raw string, limit int) []RawItem {
	if limit <= 0 {
		limit = 20
	}
	blocks := woahCardRe.FindAllStringSubmatch(raw, -1)
	out := make([]RawItem, 0, len(blocks))
	seen := make(map[string]struct{}, len(blocks))
	for _, m := range blocks {
		if len(m) < 2 {
			continue
		}
		block := m[1]
		tm := woahTitleRe.FindStringSubmatch(block)
		if tm == nil {
			continue
		}
		url := NormalizeURL(ResolveURL(woahNewsBase, tm[1]))
		title := StripHTML(tm[2])
		if url == "" || title == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		item := RawItem{Title: title, SourceURL: url}
		if dm := woahDateRe.FindStringSubmatch(block); dm != nil {
			item.PublishedAt = ParseFlexibleTime(StripHTML(dm[1]))
		}
		if ty := woahTypeRe.FindStringSubmatch(block); ty != nil {
			item.Summary = truncateRunes(StripHTML(ty[1]), 120)
		}
		if im := woahImgRe.FindStringSubmatch(block); im != nil {
			item.ImageURL = im[1]
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
