package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProspectInactiveThresholdDays — pastille rouge / stale dashboards (règle terrain).
const ProspectInactiveThresholdDays = 30

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
	UpdatedAt          time.Time  `json:"updatedAt,omitempty"`
	FirstContactedAt   *time.Time `json:"firstContactedAt,omitempty"`
	LastContactedAt    *time.Time `json:"lastContactedAt,omitempty"`
	AppointmentAt      *time.Time `json:"appointmentAt,omitempty"`
	AppointmentOutcome string     `json:"appointmentOutcome,omitempty"`
	LostReason         string     `json:"lostReason,omitempty"`
	ConvertedVetUserID string     `json:"convertedVetUserId,omitempty"`
	EmailOptOut        bool       `json:"emailOptOut"`
	CommercialName     string     `json:"commercialName,omitempty"`
	CommercialEmail    string     `json:"commercialEmail,omitempty"`
	InactiveDays       int        `json:"inactiveDays"`
	Inactive           bool       `json:"inactive"`
}

// ErrProspectOwned is returned when create/claim hits an already-owned prospect (first encoding wins).
type ErrProspectOwned struct {
	OwnerName    string   `json:"ownerName"`
	OwnerUserID  string   `json:"ownerUserId"`
	ProspectID   string   `json:"prospectId"`
	PracticeName string   `json:"practiceName,omitempty"`
	Prospect     Prospect `json:"-"`
}

func (e *ErrProspectOwned) Error() string { return "prospect_already_owned" }
func (e *ErrProspectOwned) Unwrap() error { return ErrConflict }

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
	city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at, updated_at,
	first_contacted_at, last_contacted_at, appointment_at, COALESCE(appointment_outcome,''),
	COALESCE(lost_reason,''), COALESCE(converted_vet_user_id::text,''), COALESCE(email_opt_out, false)`

const prospectSelectColsP = `
	p.id::text, COALESCE(p.commercial_user_id::text,''), p.practice_name, p.contact_name, p.contact_email, p.contact_phone,
	p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at, p.updated_at,
	p.first_contacted_at, p.last_contacted_at, p.appointment_at, COALESCE(p.appointment_outcome,''),
	COALESCE(p.lost_reason,''), COALESCE(p.converted_vet_user_id::text,''), COALESCE(p.email_opt_out, false)`

func scanProspectRow(scan func(dest ...any) error) (Prospect, error) {
	var p Prospect
	var first, last, appt *time.Time
	err := scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt, &p.UpdatedAt,
		&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID, &p.EmailOptOut,
	)
	if err != nil {
		return Prospect{}, err
	}
	p.FirstContactedAt = first
	p.LastContactedAt = last
	p.AppointmentAt = appt
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	applyProspectActivity(&p)
	return p, nil
}

func applyProspectActivity(p *Prospect) {
	last := p.StatusChangedAt
	if p.LastContactedAt != nil && p.LastContactedAt.After(last) {
		last = *p.LastContactedAt
	}
	if !p.UpdatedAt.IsZero() && p.UpdatedAt.After(last) {
		last = p.UpdatedAt
	}
	p.InactiveDays = daysSince(last)
	open := p.Status == "new" || p.Status == "contacted" || p.Status == "qualified"
	p.Inactive = open && p.InactiveDays >= ProspectInactiveThresholdDays
}

func normalizeProspectPhone(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
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
	if owned, err := s.findOwnedDuplicateProspect(ctx, in.PracticeName, in.ContactPhone, in.City); err != nil {
		return Prospect{}, err
	} else if owned != nil {
		return Prospect{}, owned
	}
	// Free match (unclaimed) → claim instead of creating a duplicate.
	if free, err := s.findFreeDuplicateProspect(ctx, in.PracticeName, in.ContactPhone, in.City); err != nil {
		return Prospect{}, err
	} else if free != nil {
		claimed, err := s.ClaimProspect(ctx, free.ID, commercialUserID)
		if err != nil {
			return Prospect{}, err
		}
		return claimed, nil
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

func (s *Store) findOwnedDuplicateProspect(ctx context.Context, practiceName, phone, city string) (*ErrProspectOwned, error) {
	practiceName = strings.TrimSpace(practiceName)
	city = strings.TrimSpace(city)
	phoneDigits := normalizeProspectPhone(phone)
	if practiceName == "" && phoneDigits == "" {
		return nil, nil
	}
	row := s.pool.QueryRow(ctx, `
		SELECT `+prospectSelectColsP+`, COALESCE(u.full_name,''), COALESCE(u.id::text,'')
		FROM sales.prospects p
		JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE p.commercial_user_id IS NOT NULL
		  AND p.status NOT IN ('lost')
		  AND (
			($1 <> '' AND regexp_replace(p.contact_phone, '[^0-9]', '', 'g') = $1 AND length($1) >= 6)
			OR (
				$2 <> '' AND lower(trim(p.practice_name)) = lower(trim($2))
				AND ($3 = '' OR lower(trim(p.city)) = lower(trim($3)))
			)
		  )
		ORDER BY p.created_at ASC
		LIMIT 1`, phoneDigits, practiceName, city)
	var p Prospect
	var first, last, appt *time.Time
	var ownerName, ownerID string
	err := row.Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt, &p.UpdatedAt,
		&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID, &p.EmailOptOut,
		&ownerName, &ownerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.FirstContactedAt = first
	p.LastContactedAt = last
	p.AppointmentAt = appt
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	applyProspectActivity(&p)
	p.CommercialName = ownerName
	return &ErrProspectOwned{
		OwnerName:    ownerName,
		OwnerUserID:  ownerID,
		ProspectID:   p.ID,
		PracticeName: p.PracticeName,
		Prospect:     p,
	}, nil
}

func (s *Store) findFreeDuplicateProspect(ctx context.Context, practiceName, phone, city string) (*Prospect, error) {
	practiceName = strings.TrimSpace(practiceName)
	city = strings.TrimSpace(city)
	phoneDigits := normalizeProspectPhone(phone)
	if practiceName == "" && phoneDigits == "" {
		return nil, nil
	}
	row := s.pool.QueryRow(ctx, `
		SELECT `+prospectSelectCols+`
		FROM sales.prospects
		WHERE commercial_user_id IS NULL
		  AND status NOT IN ('converted')
		  AND (
			($1 <> '' AND regexp_replace(contact_phone, '[^0-9]', '', 'g') = $1 AND length($1) >= 6)
			OR (
				$2 <> '' AND lower(trim(practice_name)) = lower(trim($2))
				AND ($3 = '' OR lower(trim(city)) = lower(trim($3)))
			)
		  )
		ORDER BY created_at ASC
		LIMIT 1`, phoneDigits, practiceName, city)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ClaimProspect atomically assigns an unclaimed prospect to a commercial (first encoding wins).
func (s *Store) ClaimProspect(ctx context.Context, prospectID, commercialUserID string) (Prospect, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE sales.prospects
		SET commercial_user_id=$2::uuid, updated_at=NOW()
		WHERE id=$1 AND commercial_user_id IS NULL AND status <> 'converted'
		RETURNING `+prospectSelectCols, prospectID, commercialUserID)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, gerr := s.GetProspectByID(ctx, prospectID)
		if gerr != nil {
			return Prospect{}, gerr
		}
		if existing.CommercialUserID != "" && existing.CommercialUserID != commercialUserID {
			name, _ := s.userDisplayName(ctx, existing.CommercialUserID)
			return Prospect{}, &ErrProspectOwned{
				OwnerName:    name,
				OwnerUserID:  existing.CommercialUserID,
				ProspectID:   existing.ID,
				PracticeName: existing.PracticeName,
				Prospect:     existing,
			}
		}
		return Prospect{}, ErrConflict
	}
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

func (s *Store) userDisplayName(ctx context.Context, userID string) (string, error) {
	var name string
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(full_name,'') FROM identity.users WHERE id=$1`, userID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return name, err
}

// ProspectLookupResult is the terrain pre-call check (first encoding wins).
type ProspectLookupResult struct {
	Status      string    `json:"status"` // free | owned | not_found
	OwnerName   string    `json:"ownerName,omitempty"`
	OwnerUserID string    `json:"ownerUserId,omitempty"`
	Prospect    *Prospect `json:"prospect,omitempty"`
}

// LookupProspect searches globally by name/phone/city before dialing.
func (s *Store) LookupProspect(ctx context.Context, q string) (ProspectLookupResult, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return ProspectLookupResult{Status: "not_found"}, nil
	}
	like := "%" + q + "%"
	phoneDigits := normalizeProspectPhone(q)
	row := s.pool.QueryRow(ctx, `
		SELECT `+prospectSelectColsP+`, COALESCE(u.full_name,''), COALESCE(u.id::text,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE p.practice_name ILIKE $1
		   OR p.city ILIKE $1
		   OR p.contact_name ILIKE $1
		   OR p.contact_phone ILIKE $1
		   OR ($2 <> '' AND length($2) >= 6 AND regexp_replace(p.contact_phone, '[^0-9]', '', 'g') LIKE '%' || $2 || '%')
		ORDER BY
			CASE WHEN p.commercial_user_id IS NOT NULL THEN 0 ELSE 1 END,
			p.created_at ASC
		LIMIT 1`, like, phoneDigits)
	var p Prospect
	var first, last, appt *time.Time
	var ownerName, ownerID string
	err := row.Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt, &p.UpdatedAt,
		&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID, &p.EmailOptOut,
		&ownerName, &ownerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProspectLookupResult{Status: "not_found"}, nil
	}
	if err != nil {
		return ProspectLookupResult{}, err
	}
	p.FirstContactedAt = first
	p.LastContactedAt = last
	p.AppointmentAt = appt
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	applyProspectActivity(&p)
	p.CommercialName = ownerName
	if p.CommercialUserID == "" {
		return ProspectLookupResult{Status: "free", Prospect: &p}, nil
	}
	return ProspectLookupResult{
		Status:      "owned",
		OwnerName:   ownerName,
		OwnerUserID: ownerID,
		Prospect:    &p,
	}, nil
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
		WHERE id=$1 AND (commercial_user_id=$2 OR commercial_user_id IS NULL)`, id, commercialUserID)
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
		appt = nil
	}

	// First-contact / note update on a free prospect claims it atomically.
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
			commercial_user_id=COALESCE(commercial_user_id, $2::uuid),
			updated_at=NOW()
		WHERE id=$1 AND (commercial_user_id=$2 OR commercial_user_id IS NULL)
		RETURNING `+prospectSelectCols,
		id, commercialUserID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone, in.City, in.Notes, status,
		appt, outcome, in.ClearAppointment, in.LostReason)
	p, err := scanProspectRow(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, gerr := s.GetProspectByID(ctx, id)
		if gerr == nil && existing.CommercialUserID != "" && existing.CommercialUserID != commercialUserID {
			name, _ := s.userDisplayName(ctx, existing.CommercialUserID)
			return Prospect{}, &ErrProspectOwned{
				OwnerName:    name,
				OwnerUserID:  existing.CommercialUserID,
				ProspectID:   existing.ID,
				PracticeName: existing.PracticeName,
				Prospect:     existing,
			}
		}
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	return p, nil
}

// UpdateProspectAsManager updates a prospect owned by a team member or free pool.
func (s *Store) UpdateProspectAsManager(ctx context.Context, managerUserID, id string, in ProspectInput) (Prospect, error) {
	existing, err := s.GetProspectByID(ctx, id)
	if err != nil {
		return Prospect{}, err
	}
	if existing.CommercialUserID != "" {
		ok, err := s.CommercialBelongsToManager(ctx, existing.CommercialUserID, managerUserID)
		if err != nil {
			return Prospect{}, err
		}
		if !ok {
			return Prospect{}, ErrNotFound
		}
	}
	owner := existing.CommercialUserID
	if owner == "" {
		owner = managerUserID
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
		WHERE (commercial_user_id=$1 OR (commercial_user_id IS NULL AND status <> 'converted'))
		  AND ($2='' OR status=$2)
		  AND ($3='' OR source=$3)
		  AND ($4='' OR practice_name ILIKE $5 OR city ILIKE $5 OR contact_name ILIKE $5 OR notes ILIKE $5 OR contact_phone ILIKE $5)`,
		commercialUserID, f.Status, f.Source, q, like,
	).Scan(&total); err != nil {
		return ProspectListPage{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+prospectSelectCols+`
		FROM sales.prospects
		WHERE (commercial_user_id=$1 OR (commercial_user_id IS NULL AND status <> 'converted'))
		  AND ($2='' OR status=$2)
		  AND ($3='' OR source=$3)
		  AND ($4='' OR practice_name ILIKE $5 OR city ILIKE $5 OR contact_name ILIKE $5 OR notes ILIKE $5 OR contact_phone ILIKE $5)
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
			p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at, p.updated_at,
			p.first_contacted_at, p.last_contacted_at, p.appointment_at, COALESCE(p.appointment_outcome,''),
			COALESCE(p.lost_reason,''), COALESCE(p.converted_vet_user_id::text,''), COALESCE(p.email_opt_out, false),
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
			&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt, &p.UpdatedAt,
			&first, &last, &appt, &p.AppointmentOutcome, &p.LostReason, &p.ConvertedVetUserID, &p.EmailOptOut,
			&p.CommercialName, &p.CommercialEmail); err != nil {
			return nil, err
		}
		p.FirstContactedAt = first
		p.LastContactedAt = last
		p.AppointmentAt = appt
		p.DaysInStatus = daysSince(p.StatusChangedAt)
		applyProspectActivity(&p)
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
		WHERE id=$1 AND (commercial_user_id=$2 OR commercial_user_id IS NULL)`,
		prospectID, commercialUserID, vetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ReleaseProspect clears ownership so the prospect returns to the free pool (manager/admin).
// Converted prospects cannot be released.
func (s *Store) ReleaseProspect(ctx context.Context, prospectID, managerUserID string) error {
	existing, err := s.GetProspectByID(ctx, prospectID)
	if err != nil {
		return err
	}
	if existing.Status == "converted" {
		return ErrValidation
	}
	if existing.CommercialUserID == "" {
		return nil
	}
	ok, err := s.CommercialBelongsToManager(ctx, existing.CommercialUserID, managerUserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.prospects
		SET commercial_user_id=NULL, updated_at=NOW()
		WHERE id=$1 AND status <> 'converted'`, prospectID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE sales.activities SET status='cancelled', updated_at=NOW()
		WHERE prospect_id=$1 AND status='open'`, prospectID)
	return nil
}

// ReleaseInactiveProspectsForManager frees open inactive prospects owned by the manager's team.
func (s *Store) ReleaseInactiveProspectsForManager(ctx context.Context, managerUserID string) (int, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.prospects p
		SET commercial_user_id=NULL, updated_at=NOW()
		FROM identity.users u
		WHERE p.commercial_user_id = u.id
		  AND u.manager_user_id = $1::uuid
		  AND p.status IN ('new','contacted','qualified')
		  AND GREATEST(p.status_changed_at, COALESCE(p.last_contacted_at, p.status_changed_at), p.updated_at)
		      < NOW() - make_interval(days => $2)`, managerUserID, ProspectInactiveThresholdDays)
	if err != nil {
		return 0, err
	}
	return int(ct.RowsAffected()), nil
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
