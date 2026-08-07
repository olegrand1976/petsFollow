package vetnews

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 15 * time.Second
	maxBodyBytes       = 3 << 20 // 3 MiB
	userAgent          = "petsFollow/1.0 (+https://petsfollow.app; vet-news-ingest)"
)

// HTTPClient is shared by providers (injectable in tests).
type HTTPClient struct {
	Client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{Client: &http.Client{Timeout: defaultHTTPTimeout}}
}

func (c *HTTPClient) http() *http.Client {
	if c != nil && c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: defaultHTTPTimeout}
}

// Get fetches a URL and returns body bytes. Non-200 → error.
func (c *HTTPClient) Get(ctx context.Context, url, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	if accept != "" {
		req.Header.Set("Accept", accept)
	} else {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/rss+xml,application/xml;q=0.9,*/*;q=0.8")
	}
	res, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
		return nil, fmt.Errorf("http_%d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBodyBytes))
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ResolveURL joins relative href against base.
func ResolveURL(base, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasPrefix(href, "//") {
		if strings.HasPrefix(base, "https:") {
			return "https:" + href
		}
		return "http:" + href
	}
	if strings.HasPrefix(href, "/") {
		// scheme+host from base
		schemeEnd := strings.Index(base, "://")
		if schemeEnd < 0 {
			return base + href
		}
		rest := base[schemeEnd+3:]
		slash := strings.Index(rest, "/")
		host := rest
		if slash >= 0 {
			host = rest[:slash]
		}
		return base[:schemeEnd+3] + host + href
	}
	return base + "/" + href
}
