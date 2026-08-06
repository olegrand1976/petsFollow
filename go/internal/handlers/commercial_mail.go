package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

var hrefRewriteRE = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)

func (a *API) registerCommercialMailRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/commercial/email-templates", a.commercialListEmailTemplates)
		pr.Post("/commercial/email-templates", a.commercialCreateEmailTemplate)
		pr.Get("/commercial/email-templates/{id}", a.commercialGetEmailTemplate)
		pr.Patch("/commercial/email-templates/{id}", a.commercialPatchEmailTemplate)
		pr.Post("/commercial/email-templates/{id}/preview", a.commercialPreviewEmailTemplate)
		pr.Get("/commercial/emails", a.commercialListEmails)
		pr.Get("/commercial/emails/{id}", a.commercialGetEmail)
		pr.Get("/commercial/prospects/{id}/emails", a.commercialListProspectEmails)
		pr.Post("/commercial/prospects/{id}/emails", a.commercialSendProspectEmail)
	})
}

func (a *API) registerCommercialMailPublicRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(rateLimit)
		pr.Get("/public/commercial-mail/o/{token}", a.publicCommercialMailOpen)
		pr.Get("/public/commercial-mail/c/{token}", a.publicCommercialMailClick)
		pr.Get("/public/commercial-mail/unsubscribe/{token}", a.publicCommercialMailUnsubscribe)
	})
}

func (a *API) commercialListEmailTemplates(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCommercial(w, r); !ok {
		return
	}
	activeOnly := r.URL.Query().Get("active") == "1" || r.URL.Query().Get("active") == "true"
	items, err := a.store.ListEmailTemplates(r.Context(), activeOnly)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type emailTemplateReq struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Locale   string `json:"locale"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"bodyHtml"`
	IsActive *bool  `json:"isActive"`
}

func (a *API) commercialCreateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	var req emailTemplateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.Slug) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.BodyHTML) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if req.Category != "" && !store.ValidEmailTemplateCategory(req.Category) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_category")
		return
	}
	tpl, err := a.store.CreateEmailTemplate(r.Context(), store.EmailTemplateInput{
		Slug: req.Slug, Name: req.Name, Category: req.Category, Locale: req.Locale,
		Subject: req.Subject, BodyHTML: req.BodyHTML, IsActive: req.IsActive,
	}, id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "slug_taken")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, tpl)
}

func (a *API) commercialGetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCommercial(w, r); !ok {
		return
	}
	tpl, err := a.store.GetEmailTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, tpl)
}

func (a *API) commercialPatchEmailTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	var req emailTemplateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Category != "" && !store.ValidEmailTemplateCategory(req.Category) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_category")
		return
	}
	tpl, err := a.store.UpdateEmailTemplate(r.Context(), chi.URLParam(r, "id"), store.EmailTemplateInput{
		Name: req.Name, Category: req.Category, Locale: req.Locale,
		Subject: req.Subject, BodyHTML: req.BodyHTML, IsActive: req.IsActive,
	}, id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, tpl)
}

type emailPreviewReq struct {
	ProspectID string            `json:"prospectId"`
	Subject    string            `json:"subject"`
	BodyHTML   string            `json:"bodyHtml"`
	Vars       map[string]string `json:"vars"`
}

func (a *API) commercialPreviewEmailTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	tpl, err := a.store.GetEmailTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var req emailPreviewReq
	_ = httpx.DecodeJSON(r, &req)
	subject := tpl.Subject
	body := tpl.BodyHTML
	if strings.TrimSpace(req.Subject) != "" {
		subject = req.Subject
	}
	if strings.TrimSpace(req.BodyHTML) != "" {
		body = req.BodyHTML
	}
	vars := a.commercialMailVars(r, id, "", req.Vars)
	if pid := strings.TrimSpace(req.ProspectID); pid != "" {
		if _, ok := a.loadProspectForMail(w, r, id, pid); !ok {
			return
		}
		vars = a.commercialMailVars(r, id, pid, req.Vars)
	}
	subject = substituteMailVars(subject, vars)
	body = substituteMailVarsHTML(body, vars)
	full := ""
	if a.notifier != nil {
		full = a.notifier.RenderBrandedHTML(body, email.BrandedHTMLOpts{
			Locale:           tpl.Locale,
			Tagline:          "Continuité de soins prescrite",
			Preheader:        subject,
			UnsubscribeLabel: "Se désinscrire de ces e-mails",
			UnsubscribeURL:   "#",
		})
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"subject":  subject,
		"bodyHtml": body,
		"fullHtml": full,
		"vars":     vars,
	})
}

func (a *API) commercialListEmails(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	filterCommercial := id.UserID
	if id.Role == kernel.RoleCommercialManager {
		if q := strings.TrimSpace(r.URL.Query().Get("commercialUserId")); q != "" {
			okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, q)
			if err != nil || !okTeam {
				writeErr(w, r, http.StatusForbidden, "forbidden", "not_team_member")
				return
			}
			filterCommercial = q
		} else {
			// Manager without filter: own sends only (managers can also prospect).
			filterCommercial = id.UserID
		}
	}
	items, total, err := a.store.ListEmailSends(r.Context(), id.UserID, filterCommercial, limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (a *API) commercialGetEmail(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	send, err := a.store.GetEmailSend(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !a.canAccessCommercialMail(r, id, send.CommercialUserID) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	send.BodyHTMLRendered = scrubTrackingFromHTML(send.BodyHTMLRendered)
	httpx.WriteData(w, http.StatusOK, send)
}

func (a *API) commercialListProspectEmails(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	if !a.canAccessProspectMail(w, r, id, prospectID) {
		return
	}
	items, err := a.store.ListEmailSendsByProspect(r.Context(), prospectID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type sendProspectEmailReq struct {
	TemplateID string `json:"templateId"`
	Subject    string `json:"subject"`
	BodyHTML   string `json:"bodyHtml"`
}

func (a *API) commercialSendProspectEmail(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	if a.notifier == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "mailer_unavailable")
		return
	}
	prospectID := chi.URLParam(r, "id")
	prospect, ok := a.loadProspectForMail(w, r, id, prospectID)
	if !ok {
		return
	}
	if strings.TrimSpace(prospect.ContactEmail) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "contact_email_required")
		return
	}
	if prospect.CommercialUserID == "" {
		writeErr(w, r, http.StatusConflict, "conflict", "prospect_unclaimed")
		return
	}
	if prospect.CommercialUserID != id.UserID && id.Role != kernel.RoleCommercialManager {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if id.Role == kernel.RoleCommercialManager && prospect.CommercialUserID != id.UserID {
		okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, prospect.CommercialUserID)
		if err != nil || !okTeam {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
	}
	if prospect.EmailOptOut {
		writeErr(w, r, http.StatusConflict, "conflict", "email_opt_out")
		return
	}
	count, err := a.store.CountCommercialMailsToday(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if count >= store.CommercialMailDailyLimit {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "daily_limit")
		return
	}

	var req sendProspectEmailReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.TemplateID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "template_required")
		return
	}
	tpl, err := a.store.GetEmailTemplate(r.Context(), req.TemplateID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "template_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !tpl.IsActive {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "template_inactive")
		return
	}

	openToken, err := randomMailToken()
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	apiBase := strings.TrimRight(a.cfg.APIPublicURL, "/")
	unsubURL := apiBase + "/api/v1/public/commercial-mail/unsubscribe/" + openToken
	vars := a.commercialMailVars(r, id, prospectID, map[string]string{
		"unsubscribe_url": unsubURL,
	})
	subject := tpl.Subject
	body := tpl.BodyHTML
	if strings.TrimSpace(req.Subject) != "" {
		subject = req.Subject
	}
	if strings.TrimSpace(req.BodyHTML) != "" {
		body = req.BodyHTML
	}
	subject = substituteMailVars(subject, vars)
	body = substituteMailVarsHTML(body, vars)

	body, clicks := rewriteTrackedLinks(body, apiBase)
	pixel := fmt.Sprintf(`<img src="%s/api/v1/public/commercial-mail/o/%s" width="1" height="1" alt="" style="display:block;width:1px;height:1px;border:0;" />`,
		apiBase, openToken)
	body = body + "\n" + pixel

	sender, _ := a.store.GetUserByID(r.Context(), id.UserID)
	replyTo := strings.TrimSpace(sender.Email)

	opts := email.BrandedHTMLOpts{
		Locale:           tpl.Locale,
		Tagline:          "Continuité de soins prescrite",
		Preheader:        subject,
		UnsubscribeLabel: "Se désinscrire de ces e-mails",
		UnsubscribeURL:   unsubURL,
		ReplyTo:          replyTo,
		SoftFail:         false,
	}
	fullHTML := a.notifier.RenderBrandedHTML(body, opts)
	sendErr := a.notifier.SendHTMLWithReplyTo(prospect.ContactEmail, subject, fullHTML, replyTo, false)

	status := "sent"
	errMsg := ""
	if sendErr != nil {
		status = "failed"
		errMsg = sendErr.Error()
	}

	clickCreates := make([]store.EmailClickCreate, 0, len(clicks))
	for _, c := range clicks {
		clickCreates = append(clickCreates, store.EmailClickCreate{ClickToken: c.Token, TargetURL: c.URL})
	}
	send, err := a.store.CreateEmailSend(r.Context(), store.EmailSendCreate{
		ProspectID:       prospectID,
		TemplateID:       tpl.ID,
		CommercialUserID: id.UserID,
		ToEmail:          prospect.ContactEmail,
		Subject:          subject,
		BodyHTMLRendered: fullHTML,
		Status:           status,
		Error:            errMsg,
		OpenToken:        openToken,
		Clicks:           clickCreates,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if status == "sent" {
		_ = a.store.TouchProspectContacted(r.Context(), prospectID, id.UserID)
		_, _ = a.store.CreateProspectEvent(r.Context(), prospectID, id.UserID, string(store.EventEmailSent),
			"E-mail envoyé : "+subject, map[string]any{"sendId": send.ID, "templateId": tpl.ID})
	}
	// Never return capability tokens to the client.
	for i := range send.Clicks {
		send.Clicks[i].ClickToken = ""
	}
	send.BodyHTMLRendered = scrubTrackingFromHTML(send.BodyHTMLRendered)
	if status == "failed" {
		httpx.WriteData(w, http.StatusBadGateway, send)
		return
	}
	httpx.WriteData(w, http.StatusCreated, send)
}

func (a *API) publicCommercialMailOpen(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	_ = a.store.RecordEmailOpen(r.Context(), token)
	// 1x1 transparent GIF
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
	})
}

func (a *API) publicCommercialMailClick(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	target, err := a.store.RecordEmailClick(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, a.cfg.ProPublicSiteURL, http.StatusFound)
		return
	}
	if !isSafeRedirectURL(target) {
		http.Redirect(w, r, a.cfg.ProPublicSiteURL, http.StatusFound)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (a *API) publicCommercialMailUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	_, err := a.store.SetProspectEmailOptOutByOpenToken(r.Context(), token)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html lang="fr"><body style="font-family:sans-serif;padding:40px;text-align:center;">
<p>Lien invalide ou déjà traité.</p></body></html>`))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html><html lang="fr"><body style="font-family:sans-serif;padding:40px;text-align:center;">
<h1>Désinscription confirmée</h1>
<p>Vous ne recevrez plus d’e-mails commerciaux petsFollow concernant ce cabinet.</p>
</body></html>`))
}

func (a *API) canAccessCommercialMail(r *http.Request, id authx.Identity, commercialUserID string) bool {
	if id.UserID == commercialUserID {
		return true
	}
	if id.Role == kernel.RoleCommercialManager {
		ok, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, commercialUserID)
		return err == nil && ok
	}
	return false
}

func (a *API) canAccessProspectMail(w http.ResponseWriter, r *http.Request, id authx.Identity, prospectID string) bool {
	_, ok := a.loadProspectForMail(w, r, id, prospectID)
	return ok
}

func (a *API) loadProspectForMail(w http.ResponseWriter, r *http.Request, id authx.Identity, prospectID string) (store.Prospect, bool) {
	if id.Role == kernel.RoleCommercialManager {
		p, err := a.store.GetProspectByID(r.Context(), prospectID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
				return store.Prospect{}, false
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return store.Prospect{}, false
		}
		if p.CommercialUserID != "" && p.CommercialUserID != id.UserID {
			okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, p.CommercialUserID)
			if err != nil || !okTeam {
				writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
				return store.Prospect{}, false
			}
		}
		return p, true
	}
	p, err := a.store.GetProspect(r.Context(), id.UserID, prospectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return store.Prospect{}, false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return store.Prospect{}, false
	}
	if p.CommercialUserID != "" && p.CommercialUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return store.Prospect{}, false
	}
	return p, true
}

func (a *API) commercialMailVars(r *http.Request, id authx.Identity, prospectID string, extra map[string]string) map[string]string {
	vars := map[string]string{
		"contact_name":     "",
		"practice_name":    "",
		"city":             "",
		"commercial_name":  "",
		"commercial_phone": "",
		"commercial_email": "",
		"appointment_at":   "",
		"appointment_mode": "",
		"cta_demo_url":     strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/commercial/brochure",
		"unsubscribe_url":  "#",
	}
	user, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err == nil {
		vars["commercial_name"] = user.FullName
		vars["commercial_email"] = user.Email
		vars["commercial_phone"] = user.ContactPhone
	}
	if prospectID != "" {
		p, err := a.store.GetProspectByID(r.Context(), prospectID)
		if err == nil {
			vars["contact_name"] = p.ContactName
			if vars["contact_name"] == "" {
				vars["contact_name"] = p.PracticeName
			}
			vars["practice_name"] = p.PracticeName
			vars["city"] = p.City
			if p.AppointmentAt != nil {
				vars["appointment_at"] = p.AppointmentAt.In(time.Local).Format("02/01/2006 15:04")
			}
			vars["appointment_mode"] = p.AppointmentOutcome
		}
	}
	for k, v := range extra {
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}
	return vars
}

func substituteMailVars(s string, vars map[string]string) string {
	out := s
	for k, v := range vars {
		placeholder := "{{" + k + "}" + "}"
		out = strings.ReplaceAll(out, placeholder, v)
	}
	return out
}

// substituteMailVarsHTML escapes non-URL vars for safe injection into HTML fragments.
func substituteMailVarsHTML(s string, vars map[string]string) string {
	out := s
	for k, v := range vars {
		placeholder := "{{" + k + "}" + "}"
		if strings.HasSuffix(k, "_url") {
			out = strings.ReplaceAll(out, placeholder, v)
			continue
		}
		out = strings.ReplaceAll(out, placeholder, html.EscapeString(v))
	}
	return out
}

// scrubTrackingFromHTML removes capability tracking / unsubscribe URLs from stored HTML
// before returning it to authenticated clients (tokens remain in DB for public endpoints).
func scrubTrackingFromHTML(body string) string {
	re := regexp.MustCompile(`(?i)https?://[^"'>\s]*/api/v1/public/commercial-mail/(?:o|c|unsubscribe)/[a-f0-9]+`)
	return re.ReplaceAllString(body, "#")
}

type trackedLink struct {
	Token string
	URL   string
}

func rewriteTrackedLinks(body, apiBase string) (string, []trackedLink) {
	clicks := make([]trackedLink, 0)
	out := hrefRewriteRE.ReplaceAllStringFunc(body, func(m string) string {
		sub := hrefRewriteRE.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		href := sub[1]
		quote := `"`
		if i := strings.IndexAny(m, `"'`); i >= 0 {
			quote = string(m[i])
		}
		lower := strings.ToLower(href)
		if strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "#") || strings.Contains(lower, "/public/commercial-mail/") {
			return m
		}
		if !isSafeRedirectURL(href) {
			return m
		}
		tok, err := randomMailToken()
		if err != nil {
			return m
		}
		clicks = append(clicks, trackedLink{Token: tok, URL: href})
		tracked := apiBase + "/api/v1/public/commercial-mail/c/" + tok
		return "href=" + quote + tracked + quote
	})
	return out, clicks
}

func isSafeRedirectURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func randomMailToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
