package vetnews

import (
	"encoding/xml"
	"html"
	"strings"
)

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Enclosure   struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`
	// media:content / media:thumbnail often appear; ignore if not in default ns.
}

// ParseRSS extracts items from an RSS 2.0 document.
func ParseRSS(body []byte, limit int) ([]RawItem, error) {
	if limit <= 0 {
		limit = 20
	}
	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	out := make([]RawItem, 0, len(feed.Channel.Items))
	seen := make(map[string]struct{}, len(feed.Channel.Items))
	for _, it := range feed.Channel.Items {
		title := StripHTML(it.Title)
		link := NormalizeURL(strings.TrimSpace(html.UnescapeString(it.Link)))
		if link == "" {
			link = NormalizeURL(strings.TrimSpace(html.UnescapeString(it.GUID)))
		}
		if title == "" || link == "" {
			continue
		}
		if _, ok := seen[link]; ok {
			continue
		}
		seen[link] = struct{}{}
		summary := truncateRunes(StripHTML(it.Description), 500)
		item := RawItem{
			Title:     title,
			Summary:   summary,
			SourceURL: link,
			ImageURL:  strings.TrimSpace(it.Enclosure.URL),
		}
		if t := ParseFlexibleTime(it.PubDate); t != nil {
			item.PublishedAt = t
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
