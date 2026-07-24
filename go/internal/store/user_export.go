package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ExportUserData — portabilité RGPD (art. 20) : agrège les données personnelles
// de l'utilisateur en JSON. Les clients reçoivent l'ensemble de leurs données ;
// les comptes Pro reçoivent leur profil (les données cliniques appartiennent au cabinet).
func (s *Store) ExportUserData(ctx context.Context, userID string, fullClientExport bool) (map[string]any, error) {
	out := map[string]any{}

	queries := map[string]string{
		"profile": `SELECT to_jsonb(u) - 'password_hash' - 'totp_secret' - 'google_sub'
			FROM identity.users u WHERE id = $1`,
	}
	if fullClientExport {
		queries["pets"] = `SELECT COALESCE(jsonb_agg(to_jsonb(p) ORDER BY p.created_at), '[]'::jsonb)
			FROM pets.pets p WHERE p.owner_user_id = $1`
		queries["heartRateSessions"] = `SELECT COALESCE(jsonb_agg(to_jsonb(h) ORDER BY h.started_at), '[]'::jsonb)
			FROM heartrate.sessions h JOIN pets.pets p ON p.id = h.pet_id WHERE p.owner_user_id = $1`
		queries["visits"] = `SELECT COALESCE(jsonb_agg(to_jsonb(v) ORDER BY v.created_at), '[]'::jsonb)
			FROM visits.visits v JOIN pets.pets p ON p.id = v.pet_id WHERE p.owner_user_id = $1`
		queries["messages"] = `SELECT COALESCE(jsonb_agg(to_jsonb(m) ORDER BY m.created_at), '[]'::jsonb)
			FROM messaging.messages m JOIN messaging.threads t ON t.id = m.thread_id
			WHERE t.client_user_id = $1`
		queries["careReminders"] = `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY c.due_at), '[]'::jsonb)
			FROM care.reminders c JOIN pets.pets p ON p.id = c.pet_id WHERE p.owner_user_id = $1`
		queries["entitlements"] = `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM billing.pet_entitlements e WHERE e.owner_user_id = $1`
		queries["addons"] = `SELECT COALESCE(jsonb_agg(to_jsonb(a) ORDER BY a.created_at), '[]'::jsonb)
			FROM billing.addon_entitlements a WHERE a.owner_user_id = $1`
	}

	for key, q := range queries {
		var raw []byte
		err := s.pool.QueryRow(ctx, q, userID).Scan(&raw)
		if errors.Is(err, pgx.ErrNoRows) {
			if key == "profile" {
				return nil, ErrNotFound
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		out[key] = v
	}
	return out, nil
}
