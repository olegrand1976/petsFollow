package vetnews

import (
	"log"
	"sync"
)

// Registry holds named providers.
type Registry struct {
	mu        sync.RWMutex
	providers []Provider
}

func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{}
	for _, p := range providers {
		if p != nil {
			r.providers = append(r.providers, p)
		}
	}
	return r
}

func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Provider, len(r.providers))
	copy(out, r.providers)
	return out
}

// SourceInfo is a static catalog entry (no HTTP client).
type SourceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SourceCatalog lists configured sources without instantiating HTTP clients.
func SourceCatalog() []SourceInfo {
	return []SourceInfo{
		{ID: "anses", Name: "Anses — Santé animale"},
		{ID: "veterinary_evidence", Name: "Veterinary Evidence (RCVS Knowledge)"},
		{ID: "clinicians_brief", Name: "Clinician's Brief"},
		{ID: "todays_vet_practice", Name: "Today's Veterinary Practice"},
		{ID: "point_veterinaire", Name: "Le Point Vétérinaire"},
		{ID: "woah", Name: "WOAH / OMSA"},
	}
}

// DefaultProviders returns the production source set.
func DefaultProviders(http *HTTPClient) []Provider {
	if http == nil {
		http = NewHTTPClient()
	}
	return []Provider{
		NewAnsesProvider(http),
		NewVeterinaryEvidenceProvider(http),
		NewCliniciansBriefProvider(http),
		NewTodaysVetPracticeProvider(http),
		NewPointVeterinaireProvider(http),
		NewWOAHProvider(http),
	}
}

// LogSourceResult emits a structured CRON line.
func LogSourceResult(prefix string, sr SourceResult) {
	if sr.Error != "" {
		log.Printf("%s source=%s fetched=%d inserted=%d updated=%d unchanged=%d skipped=%d err=%s",
			prefix, sr.SourceID, sr.Fetched, sr.Inserted, sr.Updated, sr.Unchanged, sr.Skipped, sr.Error)
		return
	}
	log.Printf("%s source=%s fetched=%d inserted=%d updated=%d unchanged=%d skipped=%d",
		prefix, sr.SourceID, sr.Fetched, sr.Inserted, sr.Updated, sr.Unchanged, sr.Skipped)
}
