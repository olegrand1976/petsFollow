package vetnews

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	tagRe    = regexp.MustCompile(`(?is)<[^>]+>`)
	scriptRe = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	wsRe     = regexp.MustCompile(`\s+`)
)

// StripHTML removes tags and collapses whitespace.
func StripHTML(raw string) string {
	s := scriptRe.ReplaceAllString(raw, " ")
	s = tagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = wsRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// NormalizeURL canonicalizes for dedup (trim, drop fragment, lowercase host scheme).
// Non-https URLs are rejected (empty) to avoid javascript:/data: poisoning in UI links.
func NormalizeURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if i := strings.Index(u, "#"); i >= 0 {
		u = u[:i]
	}
	u = strings.TrimRight(u, "/")
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return ""
	}
	parsed.Scheme = "https"
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}

// ContentHash fingerprints title/url/summary/category/importance for change detection.
func ContentHash(title, url, summary, category, importance string) string {
	payload := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(title)),
		NormalizeURL(url),
		strings.ToLower(strings.TrimSpace(summary)),
		strings.TrimSpace(category),
		strings.TrimSpace(importance),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}

// ParseFlexibleTime tries common RSS / HTML date formats → UTC.
func ParseFlexibleTime(raw string) *time.Time {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return nil
	}
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"02.01.2006",
		"01/02/2006",
		"2 January 2006",
		"January 2, 2006",
		"02 Jan 2006",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Published on 02/01/2006",
	}
	// Strip "Published on " prefix common on WOAH.
	raw2 := strings.TrimSpace(strings.TrimPrefix(raw, "Published on "))
	raw2 = strings.TrimSpace(strings.TrimPrefix(raw2, "Publié le "))
	for _, cand := range []string{raw, raw2} {
		for _, layout := range layouts {
			if t, err := time.Parse(layout, cand); err == nil {
				u := t.UTC()
				return &u
			}
		}
		// DMY with dots: 07.08.2026
		if t, err := time.Parse("02.01.2006", cand); err == nil {
			u := t.UTC()
			return &u
		}
	}
	return nil
}

func uniqueNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
				return r
			}
			if r == ' ' || r == '_' {
				return '-'
			}
			return -1
		}, s)
		s = strings.Trim(s, "-")
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
