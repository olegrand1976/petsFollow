package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/afsca"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const (
	afscaNewsCacheTTL = 6 * time.Hour
	afscaNewsFailTTL  = 10 * time.Minute
	afscaCacheKeyPref = "afsca:newsletters:v2:"
)

// SetAfscaClient injects the AFSCA HTTP client (tests / custom base URL).
func (a *API) SetAfscaClient(c *afsca.Client) {
	if c == nil {
		a.afsca = afsca.NewClient()
		return
	}
	a.afsca = c
}

func (a *API) afscaClient() *afsca.Client {
	if a.afsca == nil {
		a.afsca = afsca.NewClient()
	}
	return a.afsca
}

// listAfscaNewsletters returns recent AFSCA veterinary newsletters for BE practices only.
// Items are tagged small|large|both (Gemini lite when configured, else heuristics) and
// filtered by practice.animal_scope.
func (a *API) listAfscaNewsletters(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.read")
	if !ok {
		return
	}
	contact, err := a.store.GetPracticeContact(r.Context(), id.PracticeID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "practice_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if store.NormalizeCountryCode(contact.CountryCode) != "BE" {
		writeErr(w, r, http.StatusNotFound, "afsca_not_available", "not_found")
		return
	}
	practiceScope, err := a.store.GetPracticeAnimalScope(r.Context(), id.PracticeID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	locale := i18n.FromContext(r.Context())
	if q := strings.TrimSpace(r.URL.Query().Get("locale")); q != "" {
		locale = i18n.NormalizeLocale(q)
	}
	limit := afsca.ClampLimit(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = afsca.ClampLimit(n)
		}
	}

	feed, err := a.loadAfscaFeed(r.Context(), locale, limit, practiceScope)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, feed)
}

func (a *API) loadAfscaFeed(ctx context.Context, locale string, limit int, practiceScope string) (afsca.Feed, error) {
	limit = afsca.ClampLimit(limit)
	afscaLocale := afsca.ResolveLocale(locale)
	cacheKey := afscaCacheKeyPref + afscaLocale
	client := a.afscaClient()
	practiceScope = afsca.NormalizeScope(practiceScope)

	if a.redis != nil {
		if raw, err := a.redis.Get(ctx, cacheKey); err == nil && raw != "" {
			var cached afsca.Feed
			if json.Unmarshal([]byte(raw), &cached) == nil {
				if cached.SourceURL == "" {
					cached.SourceURL = client.IndexURL(afscaLocale)
				}
				if cached.Locale == "" {
					cached.Locale = afscaLocale
				}
				cached.Items = afsca.FilterByScope(cached.Items, practiceScope)
				if len(cached.Items) > limit {
					cached.Items = cached.Items[:limit]
				}
				return cached, nil
			}
		}
	}

	feed, err := client.Fetch(ctx, afscaLocale, afsca.FetchPool)
	if err != nil {
		log.Printf("afsca fetch locale=%s err=%v", afscaLocale, err)
		empty := afsca.Feed{
			Items:     []afsca.Item{},
			SourceURL: client.IndexURL(afscaLocale),
			Locale:    afscaLocale,
		}
		a.cacheAfscaFeed(ctx, cacheKey, empty, afscaNewsFailTTL)
		return empty, nil
	}

	feed.Items = a.tagAfscaItems(ctx, feed.Items)
	ttl := afscaNewsCacheTTL
	if len(feed.Items) == 0 {
		ttl = afscaNewsFailTTL
	}
	a.cacheAfscaFeed(ctx, cacheKey, feed, ttl)

	feed.Items = afsca.FilterByScope(feed.Items, practiceScope)
	if len(feed.Items) > limit {
		feed.Items = feed.Items[:limit]
	}
	return feed, nil
}

func (a *API) tagAfscaItems(ctx context.Context, items []afsca.Item) []afsca.Item {
	if len(items) == 0 {
		return items
	}
	allTagged := true
	for _, it := range items {
		if it.Scope == "" {
			allTagged = false
			break
		}
	}
	if allTagged {
		return items
	}
	if a.gemini != nil {
		return a.gemini.ClassifyAfscaScopes(ctx, items)
	}
	return afsca.HeuristicClassify(items)
}

func (a *API) cacheAfscaFeed(ctx context.Context, key string, feed afsca.Feed, ttl time.Duration) {
	if a.redis == nil {
		return
	}
	raw, err := json.Marshal(feed)
	if err != nil {
		return
	}
	_ = a.redis.Set(ctx, key, string(raw), ttl)
}
