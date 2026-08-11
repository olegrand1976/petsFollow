package vetnews

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 15 * time.Second
	maxBodyBytes       = 3 << 20 // 3 MiB
	userAgent          = "petsFollow/1.0 (+https://petsfollow.app; vet-news-ingest)"
	maxRedirects       = 5
)

// HTTPClient is shared by providers (injectable in tests).
type HTTPClient struct {
	Client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{Client: newSafeHTTPClient()}
}

func newSafeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultHTTPTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("too_many_redirects")
			}
			if err := rejectPrivateOrNonHTTPS(req.URL); err != nil {
				return err
			}
			return nil
		},
	}
}

func (c *HTTPClient) http() *http.Client {
	if c != nil && c.Client != nil {
		return c.Client
	}
	return newSafeHTTPClient()
}

// Get fetches a URL and returns body bytes. Non-200 → error.
func (c *HTTPClient) Get(ctx context.Context, urlStr, accept string) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(urlStr))
	if err != nil {
		return nil, err
	}
	if err := rejectPrivateOrNonHTTPS(u); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

func rejectPrivateOrNonHTTPS(u *url.URL) error {
	if u == nil {
		return fmt.Errorf("invalid_url")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("https_required")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("invalid_host")
	}
	// Literal IPs / localhost only — DNS rebinding to RFC1918 is out of scope here
	// (ingest URLs are provider-fixed, not user-controlled).
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("private_ip_forbidden")
		}
		return nil
	}
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") || strings.HasSuffix(lower, ".local") {
		return fmt.Errorf("private_host_forbidden")
	}
	return nil
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
