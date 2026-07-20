package journey

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const advisoryLockKey int64 = 824719001

// Mailer sends a journey step email.
type Mailer interface {
	SendJourneyStep(to, locale, fullName, stepKey, ctaURL, unsubscribeURL string, vars map[string]string) error
}

type Config struct {
	AppDownloadURL string
	APIPublicURL   string
	Interval       time.Duration
	Enabled        bool
	BatchLimit     int
}

type Runner struct {
	store  *store.Store
	mailer Mailer
	tokens *authx.TokenIssuer
	cfg    Config
}

func NewRunner(st *store.Store, mailer Mailer, tokens *authx.TokenIssuer, cfg Config) *Runner {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Hour
	}
	if cfg.BatchLimit <= 0 {
		cfg.BatchLimit = 500
	}
	if cfg.APIPublicURL == "" {
		cfg.APIPublicURL = "http://localhost:8291"
	}
	return &Runner{store: st, mailer: mailer, tokens: tokens, cfg: cfg}
}

// Start runs backfill once then ticks until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) {
	if !r.cfg.Enabled {
		log.Printf("journey: email runner disabled")
		return
	}
	if n, err := r.store.BackfillEmailJourneys(ctx); err != nil {
		log.Printf("journey: backfill error: %v", err)
	} else if n > 0 {
		log.Printf("journey: backfilled %d clients", n)
	}
	// First pass shortly after boot (useful in dev / MailHog).
	r.RunOnce(ctx)
	t := time.NewTicker(r.cfg.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.RunOnce(ctx)
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context) {
	if err := r.store.WithAdvisoryLock(ctx, advisoryLockKey, func(ctx context.Context) error {
		return r.runLocked(ctx)
	}); err != nil {
		log.Printf("journey: run error: %v", err)
	}
}

func (r *Runner) runLocked(ctx context.Context) error {
	journeys, err := r.store.ListActiveEmailJourneys(ctx, r.cfg.BatchLimit)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, j := range journeys {
		if err := r.processJourney(ctx, j, now); err != nil {
			log.Printf("journey: user %s: %v", j.UserID, err)
		}
	}
	return nil
}

func (r *Runner) processJourney(ctx context.Context, j store.EmailJourney, now time.Time) error {
	seg, err := r.store.LoadJourneyClientSegment(ctx, j.UserID, j.AnchorAt, now)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}
	for _, step := range TimedSteps() {
		if err := r.maybeSendTimed(ctx, j, seg, step, now); err != nil {
			return err
		}
	}
	for _, step := range EventSteps() {
		if err := r.maybeSendEvent(ctx, j, seg, step, now); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) maybeSendTimed(ctx context.Context, j store.EmailJourney, seg store.JourneyClientSegment, step Step, now time.Time) error {
	dueAt := j.AnchorAt.UTC().AddDate(0, 0, step.OffsetDays)
	if now.Before(dueAt) {
		return nil
	}
	exists, err := r.store.HasEmailSend(ctx, j.UserID, step.Key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !PrefEnabled(seg, step.Pref) {
		return r.store.RecordEmailSend(ctx, j.UserID, step.Key, "skipped", map[string]any{"reason": "pref_off"})
	}
	ok, reason := step.Eligible(seg, now)
	if !ok {
		return r.store.RecordEmailSend(ctx, j.UserID, step.Key, "skipped", map[string]any{"reason": reason})
	}
	if err := r.send(ctx, seg, step.Key); err != nil {
		return err
	}
	if step.Key == "d365_anniversary" {
		_ = r.store.CompleteEmailJourney(ctx, j.UserID)
	}
	return nil
}

// TriggerPastDue sends evt_past_due immediately after a Stripe past_due transition (best-effort).
func (r *Runner) TriggerPastDue(ctx context.Context, ownerUserID string) {
	if !r.cfg.Enabled || ownerUserID == "" {
		return
	}
	j, err := r.store.GetEmailJourney(ctx, ownerUserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			_ = r.store.EnrollEmailJourney(ctx, ownerUserID, time.Now().UTC())
			j, err = r.store.GetEmailJourney(ctx, ownerUserID)
		}
		if err != nil {
			log.Printf("journey: TriggerPastDue enroll %s: %v", ownerUserID, err)
			return
		}
	}
	if j.Status != store.JourneyStatusActive && j.Status != "" {
		// Still allow billing emails when paused — past_due is gated by billing pref only.
	}
	now := time.Now().UTC()
	seg, err := r.store.LoadJourneyClientSegment(ctx, ownerUserID, j.AnchorAt, now)
	if err != nil {
		log.Printf("journey: TriggerPastDue segment %s: %v", ownerUserID, err)
		return
	}
	for _, step := range EventSteps() {
		if step.Key != "evt_past_due" {
			continue
		}
		if err := r.maybeSendEvent(ctx, j, seg, step, now); err != nil {
			log.Printf("journey: TriggerPastDue send %s: %v", ownerUserID, err)
		}
		return
	}
}

func (r *Runner) maybeSendEvent(ctx context.Context, j store.EmailJourney, seg store.JourneyClientSegment, step Step, now time.Time) error {
	ok, _ := step.Eligible(seg, now)
	if !ok {
		return nil
	}
	if !PrefEnabled(seg, step.Pref) {
		return nil
	}
	last, err := r.store.LastEmailSendAt(ctx, j.UserID, step.Key)
	if err != nil {
		return err
	}
	if !last.IsZero() {
		if step.Cooldown <= 0 {
			return nil // once forever
		}
		if now.Sub(last) < step.Cooldown {
			return nil
		}
	}
	return r.send(ctx, seg, step.Key)
}

func (r *Runner) send(ctx context.Context, seg store.JourneyClientSegment, stepKey string) error {
	cta := appendUTM(r.cfg.AppDownloadURL, stepKey)
	unsub := ""
	if r.tokens != nil {
		tok, err := r.tokens.IssueJourneyUnsubscribe(seg.UserID, seg.Email)
		if err == nil {
			unsub = strings.TrimRight(r.cfg.APIPublicURL, "/") + "/api/v1/public/journey/unsubscribe?token=" + url.QueryEscape(tok)
		}
	}
	name := seg.FullName
	if name == "" {
		name = seg.Email
	}
	vars := map[string]string{
		"fullName": name,
	}
	// Soft upsells live in the detail block — omit when not contextual.
	switch stepKey {
	case "d4_routine":
		if seg.ActiveAddons["care_plus"] {
			vars["_omitDetail"] = "1"
		}
	case "d30_habit":
		if !FamilySoftEligible(seg) {
			vars["_omitDetail"] = "1"
		}
	case "d90_quarter":
		if !QuarterFamilySoftEligible(seg) {
			vars["_omitDetail"] = "1"
		}
	case "d330_prerenew":
		if AnnualNearRenewal(seg, time.Now().UTC()) {
			vars["_introNear"] = "1"
		}
	}
	if err := r.mailer.SendJourneyStep(seg.Email, seg.Locale, name, stepKey, cta, unsub, vars); err != nil {
		return err
	}
	return r.store.RecordEmailSend(ctx, seg.UserID, stepKey, "sent", map[string]any{})
}

// MarkStepSkipped records a step as skipped (e.g. app invite covers d0_welcome).
func (r *Runner) MarkStepSkipped(ctx context.Context, userID, stepKey, reason string) error {
	exists, err := r.store.HasEmailSend(ctx, userID, stepKey)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return r.store.RecordEmailSend(ctx, userID, stepKey, "skipped", map[string]any{"reason": reason})
}

func appendUTM(raw, stepKey string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set("utm_source", "petsfollow")
	q.Set("utm_medium", "email")
	q.Set("utm_campaign", "client_journey")
	q.Set("utm_content", stepKey)
	u.RawQuery = q.Encode()
	return u.String()
}
