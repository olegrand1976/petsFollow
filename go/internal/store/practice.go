package store

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/headerlinks"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type PracticeProfile struct {
	PracticeID            string     `json:"practiceId"`
	PracticeName          string     `json:"practiceName"`
	Phone                 string     `json:"phone"`
	ContactEmail          string     `json:"contactEmail"`
	AddressLine1          string     `json:"addressLine1"`
	AddressLine2          string     `json:"addressLine2"`
	City                  string     `json:"city"`
	PostalCode            string     `json:"postalCode"`
	CountryCode           string     `json:"countryCode"`
	AnimalScope           string     `json:"animalScope"`
	Website               string     `json:"website"`
	ProfileCompletedAt    *time.Time `json:"profileCompletedAt,omitempty"`
	VetFullName           string     `json:"vetFullName"`
	VetEmail              string     `json:"vetEmail"`
	HeartRateDurationsSec []int      `json:"heartrateDurationsSec"`
	// DeskIdleMinutes: shared-desk auto-lock delay (1|2|5|10|15|30). Default 2.
	DeskIdleMinutes int `json:"deskIdleMinutes"`
	// Company / payout (for commission sheets) — not required for onboarding.
	CompanyLegalName       string `json:"companyLegalName"`
	VATNumber              string `json:"vatNumber"`
	CompanyNumber          string `json:"companyNumber"`
	LegalForm              string `json:"legalForm"`
	BillingSameAsPractice  bool   `json:"billingSameAsPractice"`
	BillingAddressLine1    string `json:"billingAddressLine1"`
	BillingAddressLine2    string `json:"billingAddressLine2"`
	BillingPostalCode      string `json:"billingPostalCode"`
	BillingCity            string `json:"billingCity"`
	PayoutIBAN             string `json:"payoutIban"`
	PayoutBIC              string `json:"payoutBic"`
	PayoutAccountHolder    string `json:"payoutAccountHolder"`
	PayoutProfileComplete  bool   `json:"payoutProfileComplete"`
	// HeaderLinks: catalog toggles + custom URLs for the Pro topbar.
	HeaderLinks headerlinks.Prefs `json:"headerLinks"`
}

// IsVetPayoutProfileComplete reports whether company + bank fields are sufficient for payout.
func IsVetPayoutProfileComplete(p PracticeProfile) bool {
	if strings.TrimSpace(p.CompanyLegalName) == "" || strings.TrimSpace(p.VATNumber) == "" ||
		strings.TrimSpace(p.CompanyNumber) == "" || strings.TrimSpace(p.LegalForm) == "" {
		return false
	}
	if strings.TrimSpace(p.PayoutIBAN) == "" || strings.TrimSpace(p.PayoutAccountHolder) == "" {
		return false
	}
	if p.BillingSameAsPractice {
		return strings.TrimSpace(p.AddressLine1) != "" && strings.TrimSpace(p.City) != "" && strings.TrimSpace(p.PostalCode) != ""
	}
	return strings.TrimSpace(p.BillingAddressLine1) != "" &&
		strings.TrimSpace(p.BillingCity) != "" &&
		strings.TrimSpace(p.BillingPostalCode) != ""
}

type RegisterVetInput struct {
	Email            string
	Password         string
	FullName         string
	PracticeName     string
	PreferredLocale  string
	AutoReplyDefault string
	// TermsAccepted horodate le consentement CGU/privacy (RGPD art. 7).
	TermsAccepted bool
	// AssignedCommercialID optional commercial from invite code at signup.
	AssignedCommercialID string
}

type RegisterVetResult struct {
	UserID string
	Token  string
}

// PracticeContact is the minimal public contact info for client booking UX.
type PracticeContact struct {
	PracticeID   string
	PracticeName string
	Phone        string
	CountryCode  string
}

func (s *Store) GetPracticeContact(ctx context.Context, practiceID string) (PracticeContact, error) {
	var c PracticeContact
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, name, COALESCE(phone,''), COALESCE(country_code,'BE')
		FROM practice.practices WHERE id = $1`, practiceID,
	).Scan(&c.PracticeID, &c.PracticeName, &c.Phone, &c.CountryCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return PracticeContact{}, ErrNotFound
	}
	return c, err
}

// GetPracticeAnimalScope returns the AFSCA dashboard filter preference (small|large|both).
func (s *Store) GetPracticeAnimalScope(ctx context.Context, practiceID string) (string, error) {
	var scope string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(animal_scope,'both') FROM practice.practices WHERE id = $1`, practiceID,
	).Scan(&scope)
	if errors.Is(err, pgx.ErrNoRows) {
		return "both", ErrNotFound
	}
	if err != nil {
		return "both", err
	}
	return NormalizeAnimalScope(scope), nil
}

// NormalizeAnimalScope returns small|large|both (default both).
// Kept in sync with afsca.NormalizeScope (same values).
func NormalizeAnimalScope(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "small", "large":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "both"
	}
}

// NormalizeCountryCode returns a 2-letter ISO code (default BE).
func NormalizeCountryCode(code string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	if len(c) != 2 {
		return "BE"
	}
	switch c {
	case "BE", "FR", "NL", "LU", "DE", "ES", "IT", "PT", "AT", "CH", "GB", "IE", "PL", "EE", "US":
		return c
	default:
		return "BE"
	}
}

func (s *Store) GetPracticeProfile(ctx context.Context, practiceID, vetUserID string) (PracticeProfile, error) {
	var p PracticeProfile
	var completedAt *time.Time
	var durations []int32
	var headerLinksRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT pr.id::text, pr.name, COALESCE(pr.phone,''), COALESCE(pr.contact_email,''),
			COALESCE(pr.address_line1,''), COALESCE(pr.address_line2,''), COALESCE(pr.city,''),
			COALESCE(pr.postal_code,''), COALESCE(pr.country_code,'BE'), COALESCE(pr.animal_scope,'both'),
			COALESCE(pr.website,''), pr.profile_completed_at,
			u.full_name, u.email, pr.heartrate_durations_sec, pr.desk_idle_minutes,
			COALESCE(pr.company_legal_name,''), COALESCE(pr.vat_number,''), COALESCE(pr.company_number,''),
			COALESCE(pr.legal_form,''), COALESCE(pr.billing_same_as_practice, true),
			COALESCE(pr.billing_address_line1,''), COALESCE(pr.billing_address_line2,''),
			COALESCE(pr.billing_postal_code,''), COALESCE(pr.billing_city,''),
			COALESCE(pr.payout_iban,''), COALESCE(pr.payout_bic,''), COALESCE(pr.payout_account_holder,''),
			COALESCE(pr.header_links, '{}'::jsonb)
		FROM practice.practices pr
		JOIN identity.users u ON u.id = $2 AND u.practice_id = pr.id
		WHERE pr.id = $1`, practiceID, vetUserID).Scan(
		&p.PracticeID, &p.PracticeName, &p.Phone, &p.ContactEmail,
		&p.AddressLine1, &p.AddressLine2, &p.City, &p.PostalCode, &p.CountryCode, &p.AnimalScope, &p.Website, &completedAt,
		&p.VetFullName, &p.VetEmail, &durations, &p.DeskIdleMinutes,
		&p.CompanyLegalName, &p.VATNumber, &p.CompanyNumber, &p.LegalForm, &p.BillingSameAsPractice,
		&p.BillingAddressLine1, &p.BillingAddressLine2, &p.BillingPostalCode, &p.BillingCity,
		&p.PayoutIBAN, &p.PayoutBIC, &p.PayoutAccountHolder,
		&headerLinksRaw,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PracticeProfile{}, ErrNotFound
	}
	if err != nil {
		return PracticeProfile{}, err
	}
	p.ProfileCompletedAt = completedAt
	p.CountryCode = NormalizeCountryCode(p.CountryCode)
	p.AnimalScope = NormalizeAnimalScope(p.AnimalScope)
	p.HeartRateDurationsSec = int32SliceToInts(durations)
	p.DeskIdleMinutes = kernel.NormalizeDeskIdleMinutes(p.DeskIdleMinutes)
	p.PayoutProfileComplete = IsVetPayoutProfileComplete(p)
	p.HeaderLinks = headerlinks.ParsePrefs(headerLinksRaw)
	return p, nil
}

// GetPracticeDeskIdleMinutes returns the shared-desk idle lock delay for a practice.
func (s *Store) GetPracticeDeskIdleMinutes(ctx context.Context, practiceID string) (int, error) {
	var minutes int
	err := s.pool.QueryRow(ctx, `
		SELECT desk_idle_minutes FROM practice.practices WHERE id = $1`, practiceID).Scan(&minutes)
	if errors.Is(err, pgx.ErrNoRows) {
		return kernel.DefaultDeskIdleMinutes, ErrNotFound
	}
	if err != nil {
		return kernel.DefaultDeskIdleMinutes, err
	}
	return kernel.NormalizeDeskIdleMinutes(minutes), nil
}

func int32SliceToInts(in []int32) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

func (s *Store) GetPracticeHeartRateDurations(ctx context.Context, practiceID string) ([]int, error) {
	var durations []int32
	err := s.pool.QueryRow(ctx, `
		SELECT heartrate_durations_sec FROM practice.practices WHERE id = $1`, practiceID).Scan(&durations)
	if errors.Is(err, pgx.ErrNoRows) {
		return []int{60}, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(durations) == 0 {
		return []int{60}, nil
	}
	return int32SliceToInts(durations), nil
}

// GetPracticeName returns the practice display name or ErrNotFound.
func (s *Store) GetPracticeName(ctx context.Context, practiceID string) (string, error) {
	var name string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(name,'') FROM practice.practices WHERE id = $1`, practiceID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return name, err
}

func (s *Store) UpdatePracticeProfile(ctx context.Context, practiceID, vetUserID string, p PracticeProfile, markComplete bool, heartRateDurationsSec *[]int, deskIdleMinutes *int, animalScope *string, headerLinks *headerlinks.Prefs) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET full_name = $2 WHERE id = $1 AND practice_id = $3`,
		vetUserID, p.VetFullName, practiceID); err != nil {
		return err
	}

	q := `
		UPDATE practice.practices
		SET name = $2, phone = $3, contact_email = $4, address_line1 = $5, address_line2 = $6,
			city = $7, postal_code = $8, website = $9, country_code = $10,
			company_legal_name = $11, vat_number = $12, company_number = $13, legal_form = $14,
			billing_same_as_practice = $15, billing_address_line1 = $16, billing_address_line2 = $17,
			billing_postal_code = $18, billing_city = $19,
			payout_iban = $20, payout_bic = $21, payout_account_holder = $22`
	args := []any{
		practiceID, p.PracticeName, p.Phone, p.ContactEmail, p.AddressLine1, p.AddressLine2,
		p.City, p.PostalCode, p.Website, NormalizeCountryCode(p.CountryCode),
		p.CompanyLegalName, p.VATNumber, p.CompanyNumber, p.LegalForm,
		p.BillingSameAsPractice, p.BillingAddressLine1, p.BillingAddressLine2,
		p.BillingPostalCode, p.BillingCity,
		p.PayoutIBAN, p.PayoutBIC, p.PayoutAccountHolder,
	}
	if animalScope != nil {
		args = append(args, NormalizeAnimalScope(*animalScope))
		q += `, animal_scope = $` + strconv.Itoa(len(args))
	}
	if heartRateDurationsSec != nil {
		args = append(args, *heartRateDurationsSec)
		q += `, heartrate_durations_sec = $` + strconv.Itoa(len(args))
	}
	if deskIdleMinutes != nil {
		args = append(args, kernel.NormalizeDeskIdleMinutes(*deskIdleMinutes))
		q += `, desk_idle_minutes = $` + strconv.Itoa(len(args))
	}
	if headerLinks != nil {
		raw, err := headerlinks.MarshalPrefs(*headerLinks)
		if err != nil {
			return err
		}
		args = append(args, json.RawMessage(raw))
		q += `, header_links = $` + strconv.Itoa(len(args))
	}
	if markComplete {
		q += `, profile_completed_at = COALESCE(profile_completed_at, NOW())`
	}
	q += ` WHERE id = $1`
	if _, err := tx.Exec(ctx, q, args...); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = s.RefreshVetPayoutLineStatusesForPractice(ctx, practiceID)
	return nil
}

// GetPracticeHeaderLinks returns raw header link prefs for a practice.
func (s *Store) GetPracticeHeaderLinks(ctx context.Context, practiceID string) (headerlinks.Prefs, string, error) {
	var raw []byte
	var country string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(header_links, '{}'::jsonb), COALESCE(country_code, 'BE')
		FROM practice.practices WHERE id = $1`, practiceID).Scan(&raw, &country)
	if errors.Is(err, pgx.ErrNoRows) {
		return headerlinks.EmptyPrefs(), "BE", ErrNotFound
	}
	if err != nil {
		return headerlinks.EmptyPrefs(), "BE", err
	}
	return headerlinks.ParsePrefs(raw), NormalizeCountryCode(country), nil
}

func (s *Store) IsProfileComplete(ctx context.Context, practiceID string) (bool, error) {
	var complete bool
	err := s.pool.QueryRow(ctx, `
		SELECT profile_completed_at IS NOT NULL
			AND phone <> '' AND address_line1 <> '' AND city <> '' AND postal_code <> '' AND contact_email <> ''
		FROM practice.practices WHERE id = $1`, practiceID).Scan(&complete)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return complete, err
}

func (s *Store) RegisterVet(ctx context.Context, in RegisterVetInput) (RegisterVetResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterVetResult{}, err
	}

	practiceID := uuid.NewString()
	userID := uuid.NewString()
	token := uuid.NewString()
	expires := time.Now().Add(48 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RegisterVetResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO practice.practices (id, name, contact_email) VALUES ($1, $2, $3)`,
		practiceID, in.PracticeName, in.Email); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := insertPrimarySiteTx(ctx, tx, practiceID); err != nil {
		return RegisterVetResult{}, err
	}
	var assignedCommercial any
	if in.AssignedCommercialID != "" {
		assignedCommercial = in.AssignedCommercialID
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, preferred_locale, terms_accepted_at, assigned_commercial_id)
		VALUES ($1, $2, $3, $4, 'vet', $5, $6, CASE WHEN $7 THEN NOW() END, $8)`,
		userID, in.Email, string(hash), in.FullName, practiceID, i18n.NormalizeLocale(in.PreferredLocale), in.TermsAccepted, assignedCommercial); err != nil {
		return RegisterVetResult{}, err
	}
	autoReply := in.AutoReplyDefault
	if autoReply == "" {
		autoReply = "Je suis indisponible, je reviens vers vous rapidement."
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, 'available', $3)`,
		userID, practiceID, autoReply); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, true, true)`, userID); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), userID, token, expires); err != nil {
		return RegisterVetResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RegisterVetResult{}, err
	}
	_ = s.EnsureUserProfiles(ctx, userID)
	_ = s.EnsureReferenceTeamMembership(ctx, practiceID, userID)
	if in.AssignedCommercialID != "" {
		_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
			EventType:        FiliationEventVetAssigned,
			CommercialUserID: in.AssignedCommercialID,
			VetUserID:        userID,
			PracticeID:       practiceID,
			ActorUserID:      in.AssignedCommercialID,
			Meta:             map[string]any{"source": "register_invite"},
		})
	}
	return RegisterVetResult{UserID: userID, Token: token}, nil
}

func (s *Store) ConfirmEmail(ctx context.Context, token string) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id::text, expires_at, used_at FROM identity.email_verification_tokens WHERE token = $1`, token).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if usedAt != nil {
		return User{}, errors.New("token already used")
	}
	if time.Now().After(expiresAt) {
		return User{}, errors.New("token expired")
	}

	if _, err := tx.Exec(ctx, `UPDATE identity.users SET email_verified_at = NOW() WHERE id = $1`, userID); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.email_verification_tokens SET used_at = NOW() WHERE token = $1`, token); err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Store) GetUserMe(ctx context.Context, userID string) (map[string]any, error) {
	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = s.EnsureUserProfiles(ctx, userID)
	profiles, _ := s.ListProfiles(ctx, userID)
	active, _ := s.GetActiveProfile(ctx, userID)
	out := map[string]any{
		"userId":             u.ID,
		"email":              u.Email,
		"role":               u.Role,
		"fullName":           u.FullName,
		"avatarUrl":          u.AvatarURL,
		"emailVerified":      u.EmailVerifiedAt != nil,
		"authProvider":       u.AuthProvider,
		"googleLinked":       u.GoogleSub != "",
		"twoFactorEnabled":   u.TOTPEnabled,
		"preferredLocale":    u.PreferredLocale,
		"mustChangePassword": u.MustChangePassword,
		"contactPhone":       u.ContactPhone,
		"profiles":           profiles,
	}
	if u.TermsAcceptedAt != nil {
		out["termsAcceptedAt"] = u.TermsAcceptedAt.UTC().Format(time.RFC3339)
	} else {
		out["termsAcceptedAt"] = nil
	}
	if active.ID != "" {
		out["activeProfileId"] = active.ID
	}
	if u.ProfessionalSpecialty != "" {
		out["professionalSpecialty"] = u.ProfessionalSpecialty
	}
	if u.PracticeID != "" {
		out["practiceId"] = u.PracticeID
		var practiceName string
		_ = s.pool.QueryRow(ctx, `SELECT name FROM practice.practices WHERE id = $1`, u.PracticeID).Scan(&practiceName)
		out["practiceName"] = practiceName
		complete, _ := s.IsProfileComplete(ctx, u.PracticeID)
		out["profileComplete"] = complete
		if ref, _ := s.IsReferenceVet(ctx, u.PracticeID, userID); ref {
			out["isReferenceVet"] = true
		}
		if kernel.IsPracticeStaff(u.Role) {
			perms, perr := s.PracticePermissions(ctx, u.PracticeID, userID)
			if perr != nil {
				// Ne pas faire échouer /me : la nav Nuxt fail-open tant que la clé est absente.
				log.Printf("GetUserMe practicePermissions user=%s practice=%s: %v", userID, u.PracticeID, perr)
			} else {
				// Toujours exposer la map (même vide) pour que le client distingue « erreur » vs « aucun droit ».
				out["practicePermissions"] = perms
			}
			if _, err := s.EnsurePrimarySite(ctx, u.PracticeID); err != nil {
				log.Printf("GetUserMe EnsurePrimarySite practice=%s: %v", u.PracticeID, err)
			}
			if summaries, serr := s.ListSiteSummaries(ctx, u.PracticeID); serr == nil {
				out["sites"] = summaries
			}
			if defSite, derr := s.GetTeamDefaultSiteID(ctx, u.PracticeID, userID); derr == nil && defSite != "" {
				out["defaultSiteId"] = defSite
			}
		}
	}
	return out, nil
}
