package store

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// Applicabilité du DAF pour une espèce dans un pays donné.
//
//	never      — espèce non productrice de denrées alimentaires
//	always     — espèce productrice de denrées par nature
//	per_animal — statut décidé par individu (équidés : passeport) ; le DAF reste
//	             possible sauf si le patient est excluded_from_food_chain
const (
	DAFRequiredNever     = "never"
	DAFRequiredAlways    = "always"
	DAFRequiredPerAnimal = "per_animal"
)

// speciesCacheTTL borne la fraîcheur du catalogue en lecture. Les mutations admin
// invalident le cache explicitement, ce TTL n'est qu'un filet.
const speciesCacheTTL = 60 * time.Second

var speciesCodeRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,39}$`)

// Species est une entrée du catalogue, indépendante de tout pays.
type Species struct {
	Code              string            `json:"code"`
	Labels            map[string]string `json:"labels"`
	SortOrder         int               `json:"sortOrder"`
	IsActive          bool              `json:"isActive"`
	SupportsHeartRate bool              `json:"supportsHeartRate"`
}

// SpeciesRule porte les règles réglementaires d'une espèce dans un pays.
type SpeciesRule struct {
	CountryCode            string `json:"countryCode"`
	IsLargeAnimal          bool   `json:"isLargeAnimal"`
	DAFRequired            string `json:"dafRequired"`
	IsFoodChain            bool   `json:"isFoodChain"`
	DefaultFoodChainStatus string `json:"defaultFoodChainStatus"`
}

// SpeciesWithRule est ce que consomment les fronts : catalogue + règles résolues.
type SpeciesWithRule struct {
	Species
	Rule SpeciesRule `json:"rule"`
}

type speciesCacheEntry struct {
	rows     []SpeciesWithRule
	fetchedA time.Time
}

// ValidSpeciesDAFRequired reports whether v is an accepted daf_required value.
func ValidSpeciesDAFRequired(v string) bool {
	switch v {
	case DAFRequiredNever, DAFRequiredAlways, DAFRequiredPerAnimal:
		return true
	default:
		return false
	}
}

// ValidFoodChainStatus reports whether v is an accepted food_chain_status value.
func ValidFoodChainStatus(v string) bool {
	switch v {
	case "companion", "food_producing", "excluded_from_food_chain":
		return true
	default:
		return false
	}
}

// fallbackSpeciesRule est la règle appliquée quand la table ne connaît pas le couple
// (espèce, pays) : espèce libre héritée, ou pays sans jeu de règles seedé.
//
// Elle reproduit le comportement historique de go/pkg/kernel pour la chaîne alimentaire,
// et autorise le DAF « selon l'animal » sur les espèces de rente — de sorte qu'un pays
// non seedé ne bloque pas les flux légitimes, sans jamais ouvrir le DAF aux animaux de
// compagnie.
func fallbackSpeciesRule(species, countryCode string) SpeciesRule {
	foodChain := kernel.IsFoodChainSpecies(species)
	daf := DAFRequiredNever
	if foodChain {
		daf = DAFRequiredPerAnimal
	}
	return SpeciesRule{
		CountryCode:            countryCode,
		IsLargeAnimal:          false,
		DAFRequired:            daf,
		IsFoodChain:            foodChain,
		DefaultFoodChainStatus: kernel.DefaultFoodChainStatus(species),
	}
}

func (s *Store) invalidateSpeciesCache() {
	s.speciesMu.Lock()
	s.speciesCache = nil
	s.speciesMu.Unlock()
}

func (s *Store) cachedSpecies(countryCode string) ([]SpeciesWithRule, bool) {
	s.speciesMu.RLock()
	defer s.speciesMu.RUnlock()
	e, ok := s.speciesCache[countryCode]
	if !ok || time.Since(e.fetchedA) > speciesCacheTTL {
		return nil, false
	}
	return e.rows, true
}

func (s *Store) storeSpeciesCache(countryCode string, rows []SpeciesWithRule) {
	s.speciesMu.Lock()
	defer s.speciesMu.Unlock()
	if s.speciesCache == nil {
		s.speciesCache = map[string]speciesCacheEntry{}
	}
	s.speciesCache[countryCode] = speciesCacheEntry{rows: rows, fetchedA: time.Now()}
}

// ListSpecies returns the whole catalogue with the rules resolved for countryCode.
// Species without a rule row for that country fall back to kernel behaviour.
// Rows are ordered by sort_order — `other` is seeded last on purpose.
func (s *Store) ListSpecies(ctx context.Context, countryCode string) ([]SpeciesWithRule, error) {
	countryCode = NormalizeCountryCode(countryCode)
	if rows, ok := s.cachedSpecies(countryCode); ok {
		return rows, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT s.code, s.labels, s.sort_order, s.is_active, s.supports_heart_rate,
			r.is_large_animal, r.daf_required, r.is_food_chain, r.default_food_chain_status
		FROM pets.species s
		LEFT JOIN pets.species_country_rules r
			ON r.species_code = s.code AND r.country_code = $1
		ORDER BY s.sort_order, s.code`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SpeciesWithRule{}
	for rows.Next() {
		var it SpeciesWithRule
		var rawLabels []byte
		var large, foodChain *bool
		var daf, status *string
		if err := rows.Scan(
			&it.Code, &rawLabels, &it.SortOrder, &it.IsActive, &it.SupportsHeartRate,
			&large, &daf, &foodChain, &status,
		); err != nil {
			return nil, err
		}
		it.Labels = map[string]string{}
		if len(rawLabels) > 0 {
			// Un JSONB corrompu ne doit pas casser la liste : libellés vides, le front
			// retombe sur la clé i18n compilée puis sur le code.
			_ = json.Unmarshal(rawLabels, &it.Labels)
		}
		if daf == nil {
			it.Rule = fallbackSpeciesRule(it.Code, countryCode)
		} else {
			it.Rule = SpeciesRule{
				CountryCode:            countryCode,
				IsLargeAnimal:          large != nil && *large,
				DAFRequired:            *daf,
				IsFoodChain:            foodChain != nil && *foodChain,
				DefaultFoodChainStatus: *status,
			}
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.storeSpeciesCache(countryCode, out)
	return out, nil
}

// ListActiveSpecies filters ListSpecies down to what a picker should offer.
func (s *Store) ListActiveSpecies(ctx context.Context, countryCode string) ([]SpeciesWithRule, error) {
	all, err := s.ListSpecies(ctx, countryCode)
	if err != nil {
		return nil, err
	}
	out := make([]SpeciesWithRule, 0, len(all))
	for _, it := range all {
		if it.IsActive {
			out = append(out, it)
		}
	}
	return out, nil
}

// ResolveSpeciesRule returns the regulatory rule for a species in a country,
// falling back to kernel behaviour when the catalogue does not know the pair.
func (s *Store) ResolveSpeciesRule(ctx context.Context, species, countryCode string) (SpeciesRule, error) {
	species = strings.ToLower(strings.TrimSpace(species))
	countryCode = NormalizeCountryCode(countryCode)
	if species == "" {
		return fallbackSpeciesRule(species, countryCode), nil
	}
	all, err := s.ListSpecies(ctx, countryCode)
	if err != nil {
		return SpeciesRule{}, err
	}
	for _, it := range all {
		if it.Code == species {
			return it.Rule, nil
		}
	}
	return fallbackSpeciesRule(species, countryCode), nil
}

// GetSpecies returns one catalogue entry (rules excluded).
func (s *Store) GetSpecies(ctx context.Context, code string) (Species, error) {
	var sp Species
	var rawLabels []byte
	err := s.pool.QueryRow(ctx, `
		SELECT code, labels, sort_order, is_active, supports_heart_rate
		FROM pets.species WHERE code = $1`, strings.TrimSpace(code)).
		Scan(&sp.Code, &rawLabels, &sp.SortOrder, &sp.IsActive, &sp.SupportsHeartRate)
	if errors.Is(err, pgx.ErrNoRows) {
		return Species{}, ErrNotFound
	}
	if err != nil {
		return Species{}, err
	}
	sp.Labels = map[string]string{}
	if len(rawLabels) > 0 {
		_ = json.Unmarshal(rawLabels, &sp.Labels)
	}
	return sp, nil
}

// CreateSpecies inserts a new catalogue entry. Codes are immutable once created.
func (s *Store) CreateSpecies(ctx context.Context, sp Species) (Species, error) {
	sp.Code = strings.ToLower(strings.TrimSpace(sp.Code))
	if !speciesCodeRe.MatchString(sp.Code) {
		return Species{}, ErrValidation
	}
	if len(sp.Labels) == 0 {
		return Species{}, ErrValidation
	}
	labels, err := json.Marshal(sp.Labels)
	if err != nil {
		return Species{}, ErrValidation
	}
	if sp.SortOrder <= 0 {
		sp.SortOrder = 100
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO pets.species (code, labels, sort_order, is_active, supports_heart_rate)
		VALUES ($1, $2::jsonb, $3, $4, $5)
		ON CONFLICT (code) DO NOTHING`,
		sp.Code, string(labels), sp.SortOrder, sp.IsActive, sp.SupportsHeartRate)
	if err != nil {
		return Species{}, err
	}
	if tag.RowsAffected() == 0 {
		return Species{}, ErrConflict
	}
	s.invalidateSpeciesCache()
	return s.GetSpecies(ctx, sp.Code)
}

// SpeciesPatch carries the admin-editable catalogue fields; nil means "leave as is".
type SpeciesPatch struct {
	Labels            map[string]string
	SortOrder         *int
	IsActive          *bool
	SupportsHeartRate *bool
}

// UpdateSpecies applies a partial update to an existing catalogue entry.
func (s *Store) UpdateSpecies(ctx context.Context, code string, p SpeciesPatch) (Species, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if _, err := s.GetSpecies(ctx, code); err != nil {
		return Species{}, err
	}
	var labels *string
	if p.Labels != nil {
		if len(p.Labels) == 0 {
			return Species{}, ErrValidation
		}
		raw, err := json.Marshal(p.Labels)
		if err != nil {
			return Species{}, ErrValidation
		}
		v := string(raw)
		labels = &v
	}
	if p.SortOrder != nil && *p.SortOrder <= 0 {
		return Species{}, ErrValidation
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE pets.species SET
			labels = COALESCE($2::jsonb, labels),
			sort_order = COALESCE($3, sort_order),
			is_active = COALESCE($4, is_active),
			supports_heart_rate = COALESCE($5, supports_heart_rate),
			updated_at = NOW()
		WHERE code = $1`, code, labels, p.SortOrder, p.IsActive, p.SupportsHeartRate)
	if err != nil {
		return Species{}, err
	}
	s.invalidateSpeciesCache()
	return s.GetSpecies(ctx, code)
}

// UpsertSpeciesRule writes the regulatory rule for (species, country).
// The species must already exist in the catalogue.
func (s *Store) UpsertSpeciesRule(ctx context.Context, speciesCode string, r SpeciesRule) error {
	speciesCode = strings.ToLower(strings.TrimSpace(speciesCode))
	country := NormalizeCountryCode(r.CountryCode)
	if strings.TrimSpace(r.CountryCode) != "" && country != strings.ToUpper(strings.TrimSpace(r.CountryCode)) {
		// Pays hors allowlist : refuser plutôt que retomber silencieusement sur BE.
		return ErrValidation
	}
	if !ValidSpeciesDAFRequired(r.DAFRequired) || !ValidFoodChainStatus(r.DefaultFoodChainStatus) {
		return ErrValidation
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pets.species WHERE code = $1)`, speciesCode).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO pets.species_country_rules (
			species_code, country_code, is_large_animal, daf_required, is_food_chain, default_food_chain_status
		) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (species_code, country_code) DO UPDATE SET
			is_large_animal = EXCLUDED.is_large_animal,
			daf_required = EXCLUDED.daf_required,
			is_food_chain = EXCLUDED.is_food_chain,
			default_food_chain_status = EXCLUDED.default_food_chain_status,
			updated_at = NOW()`,
		speciesCode, country, r.IsLargeAnimal, r.DAFRequired, r.IsFoodChain, r.DefaultFoodChainStatus,
	); err != nil {
		return err
	}
	s.invalidateSpeciesCache()
	return nil
}

// GetPracticeCountryCode returns the practice ISO country (default BE).
func (s *Store) GetPracticeCountryCode(ctx context.Context, practiceID string) (string, error) {
	var c string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(country_code,'BE') FROM practice.practices WHERE id = $1`, practiceID).Scan(&c)
	if errors.Is(err, pgx.ErrNoRows) {
		return "BE", ErrNotFound
	}
	if err != nil {
		return "BE", err
	}
	return NormalizeCountryCode(c), nil
}

// DefaultFoodChainStatusFor resolves the food-chain status to apply on pet create,
// preferring the country catalogue over the hardcoded kernel defaults.
func (s *Store) DefaultFoodChainStatusFor(ctx context.Context, species, countryCode string) string {
	rule, err := s.ResolveSpeciesRule(ctx, species, countryCode)
	if err != nil {
		return kernel.DefaultFoodChainStatus(species)
	}
	return rule.DefaultFoodChainStatus
}
