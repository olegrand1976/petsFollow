package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// BranchAutoCandidate is a commercial eligible for automatic branch creation.
type BranchAutoCandidate struct {
	UserID        string `json:"userId"`
	FullName      string `json:"fullName"`
	Email         string `json:"email"`
	Locale        string `json:"locale,omitempty"`
	SuggestedName string `json:"suggestedName,omitempty"`
	SuggestedCode string `json:"suggestedCode,omitempty"`
}

// AutoBranchResult is one successful auto-create + assign.
type AutoBranchResult struct {
	UserID   string      `json:"userId"`
	FullName string      `json:"fullName"`
	Email    string      `json:"email"`
	Locale   string      `json:"locale,omitempty"`
	Branch   SalesBranch `json:"branch"`
}

// AutoBranchSkip is one candidate skipped during a batch run.
type AutoBranchSkip struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName,omitempty"`
	Reason   string `json:"reason"`
}

// AutoBranchRunResult summarizes a batch auto-create run.
type AutoBranchRunResult struct {
	Created []AutoBranchResult `json:"created"`
	Skipped []AutoBranchSkip   `json:"skipped"`
}

// DeriveBranchNameAndCode builds display name "Dupont D" and base code "DUPONTD"
// from full_name ("Prénom … Nom"). Single-token names yield name=token, code=TOKEN.
func DeriveBranchNameAndCode(fullName string) (name, codeBase string) {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		last := parts[0]
		return last, branchCodeToken(last)
	}
	first := parts[0]
	last := parts[len(parts)-1]
	initial := firstLetter(first)
	name = last
	if initial != "" {
		name = last + " " + initial
	}
	codeBase = branchCodeToken(last) + branchCodeToken(initial)
	return name, codeBase
}

func firstLetter(s string) string {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return string(unicode.ToUpper(r))
		}
	}
	return ""
}

func branchCodeToken(s string) string {
	s = stripDiacritics(s)
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func stripDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return out
}

// ListBranchlessIndependentCommercials returns commercials without a branch
// who are not sponsored by another commercial (peer).
func (s *Store) ListBranchlessIndependentCommercials(ctx context.Context) ([]BranchAutoCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, COALESCE(u.full_name,''), u.email, COALESCE(u.preferred_locale,'fr')
		FROM identity.users u
		LEFT JOIN identity.users sp ON sp.id = u.sponsor_user_id
		WHERE u.role = 'commercial'
		  AND u.branch_id IS NULL
		  AND (u.sponsor_user_id IS NULL OR sp.role IS DISTINCT FROM 'commercial')
		ORDER BY u.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]BranchAutoCandidate, 0)
	for rows.Next() {
		var c BranchAutoCandidate
		if err := rows.Scan(&c.UserID, &c.FullName, &c.Email, &c.Locale); err != nil {
			return nil, err
		}
		c.SuggestedName, c.SuggestedCode = DeriveBranchNameAndCode(c.FullName)
		out = append(out, c)
	}
	return out, rows.Err()
}

func nextUniqueBranchCodeTx(ctx context.Context, tx pgx.Tx, codeBase string) (string, error) {
	codeBase = strings.ToUpper(strings.TrimSpace(codeBase))
	if codeBase == "" {
		return "", ErrValidation
	}
	for i := range 100 {
		code := codeBase
		if i > 0 {
			code = fmt.Sprintf("%s%d", codeBase, i+1)
		}
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM sales.branches WHERE code=$1)`, code,
		).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", ErrConflict
}

func skipReason(err error) string {
	switch {
	case errors.Is(err, ErrConflict):
		return "conflict"
	case errors.Is(err, ErrValidation):
		return "validation"
	case errors.Is(err, ErrNotFound):
		return "not_found"
	default:
		return "error"
	}
}

// EnsureAutoSalesBranchForUser creates a branch from the commercial's name and assigns it
// in a single transaction (no orphan branch if assign fails).
// Returns ErrConflict if the user is no longer eligible (already branched / peer-sponsored).
func (s *Store) EnsureAutoSalesBranchForUser(ctx context.Context, userID string) (AutoBranchResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AutoBranchResult{}, err
	}
	defer tx.Rollback(ctx)

	var fullName, email, locale string
	var branchID *string
	var sponsorRole *string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(u.full_name,''), u.email, COALESCE(u.preferred_locale,'fr'),
			u.branch_id::text, sp.role
		FROM identity.users u
		LEFT JOIN identity.users sp ON sp.id = u.sponsor_user_id
		WHERE u.id=$1 AND u.role='commercial'
		FOR UPDATE OF u`, userID,
	).Scan(&fullName, &email, &locale, &branchID, &sponsorRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return AutoBranchResult{}, ErrNotFound
	}
	if err != nil {
		return AutoBranchResult{}, err
	}
	if branchID != nil && *branchID != "" {
		return AutoBranchResult{}, ErrConflict
	}
	if sponsorRole != nil && *sponsorRole == "commercial" {
		return AutoBranchResult{}, ErrConflict
	}

	name, codeBase := DeriveBranchNameAndCode(fullName)
	if name == "" || codeBase == "" {
		return AutoBranchResult{}, ErrValidation
	}

	var branch SalesBranch
	for range 100 {
		code, err := nextUniqueBranchCodeTx(ctx, tx, codeBase)
		if err != nil {
			return AutoBranchResult{}, err
		}
		id := uuid.NewString()
		err = tx.QueryRow(ctx, `
			INSERT INTO sales.branches (id, name, code, external_mlm_id)
			VALUES ($1, $2, $3, NULL)
			RETURNING id::text, name, code, COALESCE(external_mlm_id,''), created_at`,
			id, name, code,
		).Scan(&branch.ID, &branch.Name, &branch.Code, &branch.ExternalMLMID, &branch.CreatedAt)
		if err == nil {
			break
		}
		if isUniqueViolation(err) {
			// Concurrent insert won the code — retry with next suffix.
			continue
		}
		return AutoBranchResult{}, err
	}
	if branch.ID == "" {
		return AutoBranchResult{}, ErrConflict
	}

	ct, err := tx.Exec(ctx, `
		UPDATE identity.users SET branch_id=$2
		WHERE id=$1 AND role='commercial' AND branch_id IS NULL`, userID, branch.ID)
	if err != nil {
		return AutoBranchResult{}, err
	}
	if ct.RowsAffected() == 0 {
		return AutoBranchResult{}, ErrConflict
	}

	if err := tx.Commit(ctx); err != nil {
		return AutoBranchResult{}, err
	}
	return AutoBranchResult{
		UserID:   userID,
		FullName: fullName,
		Email:    email,
		Locale:   locale,
		Branch:   branch,
	}, nil
}

// RunAutoSalesBranches creates branches for all eligible commercials.
func (s *Store) RunAutoSalesBranches(ctx context.Context) (AutoBranchRunResult, error) {
	cands, err := s.ListBranchlessIndependentCommercials(ctx)
	if err != nil {
		return AutoBranchRunResult{}, err
	}
	out := AutoBranchRunResult{
		Created: make([]AutoBranchResult, 0, len(cands)),
		Skipped: make([]AutoBranchSkip, 0),
	}
	for _, c := range cands {
		res, err := s.EnsureAutoSalesBranchForUser(ctx, c.UserID)
		if err != nil {
			reason := skipReason(err)
			fmt.Printf("sales-branches-auto: skip user=%s name=%q reason=%s err=%v\n",
				c.UserID, c.FullName, reason, err)
			out.Skipped = append(out.Skipped, AutoBranchSkip{
				UserID:   c.UserID,
				FullName: c.FullName,
				Reason:   reason,
			})
			continue
		}
		out.Created = append(out.Created, res)
	}
	return out, nil
}
