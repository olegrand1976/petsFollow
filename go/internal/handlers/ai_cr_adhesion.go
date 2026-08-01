package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// internalRunAiModuleFriction expires trials, runs adhesion drip, then friction alerts.
func (a *API) internalRunAiModuleFriction(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Ai-Module-Friction-Secret", a.cfg.AiModuleFrictionSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	expired, _ := a.store.ExpireStaleAiCrTrials(r.Context())
	emailsSent, emailsSkipped, adhesionScanned := a.runAiCrAdhesionDrip(r.Context())

	candidates, err := a.store.ListAiCrFrictionCandidates(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	alerts := 0
	now := time.Now().UTC()
	for _, c := range candidates {
		signals := a.detectAiCrFriction(r, c, now)
		for _, sig := range signals {
			exists, _ := a.store.RecentFrictionAlertExists(r.Context(), c.PracticeID, sig.code, 7*24*time.Hour)
			if exists {
				continue
			}
			_ = a.store.InsertFrictionAlert(r.Context(), c.PracticeID, c.CommercialUserID, sig.code, sig.detail)
			alerts++
			if c.CommercialEmail != "" && a.notifier != nil {
				_ = a.notifier.SendVetAlert(c.CommercialEmail,
					"petsFollow — friction module CR IA : "+c.PracticeName,
					"Signal : "+sig.code+"\nCabinet : "+c.PracticeName+"\n"+sig.detail+
						"\nRelancez le cabinet et proposez le mode d'emploi /commercial/ai-cr-playbook.")
			}
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"expiredTrials":   expired,
		"alertsCreated":   alerts,
		"scanned":         len(candidates),
		"adhesionScanned": adhesionScanned,
		"emailsSent":      emailsSent,
		"emailsSkipped":   emailsSkipped,
	})
}

func (a *API) runAiCrAdhesionDrip(ctx context.Context) (sent, skipped, scanned int) {
	cands, err := a.store.ListAiCrAdhesionCandidates(ctx)
	if err != nil {
		log.Printf("ai_cr adhesion list: %v", err)
		return 0, 0, 0
	}
	now := time.Now().UTC()
	cta := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/dashboard"
	for _, c := range cands {
		scanned++
		days := store.AiCrAdhesionDaysSince(c.ActivatedAt, now)
		usage := a.aiCrUsageCount(ctx, c.PracticeID, c.ActivatedAt)
		for _, step := range a.aiCrStepsForDay(days, usage) {
			ok, err := a.sendAiCrAdhesionStep(ctx, c, step, cta, days, usage)
			if err != nil {
				log.Printf("ai_cr adhesion %s %s: %v", c.PracticeID, step, err)
				continue
			}
			if ok {
				sent++
			} else {
				skipped++
			}
		}
	}
	return sent, skipped, scanned
}

func (a *API) aiCrUsageCount(ctx context.Context, practiceID string, since time.Time) int {
	tr, _ := a.store.CountAiCrEvents(ctx, practiceID, store.AiCrUsageTranscribe, since)
	im, _ := a.store.CountAiCrEvents(ctx, practiceID, store.AiCrUsageImprove, since)
	fi, _ := a.store.CountAiCrEvents(ctx, practiceID, store.AiCrUsageFinalize, since)
	return tr + im + fi
}

func (a *API) aiCrStepsForDay(days, usage int) []string {
	out := make([]string, 0, 4)
	if days >= 0 {
		out = append(out, store.AiCrStepJ0Activation)
	}
	if days >= 3 && usage == 0 {
		out = append(out, store.AiCrStepJ3Nudge)
	}
	if days >= 7 && usage == 0 {
		out = append(out, store.AiCrStepJ7Nudge)
	}
	if days >= 14 {
		out = append(out, store.AiCrStepJ14NPS)
	}
	if days >= 15 {
		out = append(out, store.AiCrStepJ15Digest)
	}
	if days >= 30 {
		out = append(out, store.AiCrStepJ30Digest)
	}
	if days >= 45 {
		out = append(out, store.AiCrStepJ45NPS)
	}
	if days >= 60 {
		out = append(out, store.AiCrStepJ60ROI)
	}
	// Paid convert drip (J75/J85/J90) retired: CR IA is included in Pro SaaS.
	return out
}

func (a *API) sendAiCrAdhesionStep(ctx context.Context, c store.AiCrAdhesionCandidate, step, cta string, days, usage int) (sent bool, err error) {
	// Claim first to avoid double-send under concurrent jobs; release on hard failure so drip can retry.
	claimed, err := a.store.TryClaimAiCrEmailSend(ctx, c.PracticeID, step, "sent", map[string]any{
		"days":  days,
		"usage": usage,
	})
	if err != nil || !claimed {
		return false, err
	}
	release := func() {
		if delErr := a.store.DeleteAiCrEmailSend(ctx, c.PracticeID, step); delErr != nil {
			log.Printf("ai_cr adhesion release %s %s: %v", c.PracticeID, step, delErr)
		}
	}
	if a.notifier == nil {
		return true, nil
	}
	vets, err := a.store.ListVetsForVisitAlert(ctx, c.PracticeID, "")
	if err != nil {
		release()
		return false, err
	}
	vars := a.aiCrAdhesionVars(ctx, c, days, usage)
	anySent := false
	for _, v := range vets {
		if strings.TrimSpace(v.Email) == "" {
			continue
		}
		if err := a.notifier.SendAiCrAdhesionStep(v.Email, v.PreferredLocale, v.FullName, step, cta, vars); err != nil {
			log.Printf("ai_cr adhesion mail %s: %v", v.Email, err)
			continue
		}
		anySent = true
	}
	if !anySent {
		// No recipient or all SMTP failed — allow next daily run to retry.
		release()
		return false, nil
	}
	return true, nil
}

func (a *API) aiCrAdhesionVars(ctx context.Context, c store.AiCrAdhesionCandidate, days, usage int) map[string]string {
	tr, _ := a.store.CountAiCrEvents(ctx, c.PracticeID, store.AiCrUsageTranscribe, c.ActivatedAt)
	im, _ := a.store.CountAiCrEvents(ctx, c.PracticeID, store.AiCrUsageImprove, c.ActivatedAt)
	fi, _ := a.store.CountAiCrEvents(ctx, c.PracticeID, store.AiCrUsageFinalize, c.ActivatedAt)
	left := int(time.Until(c.TrialEndsAt).Hours() / 24)
	if left < 0 {
		left = 0
	}
	vars := map[string]string{
		"practiceName": c.PracticeName,
		"days":         fmt.Sprintf("%d", days),
		"daysLeft":     fmt.Sprintf("%d", left),
		"trialEnds":    c.TrialEndsAt.Format("02/01/2006"),
		"usage":        fmt.Sprintf("%d", usage),
		"transcribe":   fmt.Sprintf("%d", tr),
		"improve":      fmt.Sprintf("%d", im),
		"finalize":     fmt.Sprintf("%d", fi),
	}
	roi, err := a.store.ComputeAiCrROI(ctx, c.PracticeID)
	if err == nil && roi.Unlocked {
		vars["minutesSaved"] = fmt.Sprintf("%d", roi.MinutesSaved)
		vars["euroNet"] = fmt.Sprintf("%.0f", float64(roi.NetEuroCents)/100.0)
		vars["crsIA"] = fmt.Sprintf("%d", roi.CrsIA)
	} else {
		vars["minutesSaved"] = "—"
		vars["euroNet"] = "—"
		vars["crsIA"] = fmt.Sprintf("%d", fi)
	}
	return vars
}

func (a *API) notifyPracticeAiModuleActivated(r *http.Request, practiceID string, m store.AiCrModule) {
	if a.notifier == nil {
		return
	}
	claimed, err := a.store.TryClaimAiCrEmailSend(r.Context(), practiceID, store.AiCrStepJ0Activation, "sent", map[string]any{
		"source": "activate",
	})
	if err != nil || !claimed {
		return
	}
	vets, err := a.store.ListVetsForVisitAlert(r.Context(), practiceID, "")
	if err != nil {
		log.Printf("ai_cr activation list vets %s: %v", practiceID, err)
		return
	}
	cta := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/dashboard"
	c := store.AiCrAdhesionCandidate{
		PracticeID:   practiceID,
		PracticeName: m.PracticeName,
		Status:       m.Status,
		ActivatedAt:  m.ActivatedAt,
		TrialEndsAt:  m.TrialEndsAt,
	}
	vars := a.aiCrAdhesionVars(r.Context(), c, 0, 0)
	// Keep claim even if SMTP fails: avoid re-spamming J0 on every drip/activate retry.
	for _, v := range vets {
		if strings.TrimSpace(v.Email) == "" {
			continue
		}
		if err := a.notifier.SendAiCrAdhesionStep(v.Email, v.PreferredLocale, v.FullName, store.AiCrStepJ0Activation, cta, vars); err != nil {
			log.Printf("ai_cr activation mail %s: %v", v.Email, err)
		}
	}
}
