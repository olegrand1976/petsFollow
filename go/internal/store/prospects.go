package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Prospect struct {
	ID                 string     `json:"id"`
	CommercialUserID   string     `json:"commercialUserId,omitempty"`
	PracticeName       string     `json:"practiceName"`
	ContactName        string     `json:"contactName"`
	ContactEmail       string     `json:"contactEmail"`
	ContactPhone       string     `json:"contactPhone"`
	City               string     `json:"city"`
	Notes              string     `json:"notes"`
	Source             string     `json:"source"`
	ReferringVetUserID string     `json:"referringVetUserId,omitempty"`
	Status             string     `json:"status"`
	StatusChangedAt    time.Time  `json:"statusChangedAt"`
	DaysInStatus       int        `json:"daysInStatus"`
	CreatedAt          time.Time  `json:"createdAt"`
	FirstContactedAt   *time.Time `json:"firstContactedAt,omitempty"`
	LastContactedAt    *time.Time `json:"lastContactedAt,omitempty"`
	AppointmentAt      *time.Time `json:"appointmentAt,omitempty"`
	AppointmentOutcome string     `json:"appointmentOutcome,omitempty"`
	LostReason         string     `json:"lostReason,omitempty"`
	ConvertedVetUserID string     `json:"convertedVetUserId,omitempty"`
	CommercialName     string     `json:"commercialName,omitempty"`
	CommercialEmail    string     `json:"commercialEmail,omitempty"`
}

type ProspectInput struct {
	PracticeName       string
	ContactName        string
	ContactEmail       string
	ContactPhone       string
	City               string
	Notes              string
	Status             string
	Source             string
	AppointmentAt      *time.Time
	AppointmentOutcome string
	LostReason         string
	ClearAppointment   bool
}

func ValidProspectStatus(status string) bool {
	switch status {
	case "new", "contacted", "qualified", "converted", "lost":
		return true
	default:
		return false
	}
}

func ValidProspectSource(source string) bool {
	switch source {
	case "commercial", "vet_referral", "directory":
		return true
	default:
		return false
	}
}

func ValidAppointmentOutcome(outcome string) bool {
	switch outcome {
	case "", "scheduled", "done", "no_show", "cancelled":
		return true
	default:
		return false
	}
}

const prospectSelectCols = `
	id::text, COALESCE(commercial_user_id::text,''), practice_name, contact_name, contact_email, contact_phone,
	city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at,
	first_contacted_at, last_contacted_at, appointment_at, COALESCE(appointment_outcome,''),
	COALESCE(lost_reason,''), COALESCE(converted_vet_user_id::text,'')`

func scanProspectRow(scan func(dest ...any) error) (Prospect, error) {
	var p Prospect
	var first, last, appt *time.Time
	err := scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt,
		&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID,
	)
	if err != nil {
		return Prospect{}, err
	}
	p.FirstContactedAt = first
	p.LastContactedAt = last
	p.AppointmentAt = appt
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	return p, nil
}

func (s *Store) CreateProspect(ctx context.Context, commercialUserID string, in ProspectInput) (Prospect, error) {
	status := in.Status
	if status == "" {
		status = "new"
	}
	source := in.Source
	if source == "" {
		source = "commercial"
	}
	if !ValidProspectSource(source) {
		return Prospect{}, errors.New("invalid prospect source")
	}
	id := uuid.NewString()
	outcome := in.AppointmentOutcome
	if outcome == "" && in.AppointmentAt != nil {
		outcome = "scheduled"
	}
	var firstContact, lastContact any
	if status == "contacted" || status == "qualified" || status == "converted" {
		firstContact = time.Now().UTC()
		lastContact = firstContact
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, status, appointment_at, appointment_outcome, lost_reason,
			first_contacted_at, last_contacted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING `+prospectSelectCols,
		id, commercialUserID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone,
		in.City, in.Notes, source, status, in.AppointmentAt, outcome, in.LostReason, firstContact, lastContact)
	p, err := scanProspectRow(row.Scan)
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

// CreateVetReferralProspect creates a prospect for the commercial assigned to the vet.
func (s *Store) CreateVetReferralProspect(ctx context.Context, vetUserID string, in ProspectInput) (Prospect, error) {
	commercialID, err := s.GetAssignedCommercialID(ctx, vetUserID)
	if err != nil {
		return Prospect{}, err
	}
	if commercialID == "" {
		return Prospect{}, ErrNotFound
	}
	status := in.Status
	if status == "" {
		status = "new"
	}
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, referring_vet_user_id, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'vet_referral',$9,$10)
		RETURNING `+prospectSelectCols,
		id, commercialID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone,
		in.City, in.Notes, vetUserID, status)
	p, err := scanProspectRow(row.Scan)
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

func (s *Store) GetProspect(ctx context.Context, commercialUserID, id string) (Prospect, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+prospectSelectCols+`
		FROM sales.prospects
		WHERE id=$1 AND (commercial_user_id=$2 OR source='directory')`, id, commercialUserID)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

func (s *Store) GetProspectByID(ctx context.Context, id string) (Prospect, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+prospectSelectCols+`
		FROM sales.prospects WHERE id=$1`, id)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

func (s *Store) UpdateProspect(ctx context.Context, commercialUserID, id string, in ProspectInput) (Prospect, error) {
	status := in.Status
	if status == "" {
		status = "new"
	}
	if !ValidAppointmentOutcome(in.AppointmentOutcome) {
		return Prospect{}, errors.New("invalid appointment outcome")
	}
	outcome := in.AppointmentOutcome
	var appt any
	if in.ClearAppointment {
		appt = nil
		outcome = ""
	} else if in.AppointmentAt != nil {
		appt = *in.AppointmentAt
		if outcome == "" {
			outcome = "scheduled"
		}
	} else {
		// Keep existing appointment_at unless outcome/lost/status-only update — use sentinel via SQL COALESCE
		appt = nil
	}

	// When AppointmentAt is nil and not clearing, preserve existing appointment columns in SQL.
	row := s.pool.QueryRow(ctx, `
		UPDATE sales.prospects SET
			practice_name=$3, contact_name=$4, contact_email=$5, contact_phone=$6, city=$7, notes=$8,
			status=$9,
			status_changed_at=CASE WHEN status <> $9 THEN NOW() ELSE status_changed_at END,
			first_contacted_at=CASE
				WHEN $9 IN ('contacted','qualified','converted') AND first_contacted_at IS NULL THEN NOW()
				ELSE first_contacted_at
			END,
			last_contacted_at=CASE
				WHEN $9 IN ('contacted','qualified','converted') AND status <> $9 THEN NOW()
				WHEN $9 IN ('contacted','qualified','converted') AND last_contacted_at IS NULL THEN NOW()
				ELSE last_contacted_at
			END,
			appointment_at=CASE
				WHEN $12::boolean THEN NULL
				WHEN $10::timestamptz IS NOT NULL THEN $10::timestamptz
				ELSE appointment_at
			END,
			appointment_outcome=CASE
				WHEN $12::boolean THEN ''
				WHEN $10::timestamptz IS NOT NULL OR $11::text <> '' THEN $11
				ELSE appointment_outcome
			END,
			lost_reason=CASE WHEN $9='lost' THEN $13 ELSE '' END,
			updated_at=NOW()
		WHERE id=$1 AND (commercial_user_id=$2 OR source='directory')
		RETURNING `+prospectSelectCols,
		id, commercialUserID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone, in.City, in.Notes, status,
		appt, outcome, in.ClearAppointment, in.LostReason)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

// UpdateProspectAsManager updates a prospect owned by a team member or directory (no ownership filter on self).
func (s *Store) UpdateProspectAsManager(ctx context.Context, managerUserID, id string, in ProspectInput) (Prospect, error) {
	existing, err := s.GetProspectByID(ctx, id)
	if err != nil {
		return Prospect{}, err
	}
	if existing.Source != "directory" {
		ok, err := s.CommercialBelongsToManager(ctx, existing.CommercialUserID, managerUserID)
		if err != nil {
			return Prospect{}, err
		}
		if !ok {
			return Prospect{}, ErrNotFound
		}
	}
	// Reuse UpdateProspect with commercial id when owned; for directory use empty owner + directory match via Get path.
	owner := existing.CommercialUserID
	if owner == "" {
		owner = managerUserID // directory: UpdateProspect allows source='directory'
	}
	return s.UpdateProspect(ctx, owner, id, in)
}

func (s *Store) DeleteProspect(ctx context.Context, commercialUserID, id string) error {
	ct, err := s.pool.Exec(ctx, `
		DELETE FROM sales.prospects
		WHERE id=$1 AND commercial_user_id=$2 AND source <> 'directory'`, id, commercialUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ProspectListFilter controls commercial CRM list (pagination + search).
type ProspectListFilter struct {
	Status string
	Source string
	Q      string
	Limit  int
	Offset int
}

// ProspectListPage is a paginated prospect list.
type ProspectListPage struct {
	Items []Prospect `json:"items"`
	Total int        `json:"total"`
}

func (s *Store) ListProspects(ctx context.Context, commercialUserID string, f ProspectListFilter) (ProspectListPage, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	q := strings.TrimSpace(f.Q)
	like := "%" + q + "%"

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM sales.prospects
		WHERE (commercial_user_id=$1 OR source='directory')
		  AND ($2='' OR status=$2)
		  AND ($3='' OR source=$3)
		  AND ($4='' OR practice_name ILIKE $5 OR city ILIKE $5 OR contact_name ILIKE $5 OR notes ILIKE $5)`,
		commercialUserID, f.Status, f.Source, q, like,
	).Scan(&total); err != nil {
		return ProspectListPage{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+prospectSelectCols+`
		FROM sales.prospects
		WHERE (commercial_user_id=$1 OR source='directory')
		  AND ($2='' OR status=$2)
		  AND ($3='' OR source=$3)
		  AND ($4='' OR practice_name ILIKE $5 OR city ILIKE $5 OR contact_name ILIKE $5 OR notes ILIKE $5)
		ORDER BY practice_name ASC, created_at DESC
		LIMIT $6 OFFSET $7`,
		commercialUserID, f.Status, f.Source, q, like, f.Limit, f.Offset)
	if err != nil {
		return ProspectListPage{}, err
	}
	defer rows.Close()
	items, err := scanProspects(rows)
	if err != nil {
		return ProspectListPage{}, err
	}
	return ProspectListPage{Items: items, Total: total}, nil
}

func (s *Store) ListAllProspects(ctx context.Context, statusFilter string) ([]Prospect, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, COALESCE(p.commercial_user_id::text,''), p.practice_name, p.contact_name, p.contact_email, p.contact_phone,
			p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at,
			p.first_contacted_at, p.last_contacted_at, p.appointment_at, COALESCE(p.appointment_outcome,''),
			COALESCE(p.lost_reason,''), COALESCE(p.converted_vet_user_id::text,''),
			COALESCE(u.full_name,''), COALESCE(u.email,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE ($1='' OR p.status=$1)
		ORDER BY p.created_at DESC`, statusFilter)
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

func (s *Store) MarkProspectConverted(ctx context.Context, prospectID, commercialUserID, vetUserID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.prospects SET
			status='converted',
			status_changed_at=CASE WHEN status <> 'converted' THEN NOW() ELSE status_changed_at END,
			converted_vet_user_id=$3::uuid,
			commercial_user_id=COALESCE(commercial_user_id, $2::uuid),
			first_contacted_at=COALESCE(first_contacted_at, NOW()),
			last_contacted_at=NOW(),
			updated_at=NOW()
		WHERE id=$1 AND (commercial_user_id=$2 OR source='directory')`,
		prospectID, commercialUserID, vetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanProspects(rows pgx.Rows) ([]Prospect, error) {
	out := make([]Prospect, 0)
	for rows.Next() {
		p, err := scanProspectRow(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func daysSince(t time.Time) int {
	d := int(time.Since(t).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}
