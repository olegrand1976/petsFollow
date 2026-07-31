package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// clientPreserveSnapshot holds a staging client's owned graph across truncate + re-seed.
type clientPreserveSnapshot struct {
	UserIDs           []string
	PracticeIDToName  map[string]string
	UserIDToEmail     map[string]string
	FallbackPractices json.RawMessage
	FallbackVets      json.RawMessage

	Profiles           json.RawMessage
	Pets               json.RawMessage
	PracticeClients    json.RawMessage
	LinkRequests       json.RawMessage
	StripeCustomers    json.RawMessage
	PetEntitlements    json.RawMessage
	AddonEntitlements  json.RawMessage
	Threads            json.RawMessage
	Messages           json.RawMessage
	WeightReadings     json.RawMessage
	DossierEvents      json.RawMessage
	Documents          json.RawMessage
	HeartRateSessions  json.RawMessage
	CareReminders      json.RawMessage
	CareContacts       json.RawMessage
	CareCompetitions   json.RawMessage
	Visits             json.RawMessage
	VisitReports       json.RawMessage
	DossierShares      json.RawMessage
	ConsultationShares json.RawMessage
	DeviceTokens       json.RawMessage
	ClientPreferences  json.RawMessage
	DiscoveryProgress  json.RawMessage
	DiscoveryJourney   json.RawMessage
	DiscoverySends     json.RawMessage
	PetAccess          json.RawMessage
	ClientAccess       json.RawMessage
	UserPracticeIDs    map[string]string // userID -> old practice_id
	UserActiveProfiles map[string]string // userID -> old active_profile_id
}

func snapshotPreserveClients(ctx context.Context, tx pgx.Tx) (*clientPreserveSnapshot, error) {
	if len(preserveClientEmails) == 0 {
		return &clientPreserveSnapshot{}, nil
	}
	rows, err := tx.Query(ctx, `
		SELECT id::text, COALESCE(practice_id::text, ''), COALESCE(active_profile_id::text, '')
		FROM identity.users
		WHERE email = ANY($1::text[]) AND role = 'client'`, preserveClientEmails)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snap := &clientPreserveSnapshot{
		PracticeIDToName:   map[string]string{},
		UserIDToEmail:      map[string]string{},
		UserPracticeIDs:    map[string]string{},
		UserActiveProfiles: map[string]string{},
	}
	for rows.Next() {
		var id, practiceID, activeProfile string
		if err := rows.Scan(&id, &practiceID, &activeProfile); err != nil {
			return nil, err
		}
		snap.UserIDs = append(snap.UserIDs, id)
		if practiceID != "" {
			snap.UserPracticeIDs[id] = practiceID
		}
		if activeProfile != "" {
			snap.UserActiveProfiles[id] = activeProfile
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(snap.UserIDs) == 0 {
		return snap, nil
	}

	if err := loadJSONAgg(ctx, tx, &snap.Profiles, `
		SELECT COALESCE(json_agg(row_to_json(p)), '[]'::json)
		FROM identity.profiles p WHERE p.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.Pets, `
		SELECT COALESCE(json_agg(row_to_json(p)), '[]'::json)
		FROM pets.pets p WHERE p.owner_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}

	petIDs, err := jsonFieldIDs(snap.Pets, "id")
	if err != nil {
		return nil, err
	}
	practiceIDs := map[string]struct{}{}
	for _, pid := range snap.UserPracticeIDs {
		practiceIDs[pid] = struct{}{}
	}
	for _, pid := range jsonStringFields(snap.Pets, "practice_id") {
		if pid != "" {
			practiceIDs[pid] = struct{}{}
		}
	}

	if err := loadJSONAgg(ctx, tx, &snap.PracticeClients, `
		SELECT COALESCE(json_agg(row_to_json(pc)), '[]'::json)
		FROM practice.practice_clients pc WHERE pc.client_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	for _, pid := range jsonStringFields(snap.PracticeClients, "practice_id") {
		if pid != "" {
			practiceIDs[pid] = struct{}{}
		}
	}
	if err := loadJSONAgg(ctx, tx, &snap.LinkRequests, `
		SELECT COALESCE(json_agg(row_to_json(r)), '[]'::json)
		FROM practice.client_vet_link_requests r WHERE r.client_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	for _, pid := range jsonStringFields(snap.LinkRequests, "practice_id") {
		if pid != "" {
			practiceIDs[pid] = struct{}{}
		}
	}

	if err := loadJSONAgg(ctx, tx, &snap.Threads, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM messaging.threads t WHERE t.client_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	for _, pid := range jsonStringFields(snap.Threads, "practice_id") {
		if pid != "" {
			practiceIDs[pid] = struct{}{}
		}
	}
	threadIDs, err := jsonFieldIDs(snap.Threads, "id")
	if err != nil {
		return nil, err
	}
	if len(threadIDs) > 0 {
		if err := loadJSONAgg(ctx, tx, &snap.Messages, `
			SELECT COALESCE(json_agg(row_to_json(m)), '[]'::json)
			FROM messaging.messages m WHERE m.thread_id::text = ANY($1)`, threadIDs); err != nil {
			return nil, err
		}
	} else {
		snap.Messages = json.RawMessage("[]")
	}

	if len(petIDs) > 0 {
		loaders := []struct {
			dst *json.RawMessage
			q   string
		}{
			{&snap.WeightReadings, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM pets.weight_readings t WHERE t.pet_id::text = ANY($1)`},
			{&snap.DossierEvents, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM pets.dossier_events t WHERE t.pet_id::text = ANY($1)`},
			{&snap.Documents, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM pets.documents t WHERE t.pet_id::text = ANY($1)`},
			{&snap.HeartRateSessions, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM heartrate.sessions t WHERE t.pet_id::text = ANY($1)`},
			{&snap.CareReminders, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM care.reminders t WHERE t.pet_id::text = ANY($1)`},
			{&snap.CareContacts, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM care.professional_contacts t WHERE t.pet_id::text = ANY($1)`},
			{&snap.CareCompetitions, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM care.competitions t WHERE t.pet_id::text = ANY($1)`},
			{&snap.Visits, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM visits.visits t WHERE t.pet_id::text = ANY($1)`},
			{&snap.PetEntitlements, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM billing.pet_entitlements t WHERE t.pet_id::text = ANY($1)`},
			{&snap.DossierShares, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM pets.dossier_share_tokens t WHERE t.pet_id::text = ANY($1)`},
			{&snap.PetAccess, `SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json) FROM pets.pet_access t WHERE t.pet_id::text = ANY($1)`},
		}
		for _, l := range loaders {
			if err := loadJSONAgg(ctx, tx, l.dst, l.q, petIDs); err != nil {
				return nil, err
			}
		}
		for _, pid := range jsonStringFields(snap.CareReminders, "practice_id") {
			if pid != "" {
				practiceIDs[pid] = struct{}{}
			}
		}
		for _, pid := range jsonStringFields(snap.Visits, "practice_id") {
			if pid != "" {
				practiceIDs[pid] = struct{}{}
			}
		}
		for _, pid := range jsonStringFields(snap.HeartRateSessions, "practice_id") {
			if pid != "" {
				practiceIDs[pid] = struct{}{}
			}
		}
		for _, pid := range jsonStringFields(snap.WeightReadings, "practice_id") {
			if pid != "" {
				practiceIDs[pid] = struct{}{}
			}
		}
	} else {
		empty := json.RawMessage("[]")
		snap.WeightReadings = empty
		snap.DossierEvents = empty
		snap.Documents = empty
		snap.HeartRateSessions = empty
		snap.CareReminders = empty
		snap.CareContacts = empty
		snap.CareCompetitions = empty
		snap.Visits = empty
		snap.PetEntitlements = empty
		snap.DossierShares = empty
		snap.PetAccess = empty
	}

	visitIDs, err := jsonFieldIDs(snap.Visits, "id")
	if err != nil {
		return nil, err
	}
	if len(visitIDs) > 0 {
		if err := loadJSONAgg(ctx, tx, &snap.VisitReports, `
			SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
			FROM visits.visit_reports t WHERE t.visit_id::text = ANY($1)`, visitIDs); err != nil {
			return nil, err
		}
		if err := loadJSONAgg(ctx, tx, &snap.ConsultationShares, `
			SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
			FROM pets.consultation_share_tokens t WHERE t.visit_id::text = ANY($1)`, visitIDs); err != nil {
			return nil, err
		}
	} else {
		snap.VisitReports = json.RawMessage("[]")
		snap.ConsultationShares = json.RawMessage("[]")
	}

	if err := loadJSONAgg(ctx, tx, &snap.StripeCustomers, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM billing.stripe_customers t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.AddonEntitlements, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM billing.addon_entitlements t WHERE t.owner_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.DeviceTokens, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM notifications.device_tokens t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.ClientPreferences, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM notifications.client_preferences t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.DiscoveryProgress, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM discovery.progress t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.DiscoveryJourney, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM discovery.email_journey t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.DiscoverySends, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM discovery.email_sends t WHERE t.user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}
	if err := loadJSONAgg(ctx, tx, &snap.ClientAccess, `
		SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
		FROM practice.client_access t
		WHERE t.client_user_id::text = ANY($1) OR t.grantee_user_id::text = ANY($1)`, snap.UserIDs); err != nil {
		return nil, err
	}

	pidList := keys(practiceIDs)
	if len(pidList) > 0 {
		nameRows, err := tx.Query(ctx, `SELECT id::text, name FROM practice.practices WHERE id::text = ANY($1)`, pidList)
		if err != nil {
			return nil, err
		}
		for nameRows.Next() {
			var id, name string
			if err := nameRows.Scan(&id, &name); err != nil {
				nameRows.Close()
				return nil, err
			}
			snap.PracticeIDToName[id] = name
		}
		nameRows.Close()
		if err := nameRows.Err(); err != nil {
			return nil, err
		}
		if err := loadJSONAgg(ctx, tx, &snap.FallbackPractices, `
			SELECT COALESCE(json_agg(row_to_json(p)), '[]'::json)
			FROM practice.practices p WHERE p.id::text = ANY($1)`, pidList); err != nil {
			return nil, err
		}
	} else {
		snap.FallbackPractices = json.RawMessage("[]")
	}

	refUserIDs := map[string]struct{}{}
	for _, id := range snap.UserIDs {
		refUserIDs[id] = struct{}{}
	}
	collectUserRefs := func(raw json.RawMessage, fields ...string) {
		for _, f := range fields {
			for _, id := range jsonStringFields(raw, f) {
				if id != "" {
					refUserIDs[id] = struct{}{}
				}
			}
		}
	}
	collectUserRefs(snap.PracticeClients, "vet_user_id", "client_user_id")
	collectUserRefs(snap.LinkRequests, "vet_user_id", "client_user_id")
	collectUserRefs(snap.Threads, "vet_user_id", "client_user_id")
	collectUserRefs(snap.Messages, "sender_user_id")
	collectUserRefs(snap.DossierEvents, "author_user_id")
	collectUserRefs(snap.Documents, "uploaded_by_user_id")
	collectUserRefs(snap.WeightReadings, "owner_user_id", "author_user_id")
	collectUserRefs(snap.VisitReports, "author_user_id")
	collectUserRefs(snap.PetAccess, "grantee_user_id", "granted_by_user_id")
	collectUserRefs(snap.ClientAccess, "client_user_id", "grantee_user_id", "granted_by_user_id")
	collectUserRefs(snap.CareContacts, "owner_user_id")
	collectUserRefs(snap.CareCompetitions, "owner_user_id")

	uidList := keys(refUserIDs)
	if len(uidList) > 0 {
		urows, err := tx.Query(ctx, `SELECT id::text, email FROM identity.users WHERE id::text = ANY($1)`, uidList)
		if err != nil {
			return nil, err
		}
		for urows.Next() {
			var id, email string
			if err := urows.Scan(&id, &email); err != nil {
				urows.Close()
				return nil, err
			}
			snap.UserIDToEmail[id] = email
		}
		urows.Close()
		if err := urows.Err(); err != nil {
			return nil, err
		}
		// Vets linked via practice_clients / threads — keep a fallback copy if remap by email fails.
		if err := loadJSONAgg(ctx, tx, &snap.FallbackVets, `
			SELECT COALESCE(json_agg(row_to_json(u)), '[]'::json)
			FROM identity.users u
			WHERE u.id::text = ANY($1) AND u.role IN ('vet','vet_assistant','secretary','care_pro')`, uidList); err != nil {
			return nil, err
		}
	} else {
		snap.FallbackVets = json.RawMessage("[]")
	}

	return snap, nil
}

// restorePreserveClients re-inserts the snapshotted client graph into the same seed transaction
// (after demo practices/users exist) so a restore failure rolls back the whole reset.
func restorePreserveClients(ctx context.Context, tx pgx.Tx, snap *clientPreserveSnapshot) error {
	if snap == nil || len(snap.UserIDs) == 0 {
		return nil
	}

	practiceMap, err := buildPracticeRemap(ctx, tx, snap)
	if err != nil {
		return err
	}
	userMap, err := buildUserRemap(ctx, tx, snap, practiceMap)
	if err != nil {
		return err
	}
	// Preserve clients keep their IDs.
	for _, id := range snap.UserIDs {
		userMap[id] = id
	}

	maps := map[string]map[string]string{
		"practice_id":           practiceMap,
		"vet_user_id":           userMap,
		"client_user_id":        userMap,
		"owner_user_id":         userMap,
		"author_user_id":        userMap,
		"sender_user_id":        userMap,
		"uploaded_by_user_id":   userMap,
		"grantee_user_id":       userMap,
		"granted_by_user_id":    userMap,
		"user_id":               userMap,
		"reference_vet_user_id": userMap,
	}

	type step struct {
		table    string
		raw      json.RawMessage
		required []string // unmapped non-null → drop row (no PHI re-attribution)
		optional []string // unmapped non-null → set NULL
	}

	run := func(s step) error {
		remapped, dropped, err := remapJSONUUIDFields(s.raw, maps, s.required, s.optional)
		if err != nil {
			return err
		}
		if dropped > 0 {
			log.Printf("seed preserve: dropped %d %s row(s) with unmapped FK", dropped, s.table)
		}
		if err := insertJSONRecords(ctx, tx, s.table, remapped); err != nil {
			return fmt.Errorf("%s: %w", s.table, err)
		}
		return nil
	}

	if err := run(step{"identity.profiles", snap.Profiles,
		[]string{"user_id"}, []string{"practice_id"}}); err != nil {
		return err
	}
	if err := run(step{"pets.pets", snap.Pets,
		[]string{"owner_user_id"}, []string{"practice_id"}}); err != nil {
		return err
	}
	if err := run(step{"practice.practice_clients", snap.PracticeClients,
		[]string{"practice_id", "client_user_id", "vet_user_id"}, nil}); err != nil {
		return err
	}
	if err := run(step{"practice.client_vet_link_requests", snap.LinkRequests,
		[]string{"practice_id", "client_user_id", "vet_user_id"}, nil}); err != nil {
		return err
	}
	if err := run(step{"billing.stripe_customers", snap.StripeCustomers,
		[]string{"user_id"}, nil}); err != nil {
		return err
	}
	if err := run(step{"billing.pet_entitlements", snap.PetEntitlements,
		[]string{"owner_user_id"}, nil}); err != nil {
		return err
	}
	if err := run(step{"billing.addon_entitlements", snap.AddonEntitlements,
		[]string{"owner_user_id"}, nil}); err != nil {
		return err
	}
	if err := run(step{"messaging.threads", snap.Threads,
		[]string{"client_user_id", "vet_user_id"}, []string{"practice_id"}}); err != nil {
		return err
	}
	if err := run(step{"messaging.messages", snap.Messages,
		[]string{"sender_user_id"}, nil}); err != nil {
		return err
	}

	for _, s := range []step{
		{"pets.weight_readings", snap.WeightReadings, []string{"owner_user_id", "author_user_id"}, []string{"practice_id"}},
		{"pets.dossier_events", snap.DossierEvents, []string{"author_user_id"}, nil},
		{"pets.documents", snap.Documents, []string{"uploaded_by_user_id"}, nil},
		{"heartrate.sessions", snap.HeartRateSessions, []string{"owner_user_id"}, []string{"practice_id"}},
		{"care.reminders", snap.CareReminders, []string{"practice_id"}, nil},
		{"care.professional_contacts", snap.CareContacts, []string{"owner_user_id"}, nil},
		{"care.competitions", snap.CareCompetitions, []string{"owner_user_id"}, nil},
		{"visits.visits", snap.Visits, []string{"practice_id"}, nil},
		{"visits.visit_reports", snap.VisitReports, []string{"author_user_id"}, nil},
		{"pets.dossier_share_tokens", snap.DossierShares, []string{"owner_user_id"}, nil},
		{"pets.consultation_share_tokens", snap.ConsultationShares, []string{"owner_user_id"}, nil},
		{"notifications.device_tokens", snap.DeviceTokens, []string{"user_id"}, nil},
		{"notifications.client_preferences", snap.ClientPreferences, []string{"user_id"}, nil},
		{"discovery.progress", snap.DiscoveryProgress, []string{"user_id"}, nil},
		{"discovery.email_journey", snap.DiscoveryJourney, []string{"user_id"}, nil},
		{"discovery.email_sends", snap.DiscoverySends, []string{"user_id"}, nil},
		{"pets.pet_access", snap.PetAccess, []string{"grantee_user_id", "granted_by_user_id"}, nil},
		{"practice.client_access", snap.ClientAccess, []string{"client_user_id", "grantee_user_id", "granted_by_user_id"}, nil},
	} {
		if err := run(s); err != nil {
			return err
		}
	}

	for userID, oldPracticeID := range snap.UserPracticeIDs {
		newPracticeID := practiceMap[oldPracticeID]
		if newPracticeID == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE identity.users SET practice_id = $2::uuid WHERE id = $1::uuid`,
			userID, newPracticeID); err != nil {
			return err
		}
	}
	for userID, oldProfileID := range snap.UserActiveProfiles {
		if _, err := tx.Exec(ctx, `
			UPDATE identity.users SET active_profile_id = $2::uuid
			WHERE id = $1::uuid AND EXISTS (SELECT 1 FROM identity.profiles WHERE id = $2::uuid)`,
			userID, oldProfileID); err != nil {
			return err
		}
	}

	return nil
}

func buildPracticeRemap(ctx context.Context, tx pgx.Tx, snap *clientPreserveSnapshot) (map[string]string, error) {
	out := map[string]string{}
	nameToNew := map[string]string{}
	rows, err := tx.Query(ctx, `SELECT id::text, name FROM practice.practices`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, err
		}
		// First match wins (seed order); demos have unique names.
		if _, ok := nameToNew[name]; !ok {
			nameToNew[name] = id
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	missing := map[string]struct{}{}
	for oldID, name := range snap.PracticeIDToName {
		if newID, ok := nameToNew[name]; ok {
			out[oldID] = newID
		} else {
			missing[oldID] = struct{}{}
		}
	}
	if len(missing) == 0 {
		return out, nil
	}

	// Fallback: re-insert practices that are not in the demo seed (custom staging cabinets).
	var practices []map[string]any
	if err := json.Unmarshal(snap.FallbackPractices, &practices); err != nil {
		return nil, err
	}
	for _, p := range practices {
		oldID, _ := p["id"].(string)
		if _, need := missing[oldID]; !need {
			continue
		}
		// Clear reference_vet until users exist.
		p["reference_vet_user_id"] = nil
		raw, err := json.Marshal([]map[string]any{p})
		if err != nil {
			return nil, err
		}
		if err := insertJSONRecords(ctx, tx, "practice.practices", raw); err != nil {
			return nil, fmt.Errorf("fallback practice %s: %w", oldID, err)
		}
		out[oldID] = oldID
		if name, _ := p["name"].(string); name != "" {
			nameToNew[name] = oldID
		}
	}
	return out, nil
}

func buildUserRemap(ctx context.Context, tx pgx.Tx, snap *clientPreserveSnapshot, practiceMap map[string]string) (map[string]string, error) {
	out := map[string]string{}
	emailToID := map[string]string{}
	rows, err := tx.Query(ctx, `SELECT id::text, email FROM identity.users`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, email string
		if err := rows.Scan(&id, &email); err != nil {
			rows.Close()
			return nil, err
		}
		emailToID[email] = id
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	missing := map[string]struct{}{}
	for oldID, email := range snap.UserIDToEmail {
		if newID, ok := emailToID[email]; ok {
			out[oldID] = newID
		} else {
			missing[oldID] = struct{}{}
		}
	}

	var vets []map[string]any
	if err := json.Unmarshal(snap.FallbackVets, &vets); err != nil {
		return nil, err
	}
	for _, u := range vets {
		oldID, _ := u["id"].(string)
		if _, need := missing[oldID]; !need {
			continue
		}
		if email, _ := u["email"].(string); email != "" {
			if newID, ok := emailToID[email]; ok {
				out[oldID] = newID
				delete(missing, oldID)
				continue
			}
		}
		if pid, ok := u["practice_id"].(string); ok && pid != "" {
			if np, ok := practiceMap[pid]; ok {
				u["practice_id"] = np
			}
		}
		u["active_profile_id"] = nil
		u["assigned_commercial_id"] = nil
		u["manager_user_id"] = nil
		u["sponsor_user_id"] = nil
		raw, err := json.Marshal([]map[string]any{u})
		if err != nil {
			return nil, err
		}
		if err := insertJSONRecords(ctx, tx, "identity.users", raw); err != nil {
			// Email conflict with a protected role — map to that user if possible.
			if email, _ := u["email"].(string); email != "" {
				if newID, ok := emailToID[email]; ok {
					out[oldID] = newID
					delete(missing, oldID)
					continue
				}
			}
			return nil, fmt.Errorf("fallback vet %s: %w", oldID, err)
		}
		out[oldID] = oldID
		if email, _ := u["email"].(string); email != "" {
			emailToID[email] = oldID
		}
		delete(missing, oldID)
	}

	// Remaining missing IDs stay unmapped: callers drop rows with required FKs
	// or null optional FKs — never re-attribute authorship to the preserved client.
	return out, nil
}

func loadJSONAgg(ctx context.Context, tx pgx.Tx, dst *json.RawMessage, query string, arg any) error {
	var raw []byte
	if err := tx.QueryRow(ctx, query, arg).Scan(&raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		*dst = json.RawMessage("[]")
		return nil
	}
	*dst = json.RawMessage(raw)
	return nil
}

func insertJSONRecords(ctx context.Context, tx pgx.Tx, table string, raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		return nil
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	// Use json_populate_recordset so new columns are ignored gracefully when absent from snapshot.
	_, err := tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s
		SELECT * FROM json_populate_recordset(NULL::%s, $1::json)
		ON CONFLICT DO NOTHING`, table, table), string(raw))
	return err
}

// remapJSONUUIDFields remaps UUID FK fields.
// required: non-null value not in map → drop the whole row (no silent re-attribution).
// optional: non-null value not in map → set NULL.
func remapJSONUUIDFields(raw json.RawMessage, maps map[string]map[string]string, required, optional []string) (json.RawMessage, int, error) {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		return json.RawMessage("[]"), 0, nil
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, err
	}
	reqSet := map[string]struct{}{}
	for _, f := range required {
		reqSet[f] = struct{}{}
	}
	optSet := map[string]struct{}{}
	for _, f := range optional {
		optSet[f] = struct{}{}
	}

	out := make([]map[string]any, 0, len(rows))
	dropped := 0
	for _, row := range rows {
		drop := false
		for field := range reqSet {
			m := maps[field]
			v, ok := row[field]
			if !ok || v == nil {
				continue
			}
			oldID, ok := v.(string)
			if !ok || oldID == "" {
				continue
			}
			newID, ok := m[oldID]
			if !ok {
				drop = true
				break
			}
			row[field] = newID
		}
		if drop {
			dropped++
			continue
		}
		for field := range optSet {
			m := maps[field]
			v, ok := row[field]
			if !ok || v == nil {
				continue
			}
			oldID, ok := v.(string)
			if !ok || oldID == "" {
				continue
			}
			if newID, ok := m[oldID]; ok {
				row[field] = newID
			} else {
				row[field] = nil
			}
		}
		out = append(out, row)
	}
	b, err := json.Marshal(out)
	return b, dropped, err
}

func jsonFieldIDs(raw json.RawMessage, field string) ([]string, error) {
	ids := jsonStringFields(raw, field)
	out := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func jsonStringFields(raw json.RawMessage, field string) []string {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		return nil
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil
	}
	var out []string
	for _, row := range rows {
		v, ok := row[field]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			out = append(out, t)
		default:
			out = append(out, fmt.Sprint(t))
		}
	}
	return out
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
