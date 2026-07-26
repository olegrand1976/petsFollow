package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SalesBranch struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	ExternalMLMID string    `json:"externalMlmId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	MemberCount   int       `json:"memberCount,omitempty"`
}

type SalesNetworkInfo struct {
	MLMOrgEnabled bool           `json:"mlmOrgEnabled"`
	Branch        *SalesBranch   `json:"branch,omitempty"`
	Sponsor       *NetworkPerson `json:"sponsor,omitempty"`
	Manager       *NetworkPerson `json:"manager,omitempty"`
	SalesRank     int            `json:"salesRank"`
	Downline      []DownlineNode `json:"downline"`
	DownlineNote  string         `json:"downlineNote,omitempty"`
}

type NetworkPerson struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Role     string `json:"role,omitempty"`
}

type DownlineNode struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Depth    int    `json:"depth"`
	Role     string `json:"role"`
}

type CommercialAdminRow struct {
	UserID        string  `json:"userId"`
	FullName      string  `json:"fullName"`
	Email         string  `json:"email"`
	ClientCount   int     `json:"clientCount"`
	ManagerUserID string  `json:"managerUserId,omitempty"`
	ManagerName   string  `json:"managerName,omitempty"`
	SponsorUserID string  `json:"sponsorUserId,omitempty"`
	SponsorName   string  `json:"sponsorName,omitempty"`
	BranchID      string  `json:"branchId,omitempty"`
	BranchName    string  `json:"branchName,omitempty"`
	BranchCode    string  `json:"branchCode,omitempty"`
	SalesRank     int     `json:"salesRank"`
	BaseCity      string  `json:"baseCity,omitempty"`
	BasePostalCode string `json:"basePostalCode,omitempty"`
}

type CommercialQuota struct {
	UserID             string `json:"userId"`
	PeriodYM           string `json:"periodYm"`
	TargetActivations  int    `json:"targetActivations"`
	TargetEarnedCents  int    `json:"targetEarnedCents"`
	ActualActivations  int    `json:"actualActivations"`
	ActualEarnedCents  int    `json:"actualEarnedCents"`
	FullName           string `json:"fullName,omitempty"`
}

func (s *Store) ListSalesBranches(ctx context.Context) ([]SalesBranch, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id::text, b.name, b.code, COALESCE(b.external_mlm_id,''), b.created_at,
			COALESCE((SELECT COUNT(*)::int FROM identity.users u
				WHERE u.branch_id=b.id AND u.role IN ('commercial','commercial_manager')), 0)
		FROM sales.branches b
		ORDER BY b.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SalesBranch, 0)
	for rows.Next() {
		var b SalesBranch
		if err := rows.Scan(&b.ID, &b.Name, &b.Code, &b.ExternalMLMID, &b.CreatedAt, &b.MemberCount); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) CreateSalesBranch(ctx context.Context, name, code, externalMLMID string) (SalesBranch, error) {
	name = strings.TrimSpace(name)
	code = strings.ToUpper(strings.TrimSpace(code))
	if name == "" || code == "" {
		return SalesBranch{}, ErrValidation
	}
	id := uuid.NewString()
	var b SalesBranch
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sales.branches (id, name, code, external_mlm_id)
		VALUES ($1, $2, $3, NULLIF($4,''))
		RETURNING id::text, name, code, COALESCE(external_mlm_id,''), created_at`,
		id, name, code, strings.TrimSpace(externalMLMID),
	).Scan(&b.ID, &b.Name, &b.Code, &b.ExternalMLMID, &b.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return SalesBranch{}, ErrConflict
		}
		return SalesBranch{}, err
	}
	return b, nil
}

func (s *Store) SetUserBranch(ctx context.Context, userID, branchID string) error {
	if branchID == "" {
		ct, err := s.pool.Exec(ctx, `
			UPDATE identity.users SET branch_id=NULL
			WHERE id=$1 AND role IN ('commercial','commercial_manager')`, userID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sales.branches WHERE id=$1)`, branchID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET branch_id=$2
		WHERE id=$1 AND role IN ('commercial','commercial_manager')`, userID, branchID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDownline returns sponsored/managed reps up to maxDepth.
// When maxDepth<=1 (current SFM), uses manager_user_id / sponsor_user_id direct children only.
func (s *Store) ListDownline(ctx context.Context, userID string, maxDepth int) ([]DownlineNode, error) {
	if maxDepth < 1 {
		maxDepth = 1
	}
	if maxDepth > 10 {
		maxDepth = 10
	}
	// Recursive CTE on sponsor_user_id with fallback to manager_user_id for legacy rows.
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE tree AS (
			SELECT u.id, u.full_name, u.email, u.role, 1 AS depth
			FROM identity.users u
			WHERE u.role IN ('commercial','commercial_manager')
			  AND (u.sponsor_user_id=$1 OR (u.sponsor_user_id IS NULL AND u.manager_user_id=$1))
			UNION ALL
			SELECT c.id, c.full_name, c.email, c.role, t.depth+1
			FROM identity.users c
			JOIN tree t ON (c.sponsor_user_id=t.id OR (c.sponsor_user_id IS NULL AND c.manager_user_id=t.id))
			WHERE t.depth < $2
			  AND c.role IN ('commercial','commercial_manager')
		)
		SELECT id::text, full_name, email, role, depth FROM tree
		ORDER BY depth, full_name`, userID, maxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DownlineNode, 0)
	for rows.Next() {
		var n DownlineNode
		if err := rows.Scan(&n.UserID, &n.FullName, &n.Email, &n.Role, &n.Depth); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) GetSalesNetworkInfo(ctx context.Context, userID string, mlmOrgEnabled bool) (SalesNetworkInfo, error) {
	info := SalesNetworkInfo{
		MLMOrgEnabled: mlmOrgEnabled,
		Downline:      []DownlineNode{},
		DownlineNote:  "mlm_integration_pending",
	}
	var branchID, sponsorID, managerID *string
	var salesRank int
	err := s.pool.QueryRow(ctx, `
		SELECT branch_id::text, sponsor_user_id::text, manager_user_id::text, sales_rank
		FROM identity.users WHERE id=$1`, userID,
	).Scan(&branchID, &sponsorID, &managerID, &salesRank)
	if errors.Is(err, pgx.ErrNoRows) {
		return info, ErrNotFound
	}
	if err != nil {
		return info, err
	}
	info.SalesRank = salesRank
	if branchID != nil && *branchID != "" {
		var b SalesBranch
		err := s.pool.QueryRow(ctx, `
			SELECT id::text, name, code, COALESCE(external_mlm_id,''), created_at
			FROM sales.branches WHERE id=$1`, *branchID,
		).Scan(&b.ID, &b.Name, &b.Code, &b.ExternalMLMID, &b.CreatedAt)
		if err == nil {
			info.Branch = &b
		}
	}
	loadPerson := func(id *string) (*NetworkPerson, error) {
		if id == nil || *id == "" {
			return nil, nil
		}
		var p NetworkPerson
		err := s.pool.QueryRow(ctx, `
			SELECT id::text, full_name, email, role FROM identity.users WHERE id=$1`, *id,
		).Scan(&p.UserID, &p.FullName, &p.Email, &p.Role)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &p, nil
	}
	if info.Sponsor, err = loadPerson(sponsorID); err != nil {
		return info, err
	}
	if info.Manager, err = loadPerson(managerID); err != nil {
		return info, err
	}
	maxDepth := 1
	if mlmOrgEnabled {
		maxDepth = 5
	}
	downline, err := s.ListDownline(ctx, userID, maxDepth)
	if err != nil {
		return info, err
	}
	info.Downline = downline
	if len(downline) > 0 {
		info.DownlineNote = ""
	}
	return info, nil
}

func (s *Store) ListAllCommercialsAdmin(ctx context.Context) ([]CommercialAdminRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE((SELECT COUNT(*)::int FROM identity.users v WHERE v.role='vet' AND v.assigned_commercial_id=u.id), 0),
			COALESCE(u.manager_user_id::text,''), COALESCE(m.full_name,''),
			COALESCE(u.sponsor_user_id::text,''), COALESCE(sp.full_name,''),
			COALESCE(u.branch_id::text,''), COALESCE(b.name,''), COALESCE(b.code,''),
			u.sales_rank,
			COALESCE(u.base_city,''), COALESCE(u.base_postal_code,'')
		FROM identity.users u
		LEFT JOIN identity.users m ON m.id=u.manager_user_id
		LEFT JOIN identity.users sp ON sp.id=u.sponsor_user_id
		LEFT JOIN sales.branches b ON b.id=u.branch_id
		WHERE u.role='commercial'
		ORDER BY u.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialAdminRow, 0)
	for rows.Next() {
		var c CommercialAdminRow
		if err := rows.Scan(
			&c.UserID, &c.FullName, &c.Email, &c.ClientCount,
			&c.ManagerUserID, &c.ManagerName,
			&c.SponsorUserID, &c.SponsorName,
			&c.BranchID, &c.BranchName, &c.BranchCode,
			&c.SalesRank, &c.BaseCity, &c.BasePostalCode,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ReassignProspectCommercial(ctx context.Context, prospectID, commercialUserID, managerUserID string) error {
	existing, err := s.GetProspectByID(ctx, prospectID)
	if err != nil {
		return err
	}
	// Unassign only allowed for directory pool (avoid orphaning owned CRM rows).
	if commercialUserID == "" {
		if existing.Source != "directory" {
			return ErrValidation
		}
	}
	if existing.Source != "directory" {
		ok, err := s.CommercialBelongsToManager(ctx, existing.CommercialUserID, managerUserID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNotFound
		}
	} else if existing.CommercialUserID != "" {
		// Directory already claimed by another team → only that manager may reassign.
		ok, err := s.CommercialBelongsToManager(ctx, existing.CommercialUserID, managerUserID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNotFound
		}
	}
	if commercialUserID != "" {
		ok, err := s.CommercialBelongsToManager(ctx, commercialUserID, managerUserID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("invalid_commercial")
		}
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.prospects
		SET commercial_user_id=NULLIF($2::text,'')::uuid, updated_at=NOW()
		WHERE id=$1`, prospectID, commercialUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpsertCommercialQuota(ctx context.Context, userID, periodYM string, targetActivations, targetEarnedCents int) error {
	if periodYM == "" {
		periodYM = PeriodYM(time.Now())
	}
	if targetActivations < 0 || targetEarnedCents < 0 {
		return ErrValidation
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sales.commercial_quotas (user_id, period_ym, target_activations, target_earned_cents, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, period_ym) DO UPDATE SET
			target_activations=$3, target_earned_cents=$4, updated_at=NOW()`,
		userID, periodYM, targetActivations, targetEarnedCents)
	return err
}

func (s *Store) ListTeamQuotas(ctx context.Context, managerUserID, periodYM string) ([]CommercialQuota, error) {
	if periodYM == "" {
		periodYM = PeriodYM(time.Now())
	}
	if !ValidPeriodYM(periodYM) {
		return nil, ErrValidation
	}
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, $2,
			COALESCE(q.target_activations, 0), COALESCE(q.target_earned_cents, 0),
			COALESCE((
				SELECT COUNT(*)::int FROM billing.commercial_commission_ledger cl
				WHERE cl.commercial_user_id=u.id AND cl.period_ym=$2
				  AND cl.source_type IN ('subscription_pct','subscription_mirror')
			), 0),
			COALESCE((
				SELECT SUM(cl.commission_cents)::int FROM billing.commercial_commission_ledger cl
				WHERE cl.commercial_user_id=u.id AND cl.period_ym=$2
			), 0)
		FROM identity.users u
		LEFT JOIN sales.commercial_quotas q ON q.user_id=u.id AND q.period_ym=$2
		WHERE u.role='commercial' AND u.manager_user_id=$1
		ORDER BY u.full_name`, managerUserID, periodYM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialQuota, 0)
	for rows.Next() {
		var q CommercialQuota
		if err := rows.Scan(&q.UserID, &q.FullName, &q.PeriodYM,
			&q.TargetActivations, &q.TargetEarnedCents,
			&q.ActualActivations, &q.ActualEarnedCents); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) ManagerLeaderboard(ctx context.Context, managerUserID, periodYM string) ([]ManagerTeamMember, error) {
	if periodYM == "" {
		periodYM = PeriodYM(time.Now())
	}
	if !ValidPeriodYM(periodYM) {
		return nil, ErrValidation
	}
	team, err := s.ListManagerTeamForPeriod(ctx, managerUserID, periodYM)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(team); i++ {
		for j := i + 1; j < len(team); j++ {
			if team[j].MonthEarnedCents > team[i].MonthEarnedCents {
				team[i], team[j] = team[j], team[i]
			}
		}
	}
	return team, nil
}
