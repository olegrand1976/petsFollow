package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ExportUserData — portabilité RGPD (art. 20) : agrège les données personnelles
// de l'utilisateur. Toujours inclut les agrégats « client-owned » (vides si
// absents) + les agrégats pro (factures / filiation élargie) — dual care_pro inclus.
func (s *Store) ExportUserData(ctx context.Context, userID string) (map[string]any, error) {
	out := map[string]any{}

	queries := map[string]string{
		"profile": `SELECT to_jsonb(u) - 'password_hash' - 'totp_secret' - 'google_sub'
			FROM identity.users u WHERE id = $1`,
		"visitReportsAuthored": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(vr) - 'audio_object_key' - 'audio_url') ORDER BY vr.created_at), '[]'::jsonb)
			FROM visits.visit_reports vr WHERE vr.author_user_id = $1`,
		"aiCrUsageEvents": `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM practice.ai_cr_usage_events e WHERE e.user_id = $1`,
		"aiCrFeedback": `SELECT COALESCE(jsonb_agg(to_jsonb(f) ORDER BY f.created_at), '[]'::jsonb)
			FROM practice.ai_cr_feedback f WHERE f.user_id = $1`,
		"pets": `SELECT COALESCE(jsonb_agg(to_jsonb(p) ORDER BY p.created_at), '[]'::jsonb)
			FROM pets.pets p WHERE p.owner_user_id = $1`,
		"heartRateSessions": `SELECT COALESCE(jsonb_agg(to_jsonb(h) ORDER BY h.started_at), '[]'::jsonb)
			FROM heartrate.sessions h JOIN pets.pets p ON p.id = h.pet_id WHERE p.owner_user_id = $1`,
		"weightReadings": `SELECT COALESCE(jsonb_agg(to_jsonb(w) ORDER BY w.recorded_at), '[]'::jsonb)
			FROM pets.weight_readings w JOIN pets.pets p ON p.id = w.pet_id WHERE p.owner_user_id = $1`,
		"bloodPressureReadings": `SELECT COALESCE(jsonb_agg(to_jsonb(b) ORDER BY b.recorded_at), '[]'::jsonb)
			FROM pets.blood_pressure_readings b JOIN pets.pets p ON p.id = b.pet_id WHERE p.owner_user_id = $1`,
		"labPanels": `SELECT COALESCE(jsonb_agg(to_jsonb(lp) ORDER BY lp.collected_at), '[]'::jsonb)
			FROM labs.panels lp JOIN pets.pets p ON p.id = lp.pet_id WHERE p.owner_user_id = $1`,
		"labPanelResults": `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.analyte_code), '[]'::jsonb)
			FROM labs.panel_results r
			JOIN labs.panels lp ON lp.id = r.panel_id
			JOIN pets.pets p ON p.id = lp.pet_id WHERE p.owner_user_id = $1`,
		"visits": `SELECT COALESCE(jsonb_agg(to_jsonb(v) ORDER BY v.created_at), '[]'::jsonb)
			FROM visits.visits v JOIN pets.pets p ON p.id = v.pet_id WHERE p.owner_user_id = $1`,
		"preconsultIntakes": `SELECT COALESCE(jsonb_agg(to_jsonb(i) ORDER BY i.created_at), '[]'::jsonb)
			FROM visits.preconsult_intakes i
			JOIN visits.visits v ON v.id = i.visit_id
			JOIN pets.pets p ON p.id = v.pet_id WHERE p.owner_user_id = $1`,
		"visitReportExplanations": `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM visits.visit_report_explanations e
			JOIN visits.visits v ON v.id = e.visit_id
			JOIN pets.pets p ON p.id = v.pet_id WHERE p.owner_user_id = $1`,
		"clientAiTriageSessions": `SELECT COALESCE(jsonb_agg(to_jsonb(s) ORDER BY s.created_at), '[]'::jsonb)
			FROM (
				SELECT ts.id, ts.pet_id, ts.practice_id, ts.created_at, ts.updated_at,
					COALESCE((
						SELECT jsonb_agg(jsonb_build_object(
							'id', m.id,
							'role', m.role,
							'body', m.body,
							'level', m.level,
							'watchSigns', m.watch_signs,
							'recommendedAction', m.recommended_action,
							'createdAt', m.created_at
						) ORDER BY m.created_at)
						FROM client_ai.triage_messages m WHERE m.session_id = ts.id
					), '[]'::jsonb) AS messages
				FROM client_ai.triage_sessions ts
				WHERE ts.user_id = $1
			) s`,
		"clientAiUsageEvents": `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM client_ai.usage_events e WHERE e.user_id = $1`,
		"messages": `SELECT COALESCE(jsonb_agg(to_jsonb(m) ORDER BY m.created_at), '[]'::jsonb)
			FROM messaging.messages m JOIN messaging.threads t ON t.id = m.thread_id
			WHERE t.client_user_id = $1`,
		"careReminders": `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY c.due_at), '[]'::jsonb)
			FROM care.reminders c JOIN pets.pets p ON p.id = c.pet_id WHERE p.owner_user_id = $1`,
		"entitlements": `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM billing.pet_entitlements e WHERE e.owner_user_id = $1`,
		"addons": `SELECT COALESCE(jsonb_agg(to_jsonb(a) ORDER BY a.created_at), '[]'::jsonb)
			FROM billing.addon_entitlements a WHERE a.owner_user_id = $1`,
		"vetLinks": `SELECT COALESCE(jsonb_agg(to_jsonb(pc) ORDER BY pc.created_at), '[]'::jsonb)
			FROM practice.practice_clients pc WHERE pc.client_user_id = $1`,
		"vetLinkRequests": `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.created_at), '[]'::jsonb)
			FROM practice.client_vet_link_requests r WHERE r.client_user_id = $1`,
		"vetLeads": `SELECT COALESCE(jsonb_agg(to_jsonb(l) ORDER BY l.created_at), '[]'::jsonb)
			FROM practice.vet_leads l WHERE l.client_user_id = $1`,
		"commercialReferrals": `SELECT COALESCE(jsonb_agg(to_jsonb(cr) ORDER BY cr.created_at), '[]'::jsonb)
			FROM practice.commercial_referrals cr WHERE cr.client_user_id = $1`,
		"clientReferrals": `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.created_at), '[]'::jsonb)
			FROM practice.client_referrals r
			WHERE r.referred_client_user_id = $1 OR r.sponsor_client_user_id = $1`,
		// Union client + pro (dual care_pro / commercial actor).
		"filiationEvents": `SELECT COALESCE(jsonb_agg(to_jsonb(e) ORDER BY e.created_at), '[]'::jsonb)
			FROM practice.filiation_events e
			WHERE e.client_user_id = $1 OR e.actor_user_id = $1
			   OR e.commercial_user_id = $1 OR e.vet_user_id = $1`,
		"dossierShares": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(t) - 'object_key' - 'token') ORDER BY t.created_at), '[]'::jsonb)
			FROM pets.dossier_share_tokens t WHERE t.owner_user_id = $1`,
		"consultationShares": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(t) - 'object_key' - 'token') ORDER BY t.created_at), '[]'::jsonb)
			FROM pets.consultation_share_tokens t WHERE t.owner_user_id = $1`,
		"prescriptions": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(rx) - 'pdf_object_key' - 'pdf_sha256' - 'signature_id') ORDER BY rx.created_at), '[]'::jsonb)
			FROM prescriptions.prescriptions rx WHERE rx.owner_id = $1`,
		"petDocuments": `SELECT COALESCE(jsonb_agg(jsonb_build_object(
			'id', d.id,
			'petId', d.pet_id,
			'title', d.title,
			'fileName', d.file_name,
			'contentType', d.content_type,
			'sizeBytes', d.size_bytes,
			'createdAt', d.created_at
		) ORDER BY d.created_at), '[]'::jsonb)
			FROM pets.documents d
			JOIN pets.pets p ON p.id = d.pet_id WHERE p.owner_user_id = $1`,
		"imagingStudies": `SELECT COALESCE(jsonb_agg(jsonb_build_object(
			'id', ps.id,
			'petId', ps.pet_id,
			'practiceId', ps.practice_id,
			'orthancStudyId', ps.orthanc_study_id,
			'studyInstanceUid', ps.study_instance_uid,
			'orthancSeriesId', ps.orthanc_series_id,
			'description', ps.description,
			'modality', ps.modality,
			'createdAt', ps.created_at,
			'comments', COALESCE((
				SELECT jsonb_agg(jsonb_build_object(
					'id', c.id,
					'body', c.body,
					'createdAt', c.created_at,
					'authorUserId', c.author_user_id
				) ORDER BY c.created_at)
				FROM imaging.pet_study_comments c
				WHERE c.pet_study_id = ps.id
			), '[]'::jsonb)
		) ORDER BY ps.created_at), '[]'::jsonb)
			FROM imaging.pet_studies ps
			JOIN pets.pets p ON p.id = ps.pet_id WHERE p.owner_user_id = $1`,
		"deviceTokens": `SELECT COALESCE(jsonb_agg(jsonb_build_object(
			'platform', t.platform,
			'updatedAt', t.updated_at
		) ORDER BY t.updated_at), '[]'::jsonb)
			FROM notifications.device_tokens t WHERE t.user_id = $1`,
		"invoicingDocuments": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(d) - 'idempotency_key') ORDER BY d.created_at), '[]'::jsonb)
			FROM invoicing.documents d WHERE d.created_by = $1`,
		"pharmacyDafAsClient": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(d) - 'pdf_object_key') ORDER BY d.created_at), '[]'::jsonb)
			FROM pharmacy.daf_documents d WHERE d.client_user_id = $1`,
		"pharmacyDafAsPetOwner": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(d) - 'pdf_object_key') ORDER BY d.created_at), '[]'::jsonb)
			FROM pharmacy.daf_documents d
			JOIN pets.pets p ON p.id = d.pet_id
			WHERE p.owner_user_id = $1`,
		"pharmacyDafAsPrescriber": `SELECT COALESCE(jsonb_agg(
			(to_jsonb(d) - 'pdf_object_key') ORDER BY d.created_at), '[]'::jsonb)
			FROM pharmacy.daf_documents d WHERE d.prescriber_user_id = $1`,
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

	supportRaw, err := s.ListSupportTicketsForExport(ctx, userID)
	if err != nil {
		return nil, err
	}
	var support any
	if err := json.Unmarshal(supportRaw, &support); err != nil {
		return nil, err
	}
	out["supportTickets"] = support

	return out, nil
}
