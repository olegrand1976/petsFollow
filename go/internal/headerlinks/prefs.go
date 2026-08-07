package headerlinks

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// Prefs is the JSONB stored on practice.practices.header_links.
type Prefs struct {
	// Configured distinguishes never-touched ({}) from an explicit empty selection.
	Configured bool         `json:"configured,omitempty"`
	Enabled    []string     `json:"enabled"`
	Order      []string     `json:"order"`
	Custom     []CustomLink `json:"custom"`
}

// CustomLink is a practice-defined external link.
type CustomLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ResolvedLink is a topbar-ready link.
type ResolvedLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind"` // catalog | custom
}

// CatalogItem is a catalog entry with enabled flag for settings UI.
type CatalogItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// EmptyPrefs returns a zero prefs object.
func EmptyPrefs() Prefs {
	return Prefs{Enabled: []string{}, Order: []string{}, Custom: []CustomLink{}}
}

// ParsePrefs unmarshals JSONB (empty / null → empty prefs).
func ParsePrefs(raw []byte) Prefs {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return EmptyPrefs()
	}
	var p Prefs
	if err := json.Unmarshal(raw, &p); err != nil {
		return EmptyPrefs()
	}
	if p.Enabled == nil {
		p.Enabled = []string{}
	}
	if p.Order == nil {
		p.Order = []string{}
	}
	if p.Custom == nil {
		p.Custom = []CustomLink{}
	}
	return p
}

// NormalizeAndValidate cleans prefs for persistence. countryCode drives catalog allowlist.
func NormalizeAndValidate(countryCode string, in Prefs) (Prefs, error) {
	out := EmptyPrefs()
	catalog := CatalogForCountry(countryCode)
	allowed := map[string]struct{}{}
	for _, e := range catalog {
		allowed[e.ID] = struct{}{}
	}

	customByID := map[string]CustomLink{}
	seenURL := map[string]struct{}{}
	for _, c := range in.Custom {
		label := strings.TrimSpace(c.Label)
		rawURL := strings.TrimSpace(c.URL)
		if label == "" && rawURL == "" {
			continue
		}
		if label == "" || len(label) > MaxLabelLen {
			return Prefs{}, fmt.Errorf("invalid_custom_label")
		}
		normURL, err := normalizeHTTPSURL(rawURL)
		if err != nil {
			return Prefs{}, err
		}
		key := strings.ToLower(normURL)
		if _, dup := seenURL[key]; dup {
			return Prefs{}, fmt.Errorf("duplicate_custom_url")
		}
		seenURL[key] = struct{}{}
		id := strings.TrimSpace(c.ID)
		if id == "" || !strings.HasPrefix(id, "custom_") {
			id = "custom_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
		}
		if _, exists := customByID[id]; exists {
			id = "custom_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
		}
		customByID[id] = CustomLink{ID: id, Label: label, URL: normURL}
		if len(customByID) > MaxCustomLinks {
			return Prefs{}, fmt.Errorf("too_many_custom_links")
		}
	}
	for _, c := range customByID {
		out.Custom = append(out.Custom, c)
	}
	// Stable order of custom slice by walking input order then leftovers.
	out.Custom = reorderCustoms(in.Custom, customByID)

	enabledSet := map[string]struct{}{}
	for _, id := range in.Enabled {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if strings.HasPrefix(id, "custom_") {
			if _, ok := customByID[id]; !ok {
				return Prefs{}, fmt.Errorf("unknown_custom_id")
			}
			enabledSet[id] = struct{}{}
			continue
		}
		if _, ok := allowed[id]; !ok {
			return Prefs{}, fmt.Errorf("invalid_catalog_id")
		}
		enabledSet[id] = struct{}{}
	}
	// Customs are always "enabled" when present.
	for id := range customByID {
		enabledSet[id] = struct{}{}
	}

	// Never configured → apply country defaults. Explicit save always sets Configured.
	if !in.Configured && len(in.Enabled) == 0 && len(in.Custom) == 0 && len(in.Order) == 0 && len(catalog) > 0 {
		for _, id := range DefaultEnabledIDs(countryCode) {
			enabledSet[id] = struct{}{}
		}
	}
	out.Configured = true

	for id := range enabledSet {
		out.Enabled = append(out.Enabled, id)
	}

	seenOrder := map[string]struct{}{}
	for _, id := range in.Order {
		id = strings.TrimSpace(id)
		if _, ok := enabledSet[id]; !ok {
			continue
		}
		if _, dup := seenOrder[id]; dup {
			continue
		}
		seenOrder[id] = struct{}{}
		out.Order = append(out.Order, id)
	}
	// Append any enabled not in order (catalog order then customs).
	for _, e := range catalog {
		if _, ok := enabledSet[e.ID]; !ok {
			continue
		}
		if _, done := seenOrder[e.ID]; done {
			continue
		}
		seenOrder[e.ID] = struct{}{}
		out.Order = append(out.Order, e.ID)
	}
	for _, c := range out.Custom {
		if _, done := seenOrder[c.ID]; done {
			continue
		}
		seenOrder[c.ID] = struct{}{}
		out.Order = append(out.Order, c.ID)
	}

	return out, nil
}

func reorderCustoms(input []CustomLink, byID map[string]CustomLink) []CustomLink {
	out := make([]CustomLink, 0, len(byID))
	seen := map[string]struct{}{}
	for _, c := range input {
		id := strings.TrimSpace(c.ID)
		link, ok := byID[id]
		if !ok {
			// Newly assigned id may not match input — match by URL.
			for _, v := range byID {
				if strings.EqualFold(v.URL, strings.TrimSpace(c.URL)) {
					link = v
					ok = true
					break
				}
			}
		}
		if !ok {
			continue
		}
		if _, dup := seen[link.ID]; dup {
			continue
		}
		seen[link.ID] = struct{}{}
		out = append(out, link)
	}
	for id, link := range byID {
		if _, dup := seen[id]; dup {
			continue
		}
		out = append(out, link)
	}
	return out
}

func normalizeHTTPSURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > MaxURLLen {
		return "", fmt.Errorf("invalid_custom_url")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", fmt.Errorf("invalid_custom_url")
	}
	// Reject credentials / javascript-like hosts.
	if u.User != nil {
		return "", fmt.Errorf("invalid_custom_url")
	}
	return u.String(), nil
}

// EffectiveEnabled returns catalog IDs that should show, applying defaults when prefs never configured.
func EffectiveEnabled(countryCode string, prefs Prefs) map[string]struct{} {
	out := map[string]struct{}{}
	if !prefs.Configured && len(prefs.Enabled) == 0 && len(prefs.Custom) == 0 && len(prefs.Order) == 0 {
		for _, id := range DefaultEnabledIDs(countryCode) {
			out[id] = struct{}{}
		}
		return out
	}
	for _, id := range prefs.Enabled {
		if strings.HasPrefix(id, "custom_") {
			continue
		}
		if _, ok := CatalogByID(countryCode, id); ok {
			out[id] = struct{}{}
		}
	}
	return out
}

// Resolve builds the ordered topbar list. labelFn resolves catalog labels (i18n).
func Resolve(countryCode, locale string, prefs Prefs, labelFn func(id string) string) []ResolvedLink {
	enabled := EffectiveEnabled(countryCode, prefs)
	customByID := map[string]CustomLink{}
	for _, c := range prefs.Custom {
		customByID[c.ID] = c
	}

	order := prefs.Order
	if len(order) == 0 {
		for _, e := range CatalogForCountry(countryCode) {
			if _, ok := enabled[e.ID]; ok {
				order = append(order, e.ID)
			}
		}
		for _, c := range prefs.Custom {
			order = append(order, c.ID)
		}
	}

	seen := map[string]struct{}{}
	var items []ResolvedLink
	appendCatalog := func(id string) {
		e, ok := CatalogByID(countryCode, id)
		if !ok {
			return
		}
		if _, on := enabled[id]; !on {
			return
		}
		if _, dup := seen[id]; dup {
			return
		}
		seen[id] = struct{}{}
		label := id
		if labelFn != nil {
			label = labelFn(id)
		}
		items = append(items, ResolvedLink{
			ID: id, Label: label, URL: e.URLFor(locale), Kind: "catalog",
		})
	}
	appendCustom := func(id string) {
		c, ok := customByID[id]
		if !ok {
			return
		}
		if _, dup := seen[id]; dup {
			return
		}
		seen[id] = struct{}{}
		items = append(items, ResolvedLink{
			ID: c.ID, Label: c.Label, URL: c.URL, Kind: "custom",
		})
	}

	for _, id := range order {
		if strings.HasPrefix(id, "custom_") {
			appendCustom(id)
		} else {
			appendCatalog(id)
		}
	}
	for id := range enabled {
		appendCatalog(id)
	}
	for _, c := range prefs.Custom {
		appendCustom(c.ID)
	}
	return items
}

// CatalogForSettings returns catalog rows with enabled flags for the settings UI.
func CatalogForSettings(countryCode, locale string, prefs Prefs, labelFn func(id string) string) []CatalogItem {
	enabled := EffectiveEnabled(countryCode, prefs)
	var out []CatalogItem
	for _, e := range CatalogForCountry(countryCode) {
		label := e.ID
		if labelFn != nil {
			label = labelFn(e.ID)
		}
		_, on := enabled[e.ID]
		out = append(out, CatalogItem{
			ID: e.ID, Label: label, URL: e.URLFor(locale), Enabled: on,
		})
	}
	return out
}

// MarshalPrefs encodes prefs for DB storage.
func MarshalPrefs(p Prefs) ([]byte, error) {
	if p.Enabled == nil {
		p.Enabled = []string{}
	}
	if p.Order == nil {
		p.Order = []string{}
	}
	if p.Custom == nil {
		p.Custom = []CustomLink{}
	}
	return json.Marshal(p)
}
