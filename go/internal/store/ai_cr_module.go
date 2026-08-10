package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	AiCrStatusTrial    = "trial"
	AiCrStatusActive   = "active"
	AiCrStatusExpired  = "expired"
	AiCrStatusDisabled = "disabled"

	AiCrPlanMonthly39 = "monthly_39"
	AiCrPlanAnnual390 = "annual_390"

	AiCrTrialDays = 90
	AiCrROIDay    = 60

	AiCrDefaultBaselineMin = 10
	AiCrDefaultHourlyCents = 8000
	AiCrDefaultAIMinutes   = 4
)

type AiCrModule struct {
	PracticeID           string     `json:"practiceId"`
	PracticeName         string     `json:"practiceName,omitempty"`
	Status               string     `json:"status"`
	ActivatedAt          time.Time  `json:"activatedAt"`
	TrialEndsAt          time.Time  `json:"trialEndsAt"`
	ActivatedByUserID    string     `json:"activatedByUserId,omitempty"`
	ConvertedAt          *time.Time `json:"convertedAt,omitempty"`
	PricePlan            string     `json:"pricePlan"`
	BaselineMinutesPerCR int        `json:"baselineMinutesPerCr"`
	HourlyCostCents      int        `json:"hourlyCostCents"`
	// Derived
	DaysSinceActivation int  `json:"daysSinceActivation"`
	DaysRemainingTrial  int  `json:"daysRemainingTrial"`
	Allowed             bool `json:"allowed"`
	RoiUnlocked         bool `json:"roiUnlocked"`
}

type AiCrUsageKind string

const (
	AiCrUsageTranscribe      AiCrUsageKind = "transcribe"
	AiCrUsageImprove         AiCrUsageKind = "improve"
	AiCrUsageImproveAdvanced AiCrUsageKind = "improve_advanced"
	AiCrUsageFinalize        AiCrUsageKind = "finalize"
	AiCrUsageError           AiCrUsageKind = "error"
)

type AiCrROI struct {
	Unlocked            bool    `json:"unlocked"`
	DaysSinceActivation int     `json:"daysSinceActivation"`
	CrsIA               int     `json:"crsIa"`
	BaselineMinutes     int     `json:"baselineMinutes"`
	AiMinutesAssumed    int     `json:"aiMinutesAssumed"`
	MinutesSaved        int     `json:"minutesSaved"`
	HoursSaved          float64 `json:"hoursSaved"`
	EuroEquivCents      int     `json:"euroEquivCents"`
	ModuleCostCents     int     `json:"moduleCostCents"`
	NetEuroCents        int     `json:"netEuroCents"`
	Disclaimer          string  `json:"disclaimer"`
}

func (m *AiCrModule) refreshDerived(now time.Time) {
	if m.ActivatedAt.IsZero() {
		return
	}
	days := max(int(now.Sub(m.ActivatedAt).Hours()/24), 0)
	m.DaysSinceActivation = days
	m.RoiUnlocked = days >= AiCrROIDay
	remain := max(int(m.TrialEndsAt.Sub(now).Hours()/24), 0)
	m.DaysRemainingTrial = remain
	m.Allowed = m.isAllowedAt(now)
}

func (m *AiCrModule) isAllowedAt(now time.Time) bool {
	_ = now
	// Native IA in Pro: allowed unless the practice was explicitly disabled.
	return m.Status != AiCrStatusDisabled
}

func NormalizeAiCrPricePlan(p string) string {
	switch strings.TrimSpace(p) {
	case AiCrPlanAnnual390:
		return AiCrPlanAnnual390
	default:
		return AiCrPlanMonthly39
	}
}

func (s *Store) GetAiCrModule(ctx context.Context, practiceID string) (AiCrModule, error) {
	var m AiCrModule
	err := s.pool.QueryRow(ctx, `
		SELECT practice_id::text, status, activated_at, trial_ends_at,
			COALESCE(activated_by_user_id::text,''), converted_at, price_plan,
			baseline_minutes_per_cr, hourly_cost_cents
		FROM practice.ai_cr_modules WHERE practice_id = $1`, practiceID).Scan(
		&m.PracticeID, &m.Status, &m.ActivatedAt, &m.TrialEndsAt,
		&m.ActivatedByUserID, &m.ConvertedAt, &m.PricePlan,
		&m.BaselineMinutesPerCR, &m.HourlyCostCents,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AiCrModule{}, ErrNotFound
	}
	if err != nil {
		return AiCrModule{}, err
	}
	now := time.Now().UTC()
	// Persist soft-expiry so UI/lists don't stay on stale "trial".
	if m.Status == AiCrStatusTrial && now.After(m.TrialEndsAt) {
		if _, uerr := s.pool.Exec(ctx, `
			UPDATE practice.ai_cr_modules SET status = 'expired', updated_at = NOW()
			WHERE practice_id = $1 AND status = 'trial' AND trial_ends_at < NOW()`, practiceID); uerr == nil {
			m.Status = AiCrStatusExpired
		}
	}
	m.refreshDerived(now)
	return m, nil
}

// AiCrModuleAllowed reports whether IA transcribe/improve is permitted for the practice.
// CR IA is included in Pro: any practice is allowed unless explicitly disabled.
func (s *Store) AiCrModuleAllowed(ctx context.Context, practiceID string) (bool, error) {
	if strings.TrimSpace(practiceID) == "" {
		return false, nil
	}
	m, err := s.GetAiCrModule(ctx, practiceID)
	if errors.Is(err, ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return m.Allowed, nil
}

// ActivateAiCrModule enables ROI/adoption tracking for a practice.
// CR IA is included in Pro: status is always active (not a paid trial).
func (s *Store) ActivateAiCrModule(ctx context.Context, practiceID, byUserID string) (AiCrModule, error) {
	now := time.Now().UTC()
	existing, err := s.GetAiCrModule(ctx, practiceID)
	if err == nil {
		if existing.Status == AiCrStatusActive {
			return existing, nil
		}
		// Keep activated_at for in-progress ROI windows (trial/expired); only re-stamp when disabled/missing.
		if existing.Status == AiCrStatusTrial || existing.Status == AiCrStatusExpired {
			tag, uerr := s.pool.Exec(ctx, `
				UPDATE practice.ai_cr_modules
				SET status = 'active', converted_at = NULL, updated_at = NOW()
				WHERE practice_id = $1`, practiceID)
			if uerr != nil {
				return AiCrModule{}, uerr
			}
			if tag.RowsAffected() == 0 {
				return AiCrModule{}, ErrNotFound
			}
			return s.GetAiCrModule(ctx, practiceID)
		}
		// disabled → fresh active window below
	} else if !errors.Is(err, ErrNotFound) {
		return AiCrModule{}, err
	}
	roiWindowEnds := now.AddDate(0, 0, AiCrTrialDays)
	var activatedBy any
	if strings.TrimSpace(byUserID) != "" {
		activatedBy = byUserID
	}
	// price_plan kept for schema CHECK only — CR IA is included (no module billing).
	_, err = s.pool.Exec(ctx, `
		INSERT INTO practice.ai_cr_modules (
			practice_id, status, activated_at, trial_ends_at, activated_by_user_id,
			converted_at, price_plan, baseline_minutes_per_cr, hourly_cost_cents, updated_at
		) VALUES ($1, 'active', $2, $3, $4, NULL, 'monthly_39', $5, $6, NOW())
		ON CONFLICT (practice_id) DO UPDATE SET
			status = 'active',
			activated_at = EXCLUDED.activated_at,
			trial_ends_at = EXCLUDED.trial_ends_at,
			activated_by_user_id = EXCLUDED.activated_by_user_id,
			converted_at = NULL,
			updated_at = NOW()`,
		practiceID, now, roiWindowEnds, activatedBy, AiCrDefaultBaselineMin, AiCrDefaultHourlyCents)
	if err != nil {
		return AiCrModule{}, err
	}
	return s.GetAiCrModule(ctx, practiceID)
}

// ConvertAiCrModule marks the module as paid (post-trial or early conversion).
func (s *Store) ConvertAiCrModule(ctx context.Context, practiceID, pricePlan string) (AiCrModule, error) {
	plan := NormalizeAiCrPricePlan(pricePlan)
	now := time.Now().UTC()
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.ai_cr_modules
		SET status = 'active', converted_at = $2, price_plan = $3, updated_at = NOW()
		WHERE practice_id = $1 AND status IN ('trial', 'expired', 'active')`,
		practiceID, now, plan)
	if err != nil {
		return AiCrModule{}, err
	}
	if tag.RowsAffected() == 0 {
		return AiCrModule{}, ErrNotFound
	}
	return s.GetAiCrModule(ctx, practiceID)
}

func (s *Store) SetAiCrModuleStatus(ctx context.Context, practiceID, status string) (AiCrModule, error) {
	switch status {
	case AiCrStatusExpired, AiCrStatusDisabled, AiCrStatusActive, AiCrStatusTrial:
	default:
		return AiCrModule{}, ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.ai_cr_modules SET status = $2, updated_at = NOW()
		WHERE practice_id = $1`, practiceID, status)
	if err != nil {
		return AiCrModule{}, err
	}
	if tag.RowsAffected() == 0 {
		return AiCrModule{}, ErrNotFound
	}
	return s.GetAiCrModule(ctx, practiceID)
}

func (s *Store) PatchAiCrModuleBaseline(ctx context.Context, practiceID string, baselineMin, hourlyCents int) (AiCrModule, error) {
	if baselineMin <= 0 || baselineMin > 120 || hourlyCents < 0 {
		return AiCrModule{}, ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.ai_cr_modules
		SET baseline_minutes_per_cr = $2, hourly_cost_cents = $3, updated_at = NOW()
		WHERE practice_id = $1`, practiceID, baselineMin, hourlyCents)
	if err != nil {
		return AiCrModule{}, err
	}
	if tag.RowsAffected() == 0 {
		return AiCrModule{}, ErrNotFound
	}
	return s.GetAiCrModule(ctx, practiceID)
}

func (s *Store) ListAiCrModules(ctx context.Context) ([]AiCrModule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.practice_id::text, COALESCE(p.name,''), m.status, m.activated_at, m.trial_ends_at,
			COALESCE(m.activated_by_user_id::text,''), m.converted_at, m.price_plan,
			m.baseline_minutes_per_cr, m.hourly_cost_cents
		FROM practice.ai_cr_modules m
		JOIN practice.practices p ON p.id = m.practice_id
		ORDER BY m.activated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now().UTC()
	out := make([]AiCrModule, 0)
	for rows.Next() {
		var m AiCrModule
		if err := rows.Scan(
			&m.PracticeID, &m.PracticeName, &m.Status, &m.ActivatedAt, &m.TrialEndsAt,
			&m.ActivatedByUserID, &m.ConvertedAt, &m.PricePlan,
			&m.BaselineMinutesPerCR, &m.HourlyCostCents,
		); err != nil {
			return nil, err
		}
		m.refreshDerived(now)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ListAiCrModulesForCommercial(ctx context.Context, commercialUserID string) ([]AiCrModule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.practice_id::text, COALESCE(p.name,''), m.status, m.activated_at, m.trial_ends_at,
			COALESCE(m.activated_by_user_id::text,''), m.converted_at, m.price_plan,
			m.baseline_minutes_per_cr, m.hourly_cost_cents
		FROM practice.ai_cr_modules m
		JOIN practice.practices p ON p.id = m.practice_id
		WHERE EXISTS (
			SELECT 1 FROM identity.users u
			WHERE u.practice_id = m.practice_id AND u.role = 'vet' AND u.assigned_commercial_id = $1
		)
		ORDER BY m.activated_at DESC`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now().UTC()
	out := make([]AiCrModule, 0)
	for rows.Next() {
		var m AiCrModule
		if err := rows.Scan(
			&m.PracticeID, &m.PracticeName, &m.Status, &m.ActivatedAt, &m.TrialEndsAt,
			&m.ActivatedByUserID, &m.ConvertedAt, &m.PricePlan,
			&m.BaselineMinutesPerCR, &m.HourlyCostCents,
		); err != nil {
			return nil, err
		}
		m.refreshDerived(now)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ListAiCrModulesForTeam(ctx context.Context, managerUserID string) ([]AiCrModule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.practice_id::text, COALESCE(p.name,''), m.status, m.activated_at, m.trial_ends_at,
			COALESCE(m.activated_by_user_id::text,''), m.converted_at, m.price_plan,
			m.baseline_minutes_per_cr, m.hourly_cost_cents
		FROM practice.ai_cr_modules m
		JOIN practice.practices p ON p.id = m.practice_id
		WHERE EXISTS (
			SELECT 1 FROM identity.users u
			WHERE u.practice_id = m.practice_id AND u.role = 'vet'
			  AND (
				u.assigned_commercial_id = $1
				OR u.assigned_commercial_id IN (
					SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
				)
			)
		)
		ORDER BY m.activated_at DESC`, managerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now().UTC()
	out := make([]AiCrModule, 0)
	for rows.Next() {
		var m AiCrModule
		if err := rows.Scan(
			&m.PracticeID, &m.PracticeName, &m.Status, &m.ActivatedAt, &m.TrialEndsAt,
			&m.ActivatedByUserID, &m.ConvertedAt, &m.PricePlan,
			&m.BaselineMinutesPerCR, &m.HourlyCostCents,
		); err != nil {
			return nil, err
		}
		m.refreshDerived(now)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) CommercialOwnsPractice(ctx context.Context, commercialUserID, practiceID string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM identity.users
		WHERE role = 'vet' AND practice_id = $1 AND assigned_commercial_id = $2`,
		practiceID, commercialUserID).Scan(&n)
	return n > 0, err
}

// ResolvePracticeCommercialID returns the first assigned commercial for a practice's vets.
func (s *Store) ResolvePracticeCommercialID(ctx context.Context, practiceID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT assigned_commercial_id::text FROM identity.users
		WHERE practice_id = $1 AND role = 'vet' AND assigned_commercial_id IS NOT NULL
		ORDER BY created_at ASC LIMIT 1`, practiceID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) InsertAiCrUsage(ctx context.Context, practiceID, userID, visitID string, kind AiCrUsageKind) error {
	id := uuid.NewString()
	var visit any
	if strings.TrimSpace(visitID) != "" {
		visit = visitID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.ai_cr_usage_events (id, practice_id, user_id, visit_id, kind)
		VALUES ($1, $2, $3, $4, $5)`, id, practiceID, userID, visit, string(kind))
	return err
}

func (s *Store) CountAiCrUsers(ctx context.Context, practiceID string, since time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)::int FROM practice.ai_cr_usage_events
		WHERE practice_id = $1 AND created_at >= $2 AND kind IN ('transcribe','improve','finalize')`,
		practiceID, since).Scan(&n)
	return n, err
}

func (s *Store) CountAiCrEvents(ctx context.Context, practiceID string, kind AiCrUsageKind, since time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.ai_cr_usage_events
		WHERE practice_id = $1 AND kind = $2 AND created_at >= $3`,
		practiceID, string(kind), since).Scan(&n)
	return n, err
}

func (s *Store) CountAiCrEventsInRange(ctx context.Context, practiceID string, kind AiCrUsageKind, from, to time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.ai_cr_usage_events
		WHERE practice_id = $1 AND kind = $2 AND created_at >= $3 AND created_at < $4`,
		practiceID, string(kind), from, to).Scan(&n)
	return n, err
}

func (s *Store) CountAiCrIAFinalizations(ctx context.Context, practiceID string, since time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.visit_id)::int
		FROM practice.ai_cr_usage_events e
		WHERE e.practice_id = $1
		  AND e.kind = 'finalize'
		  AND e.visit_id IS NOT NULL
		  AND e.created_at >= $2
		  AND EXISTS (
			SELECT 1 FROM practice.ai_cr_usage_events ia
			WHERE ia.practice_id = e.practice_id
			  AND ia.visit_id = e.visit_id
			  AND ia.kind IN ('transcribe', 'improve')
		  )`, practiceID, since).Scan(&n)
	return n, err
}

func (s *Store) ComputeAiCrROI(ctx context.Context, practiceID string) (AiCrROI, error) {
	m, err := s.GetAiCrModule(ctx, practiceID)
	if err != nil {
		return AiCrROI{}, err
	}
	roi := AiCrROI{
		Unlocked:            m.RoiUnlocked,
		DaysSinceActivation: m.DaysSinceActivation,
		BaselineMinutes:     m.BaselineMinutesPerCR,
		AiMinutesAssumed:    AiCrDefaultAIMinutes,
		Disclaimer:          "estimation",
	}
	if !m.RoiUnlocked {
		return roi, nil
	}
	crs, err := s.CountAiCrIAFinalizations(ctx, practiceID, m.ActivatedAt)
	if err != nil {
		return AiCrROI{}, err
	}
	roi.CrsIA = crs
	per := max(m.BaselineMinutesPerCR-AiCrDefaultAIMinutes, 0)
	roi.MinutesSaved = crs * per
	roi.HoursSaved = float64(roi.MinutesSaved) / 60.0
	roi.EuroEquivCents = int(roi.HoursSaved * float64(m.HourlyCostCents))
	// CR IA is included in Pro SaaS — no separate module cost in ROI net.
	roi.ModuleCostCents = 0
	roi.NetEuroCents = roi.EuroEquivCents
	return roi, nil
}

func (s *Store) InsertAiCrFeedback(ctx context.Context, practiceID, userID string, nps int, comment, source string, tags []string) error {
	if nps < 0 || nps > 10 {
		return ErrValidation
	}
	switch source {
	case "j14", "j45", "j75", "in_app", "visit_report":
	default:
		source = "in_app"
	}
	if tags == nil {
		tags = []string{}
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 500 {
		comment = comment[:500]
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.ai_cr_feedback (id, practice_id, user_id, nps, comment, friction_tags, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.NewString(), practiceID, userID, nps, comment, tags, source)
	return err
}

func (s *Store) LatestAiCrFeedbackNPS(ctx context.Context, practiceID string) (int, bool, error) {
	var nps int
	err := s.pool.QueryRow(ctx, `
		SELECT nps FROM practice.ai_cr_feedback
		WHERE practice_id = $1 ORDER BY created_at DESC LIMIT 1`, practiceID).Scan(&nps)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return nps, true, err
}

func (s *Store) RecentFrictionAlertExists(ctx context.Context, practiceID, signal string, within time.Duration) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.ai_cr_friction_alerts
		WHERE practice_id = $1 AND signal = $2 AND created_at >= $3`,
		practiceID, signal, time.Now().UTC().Add(-within)).Scan(&n)
	return n > 0, err
}

func (s *Store) InsertFrictionAlert(ctx context.Context, practiceID, commercialUserID, signal, detail string) error {
	var commercial any
	if strings.TrimSpace(commercialUserID) != "" {
		commercial = commercialUserID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.ai_cr_friction_alerts (id, practice_id, commercial_user_id, signal, detail)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.NewString(), practiceID, commercial, signal, detail)
	return err
}

type AiCrFrictionCandidate struct {
	PracticeID       string
	PracticeName     string
	Status           string
	ActivatedAt      time.Time
	CommercialUserID string
	CommercialEmail  string
	CommercialName   string
}

func (s *Store) ListAiCrFrictionCandidates(ctx context.Context) ([]AiCrFrictionCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.practice_id::text, COALESCE(p.name,''), m.status, m.activated_at,
			COALESCE(v.assigned_commercial_id::text,''),
			COALESCE(c.email,''), COALESCE(c.full_name,'')
		FROM practice.ai_cr_modules m
		JOIN practice.practices p ON p.id = m.practice_id
		LEFT JOIN LATERAL (
			SELECT assigned_commercial_id FROM identity.users
			WHERE practice_id = m.practice_id AND role = 'vet' AND assigned_commercial_id IS NOT NULL
			ORDER BY created_at ASC LIMIT 1
		) v ON true
		LEFT JOIN identity.users c ON c.id = v.assigned_commercial_id
		WHERE m.status IN ('trial', 'active')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AiCrFrictionCandidate, 0)
	for rows.Next() {
		var c AiCrFrictionCandidate
		if err := rows.Scan(
			&c.PracticeID, &c.PracticeName, &c.Status, &c.ActivatedAt,
			&c.CommercialUserID, &c.CommercialEmail, &c.CommercialName,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListRecentFrictionAlertsForCommercial(ctx context.Context, commercialUserID string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.practice_id::text, COALESCE(p.name,''), a.signal, a.detail, a.created_at
		FROM practice.ai_cr_friction_alerts a
		JOIN practice.practices p ON p.id = a.practice_id
		WHERE a.commercial_user_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2`, commercialUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, practiceID, practiceName, signal, detail string
		var created time.Time
		if err := rows.Scan(&id, &practiceID, &practiceName, &signal, &detail, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":           id,
			"practiceId":   practiceID,
			"practiceName": practiceName,
			"signal":       signal,
			"detail":       detail,
			"createdAt":    created,
		})
	}
	return out, rows.Err()
}

func (s *Store) ListRecentFrictionAlertsForTeam(ctx context.Context, managerUserID string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		WITH team AS (
			SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			UNION
			SELECT $1::uuid
		)
		SELECT a.id::text, a.practice_id::text, COALESCE(p.name,''), a.signal, a.detail, a.created_at,
			COALESCE(owner.id::text, c.id::text, ''),
			COALESCE(owner.full_name, c.full_name, ''),
			COALESCE(owner.email, c.email, ''),
			COALESCE(p.phone,''), COALESCE(p.contact_email,'')
		FROM practice.ai_cr_friction_alerts a
		JOIN practice.practices p ON p.id = a.practice_id
		LEFT JOIN identity.users c ON c.id = a.commercial_user_id
			AND c.role = 'commercial'
			AND c.id IN (SELECT id FROM team)
		LEFT JOIN LATERAL (
			SELECT u.id, u.full_name, u.email
			FROM identity.users v
			JOIN identity.users u ON u.id = v.assigned_commercial_id
			WHERE v.practice_id = a.practice_id AND v.role='vet'
			  AND v.assigned_commercial_id IN (SELECT id FROM team)
			  AND u.role = 'commercial'
			ORDER BY v.created_at
			LIMIT 1
		) owner ON true
		WHERE a.commercial_user_id IN (SELECT id FROM team)
		   OR EXISTS (
				SELECT 1 FROM identity.users v
				WHERE v.practice_id = a.practice_id AND v.role='vet'
				  AND v.assigned_commercial_id IN (SELECT id FROM team)
			)
		ORDER BY a.created_at DESC
		LIMIT $2`, managerUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, practiceID, practiceName, signal, detail, commercialID, commercialName, commercialEmail string
		var practicePhone, practiceContactEmail string
		var created time.Time
		if err := rows.Scan(&id, &practiceID, &practiceName, &signal, &detail, &created,
			&commercialID, &commercialName, &commercialEmail,
			&practicePhone, &practiceContactEmail); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":                   id,
			"practiceId":           practiceID,
			"practiceName":         practiceName,
			"signal":               signal,
			"detail":               detail,
			"createdAt":            created,
			"commercialUserId":     commercialID,
			"commercialName":       commercialName,
			"commercialEmail":      commercialEmail,
			"practicePhone":        practicePhone,
			"practiceContactEmail": practiceContactEmail,
		})
	}
	return out, rows.Err()
}

// ExpireStaleAiCrTrials marks past trial_ends_at rows as expired.
func (s *Store) ExpireStaleAiCrTrials(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.ai_cr_modules
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'trial' AND trial_ends_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
