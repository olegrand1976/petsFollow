package vetnews

import (
	"context"
	"regexp"
)

const (
	ansesListURL = "https://www.anses.fr/fr/content/actualites?field_thematique_target_id=222"
	ansesBase    = "https://www.anses.fr"
)

var (
	ansesArticleRe = regexp.MustCompile(`(?is)<article class="actualite[^"]*"[^>]*>(.*?)</article>`)
	ansesHrefRe    = regexp.MustCompile(`(?is)href="(/fr/content/[^"]+)"`)
	ansesDateRe    = regexp.MustCompile(`(?is)infobar__item--date[^>]*>\s*(?:<svg[\s\S]*?</svg>)?\s*(\d{2}/\d{2}/\d{4})`)
	ansesTitleRe   = regexp.MustCompile(`(?is)teaser__title[^>]*>\s*<span>([\s\S]*?)</span>`)
	ansesImgRe     = regexp.MustCompile(`(?is)<img[^>]+src="([^"]+)"`)
)

type ansesProvider struct {
	http *HTTPClient
}

func NewAnsesProvider(http *HTTPClient) Provider {
	return &ansesProvider{http: http}
}

func (p *ansesProvider) ID() string   { return "anses" }
func (p *ansesProvider) Name() string { return "Anses — Santé animale" }
func (p *ansesProvider) DefaultTags() []string {
	return []string{"anses", "sante-animale", "fr"}
}

func (p *ansesProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	body, err := p.http.Get(ctx, ansesListURL, "text/html")
	if err != nil {
		return nil, err
	}
	return ParseAnsesHTML(string(body), 20), nil
}

// ParseAnsesHTML extracts teaser cards from the Anses actualités list (thematic animal health).
func ParseAnsesHTML(raw string, limit int) []RawItem {
	if limit <= 0 {
		limit = 20
	}
	blocks := ansesArticleRe.FindAllStringSubmatch(raw, -1)
	out := make([]RawItem, 0, len(blocks))
	seen := make(map[string]struct{}, len(blocks))
	for _, m := range blocks {
		if len(m) < 2 {
			continue
		}
		block := m[1]
		hrefM := ansesHrefRe.FindStringSubmatch(block)
		if hrefM == nil {
			continue
		}
		url := NormalizeURL(ResolveURL(ansesBase, hrefM[1]))
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		title := ""
		if tm := ansesTitleRe.FindStringSubmatch(block); tm != nil {
			title = StripHTML(tm[1])
		}
		if title == "" {
			continue
		}
		item := RawItem{Title: title, SourceURL: url, Summary: ""}
		if dm := ansesDateRe.FindStringSubmatch(block); dm != nil {
			item.PublishedAt = ParseFlexibleTime(dm[1])
		}
		if im := ansesImgRe.FindStringSubmatch(block); im != nil {
			item.ImageURL = ResolveURL(ansesBase, im[1])
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
