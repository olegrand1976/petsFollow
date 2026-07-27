package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FiliationEffectiveSource mirrors ResolveVetCommercial priority.
const (
	FiliationSourceVetAssignment  = "vet_assignment"
	FiliationSourceClientReferral = "client_referral"
	FiliationSourceNone           = "none"
)

// FiliationRow is a flat Commercial → Vet → Client visualization row.
type FiliationRow struct {
	BranchID                string    `json:"branchId,omitempty"`
	BranchName              string    `json:"branchName,omitempty"`
	ManagerUserID           string    `json:"managerUserId,omitempty"`
	ManagerName             string    `json:"managerName,omitempty"`
	CommercialUserID        string    `json:"commercialUserId"`
	CommercialName          string    `json:"commercialName"`
	CommercialEmail         string    `json:"commercialEmail"`
	VetUserID               string    `json:"vetUserId,omitempty"`
	VetName                 string    `json:"vetName,omitempty"`
	VetEmail                string    `json:"vetEmail,omitempty"`
	PracticeID              string    `json:"practiceId,omitempty"`
	PracticeName            string    `json:"practiceName,omitempty"`
	ClientUserID            string    `json:"clientUserId,omitempty"`
	ClientName              string    `json:"clientName,omitempty"`
	ClientEmail             string    `json:"clientEmail,omitempty"`
	ReferralCommercialID    string    `json:"referralCommercialId,omitempty"`
	InviteCode              string    `json:"inviteCode,omitempty"`
	SponsorClientID         string    `json:"sponsorClientId,omitempty"`
	SponsorClientName       string    `json:"sponsorClientName,omitempty"`
	EffectiveCommercialID   string    `json:"effectiveCommercialId,omitempty"`
	EffectiveCommercialName string    `json:"effectiveCommercialName,omitempty"`
	EffectiveSource         string    `json:"effectiveSource"` // vet_assignment | client_referral | none
	LinkedAt                time.Time `json:"linkedAt,omitempty"`
}

// FiliationFilter scopes ListFiliation.
// CommercialIDs:
//   - nil → all commercials/managers (admin global)
//   - non-nil empty → no rows (explicit empty scope)
//   - non-empty → only those reps
const (
	DefaultFiliationLimit = 2000
	MaxFiliationLimit     = 5000
)

type FiliationFilter struct {
	CommercialIDs []string
	BranchID      string
	Query         string
	Limit         int // 0 → DefaultFiliationLimit
	Offset        int
}

// FiliationPage is a capped ListFiliation result (safety + pagination).
type FiliationPage struct {
	Items     []FiliationRow `json:"items"`
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
	Truncated bool           `json:"truncated"`
}

// ListFiliation returns flat filiation rows for visualization.
// Part A: assigned vets (+ practice_clients, LEFT so vet-only rows appear).
// Part B: commercial_referrals not already shown under an assigned vet of the same commercial.
//
// Effective* mirrors ResolveVetCommercial:
//   - part_a: latest practice_clients for the row practice (Accrue-equivalent when practice set)
//   - part_b: latest practice_clients globally (= Resolve("", "")), practice column = that link's
//     practice_id (not vet.practice_id, which can diverge if the vet moved); Accrue for another
//     practice may differ — use Resolve(practiceID) / part_a for practice-scoped payee.
// Priority: assigned_commercial_id on the chosen link, else commercial_referrals, else none.
func (s *Store) ListFiliation(ctx context.Context, f FiliationFilter) (FiliationPage, error) {
	empty := FiliationPage{Items: []FiliationRow{}, Limit: DefaultFiliationLimit}
	if f.CommercialIDs != nil && len(f.CommercialIDs) == 0 {
		return empty, nil
	}

	limit := f.Limit
	if limit <= 0 {
		limit = DefaultFiliationLimit
	}
	if limit > MaxFiliationLimit {
		limit = MaxFiliationLimit
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	empty.Limit = limit
	empty.Offset = offset

	args := make([]any, 0, 6)
	argN := 1

	scopeSQL := "c.role IN ('commercial','commercial_manager')"
	if f.CommercialIDs != nil {
		scopeSQL = fmt.Sprintf("c.id = ANY($%d::uuid[])", argN)
		args = append(args, f.CommercialIDs)
		argN++
	}

	branchSQL := "TRUE"
	if strings.TrimSpace(f.BranchID) != "" {
		branchSQL = fmt.Sprintf("c.branch_id = $%d::uuid", argN)
		args = append(args, strings.TrimSpace(f.BranchID))
		argN++
	}

	q := strings.TrimSpace(f.Query)
	querySQL := "TRUE"
	if q != "" {
		querySQL = fmt.Sprintf(`(
			c.full_name ILIKE $%d OR c.email ILIKE $%d OR
			COALESCE(v.full_name,'') ILIKE $%d OR COALESCE(v.email,'') ILIKE $%d OR
			COALESCE(cli.full_name,'') ILIKE $%d OR COALESCE(cli.email,'') ILIKE $%d OR
			COALESCE(pr.name,'') ILIKE $%d
		)`, argN, argN, argN, argN, argN, argN, argN)
		args = append(args, "%"+q+"%")
		argN++
	}

	limitArg := argN
	args = append(args, limit+1) // fetch one extra to detect truncation
	argN++
	offsetArg := argN
	args = append(args, offset)

	// Resolve-aligned effective for a client row (practice_id of the row when available).
	// part_a: client row always has a practice_clients link (via JOIN); align Resolve.
	resolveEffID := fmt.Sprintf(`
		CASE
			WHEN resolve_link.assigned_commercial_id IS NOT NULL THEN resolve_link.assigned_commercial_id::text
			WHEN cr.commercial_user_id IS NOT NULL THEN cr.commercial_user_id::text
			ELSE ''
		END`)
	resolveEffSrc := fmt.Sprintf(`
		CASE
			WHEN resolve_link.assigned_commercial_id IS NOT NULL THEN '%s'
			WHEN cr.commercial_user_id IS NOT NULL THEN '%s'
			ELSE '%s'
		END`, FiliationSourceVetAssignment, FiliationSourceClientReferral, FiliationSourceNone)
	// part_b orphan referral: no practice_clients → Resolve returns empty → Effectif none
	// (still show the referral row; Accrue will not pay until a cabinet link exists).
	partBEffID := `
		CASE
			WHEN resolve_link.vet_user_id IS NULL THEN ''
			WHEN resolve_link.assigned_commercial_id IS NOT NULL THEN resolve_link.assigned_commercial_id::text
			WHEN cr.commercial_user_id IS NOT NULL THEN cr.commercial_user_id::text
			ELSE ''
		END`
	partBEffSrc := fmt.Sprintf(`
		CASE
			WHEN resolve_link.vet_user_id IS NULL THEN '%s'
			WHEN resolve_link.assigned_commercial_id IS NOT NULL THEN '%s'
			WHEN cr.commercial_user_id IS NOT NULL THEN '%s'
			ELSE '%s'
		END`, FiliationSourceNone, FiliationSourceVetAssignment, FiliationSourceClientReferral, FiliationSourceNone)

	sql := fmt.Sprintf(`
WITH scoped_comms AS (
	SELECT c.id, c.full_name, c.email, c.branch_id, c.manager_user_id
	FROM identity.users c
	WHERE %s AND %s
),
part_a AS (
	SELECT
		COALESCE(b.id::text,'') AS branch_id,
		COALESCE(b.name,'') AS branch_name,
		COALESCE(mgr.id::text,'') AS manager_user_id,
		COALESCE(mgr.full_name,'') AS manager_name,
		c.id::text AS commercial_user_id,
		c.full_name AS commercial_name,
		c.email AS commercial_email,
		v.id::text AS vet_user_id,
		v.full_name AS vet_name,
		v.email AS vet_email,
		COALESCE(v.practice_id::text,'') AS practice_id,
		COALESCE(pr.name,'') AS practice_name,
		COALESCE(cli.id::text,'') AS client_user_id,
		COALESCE(cli.full_name,'') AS client_name,
		COALESCE(cli.email,'') AS client_email,
		COALESCE(cr.commercial_user_id::text,'') AS referral_commercial_id,
		COALESCE(cr.invite_code,'') AS invite_code,
		COALESCE(clr.sponsor_client_user_id::text,'') AS sponsor_client_id,
		COALESCE(sp.full_name,'') AS sponsor_client_name,
		CASE
			WHEN cli.id IS NULL THEN v.assigned_commercial_id::text
			ELSE %s
		END AS effective_commercial_id,
		CASE
			WHEN cli.id IS NULL THEN '%s'
			ELSE %s
		END AS effective_source,
		COALESCE(pc.created_at, v.created_at) AS linked_at
	FROM scoped_comms c
	JOIN identity.users v ON v.role = 'vet' AND v.assigned_commercial_id = c.id
	LEFT JOIN practice.practices pr ON pr.id = v.practice_id
	LEFT JOIN sales.branches b ON b.id = c.branch_id
	LEFT JOIN identity.users mgr ON mgr.id = c.manager_user_id
	LEFT JOIN practice.practice_clients pc ON pc.vet_user_id = v.id
	LEFT JOIN identity.users cli ON cli.id = pc.client_user_id
	LEFT JOIN practice.commercial_referrals cr ON cr.client_user_id = cli.id
	LEFT JOIN practice.client_referrals clr ON clr.referred_client_user_id = cli.id
	LEFT JOIN identity.users sp ON sp.id = clr.sponsor_client_user_id
	LEFT JOIN LATERAL (
		SELECT u.assigned_commercial_id
		FROM practice.practice_clients rpc
		JOIN identity.users u ON u.id = rpc.vet_user_id
		WHERE cli.id IS NOT NULL
		  AND rpc.client_user_id = cli.id
		  AND (v.practice_id IS NULL OR rpc.practice_id = v.practice_id)
		ORDER BY rpc.created_at DESC NULLS LAST, rpc.vet_user_id
		LIMIT 1
	) resolve_link ON TRUE
	WHERE %s
),
part_b AS (
	SELECT
		COALESCE(b.id::text,'') AS branch_id,
		COALESCE(b.name,'') AS branch_name,
		COALESCE(mgr.id::text,'') AS manager_user_id,
		COALESCE(mgr.full_name,'') AS manager_name,
		c.id::text AS commercial_user_id,
		c.full_name AS commercial_name,
		c.email AS commercial_email,
		COALESCE(v.id::text,'') AS vet_user_id,
		COALESCE(v.full_name,'') AS vet_name,
		COALESCE(v.email,'') AS vet_email,
		COALESCE(resolve_link.practice_id::text,'') AS practice_id,
		COALESCE(pr.name,'') AS practice_name,
		cli.id::text AS client_user_id,
		cli.full_name AS client_name,
		cli.email AS client_email,
		cr.commercial_user_id::text AS referral_commercial_id,
		COALESCE(cr.invite_code,'') AS invite_code,
		COALESCE(clr.sponsor_client_user_id::text,'') AS sponsor_client_id,
		COALESCE(sp.full_name,'') AS sponsor_client_name,
		%s AS effective_commercial_id,
		%s AS effective_source,
		cr.created_at AS linked_at
	FROM scoped_comms c
	JOIN practice.commercial_referrals cr ON cr.commercial_user_id = c.id
	JOIN identity.users cli ON cli.id = cr.client_user_id
	LEFT JOIN sales.branches b ON b.id = c.branch_id
	LEFT JOIN identity.users mgr ON mgr.id = c.manager_user_id
	LEFT JOIN practice.client_referrals clr ON clr.referred_client_user_id = cli.id
	LEFT JOIN identity.users sp ON sp.id = clr.sponsor_client_user_id
	-- Resolve("",""): latest global practice_clients link; display that link's practice_id.
	LEFT JOIN LATERAL (
		SELECT rpc.vet_user_id, rpc.practice_id, u.assigned_commercial_id
		FROM practice.practice_clients rpc
		JOIN identity.users u ON u.id = rpc.vet_user_id
		WHERE rpc.client_user_id = cli.id
		ORDER BY rpc.created_at DESC NULLS LAST, rpc.vet_user_id
		LIMIT 1
	) resolve_link ON TRUE
	LEFT JOIN identity.users v ON v.id = resolve_link.vet_user_id
	LEFT JOIN practice.practices pr ON pr.id = resolve_link.practice_id
	WHERE NOT EXISTS (
		SELECT 1
		FROM practice.practice_clients pc2
		JOIN identity.users v2 ON v2.id = pc2.vet_user_id
		WHERE pc2.client_user_id = cr.client_user_id
		  AND v2.assigned_commercial_id = cr.commercial_user_id
	)
	AND %s
)
SELECT
	a.branch_id, a.branch_name, a.manager_user_id, a.manager_name,
	a.commercial_user_id, a.commercial_name, a.commercial_email,
	a.vet_user_id, a.vet_name, a.vet_email, a.practice_id, a.practice_name,
	a.client_user_id, a.client_name, a.client_email,
	a.referral_commercial_id, a.invite_code, a.sponsor_client_id, a.sponsor_client_name,
	a.effective_commercial_id,
	COALESCE(ec.full_name,''),
	a.effective_source,
	a.linked_at
FROM (
	SELECT * FROM part_a
	UNION ALL
	SELECT * FROM part_b
) a
LEFT JOIN identity.users ec ON a.effective_commercial_id <> '' AND ec.id::text = a.effective_commercial_id
ORDER BY a.commercial_name, a.vet_name NULLS LAST, a.client_name NULLS LAST, a.linked_at DESC
LIMIT $%d OFFSET $%d
`, scopeSQL, branchSQL,
		resolveEffID, FiliationSourceVetAssignment, resolveEffSrc, querySQL,
		partBEffID, partBEffSrc, querySQL, limitArg, offsetArg)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return empty, err
	}
	defer rows.Close()

	out := make([]FiliationRow, 0)
	for rows.Next() {
		var r FiliationRow
		var linkedAt *time.Time
		if err := rows.Scan(
			&r.BranchID, &r.BranchName, &r.ManagerUserID, &r.ManagerName,
			&r.CommercialUserID, &r.CommercialName, &r.CommercialEmail,
			&r.VetUserID, &r.VetName, &r.VetEmail, &r.PracticeID, &r.PracticeName,
			&r.ClientUserID, &r.ClientName, &r.ClientEmail,
			&r.ReferralCommercialID, &r.InviteCode, &r.SponsorClientID, &r.SponsorClientName,
			&r.EffectiveCommercialID, &r.EffectiveCommercialName, &r.EffectiveSource,
			&linkedAt,
		); err != nil {
			return empty, err
		}
		if linkedAt != nil {
			r.LinkedAt = *linkedAt
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return empty, err
	}
	truncated := len(out) > limit
	if truncated {
		out = out[:limit]
	}
	return FiliationPage{
		Items:     out,
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
	}, nil
}

// ListTeamCommercialIDs returns commercial user IDs managed by managerUserID.
func (s *Store) ListTeamCommercialIDs(ctx context.Context, managerUserID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text FROM identity.users
		WHERE role='commercial' AND manager_user_id=$1
		ORDER BY full_name`, managerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
