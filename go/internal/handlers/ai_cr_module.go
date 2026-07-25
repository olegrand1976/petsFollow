package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerAiCrModuleRoutes(r chi.Router) {
	r.Post("/internal/ai-module-friction/run", a.internalRunAiModuleFriction)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)

		pr.Get("/me/ai-module", a.getMyAiModule)
		pr.Get("/me/ai-module/roi", a.getMyAiModuleROI)
		pr.Post("/me/ai-module/feedback", a.postMyAiModuleFeedback)
		pr.Post("/me/ai-module/request-paid", a.requestMyAiModulePaid)

		pr.Get("/admin/ai-modules", a.adminListAiModules)
		pr.Post("/admin/ai-modules/{practiceID}/activate", a.adminActivateAiModule)
		pr.Post("/admin/ai-modules/{practiceID}/convert", a.adminConvertAiModule)
		pr.Patch("/admin/ai-modules/{practiceID}", a.adminPatchAiModule)

		pr.Get("/commercial/ai-modules", a.commercialListAiModules)
		pr.Get("/commercial/ai-modules/friction-alerts", a.commercialListAiFrictionAlerts)
		pr.Post("/commercial/ai-modules/{practiceID}/activate", a.commercialActivateAiModule)
		pr.Post("/commercial/ai-modules/{practiceID}/convert", a.commercialConvertAiModule)
	})
}

func (a *API) requireAiCrPracticeAccess(w http.ResponseWriter, r *http.Request, id authx.Identity) (practiceID string, ok bool) {
	if !kernel.IsPracticeStaff(id.Role) || id.PracticeID == "" {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return "", false
	}
	return id.PracticeID, true
}

func (a *API) getMyAiModule(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if id.Role == kernel.RoleCarePro {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":         "visit_scoped",
			"allowed":        false,
			"roiUnlocked":    false,
			"priceMonthlyHt": 39,
			"priceAnnualHt":  390,
			"hint":           "ai_depends_on_visit_practice",
		})
		return
	}
	practiceID, ok := a.requireAiCrPracticeAccess(w, r, id)
	if !ok {
		return
	}
	m, err := a.store.GetAiCrModule(r.Context(), practiceID)
	if err != nil {
		if err == store.ErrNotFound {
			httpx.WriteData(w, http.StatusOK, map[string]any{
				"practiceId":     practiceID,
				"status":         "none",
				"allowed":        false,
				"roiUnlocked":    false,
				"priceMonthlyHt": 39,
				"priceAnnualHt":  390,
			})
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, enrichAiCrDTO(m))
}

func enrichAiCrDTO(m store.AiCrModule) map[string]any {
	return map[string]any{
		"practiceId":            m.PracticeID,
		"practiceName":          m.PracticeName,
		"status":                m.Status,
		"activatedAt":           m.ActivatedAt,
		"trialEndsAt":           m.TrialEndsAt,
		"convertedAt":           m.ConvertedAt,
		"pricePlan":             m.PricePlan,
		"baselineMinutesPerCr":  m.BaselineMinutesPerCR,
		"hourlyCostCents":       m.HourlyCostCents,
		"daysSinceActivation":   m.DaysSinceActivation,
		"daysRemainingTrial":    m.DaysRemainingTrial,
		"allowed":               m.Allowed,
		"roiUnlocked":           m.RoiUnlocked,
		"priceMonthlyHt":        39,
		"priceAnnualHt":         390,
	}
}

func (a *API) getMyAiModuleROI(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	practiceID, ok := a.requireAiCrPracticeAccess(w, r, id)
	if !ok {
		return
	}
	roi, err := a.store.ComputeAiCrROI(r.Context(), practiceID)
	if err != nil {
		if err == store.ErrNotFound {
			writeErr(w, r, http.StatusNotFound, "not_found", "ai_module_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, roi)
}

type aiCrFeedbackReq struct {
	NPS          int      `json:"nps"`
	Comment      string   `json:"comment"`
	Source       string   `json:"source"`
	FrictionTags []string `json:"frictionTags"`
}

func (a *API) postMyAiModuleFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	practiceID, ok := a.requireAiCrPracticeAccess(w, r, id)
	if !ok {
		return
	}
	if _, err := a.store.GetAiCrModule(r.Context(), practiceID); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "ai_module_not_found")
		return
	}
	var req aiCrFeedbackReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.InsertAiCrFeedback(r.Context(), practiceID, id.UserID, req.NPS, req.Comment, req.Source, req.FrictionTags); err != nil {
		if err == store.ErrValidation {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_nps")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (a *API) requestMyAiModulePaid(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	practiceID, ok := a.requireAiCrPracticeAccess(w, r, id)
	if !ok {
		return
	}
	m, err := a.store.GetAiCrModule(r.Context(), practiceID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "ai_module_not_found")
		return
	}
	commercialID, _ := a.store.ResolvePracticeCommercialID(r.Context(), practiceID)
	detail := "vet_requested_paid conversion practice=" + practiceID + " status=" + m.Status
	_ = a.store.InsertFrictionAlert(r.Context(), practiceID, commercialID, "paid_request", detail)
	if commercialID != "" {
		if u, uerr := a.store.GetUserByID(r.Context(), commercialID); uerr == nil && a.notifier != nil && strings.TrimSpace(u.Email) != "" {
			_ = a.notifier.SendVetAlert(u.Email,
				"petsFollow — demande activation payante CR IA",
				"Le cabinet a demandé la conversion payante du module CR IA.\n"+detail+"\nPrix catalogue : 39 € HT/mois ou 390 € HT/an (facture externe).")
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "requested"})
}

func (a *API) adminListAiModules(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListAiCrModules(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, m := range rows {
		out = append(out, enrichAiCrDTO(m))
	}
	httpx.WriteData(w, http.StatusOK, out)
}

func (a *API) adminActivateAiModule(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	a.activateAiModuleFor(w, r, id.UserID, chi.URLParam(r, "practiceID"), true)
}

func (a *API) adminConvertAiModule(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	a.convertAiModuleFor(w, r, chi.URLParam(r, "practiceID"))
}

type patchAiModuleReq struct {
	Status               *string `json:"status"`
	BaselineMinutesPerCR *int    `json:"baselineMinutesPerCr"`
	HourlyCostCents      *int    `json:"hourlyCostCents"`
}

func (a *API) adminPatchAiModule(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceID")
	var req patchAiModuleReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var m store.AiCrModule
	var err error
	if req.Status != nil {
		m, err = a.store.SetAiCrModuleStatus(r.Context(), practiceID, *req.Status)
	} else {
		m, err = a.store.GetAiCrModule(r.Context(), practiceID)
	}
	if err != nil {
		if err == store.ErrNotFound {
			writeErr(w, r, http.StatusNotFound, "not_found", "ai_module_not_found")
			return
		}
		if err == store.ErrValidation {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if req.BaselineMinutesPerCR != nil || req.HourlyCostCents != nil {
		base := m.BaselineMinutesPerCR
		hourly := m.HourlyCostCents
		if req.BaselineMinutesPerCR != nil {
			base = *req.BaselineMinutesPerCR
		}
		if req.HourlyCostCents != nil {
			hourly = *req.HourlyCostCents
		}
		m, err = a.store.PatchAiCrModuleBaseline(r.Context(), practiceID, base, hourly)
		if err != nil {
			if err == store.ErrValidation {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_baseline")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	httpx.WriteData(w, http.StatusOK, enrichAiCrDTO(m))
}

func (a *API) commercialListAiModules(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	rows, err := a.store.ListAiCrModulesForCommercial(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, m := range rows {
		out = append(out, enrichAiCrDTO(m))
	}
	httpx.WriteData(w, http.StatusOK, out)
}

func (a *API) commercialListAiFrictionAlerts(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	rows, err := a.store.ListRecentFrictionAlertsForCommercial(r.Context(), id.UserID, 30)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) commercialActivateAiModule(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceID")
	owns, err := a.store.CommercialOwnsPractice(r.Context(), id.UserID, practiceID)
	if err != nil || !owns {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_practice")
		return
	}
	a.activateAiModuleFor(w, r, id.UserID, practiceID, false)
}

func (a *API) commercialConvertAiModule(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceID")
	owns, err := a.store.CommercialOwnsPractice(r.Context(), id.UserID, practiceID)
	if err != nil || !owns {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_practice")
		return
	}
	a.convertAiModuleFor(w, r, practiceID)
}

type convertAiModuleReq struct {
	PricePlan string `json:"pricePlan"`
}

func (a *API) convertAiModuleFor(w http.ResponseWriter, r *http.Request, practiceID string) {
	var req convertAiModuleReq
	_ = httpx.DecodeJSON(r, &req)
	m, err := a.store.ConvertAiCrModule(r.Context(), practiceID, req.PricePlan)
	if err != nil {
		if err == store.ErrNotFound {
			writeErr(w, r, http.StatusNotFound, "not_found", "ai_module_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, enrichAiCrDTO(m))
}

func (a *API) activateAiModuleFor(w http.ResponseWriter, r *http.Request, byUserID, practiceID string, isAdmin bool) {
	_ = isAdmin
	practiceID = strings.TrimSpace(practiceID)
	if practiceID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "practice_required")
		return
	}
	m, err := a.store.ActivateAiCrModule(r.Context(), practiceID, byUserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.notifyPracticeAiModuleActivated(r, practiceID, m)
	httpx.WriteData(w, http.StatusOK, enrichAiCrDTO(m))
}

func (a *API) notifyPracticeAiModuleActivated(r *http.Request, practiceID string, m store.AiCrModule) {
	if a.notifier == nil {
		return
	}
	vets, err := a.store.ListVetsForVisitAlert(r.Context(), practiceID, "")
	if err != nil {
		return
	}
	for _, v := range vets {
		if strings.TrimSpace(v.Email) == "" {
			continue
		}
		_ = a.notifier.SendVetAlert(v.Email,
			"petsFollow — module CR IA activé (90 jours offerts)",
			"Le module Compte-rendu IA est activé pour votre cabinet.\n"+
				"Essai : 90 jours jusqu'au "+m.TrialEndsAt.Format("02/01/2006")+".\n"+
				"Dictez vos CR dans Pro Light (Flutter) ou uploadez l'audio sur le calendrier Web.\n"+
				"Dès J60, votre bilan temps gagné / ROI s'affiche sur le tableau de bord.\n"+
				"Après l'essai : 39 € HT/mois ou 390 € HT/an (facture externe).")
	}
}

// requireAiCrEntitlement gates improve/transcribe; returns false after writing the error.
func (a *API) requireAiCrEntitlement(w http.ResponseWriter, r *http.Request, practiceID string) bool {
	ok, err := a.store.AiCrModuleAllowed(r.Context(), practiceID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return false
	}
	if !ok {
		writeErr(w, r, http.StatusPaymentRequired, "payment_required", "ai_module_required")
		return false
	}
	return true
}

func (a *API) trackAiCrUsage(practiceID, userID, visitID string, kind store.AiCrUsageKind) {
	go func() {
		if err := a.store.InsertAiCrUsage(context.Background(), practiceID, userID, visitID, kind); err != nil {
			log.Printf("ai_cr usage: %v", err)
		}
	}()
}

func (a *API) internalRunAiModuleFriction(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Ai-Module-Friction-Secret", a.cfg.AiModuleFrictionSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	expired, _ := a.store.ExpireStaleAiCrTrials(r.Context())
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
		"expiredTrials": expired,
		"alertsCreated": alerts,
		"scanned":       len(candidates),
	})
}

type frictionSignal struct {
	code   string
	detail string
}

func (a *API) detectAiCrFriction(r *http.Request, c store.AiCrFrictionCandidate, now time.Time) []frictionSignal {
	out := make([]frictionSignal, 0)
	days := int(now.Sub(c.ActivatedAt).Hours() / 24)
	sinceAct := c.ActivatedAt

	total, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageTranscribe, sinceAct)
	improve, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageImprove, sinceAct)
	errs, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageError, sinceAct)
	usage := total + improve

	if days >= 7 && usage == 0 {
		out = append(out, frictionSignal{"no_usage_j7", "0 événement IA 7j après activation"})
	}

	weekAgo := now.AddDate(0, 0, -7)
	twoWeeks := now.AddDate(0, 0, -14)
	w1, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageTranscribe, weekAgo)
	w1b, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageImprove, weekAgo)
	w0t, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageTranscribe, twoWeeks)
	w0i, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageImprove, twoWeeks)
	prev := (w0t + w0i) - (w1 + w1b)
	curr := w1 + w1b
	if prev >= 5 && curr*2 < prev {
		out = append(out, frictionSignal{"usage_drop", "chute >50% usage semaine vs précédente"})
	}

	recentErr, _ := a.store.CountAiCrEvents(r.Context(), c.PracticeID, store.AiCrUsageError, weekAgo)
	recentOK := curr
	if recentOK+recentErr > 0 && recentErr*100/(recentOK+recentErr) >= 30 {
		out = append(out, frictionSignal{"gemini_errors", "taux erreur Gemini ≥30% sur 7j"})
	}
	_ = errs

	if nps, ok, _ := a.store.LatestAiCrFeedbackNPS(r.Context(), c.PracticeID); ok && nps <= 6 {
		out = append(out, frictionSignal{"low_nps", "dernier NPS ≤ 6"})
	}

	if days >= 45 {
		j14End := c.ActivatedAt.AddDate(0, 0, 14)
		j30 := c.ActivatedAt.AddDate(0, 0, 30)
		j45 := c.ActivatedAt.AddDate(0, 0, 45)
		earlyT, _ := a.store.CountAiCrEventsInRange(r.Context(), c.PracticeID, store.AiCrUsageTranscribe, sinceAct, j30)
		earlyI, _ := a.store.CountAiCrEventsInRange(r.Context(), c.PracticeID, store.AiCrUsageImprove, sinceAct, j30)
		midT, _ := a.store.CountAiCrEventsInRange(r.Context(), c.PracticeID, store.AiCrUsageTranscribe, j30, j45)
		midI, _ := a.store.CountAiCrEventsInRange(r.Context(), c.PracticeID, store.AiCrUsageImprove, j30, j45)
		_ = j14End
		if earlyT+earlyI > 0 && midT+midI == 0 {
			out = append(out, frictionSignal{"drop_j30_j45", "usage présent puis nul entre J30 et J45"})
		}
	}

	return out
}
