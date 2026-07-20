package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ManagerTeamMember struct {
	UserID              string `json:"userId"`
	FullName            string `json:"fullName"`
	Email               string `json:"email"`
	AssignedVets        int    `json:"assignedVets"`
	ProspectsTotal      int    `json:"prospectsTotal"`
	ProspectsNew        int    `json:"prospectsNew"`
	ProspectsContacted  int    `json:"prospectsContacted"`
	ProspectsQualified  int    `json:"prospectsQualified"`
	ProspectsConverted  int    `json:"prospectsConverted"`
	ProspectsLost       int    `json:"prospectsLost"`
	Contacts30d         int    `json:"contacts30d"`
	AppointmentsUpcoming int   `json:"appointmentsUpcoming"`
	AppointmentsDone    int    `json:"appointmentsDone"`
	AppointmentsNoShow  int    `json:"appointmentsNoShow"`
	StaleInPipeline     int    `json:"staleInPipeline"`
	MonthEarnedCents    int    `json:"monthEarnedCents"`
	LifetimeEarnedCents int    `json:"lifetimeEarnedCents"`
}

type ManagerOverview struct {
	Team                  []ManagerTeamMember `json:"team"`
	TeamProspectsTotal    int                 `json:"teamProspectsTotal"`
	TeamProspectsNew      int                 `json:"teamProspectsNew"`
	TeamProspectsContacted int                `json:"teamProspectsContacted"`
	TeamProspectsQualified int                `json:"teamProspectsQualified"`
	TeamProspectsConverted int                `json:"teamProspectsConverted"`
	TeamProspectsLost     int                 `json:"teamProspectsLost"`
	TeamContacts30d       int                 `json:"teamContacts30d"`
	TeamAppointmentsUpcoming int              `json:"teamAppointmentsUpcoming"`
	TeamAppointmentsDone  int                 `json:"teamAppointmentsDone"`
	TeamAppointmentsNoShow int                `json:"teamAppointmentsNoShow"`
	TeamStaleInPipeline   int                 `json:"teamStaleInPipeline"`
	TeamMonthEarnedCents  int                 `json:"teamMonthEarnedCents"`
	TeamLifetimeEarnedCents int              `json:"teamLifetimeEarnedCents"`
	DirectoryTotal        int                 `json:"directoryTotal"`
	ConversionRateBps     int                 `json:"conversionRateBps"`
	Self                  map[string]any      `json:"self"`
}

func (s *Store) CreateCommercialManagerUser(ctx context.Context, email, password, fullName string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	userID := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password)
		VALUES ($1, $2, $3, $4, 'commercial_manager', NULL, NOW(), true)`,
		userID, email, string(hash), fullName)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Store) SetCommercialManager(ctx context.Context, commercialUserID, managerUserID string) error {
	if managerUserID == "" {
		ct, err := s.pool.Exec(ctx, `
			UPDATE identity.users SET manager_user_id=NULL
			WHERE id=$1 AND role='commercial'`, commercialUserID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	}
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM identity.users WHERE id=$1`, managerUserID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if role != "commercial_manager" {
		return errors.New("invalid_manager")
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET manager_user_id=$2
		WHERE id=$1 AND role='commercial'`, commercialUserID, managerUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AssignAllCommercialsToManager(ctx context.Context, managerUserID string) (int64, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET manager_user_id=$1
		WHERE role='commercial'`, managerUserID)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (s *Store) CommercialBelongsToManager(ctx context.Context, commercialUserID, managerUserID string) (bool, error) {
	if commercialUserID == "" {
		return false, nil
	}
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM identity.users
			WHERE id=$1 AND role='commercial' AND manager_user_id=$2
		)`, commercialUserID, managerUserID).Scan(&ok)
	return ok, err
}

func (s *Store) CreateCommercialUserWithManager(ctx context.Context, email, password, fullName, managerUserID string) (string, error) {
	if managerUserID != "" {
		var role string
		err := s.pool.QueryRow(ctx, `SELECT role FROM identity.users WHERE id=$1`, managerUserID).Scan(&role)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		if err != nil {
			return "", err
		}
		if role != "commercial_manager" {
			return "", errors.New("invalid_manager")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	userID := uuid.NewString()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password, manager_user_id)
		VALUES ($1, $2, $3, $4, 'commercial', NULL, NOW(), true, NULLIF($5::text,'')::uuid)`,
		userID, email, string(hash), fullName, managerUserID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Store) ListCommercialManagers(ctx context.Context) ([]CommercialRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE((SELECT COUNT(*)::int FROM identity.users c WHERE c.role='commercial' AND c.manager_user_id=u.id), 0)
		FROM identity.users u
		WHERE u.role='commercial_manager'
		ORDER BY u.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialRow, 0)
	for rows.Next() {
		var c CommercialRow
		if err := rows.Scan(&c.UserID, &c.FullName, &c.Email, &c.ClientCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListManagerTeam(ctx context.Context, managerUserID string) ([]ManagerTeamMember, error) {
	period := time.Now().UTC().Format("2006-01")
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE((SELECT COUNT(*)::int FROM identity.users v WHERE v.role='vet' AND v.assigned_commercial_id=u.id), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id AND p.status='new'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id AND p.status='contacted'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id AND p.status='qualified'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id AND p.status='converted'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id AND p.status='lost'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id
				AND p.first_contacted_at >= NOW() - INTERVAL '30 days'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id
				AND p.appointment_at IS NOT NULL AND p.appointment_at >= NOW()
				AND p.appointment_outcome IN ('','scheduled')), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id
				AND p.appointment_outcome='done'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id
				AND p.appointment_outcome='no_show'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p WHERE p.commercial_user_id=u.id
				AND p.status IN ('contacted','qualified')
				AND p.status_changed_at < NOW() - INTERVAL '7 days'), 0),
			COALESCE((SELECT SUM(l.commission_cents)::int FROM billing.commercial_commission_ledger l
				WHERE l.commercial_user_id=u.id AND l.period_ym=$2), 0),
			COALESCE((SELECT SUM(l.commission_cents)::int FROM billing.commercial_commission_ledger l
				WHERE l.commercial_user_id=u.id), 0)
		FROM identity.users u
		WHERE u.role='commercial' AND u.manager_user_id=$1
		ORDER BY u.full_name`, managerUserID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ManagerTeamMember, 0)
	for rows.Next() {
		var m ManagerTeamMember
		if err := rows.Scan(
			&m.UserID, &m.FullName, &m.Email,
			&m.AssignedVets, &m.ProspectsTotal, &m.ProspectsNew, &m.ProspectsContacted,
			&m.ProspectsQualified, &m.ProspectsConverted, &m.ProspectsLost,
			&m.Contacts30d, &m.AppointmentsUpcoming, &m.AppointmentsDone, &m.AppointmentsNoShow,
			&m.StaleInPipeline, &m.MonthEarnedCents, &m.LifetimeEarnedCents,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ManagerOverview(ctx context.Context, managerUserID string) (ManagerOverview, error) {
	team, err := s.ListManagerTeam(ctx, managerUserID)
	if err != nil {
		return ManagerOverview{}, err
	}
	ov := ManagerOverview{Team: team}
	for _, m := range team {
		ov.TeamProspectsTotal += m.ProspectsTotal
		ov.TeamProspectsNew += m.ProspectsNew
		ov.TeamProspectsContacted += m.ProspectsContacted
		ov.TeamProspectsQualified += m.ProspectsQualified
		ov.TeamProspectsConverted += m.ProspectsConverted
		ov.TeamProspectsLost += m.ProspectsLost
		ov.TeamContacts30d += m.Contacts30d
		ov.TeamAppointmentsUpcoming += m.AppointmentsUpcoming
		ov.TeamAppointmentsDone += m.AppointmentsDone
		ov.TeamAppointmentsNoShow += m.AppointmentsNoShow
		ov.TeamStaleInPipeline += m.StaleInPipeline
		ov.TeamMonthEarnedCents += m.MonthEarnedCents
		ov.TeamLifetimeEarnedCents += m.LifetimeEarnedCents
	}
	closed := ov.TeamProspectsConverted + ov.TeamProspectsLost
	if closed > 0 {
		ov.ConversionRateBps = ov.TeamProspectsConverted * 10000 / closed
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sales.prospects WHERE source='directory'`).Scan(&ov.DirectoryTotal); err != nil {
		return ManagerOverview{}, err
	}
	var directoryStale int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sales.prospects
		WHERE source='directory'
		  AND status IN ('contacted','qualified')
		  AND status_changed_at < NOW() - INTERVAL '7 days'`).Scan(&directoryStale); err != nil {
		return ManagerOverview{}, err
	}
	ov.TeamStaleInPipeline += directoryStale
	self, err := s.CommercialOverview(ctx, managerUserID)
	if err != nil {
		return ManagerOverview{}, err
	}
	ov.Self = self
	return ov, nil
}

func (s *Store) ListManagerProspects(ctx context.Context, managerUserID, statusFilter, commercialUserID string, upcomingOnly bool) ([]Prospect, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, COALESCE(p.commercial_user_id::text,''), p.practice_name, p.contact_name, p.contact_email, p.contact_phone,
			p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at,
			p.first_contacted_at, p.last_contacted_at, p.appointment_at, COALESCE(p.appointment_outcome,''),
			COALESCE(p.lost_reason,''), COALESCE(p.converted_vet_user_id::text,''),
			COALESCE(u.full_name,''), COALESCE(u.email,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE (
			p.source = 'directory'
			OR p.commercial_user_id IN (
				SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			)
		)
		AND ($2='' OR p.status=$2)
		AND ($3='' OR p.commercial_user_id::text=$3)
		AND (
			NOT $4::boolean
			OR (p.appointment_at IS NOT NULL AND p.appointment_at >= NOW() AND p.appointment_outcome IN ('','scheduled'))
		)
		ORDER BY COALESCE(p.appointment_at, p.created_at) DESC`, managerUserID, statusFilter, commercialUserID, upcomingOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Prospect, 0)
	for rows.Next() {
		var p Prospect
		var first, last, appt *time.Time
		if err := rows.Scan(&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
			&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt,
			&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID,
			&p.CommercialName, &p.CommercialEmail); err != nil {
			return nil, err
		}
		p.FirstContactedAt = first
		p.LastContactedAt = last
		p.AppointmentAt = appt
		p.DaysInStatus = daysSince(p.StatusChangedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListManagerFollowups(ctx context.Context, managerUserID string) (map[string]any, error) {
	upcoming, err := s.ListManagerProspects(ctx, managerUserID, "", "", true)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, COALESCE(p.commercial_user_id::text,''), p.practice_name, p.contact_name, p.contact_email, p.contact_phone,
			p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at,
			p.first_contacted_at, p.last_contacted_at, p.appointment_at, COALESCE(p.appointment_outcome,''),
			COALESCE(p.lost_reason,''), COALESCE(p.converted_vet_user_id::text,''),
			COALESCE(u.full_name,''), COALESCE(u.email,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE (
			p.source = 'directory'
			OR p.commercial_user_id IN (
				SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			)
		)
		AND p.status IN ('contacted','qualified')
		AND p.status_changed_at < NOW() - INTERVAL '7 days'
		ORDER BY p.status_changed_at ASC
		LIMIT 100`, managerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stale := make([]Prospect, 0)
	for rows.Next() {
		var p Prospect
		var first, last, appt *time.Time
		if err := rows.Scan(&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
			&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt,
			&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID,
			&p.CommercialName, &p.CommercialEmail); err != nil {
			return nil, err
		}
		p.FirstContactedAt = first
		p.LastContactedAt = last
		p.AppointmentAt = appt
		p.DaysInStatus = daysSince(p.StatusChangedAt)
		stale = append(stale, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{
		"upcomingAppointments": upcoming,
		"staleProspects":       stale,
	}, nil
}
