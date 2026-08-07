// Package afsca fetches and parses AFSCA veterinary newsletters (Belgium).
package afsca

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultLimit = 5
	MaxLimit     = 20      // max items returned to a client
	FetchPool    = 40      // tagged/cached pool before practice-scope filter
	maxBodyBytes = 2 << 20 // 2 MiB
	DefaultBase  = "https://favv-afsca.be"
	userAgent    = "petsFollow/1.0 (+https://petsfollow.app; afsca-newsletters)"
)

// Item is one newsletter entry from the AFSCA index page.
type Item struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Label string `json:"label"` // e.g. Newsletter N°559
	Title string `json:"title"` // subject line
	URL   string `json:"url"`
	Scope string `json:"scope,omitempty"` // small|large|both
}

// Feed is the API payload for the dashboard widget.
type Feed struct {
	Items     []Item `json:"items"`
	SourceURL string `json:"sourceUrl"`
	Locale    string `json:"locale"`
}

// Client fetches AFSCA index pages. Safe for concurrent use; no package globals.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient returns a client pointed at the public AFSCA site.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBase,
		HTTPClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *Client) base() string {
	if c == nil || strings.TrimSpace(c.BaseURL) == "" {
		return DefaultBase
	}
	return strings.TrimRight(c.BaseURL, "/")
}

func (c *Client) http() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 12 * time.Second}
}

// ResolveLocale maps app locale → AFSCA index language (fr|nl only).
func ResolveLocale(appLocale string) string {
	if strings.EqualFold(strings.TrimSpace(appLocale), "nl") {
		return "nl"
	}
	return "fr"
}

// ClampLimit bounds a requested page size for API clients.
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// ClampFetchPool bounds how many index entries we parse/tag for the shared cache.
func ClampFetchPool(n int) int {
	if n <= 0 {
		return FetchPool
	}
	if n > FetchPool {
		return FetchPool
	}
	return n
}

// IndexURL returns the public AFSCA newsletters index for fr or nl.
func (c *Client) IndexURL(locale string) string {
	locale = ResolveLocale(locale)
	base := c.base()
	if locale == "nl" {
		return base + "/nl/themas/zelfstandige-beroepen/zelfstandige-dierenartsen/nieuwsbrief-voor-dierenartsen"
	}
	return base + "/fr/themes/metiers-independants/veterinaires-independants/newsletters-pour-les-veterinaires"
}

// IndexURL is a convenience using DefaultBase (soft-fail responses, docs).
func IndexURL(locale string) string {
	return NewClient().IndexURL(locale)
}

var entryRe = regexp.MustCompile(
	`(?is)<p>\s*(\d{2}/\d{2}/\d{4})\s*-\s*<a\s+[^>]*href="([^"]+)"[^>]*>\s*([^<]+?)\s*</a>\s*<br\s*/?>\s*([^<]+?)\s*</p>`,
)

// ParseIndexHTML extracts newsletter entries from the AFSCA Drupal page body.
// Duplicates (same URL) keep the first occurrence (most recent on the page).
func ParseIndexHTML(raw string, limit int) []Item {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > FetchPool {
		limit = FetchPool
	}
	matches := entryRe.FindAllStringSubmatch(raw, -1)
	out := make([]Item, 0, limit)
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		if len(m) < 5 {
			continue
		}
		url := strings.TrimSpace(html.UnescapeString(m[2]))
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		title := strings.TrimSpace(html.UnescapeString(m[4]))
		title = strings.ReplaceAll(title, "\u00a0", " ")
		title = strings.TrimSpace(title)
		label := strings.TrimSpace(html.UnescapeString(m[3]))
		item := Item{
			Date:  parseDMY(m[1]),
			Label: label,
			Title: title,
			URL:   url,
		}
		if item.Title == "" {
			item.Title = label
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func parseDMY(s string) string {
	t, err := time.Parse("02/01/2006", strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// Fetch loads and parses the AFSCA index for the given app locale.
func (c *Client) Fetch(ctx context.Context, appLocale string, limit int) (Feed, error) {
	if c == nil {
		c = NewClient()
	}
	limit = ClampFetchPool(limit)
	locale := ResolveLocale(appLocale)
	source := c.IndexURL(locale)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return Feed{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	res, err := c.http().Do(req)
	if err != nil {
		return Feed{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
		return Feed{}, fmt.Errorf("afsca_http_%d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBodyBytes))
	if err != nil {
		return Feed{}, err
	}
	items := ParseIndexHTML(string(body), limit)
	return Feed{Items: items, SourceURL: source, Locale: locale}, nil
}
