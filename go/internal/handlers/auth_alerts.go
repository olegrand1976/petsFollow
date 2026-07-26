package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const (
	authAlertDedupWindow   = time.Hour
	authSpikeLoginFail     = 30
	authSpikeUnverified    = 10
	authSpikeRegisterFail  = 5
	authUnverifiedStuckAge = 24 * time.Hour
)

type authPulse struct {
	mu      sync.Mutex
	buckets map[string]map[int64]int // kind → unixMinute → count
}

func newAuthPulse() *authPulse {
	return &authPulse{buckets: map[string]map[int64]int{}}
}

func (p *authPulse) incr(kind string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	min := time.Now().UTC().Unix() / 60
	if p.buckets[kind] == nil {
		p.buckets[kind] = map[int64]int{}
	}
	// Drop buckets older than 3 minutes.
	for k := range p.buckets[kind] {
		if k < min-3 {
			delete(p.buckets[kind], k)
		}
	}
	p.buckets[kind][min]++
	return p.buckets[kind][min]
}

func (a *API) noteAuthSignal(kind string, threshold int, detail string) {
	if a.authPulse == nil {
		return
	}
	n := a.authPulse.incr(kind)
	if n < threshold {
		return
	}
	// Max 1 spike alert / kind / hour (fingerprint = hour UTC).
	fp := time.Now().UTC().Format("2006-01-02-15")
	a.notifyAuthAlert(context.Background(), kind, fp, fmt.Sprintf("%s (count=%d/min)", detail, n))
}

// notifyAuthAlert creates a system support ticket + ops email (ALERT/URGENT). Deduped 1h.
// Ticket is written first so admin UI works even when SMTP is down.
// Returns true when a new alert was recorded.
func (a *API) notifyAuthAlert(ctx context.Context, kind, fingerprint, detail string) bool {
	if a.store == nil {
		return false
	}
	kind = strings.TrimSpace(kind)
	fingerprint = strings.TrimSpace(fingerprint)
	if kind == "" || fingerprint == "" {
		return false
	}
	window := authAlertDedupWindow
	if kind == store.AuthAlertUnverifiedStuck {
		window = 24 * time.Hour
	}
	exists, err := a.store.RecentAuthAlertExists(ctx, kind, fingerprint, window)
	if err != nil {
		log.Printf("auth_alert dedup: %v", err)
		return false
	}
	if exists {
		return false
	}

	subject := "[ALERT/URGENT] petsFollow auth: " + kind
	if len([]rune(subject)) > store.MaxSupportSubjectLen {
		subject = string([]rune(subject)[:store.MaxSupportSubjectLen])
	}
	msg := strings.TrimSpace(detail)
	if msg == "" {
		msg = kind
	}
	diag, _ := json.Marshal(map[string]any{
		"kind":        kind,
		"fingerprint": fingerprint,
		"severity":     "ALERT/URGENT",
	})

	ticket, err := a.store.CreateSupportTicket(ctx, store.CreateSupportTicketInput{
		Source:      store.SupportSourceSystem,
		Subject:     subject,
		Message:     msg,
		Diagnostics: diag,
		Route:       "/admin/support",
	})
	ticketID := ""
	if err != nil {
		log.Printf("auth_alert ticket: %v", err)
	} else {
		ticketID = ticket.ID
	}
	if err := a.store.InsertAuthAlert(ctx, kind, fingerprint, msg, ticketID); err != nil {
		log.Printf("auth_alert insert: %v", err)
	}

	to := strings.TrimSpace(a.cfg.OpsNotifyEmail)
	if to == "" {
		to = strings.TrimSpace(a.cfg.SupportInboxEmail)
	}
	if to == "" || a.notifier == nil {
		return true
	}
	adminURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/admin/support"
	if ticketID != "" {
		adminURL += "/" + ticketID
	}
	body := fmt.Sprintf(
		"<p><strong>[ALERT/URGENT]</strong> Signal auth petsFollow</p>"+
			"<ul><li><strong>Kind</strong> : %s</li><li><strong>Detail</strong> : %s</li></ul>"+
			"<p><a href=\"%s\">Ouvrir dans l'admin</a></p>",
		htmlEsc(kind), htmlEsc(msg), htmlEsc(adminURL),
	)
	// Soft-fail SendVetAlert — never recurse into notifyAuthAlert.
	_ = a.notifier.SendVetAlert(to, subject, body)
	return true
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func (a *API) reportConfirmEmailFailure(ctx context.Context, emailAddr string, sendErr error) {
	detail := fmt.Sprintf("Échec envoi email de confirmation à %s : %v", emailAddr, sendErr)
	log.Printf("event=auth_email_fail email=%s err=%v", emailAddr, sendErr)
	fp := strings.ToLower(strings.TrimSpace(emailAddr))
	if fp == "" {
		fp = "unknown"
	}
	a.notifyAuthAlert(ctx, store.AuthAlertSMTPConfirmFail, fp, detail)
}

func (a *API) internalRunAuthHealth(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Auth-Health-Secret", a.cfg.AuthHealthSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	stuck, err := a.store.CountUnverifiedPasswordClientsOlderThan(r.Context(), authUnverifiedStuckAge)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	alerted := false
	if stuck > 0 {
		fp := time.Now().UTC().Format("2006-01-02")
		detail := fmt.Sprintf("%d client(s) password non vérifiés depuis plus de %s (hors *.petsfollow.test)",
			stuck, authUnverifiedStuckAge)
		alerted = a.notifyAuthAlert(r.Context(), store.AuthAlertUnverifiedStuck, fp, detail)
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"unverifiedStuck": stuck,
		"alerted":         alerted,
	})
}
