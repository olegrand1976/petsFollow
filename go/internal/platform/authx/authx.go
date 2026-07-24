package authx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type Identity struct {
	UserID     string      `json:"userId"`
	Email      string      `json:"email"`
	Role       kernel.Role `json:"role"`
	PracticeID string      `json:"practiceId,omitempty"`
}

type ctxKey struct{}

var ErrUnauthorized = errors.New("unauthorized")

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type MFALoginResponse struct {
	Requires2FA bool   `json:"requires2FA"`
	MFAToken    string `json:"mfaToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type TokenIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	mfaTTL     time.Duration
}

func NewTokenIssuer(secret string, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		mfaTTL:     5 * time.Minute,
	}
}

type claims struct {
	Email      string       `json:"email"`
	Role       kernel.Role  `json:"role"`
	PracticeID string       `json:"practice_id,omitempty"`
	Typ        string       `json:"typ"`
	jwt.RegisteredClaims
}

func (t *TokenIssuer) Issue(userID, email string, role kernel.Role, practiceID string) (TokenPair, error) {
	now := time.Now()
	accessClaims := claims{
		Email: email, Role: role, PracticeID: practiceID, Typ: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID, IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTTL)), ID: uuid.NewString(),
		},
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(t.secret)
	if err != nil {
		return TokenPair{}, err
	}
	refreshClaims := claims{
		Email: email, Role: role, PracticeID: practiceID, Typ: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID, IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.refreshTTL)), ID: uuid.NewString(),
		},
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(t.secret)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(t.accessTTL.Seconds())}, nil
}

func (t *TokenIssuer) IssueMFA(userID, email string, role kernel.Role, practiceID string) (MFALoginResponse, error) {
	now := time.Now()
	mfaClaims := claims{
		Email: email, Role: role, PracticeID: practiceID, Typ: "mfa_pending",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID, IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.mfaTTL)), ID: uuid.NewString(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, mfaClaims).SignedString(t.secret)
	if err != nil {
		return MFALoginResponse{}, err
	}
	return MFALoginResponse{Requires2FA: true, MFAToken: token, ExpiresIn: int64(t.mfaTTL.Seconds())}, nil
}

func (t *TokenIssuer) ParseMFA(tokenStr string) (Identity, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})
	if err != nil {
		return Identity{}, ErrUnauthorized
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.Typ != "mfa_pending" {
		return Identity{}, ErrUnauthorized
	}
	return Identity{UserID: c.Subject, Email: c.Email, Role: c.Role, PracticeID: c.PracticeID}, nil
}

func (t *TokenIssuer) Parse(tokenStr string) (Identity, error) {
	return t.parseTyped(tokenStr, "access")
}

// IssueJourneyUnsubscribe issues a long-lived token for email journey opt-out links.
func (t *TokenIssuer) IssueJourneyUnsubscribe(userID, email string) (string, error) {
	now := time.Now()
	c := claims{
		Email: email, Role: kernel.RoleClient, Typ: "journey_unsub",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(400 * 24 * time.Hour)),
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(t.secret)
}

func (t *TokenIssuer) ParseJourneyUnsubscribe(tokenStr string) (Identity, error) {
	return t.parseTyped(tokenStr, "journey_unsub")
}

func (t *TokenIssuer) ParseRefresh(tokenStr string) (Identity, error) {
	return t.parseTyped(tokenStr, "refresh")
}

func (t *TokenIssuer) parseTyped(tokenStr, typ string) (Identity, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})
	if err != nil {
		return Identity{}, ErrUnauthorized
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.Typ != typ {
		return Identity{}, ErrUnauthorized
	}
	return Identity{UserID: c.Subject, Email: c.Email, Role: c.Role, PracticeID: c.PracticeID}, nil
}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func FromContext(ctx context.Context) (Identity, error) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	if !ok || id.UserID == "" {
		return Identity{}, ErrUnauthorized
	}
	return id, nil
}

func RequireRole(id Identity, roles ...kernel.Role) error {
	for _, r := range roles {
		if id.Role == r {
			return nil
		}
	}
	return fmt.Errorf("%w: forbidden", ErrUnauthorized)
}
