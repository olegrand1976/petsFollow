package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	BonusCodeCommercialMix = "commercial_mix"

	BonusStatusEarned = "earned"
	BonusStatusPaid   = "paid"

	commercialMixTargetPct   = 55
	commercialMixAmountCents = 5000

	defaultBonusTrendMonths = 6
	maxBonusTrendMonths     = 12
)

var (
	ErrBonusNotEarned   = errors.New("bonus_not_earned")
	ErrBonusAlreadyPaid = errors.New("bonus_already_paid")
)

// CommercialBonusAward is a persisted SPIFF row (earned or paid).
type CommercialBonusAward struct {
	ID                 string     `json:"id"`
	CommercialUserID   string     `json:"commercialUserId"`
	CommercialFullName string     `json:"commercialFullName,omitempty"`
	CommercialEmail    string     `json:"commercialEmail,omitempty"`
	BonusCode          string     `json:"bonusCode"`
	AmountCents        int        `json:"amountCents"`
	Status             string     `json:"status"`
	PeriodYM           string     `json:"periodYm,omitempty"`
	VetUserID          string     `json:"vetUserId,omitempty"`
	VetEmail           string     `json:"vetEmail,omitempty"`
	VetFullName        string     `json:"vetFullName,omitempty"`
	Progress           int        `json:"progress"`
	Target             int        `json:"target"`
	EarnedAt           time.Time  `json:"earnedAt"`
	PaidAt             *time.Time `json:"paidAt,omitempty"`
}

// CommercialBonusTrackRow is an admin suivi row (live progress and/or award).
type CommercialBonusTrackRow struct {
	AwardID            string `json:"awardId,omitempty"`
	CommercialUserID   string `json:"commercialUserId"`
	CommercialFullName string `json:"commercialFullName"`
	CommercialEmail    string `json:"commercialEmail"`
	BonusCode          string `json:"bonusCode"`
	AmountCents        int    `json:"amountCents"`
	Status             string `json:"status"` // available | in_progress | earned | paid
	Progress           int    `json:"progress"`
	Target             int    `json:"target"`
	PeriodYM           string `json:"periodYm,omitempty"`
	VetUserID          string `json:"vetUserId,omitempty"`
	VetEmail           string `json:"vetEmail,omitempty"`
	VetFullName        string `json:"vetFullName,omitempty"`
	TriennialCount     int    `json:"triennialCount,omitempty"`
	SubCount           int    `json:"subCount,omitempty"`
}

// CommercialBonusTrendPoint is one month of SPIFF aggregates for the trend chart.
type CommercialBonusTrendPoint struct {
	PeriodYM          string `json:"periodYm"`
	EarnedCount       int    `json:"earnedCount"`
	PaidCount         int    `json:"paidCount"`
	InProgressCount   int    `json:"inProgressCount"`
	MetCount          int    `json:"metCount"`
	DueCents          int    `json:"dueCents"`
	PaidCents         int    `json:"paidCents"`
	AvgMixPct         int    `json:"avgMixPct"`
	ActiveCommercials int    `json:"activeCommercials"`
}

// CommercialBonusCompareRow is one commercial for the selected-month comparison chart.
type CommercialBonusCompareRow struct {
	CommercialUserID   string `json:"commercialUserId"`
	CommercialFullName string `json:"commercialFullName"`
	CommercialEmail    string `json:"commercialEmail"`
	MixPct             int    `json:"mixPct"`
	TriennialCount     int    `json:"triennialCount"`
	SubCount           int    `json:"subCount"`
	Status             string `json:"status"`
	AmountCents        int    `json:"amountCents"`
	AwardID            string `json:"awardId,omitempty"`
}

// CommercialBonusKPI summarises the selected month.
type CommercialBonusKPI struct {
	EarnedCount     int `json:"earnedCount"`
	PaidCount       int `json:"paidCount"`
	InProgressCount int `json:"inProgressCount"`
	DueCents        int `json:"dueCents"`
	PaidCents       int `json:"paidCents"`
	AvgMixPct       int `json:"avgMixPct"`
	MetCount        int `json:"metCount"`
}

// CommercialBonusAdminOverview is the admin SPIFF dashboard payload.
type CommercialBonusAdminOverview struct {
	PeriodYM    string                      `json:"periodYm"`
	Periods     []string                    `json:"periods"`
	TrendMonths int                         `json:"trendMonths"`
	TargetPct   int                         `json:"targetPct"`
	AmountCents int                         `json:"amountCents"`
	KPI         CommercialBonusKPI          `json:"kpi"`
	Trend       []CommercialBonusTrendPoint `json:"trend"`
	Comparison  []CommercialBonusCompareRow `json:"comparison"`
	Items       []CommercialBonusTrackRow   `json:"items"`
	Bonuses     []BonusRule                 `json:"bonuses"`
	PlanRates   []PlanRateInfo              `json:"planRates"`
}

func mixDedupeKey(commercialUserID, periodYM string) string {
	return fmt.Sprintf("mix:%s:%s", commercialUserID, periodYM)
}

// SyncCommercialBonusAwards persists earned Mix awards when the monthly threshold is met.
// Existing paid/earned awards are never deleted when the window slides.
func (s *Store) SyncCommercialBonusAwards(ctx context.Context, commercialUserID string) error {
	month := PeriodYM(time.Now())
	triennialN, subN, err := s.mixCountsForPeriod(ctx, commercialUserID, month)
	if err != nil {
		return err
	}
	pct := mixPct(triennialN, subN)
	if subN > 0 && pct >= commercialMixTargetPct {
		if err := s.upsertBonusAward(ctx, CommercialBonusAward{
			CommercialUserID: commercialUserID,
			BonusCode:        BonusCodeCommercialMix,
			AmountCents:      commercialMixAmountCents,
			Status:           BonusStatusEarned,
			PeriodYM:         month,
			Progress:         pct,
			Target:           commercialMixTargetPct,
		}, mixDedupeKey(commercialUserID, month)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) upsertBonusAward(ctx context.Context, a CommercialBonusAward, dedupeKey string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing.commercial_bonus_awards (
			id, commercial_user_id, bonus_code, amount_cents, status,
			period_ym, vet_user_id, progress, target, dedupe_key, earned_at
		) VALUES (
			$1,$2,$3,$4,'earned',
			NULLIF($5,''), NULLIF($6,'')::uuid, $7,$8,$9,NOW()
		)
		ON CONFLICT (dedupe_key) DO UPDATE SET
			progress = EXCLUDED.progress,
			target = EXCLUDED.target
		WHERE billing.commercial_bonus_awards.status = 'earned'`,
		uuid.NewString(), a.CommercialUserID, a.BonusCode, a.AmountCents,
		a.PeriodYM, a.VetUserID, a.Progress, a.Target, dedupeKey)
	return err
}

// MarkCommercialBonusPaid marks an earned award as paid by an admin.
func (s *Store) MarkCommercialBonusPaid(ctx context.Context, awardID, adminUserID string) (CommercialBonusAward, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT status FROM billing.commercial_bonus_awards WHERE id=$1`, awardID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialBonusAward{}, ErrNotFound
	}
	if err != nil {
		return CommercialBonusAward{}, err
	}
	if status == BonusStatusPaid {
		return CommercialBonusAward{}, ErrBonusAlreadyPaid
	}
	if status != BonusStatusEarned {
		return CommercialBonusAward{}, ErrBonusNotEarned
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE billing.commercial_bonus_awards
		SET status='paid', paid_at=NOW(), paid_by_admin_id=$2
		WHERE id=$1 AND status='earned'`, awardID, adminUserID)
	if err != nil {
		return CommercialBonusAward{}, err
	}
	return s.getCommercialBonusAward(ctx, awardID)
}

func (s *Store) getCommercialBonusAward(ctx context.Context, awardID string) (CommercialBonusAward, error) {
	var a CommercialBonusAward
	var period, vetID, vetEmail, vetName *string
	var paidAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.commercial_user_id::text, cu.full_name, cu.email,
			a.bonus_code, a.amount_cents, a.status, a.period_ym, a.vet_user_id::text,
			vu.email, vu.full_name, a.progress, a.target, a.earned_at, a.paid_at
		FROM billing.commercial_bonus_awards a
		JOIN identity.users cu ON cu.id = a.commercial_user_id
		LEFT JOIN identity.users vu ON vu.id = a.vet_user_id
		WHERE a.id=$1`, awardID).Scan(
		&a.ID, &a.CommercialUserID, &a.CommercialFullName, &a.CommercialEmail,
		&a.BonusCode, &a.AmountCents, &a.Status, &period, &vetID,
		&vetEmail, &vetName, &a.Progress, &a.Target, &a.EarnedAt, &paidAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialBonusAward{}, ErrNotFound
	}
	if err != nil {
		return CommercialBonusAward{}, err
	}
	if period != nil {
		a.PeriodYM = *period
	}
	if vetID != nil {
		a.VetUserID = *vetID
	}
	if vetEmail != nil {
		a.VetEmail = *vetEmail
	}
	if vetName != nil {
		a.VetFullName = *vetName
	}
	a.PaidAt = paidAt
	return a, nil
}

// ListCommercialBonusTrackRows returns admin suivi rows (awards + live in-progress).
// Kept for callers that only need the flat list; prefer AdminCommercialBonusesOverview.
func (s *Store) ListCommercialBonusTrackRows(ctx context.Context, statusFilter, commercialFilter string) ([]CommercialBonusTrackRow, error) {
	ov, err := s.AdminCommercialBonusesOverview(ctx, "", statusFilter, commercialFilter, defaultBonusTrendMonths)
	if err != nil {
		return nil, err
	}
	return ov.Items, nil
}

type commercialBonusCache struct {
	row    CommercialRow
	awards []CommercialBonusAward
}

// AdminCommercialBonusesOverview builds the admin SPIFF dashboard (month nav + trend + comparison).
func (s *Store) AdminCommercialBonusesOverview(
	ctx context.Context,
	periodYM, statusFilter, commercialFilter string,
	trendMonths int,
) (CommercialBonusAdminOverview, error) {
	current := PeriodYM(time.Now())
	if periodYM == "" || !ValidPeriodYM(periodYM) {
		periodYM = current
	}
	if trendMonths <= 0 {
		trendMonths = defaultBonusTrendMonths
	}
	if trendMonths > maxBonusTrendMonths {
		trendMonths = maxBonusTrendMonths
	}

	commercials, err := s.ListAllCommercials(ctx)
	if err != nil {
		return CommercialBonusAdminOverview{}, err
	}

	if periodYM == current {
		for _, c := range commercials {
			if commercialFilter != "" && c.UserID != commercialFilter {
				continue
			}
			if err := s.SyncCommercialBonusAwards(ctx, c.UserID); err != nil {
				return CommercialBonusAdminOverview{}, err
			}
		}
	}

	cached := make([]commercialBonusCache, 0, len(commercials))
	for _, c := range commercials {
		if commercialFilter != "" && c.UserID != commercialFilter {
			continue
		}
		awards, err := s.listBonusAwardsForCommercial(ctx, c.UserID)
		if err != nil {
			return CommercialBonusAdminOverview{}, err
		}
		cached = append(cached, commercialBonusCache{row: c, awards: awards})
	}

	items := make([]CommercialBonusTrackRow, 0)
	comparison := make([]CommercialBonusCompareRow, 0)
	for _, cc := range cached {
		row, ok, err := s.commercialBonusRowForPeriod(ctx, cc.row, periodYM, cc.awards)
		if err != nil {
			return CommercialBonusAdminOverview{}, err
		}
		if !ok {
			continue
		}
		comparison = append(comparison, CommercialBonusCompareRow{
			CommercialUserID:   row.CommercialUserID,
			CommercialFullName: row.CommercialFullName,
			CommercialEmail:    row.CommercialEmail,
			MixPct:             row.Progress,
			TriennialCount:     row.TriennialCount,
			SubCount:           row.SubCount,
			Status:             row.Status,
			AmountCents:        row.AmountCents,
			AwardID:            row.AwardID,
		})
		if statusFilter != "" && row.Status != statusFilter {
			continue
		}
		items = append(items, row)
	}
	sort.Slice(comparison, func(i, j int) bool {
		if comparison[i].MixPct != comparison[j].MixPct {
			return comparison[i].MixPct > comparison[j].MixPct
		}
		return comparison[i].CommercialFullName < comparison[j].CommercialFullName
	})
	sort.Slice(items, func(i, j int) bool {
		if items[i].Progress != items[j].Progress {
			return items[i].Progress > items[j].Progress
		}
		return items[i].CommercialFullName < items[j].CommercialFullName
	})

	trend, err := s.commercialBonusTrend(ctx, periodYM, trendMonths, cached)
	if err != nil {
		return CommercialBonusAdminOverview{}, err
	}

	periods, err := s.commercialBonusPeriodOptions(ctx, current)
	if err != nil {
		return CommercialBonusAdminOverview{}, err
	}

	return CommercialBonusAdminOverview{
		PeriodYM:    periodYM,
		Periods:     periods,
		TrendMonths: trendMonths,
		TargetPct:   commercialMixTargetPct,
		AmountCents: commercialMixAmountCents,
		KPI:         kpiFromComparison(comparison),
		Trend:       trend,
		Comparison:  comparison,
		Items:       items,
		Bonuses:     DefaultBonusRules(),
		PlanRates:   SubscriptionPlanRates(),
	}, nil
}

func kpiFromComparison(rows []CommercialBonusCompareRow) CommercialBonusKPI {
	var kpi CommercialBonusKPI
	mixSum := 0
	mixN := 0
	for _, r := range rows {
		switch r.Status {
		case BonusStatusPaid:
			kpi.PaidCount++
			kpi.PaidCents += r.AmountCents
			kpi.MetCount++
		case BonusStatusEarned:
			kpi.EarnedCount++
			kpi.DueCents += r.AmountCents
			kpi.MetCount++
		case "in_progress":
			kpi.InProgressCount++
		}
		if r.SubCount > 0 {
			mixSum += r.MixPct
			mixN++
		}
	}
	if mixN > 0 {
		kpi.AvgMixPct = mixSum / mixN
	}
	return kpi
}

func (s *Store) commercialBonusTrend(
	ctx context.Context,
	endPeriod string,
	months int,
	cached []commercialBonusCache,
) ([]CommercialBonusTrendPoint, error) {
	periods := make([]string, 0, months)
	p := endPeriod
	for range months {
		periods = append(periods, p)
		prev, err := PrevPeriodYM(p)
		if err != nil {
			return nil, err
		}
		p = prev
	}
	for i, j := 0, len(periods)-1; i < j; i, j = i+1, j-1 {
		periods[i], periods[j] = periods[j], periods[i]
	}

	out := make([]CommercialBonusTrendPoint, 0, len(periods))
	for _, period := range periods {
		pt := CommercialBonusTrendPoint{PeriodYM: period}
		mixSum := 0
		mixN := 0
		for _, cc := range cached {
			row, ok, err := s.commercialBonusRowForPeriod(ctx, cc.row, period, cc.awards)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			pt.ActiveCommercials++
			switch row.Status {
			case BonusStatusPaid:
				pt.PaidCount++
				pt.PaidCents += row.AmountCents
				pt.MetCount++
			case BonusStatusEarned:
				pt.EarnedCount++
				pt.DueCents += row.AmountCents
				pt.MetCount++
			case "in_progress":
				pt.InProgressCount++
			}
			if row.SubCount > 0 {
				mixSum += row.Progress
				mixN++
			}
		}
		if mixN > 0 {
			pt.AvgMixPct = mixSum / mixN
		}
		out = append(out, pt)
	}
	return out, nil
}

func (s *Store) commercialBonusPeriodOptions(ctx context.Context, current string) ([]string, error) {
	seen := map[string]struct{}{current: {}}
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT period_ym
		FROM billing.commercial_bonus_awards
		WHERE bonus_code=$1 AND period_ym IS NOT NULL AND period_ym <> ''
		ORDER BY period_ym DESC`, BonusCodeCommercialMix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		seen[p] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Always offer a rolling 12-month window ending at current.
	p := current
	for range 12 {
		seen[p] = struct{}{}
		prev, err := PrevPeriodYM(p)
		if err != nil {
			return nil, err
		}
		p = prev
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out, nil
}

func (s *Store) commercialBonusRowForPeriod(
	ctx context.Context,
	c CommercialRow,
	periodYM string,
	awards []CommercialBonusAward,
) (CommercialBonusTrackRow, bool, error) {
	var award *CommercialBonusAward
	for i := range awards {
		a := &awards[i]
		if a.BonusCode == BonusCodeCommercialMix && a.PeriodYM == periodYM {
			award = a
			break
		}
	}

	triennialN, subN, err := s.mixCountsForPeriod(ctx, c.UserID, periodYM)
	if err != nil {
		return CommercialBonusTrackRow{}, false, err
	}
	pct := mixPct(triennialN, subN)

	if award != nil {
		progress := award.Progress
		if subN > 0 {
			progress = pct
		}
		return CommercialBonusTrackRow{
			AwardID:            award.ID,
			CommercialUserID:   c.UserID,
			CommercialFullName: c.FullName,
			CommercialEmail:    c.Email,
			BonusCode:          award.BonusCode,
			AmountCents:        award.AmountCents,
			Status:             award.Status,
			Progress:           progress,
			Target:             award.Target,
			PeriodYM:           award.PeriodYM,
			VetUserID:          award.VetUserID,
			VetEmail:           award.VetEmail,
			VetFullName:        award.VetFullName,
			TriennialCount:     triennialN,
			SubCount:           subN,
		}, true, nil
	}

	if subN == 0 {
		return CommercialBonusTrackRow{}, false, nil
	}

	status := "in_progress"
	if pct >= commercialMixTargetPct {
		status = BonusStatusEarned
	}
	return CommercialBonusTrackRow{
		CommercialUserID:   c.UserID,
		CommercialFullName: c.FullName,
		CommercialEmail:    c.Email,
		BonusCode:          BonusCodeCommercialMix,
		AmountCents:        commercialMixAmountCents,
		Status:             status,
		Progress:           pct,
		Target:             commercialMixTargetPct,
		PeriodYM:           periodYM,
		TriennialCount:     triennialN,
		SubCount:           subN,
	}, true, nil
}

func (s *Store) mixCountsForPeriod(ctx context.Context, commercialUserID, periodYM string) (triennialN, subN int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE pe.plan_code='triennial')::int,
			COUNT(*)::int
		FROM billing.commercial_commission_ledger cl
		JOIN billing.pet_entitlements pe ON pe.id = cl.source_id
		WHERE cl.commercial_user_id=$1
		  AND cl.period_ym=$2
		  AND cl.source_type='subscription_pct'`, commercialUserID, periodYM).Scan(&triennialN, &subN)
	return
}

func mixPct(triennialN, subN int) int {
	if subN <= 0 {
		return 0
	}
	return triennialN * 100 / subN
}

func (s *Store) listBonusAwardsForCommercial(ctx context.Context, commercialUserID string) ([]CommercialBonusAward, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.commercial_user_id::text, cu.full_name, cu.email,
			a.bonus_code, a.amount_cents, a.status, a.period_ym, a.vet_user_id::text,
			vu.email, vu.full_name, a.progress, a.target, a.earned_at, a.paid_at
		FROM billing.commercial_bonus_awards a
		JOIN identity.users cu ON cu.id = a.commercial_user_id
		LEFT JOIN identity.users vu ON vu.id = a.vet_user_id
		WHERE a.commercial_user_id=$1
		ORDER BY a.earned_at DESC`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialBonusAward, 0)
	for rows.Next() {
		var a CommercialBonusAward
		var period, vetID, vetEmail, vetName *string
		var paidAt *time.Time
		if err := rows.Scan(
			&a.ID, &a.CommercialUserID, &a.CommercialFullName, &a.CommercialEmail,
			&a.BonusCode, &a.AmountCents, &a.Status, &period, &vetID,
			&vetEmail, &vetName, &a.Progress, &a.Target, &a.EarnedAt, &paidAt,
		); err != nil {
			return nil, err
		}
		if period != nil {
			a.PeriodYM = *period
		}
		if vetID != nil {
			a.VetUserID = *vetID
		}
		if vetEmail != nil {
			a.VetEmail = *vetEmail
		}
		if vetName != nil {
			a.VetFullName = *vetName
		}
		a.PaidAt = paidAt
		out = append(out, a)
	}
	return out, rows.Err()
}
