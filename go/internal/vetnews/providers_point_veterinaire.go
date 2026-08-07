package vetnews

import (
	"context"
	"regexp"
)

const (
	pointVetURL  = "https://www.lepointveterinaire.fr/actualites/actualites-professionnelles.html"
	pointVetBase = "https://www.lepointveterinaire.fr"
)

var (
	lpvBlockRe  = regexp.MustCompile(`(?is)<div class="row actu_horizontale">(.*?)</div>\s*</div>\s*</div>`)
	lpvHrefRe   = regexp.MustCompile(`(?is)<h3>\s*<a href="(/actualites/actualites-professionnelles/[^"]+\.html)"[^>]*>([\s\S]*?)</a>`)
	lpvChapoRe  = regexp.MustCompile(`(?is)<div class="chapo">\s*<a[^>]*>([\s\S]*?)</a>`)
	lpvDateRe   = regexp.MustCompile(`(?is)<div class="infos">[\s\S]*?\|\s*(\d{2}\.\d{2}\.\d{4})`)
	lpvImgRe    = regexp.MustCompile(`(?is)<img[^>]+src="([^"]+)"`)
)

type pointVeterinaireProvider struct {
	http *HTTPClient
}

func NewPointVeterinaireProvider(http *HTTPClient) Provider {
	return &pointVeterinaireProvider{http: http}
}

func (p *pointVeterinaireProvider) ID() string   { return "point_veterinaire" }
func (p *pointVeterinaireProvider) Name() string { return "Le Point Vétérinaire" }
func (p *pointVeterinaireProvider) DefaultTags() []string {
	return []string{"veterinaire", "fr"}
}

func (p *pointVeterinaireProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	body, err := p.http.Get(ctx, pointVetURL, "text/html")
	if err != nil {
		return nil, err
	}
	return ParsePointVeterinaireHTML(string(body), 20), nil
}

// ParsePointVeterinaireHTML extracts professional news teasers.
func ParsePointVeterinaireHTML(raw string, limit int) []RawItem {
	if limit <= 0 {
		limit = 20
	}
	blocks := lpvBlockRe.FindAllStringSubmatch(raw, -1)
	out := make([]RawItem, 0, len(blocks))
	seen := make(map[string]struct{}, len(blocks))
	for _, m := range blocks {
		if len(m) < 2 {
			continue
		}
		block := m[1]
		hm := lpvHrefRe.FindStringSubmatch(block)
		if hm == nil {
			continue
		}
		url := NormalizeURL(ResolveURL(pointVetBase, hm[1]))
		title := StripHTML(hm[2])
		if url == "" || title == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		item := RawItem{Title: title, SourceURL: url}
		if cm := lpvChapoRe.FindStringSubmatch(block); cm != nil {
			item.Summary = truncateRunes(StripHTML(cm[1]), 400)
		}
		if dm := lpvDateRe.FindStringSubmatch(block); dm != nil {
			item.PublishedAt = ParseFlexibleTime(dm[1])
		}
		if im := lpvImgRe.FindStringSubmatch(block); im != nil {
			item.ImageURL = ResolveURL(pointVetBase, im[1])
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}
