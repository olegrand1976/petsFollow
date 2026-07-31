package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

const smtpOpTimeout = 15 * time.Second

const defaultLLITWebsiteURL = "https://ll-it-sc.be"

type Notifier struct {
	host           string
	port           int
	from           string
	user           string
	pass           string
	publicSiteURL  string
	llitWebsiteURL string
}

func NewNotifier(host string, port int, from, publicSiteURL, llitWebsiteURL string) *Notifier {
	return NewNotifierAuth(host, port, from, "", "", publicSiteURL, llitWebsiteURL)
}

// NewNotifierAuth is like NewNotifier with optional SMTP PLAIN credentials (OVH :587).
func NewNotifierAuth(host string, port int, from, user, pass, publicSiteURL, llitWebsiteURL string) *Notifier {
	if strings.TrimSpace(llitWebsiteURL) == "" {
		llitWebsiteURL = defaultLLITWebsiteURL
	}
	return &Notifier{
		host:           host,
		port:           port,
		from:           from,
		user:           strings.TrimSpace(user),
		pass:           pass,
		publicSiteURL:  strings.TrimRight(publicSiteURL, "/"),
		llitWebsiteURL: strings.TrimRight(llitWebsiteURL, "/"),
	}
}

// envelopeFrom extracts the bare RFC5322 addr-spec for SMTP MAIL FROM.
// Display-form values like `petsFollow <noreply@petsfollow.app>` are rejected by
// many MTAs (OVH returns 501 5.1.7 Invalid address).
func envelopeFrom(from string) string {
	from = strings.TrimSpace(from)
	if i := strings.LastIndex(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 1 {
			addr := strings.TrimSpace(from[i+1 : i+j])
			if addr != "" {
				return addr
			}
		}
	}
	return from
}

// isDevSMTP — MailHog / localhost / sans auth : soft-fail (tests + local).
func (n *Notifier) isDevSMTP() bool {
	if strings.TrimSpace(n.user) == "" {
		return true
	}
	h := strings.ToLower(strings.TrimSpace(n.host))
	return h == "localhost" || h == "127.0.0.1" || h == "mailhog"
}

func (n *Notifier) sendHTML(to, subject, body string, softFail bool) error {
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		n.from, to, encodedSubject, body)
	return n.deliverSMTP(to, []byte(msg), softFail)
}

// deliverSMTP dials SMTP with shared auth / soft-fail / logging policy.
func (n *Notifier) deliverSMTP(to string, msg []byte, softFail bool) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	mailFrom := envelopeFrom(n.from)

	// USER set without PASS (staging secret missing) — fail fast, do not dial OVH.
	if strings.TrimSpace(n.user) != "" && strings.TrimSpace(n.pass) == "" {
		err := fmt.Errorf("smtp: SMTP_PASS empty for user %s", n.user)
		log.Printf("email send failed to=%s from=%s: %v", to, mailFrom, err)
		if softFail || n.isDevSMTP() {
			return nil
		}
		return err
	}

	var authUser, authPass string
	if n.user != "" {
		authUser, authPass = n.user, n.pass
	}
	if err := sendMailTimeout(addr, authUser, authPass, mailFrom, []string{to}, msg, smtpOpTimeout); err != nil {
		if n.isDevSMTP() {
			log.Printf("email send (mailhog/dev): %v", err)
		} else {
			log.Printf("email send failed to=%s from=%s: %v", to, mailFrom, err)
		}
		if softFail || n.isDevSMTP() {
			return nil
		}
		return err
	}
	return nil
}

// loginAuth implements SMTP AUTH LOGIN (OVH advertises LOGIN, not PLAIN, after STARTTLS).
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("unencrypted connection")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	prompt := strings.ToLower(strings.TrimSpace(string(fromServer)))
	switch {
	case strings.Contains(prompt, "user"):
		return []byte(a.username), nil
	case strings.Contains(prompt, "pass"):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected AUTH LOGIN challenge %q", string(fromServer))
	}
}

func pickSMTPAuth(host, user, pass, mechanisms string) (smtp.Auth, error) {
	mechs := strings.ToUpper(mechanisms)
	switch {
	case strings.Contains(mechs, "LOGIN"):
		return LoginAuth(user, pass), nil
	case strings.Contains(mechs, "PLAIN"):
		return smtp.PlainAuth("", user, pass, host), nil
	default:
		return nil, fmt.Errorf("smtp: no supported AUTH in %q", mechanisms)
	}
}

// sendMailTimeout is smtp.SendMail with a dial/deadline so auth paths cannot hang E2E.
func sendMailTimeout(addr, user, pass, from string, to []string, msg []byte, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	host, _, _ := net.SplitHostPort(addr)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	// OVH :587 requires STARTTLS before AUTH — PlainAuth returns
	// "unencrypted connection" otherwise (same as net/smtp.SendMail).
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if user != "" {
		ok, mechanisms := c.Extension("AUTH")
		if !ok {
			return fmt.Errorf("smtp: AUTH not advertised")
		}
		a, aerr := pickSMTPAuth(host, user, pass, mechanisms)
		if aerr != nil {
			return aerr
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func (n *Notifier) SendVetAlert(to, subject, body string) error {
	return n.sendHTML(to, subject, body, true)
}

// SendVetAlertWithCSV sends an HTML alert plus a CSV attachment (reorder / supplier order).
// Unlike soft alert digests, SMTP failures are returned (except mailhog/dev soft-fail)
// so callers can avoid marking a purchase order as sent when delivery failed.
func (n *Notifier) SendVetAlertWithCSV(to, subject, htmlBody, filename string, csv []byte) error {
	return n.sendHTMLWithCSV(to, subject, htmlBody, filename, csv, false)
}

func (n *Notifier) sendHTMLWithCSV(to, subject, htmlBody, filename string, csv []byte, softFail bool) error {
	if strings.TrimSpace(filename) == "" {
		filename = "order.csv"
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	hw, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return err
	}
	if err := writeBase64Wrapped(hw, []byte(htmlBody)); err != nil {
		return err
	}
	aw, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/csv; charset=UTF-8"},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Disposition":       {fmt.Sprintf(`attachment; filename="%s"`, filename)},
	})
	if err != nil {
		return err
	}
	if err := writeBase64Wrapped(aw, csv); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", n.from, to, encodedSubject)
	fmt.Fprintf(&msg, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", mw.Boundary())
	msg.Write(body.Bytes())
	return n.deliverSMTP(to, msg.Bytes(), softFail)
}

// writeBase64Wrapped encodes data as RFC 2045 base64 (≤76 cols + CRLF).
func writeBase64Wrapped(w io.Writer, data []byte) error {
	enc := base64.NewEncoder(base64.StdEncoding, &base64LineBreaker{w: w, max: 76})
	if _, err := enc.Write(data); err != nil {
		_ = enc.Close()
		return err
	}
	return enc.Close()
}

// base64LineBreaker inserts CRLF every max bytes (RFC 2045 / SMTP line limits).
type base64LineBreaker struct {
	w   io.Writer
	col int
	max int
}

func (b *base64LineBreaker) Write(p []byte) (int, error) {
	n := 0
	for len(p) > 0 {
		if b.col >= b.max {
			if _, err := b.w.Write([]byte("\r\n")); err != nil {
				return n, err
			}
			b.col = 0
		}
		space := b.max - b.col
		chunk := p
		if len(chunk) > space {
			chunk = p[:space]
		}
		wn, err := b.w.Write(chunk)
		n += wn
		b.col += wn
		p = p[wn:]
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// SendCritical delivers an ops/alert email and returns SMTP errors (no soft-fail in prod).
func (n *Notifier) SendCritical(to, subject, body string) error {
	return n.sendHTML(to, subject, body, false)
}

func (n *Notifier) brandURLs() brandAssets {
	site := n.publicSiteURL
	if site == "" {
		site = "https://petsfollow.ll-it-sc.be"
	}
	return brandAssets{
		LLITLogoURL:    site + "/brand/ll-it-logo.png",
		LLITWebsiteURL: n.llitWebsiteURL,
		SiteURL:        site,
	}
}

func (n *Notifier) SendConfirmRegistration(to, locale, fullName, confirmURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{"fullName": fullName}
	subject := i18n.T(locale, "emails.confirm_registration_subject", nil)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.confirm_registration_tagline"),
		Greeting:        mustT(locale, "emails.confirm_registration_greeting", vars),
		Intro:           mustT(locale, "emails.confirm_registration_intro"),
		CTALabel:        mustT(locale, "emails.confirm_registration_cta"),
		CTAURL:          confirmURL,
		Expiry:          mustT(locale, "emails.confirm_registration_expiry"),
		Disclaimer:      mustT(locale, "emails.confirm_registration_disclaimer"),
		Preheader:       mustT(locale, "emails.confirm_registration_preheader"),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	// Confirm registration is critical in staging/prod (auth path); soft-fail only on dev SMTP.
	return n.sendHTML(to, subject, body, false)
}

func (n *Notifier) SendPasswordReset(to, locale, fullName, resetURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{"fullName": fullName}
	subject := mustT(locale, "emails.password_reset_subject")
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.password_reset_tagline"),
		Greeting:        mustT(locale, "emails.password_reset_greeting", vars),
		Intro:           mustT(locale, "emails.password_reset_intro"),
		CTALabel:        mustT(locale, "emails.password_reset_cta"),
		CTAURL:          resetURL,
		Expiry:          mustT(locale, "emails.password_reset_expiry"),
		Disclaimer:      mustT(locale, "emails.password_reset_disclaimer"),
		Preheader:       mustT(locale, "emails.password_reset_preheader"),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendHeartrateValidated(to, locale string, bpm int) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{"bpm": fmt.Sprintf("%d", bpm)}
	subject := mustT(locale, "emails.heartrate_validated_subject")
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.heartrate_validated_tagline"),
		Greeting:        mustT(locale, "emails.heartrate_validated_greeting"),
		Intro:           mustT(locale, "emails.heartrate_validated_intro", vars),
		Disclaimer:      mustT(locale, "emails.heartrate_validated_disclaimer"),
		Preheader:       mustT(locale, "emails.heartrate_validated_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendHeartrateThresholdAlert notifies the vet that a validated reading rose by at least the species BPM delta.
func (n *Notifier) SendHeartrateThresholdAlert(to, locale string, bpm int) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{"bpm": fmt.Sprintf("%d", bpm)}
	subject := mustT(locale, "emails.heartrate_alert_subject")
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.heartrate_alert_tagline"),
		Greeting:        mustT(locale, "emails.heartrate_alert_greeting"),
		Intro:           mustT(locale, "emails.heartrate_alert_intro", vars),
		Disclaimer:      mustT(locale, "emails.heartrate_alert_disclaimer"),
		Preheader:       mustT(locale, "emails.heartrate_alert_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendNewMessage(to, locale, messageBody string) error {
	locale = i18n.NormalizeLocale(locale)
	subject := mustT(locale, "emails.new_message_subject")
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.new_message_tagline"),
		Greeting:        mustT(locale, "emails.new_message_greeting"),
		Intro:           mustT(locale, "emails.new_message_intro"),
		Detail:          messageBody,
		Disclaimer:      mustT(locale, "emails.new_message_disclaimer"),
		Preheader:       mustT(locale, "emails.new_message_preheader"),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendVisitRequest(to, locale, clientName, petName, when, notes, ctaURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"clientName": clientName,
		"petName":    petName,
		"when":       when,
		"notes":      notes,
	}
	if vars["when"] == "" {
		vars["when"] = "—"
	}
	if vars["notes"] == "" {
		vars["notes"] = "—"
	}
	subject := mustT(locale, "emails.visit_request_subject", vars)
	detail := mustT(locale, "emails.visit_request_detail", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.visit_request_tagline"),
		Greeting:        mustT(locale, "emails.visit_request_greeting"),
		Intro:           mustT(locale, "emails.visit_request_intro", vars),
		Detail:          detail,
		CTALabel:        mustT(locale, "emails.visit_request_cta"),
		CTAURL:          ctaURL,
		Disclaimer:      mustT(locale, "emails.visit_request_disclaimer"),
		Preheader:       mustT(locale, "emails.visit_request_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendPetDossierShare emails a pro with a 24h download link for a client pet dossier.
func (n *Notifier) SendPetDossierShare(
	to, locale, petName, clientName, downloadURL,
	commercialName, commercialPhone, commercialEmail, registerURL, siteURL string,
) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"petName":         petName,
		"clientName":      clientName,
		"commercialName":  commercialName,
		"commercialPhone": commercialPhone,
		"commercialEmail": commercialEmail,
		"registerUrl":     registerURL,
		"siteUrl":         siteURL,
	}
	if vars["commercialName"] == "" {
		vars["commercialName"] = "petsFollow"
	}
	detail := mustT(locale, "emails.pet_dossier_share_detail", vars)
	if commercialPhone != "" {
		detail += "\n" + mustT(locale, "emails.pet_dossier_share_phone", vars)
	}
	if registerURL != "" {
		detail += "\n" + mustT(locale, "emails.pet_dossier_share_register", vars)
	}
	subject := mustT(locale, "emails.pet_dossier_share_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.pet_dossier_share_tagline"),
		Greeting:        mustT(locale, "emails.pet_dossier_share_greeting"),
		Intro:           mustT(locale, "emails.pet_dossier_share_intro", vars),
		Detail:          detail,
		CTALabel:        mustT(locale, "emails.pet_dossier_share_cta"),
		CTAURL:          downloadURL,
		Expiry:          mustT(locale, "emails.pet_dossier_share_expiry"),
		Disclaimer:      mustT(locale, "emails.pet_dossier_share_disclaimer"),
		Preheader:       mustT(locale, "emails.pet_dossier_share_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	// Critical: destinataire hors plateforme — échec SMTP doit remonter (soft-fail seulement en dev/MailHog).
	return n.SendCritical(to, subject, body)
}

// SendConsultationShare emails a pro with a 24h download link for a consultation PDF.
func (n *Notifier) SendConsultationShare(
	to, locale, petName, clientName, downloadURL,
	commercialName, commercialPhone, commercialEmail, registerURL, siteURL string,
) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"petName":         petName,
		"clientName":      clientName,
		"commercialName":  commercialName,
		"commercialPhone": commercialPhone,
		"commercialEmail": commercialEmail,
		"registerUrl":     registerURL,
		"siteUrl":         siteURL,
	}
	if vars["commercialName"] == "" {
		vars["commercialName"] = "petsFollow"
	}
	detail := mustT(locale, "emails.consultation_share_detail", vars)
	if commercialPhone != "" {
		detail += "\n" + mustT(locale, "emails.consultation_share_phone", vars)
	}
	if registerURL != "" {
		detail += "\n" + mustT(locale, "emails.consultation_share_register", vars)
	}
	subject := mustT(locale, "emails.consultation_share_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.consultation_share_tagline"),
		Greeting:        mustT(locale, "emails.consultation_share_greeting"),
		Intro:           mustT(locale, "emails.consultation_share_intro", vars),
		Detail:          detail,
		CTALabel:        mustT(locale, "emails.consultation_share_cta"),
		CTAURL:          downloadURL,
		Expiry:          mustT(locale, "emails.consultation_share_expiry"),
		Disclaimer:      mustT(locale, "emails.consultation_share_disclaimer"),
		Preheader:       mustT(locale, "emails.consultation_share_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendCritical(to, subject, body)
}

func (n *Notifier) SendAppDownloadInvite(to, locale, clientName, vetName, practiceName, downloadURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName":     clientName,
		"vetName":      vetName,
		"practiceName": practiceName,
	}
	subject := mustT(locale, "emails.app_download_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.app_download_tagline"),
		Greeting:        mustT(locale, "emails.app_download_greeting", vars),
		Intro:           mustT(locale, "emails.app_download_intro", vars),
		CTALabel:        mustT(locale, "emails.app_download_cta"),
		CTAURL:          downloadURL,
		Disclaimer:      mustT(locale, "emails.app_download_disclaimer"),
		Preheader:       mustT(locale, "emails.app_download_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendProductDigest sends the daily functional platform changelog to internal staff.
func (n *Notifier) SendProductDigest(to, locale, fullName, dateLabel, headline, bodyText string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName": fullName,
		"date":     dateLabel,
		"headline": headline,
	}
	if vars["fullName"] == "" {
		vars["fullName"] = mustT(locale, "emails.product_digest_fallback_name")
	}
	subject := mustT(locale, "emails.product_digest_subject", vars)
	intro := mustT(locale, "emails.product_digest_intro", vars)
	if headline != "" {
		intro = mustT(locale, "emails.product_digest_intro_with_headline", vars)
	}
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.product_digest_tagline"),
		Greeting:        mustT(locale, "emails.product_digest_greeting", vars),
		Intro:           intro,
		Detail:          bodyText,
		Disclaimer:      mustT(locale, "emails.product_digest_disclaimer"),
		Preheader:       mustT(locale, "emails.product_digest_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendStagingSeedNotice informs internal staff that staging demo data was reset (manual admin action).
func (n *Notifier) SendStagingSeedNotice(to, locale, fullName, siteURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName": fullName,
	}
	if vars["fullName"] == "" {
		vars["fullName"] = mustT(locale, "emails.staging_seed_fallback_name")
	}
	subject := mustT(locale, "emails.staging_seed_subject")
	ctaURL := strings.TrimRight(siteURL, "/")
	if ctaURL == "" {
		ctaURL = strings.TrimRight(n.publicSiteURL, "/")
	}
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.staging_seed_tagline"),
		Greeting:        mustT(locale, "emails.staging_seed_greeting", vars),
		Intro:           mustT(locale, "emails.staging_seed_intro", vars),
		Detail:          mustT(locale, "emails.staging_seed_detail", vars),
		CTALabel:        mustT(locale, "emails.staging_seed_cta"),
		CTAURL:          ctaURL,
		Disclaimer:      mustT(locale, "emails.staging_seed_disclaimer"),
		Preheader:       mustT(locale, "emails.staging_seed_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendSalesBranchCongrats congratulates a commercial on their newly auto-created sales branch.
func (n *Notifier) SendSalesBranchCongrats(to, locale, fullName, branchName, branchCode, ctaURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName":   fullName,
		"branchName": branchName,
		"branchCode": branchCode,
	}
	if vars["fullName"] == "" {
		vars["fullName"] = mustT(locale, "emails.sales_branch_congrats_fallback_name")
	}
	subject := mustT(locale, "emails.sales_branch_congrats_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         mustT(locale, "emails.sales_branch_congrats_tagline"),
		Greeting:        mustT(locale, "emails.sales_branch_congrats_greeting", vars),
		Intro:           mustT(locale, "emails.sales_branch_congrats_intro", vars),
		Detail:          mustT(locale, "emails.sales_branch_congrats_detail", vars),
		CTALabel:        mustT(locale, "emails.sales_branch_congrats_cta"),
		CTAURL:          strings.TrimRight(ctaURL, "/"),
		Disclaimer:      mustT(locale, "emails.sales_branch_congrats_disclaimer"),
		Preheader:       mustT(locale, "emails.sales_branch_congrats_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendVisitPreconsult invites the client to fill the pre-consult form (and download the app if needed).
func (n *Notifier) SendVisitPreconsult(to, locale, clientName, petName, when, practiceName, ctaURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName":     clientName,
		"petName":      petName,
		"when":         when,
		"practiceName": practiceName,
	}
	if vars["when"] == "" {
		vars["when"] = "—"
	}
	if vars["practiceName"] == "" {
		vars["practiceName"] = "petsFollow"
	}
	if vars["fullName"] == "" {
		vars["fullName"] = vars["petName"]
	}
	subject := mustT(locale, "emails.visit_preconsult_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.visit_preconsult_tagline"),
		Greeting:        mustT(locale, "emails.visit_preconsult_greeting", vars),
		Intro:           mustT(locale, "emails.visit_preconsult_intro", vars),
		Detail:          mustT(locale, "emails.visit_preconsult_detail", vars),
		CTALabel:        mustT(locale, "emails.visit_preconsult_cta"),
		CTAURL:          ctaURL,
		Disclaimer:      mustT(locale, "emails.visit_preconsult_disclaimer"),
		Preheader:       mustT(locale, "emails.visit_preconsult_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendVisitConfirmedAffiliate confirms the visit and invites the client to affiliate via app invite.
func (n *Notifier) SendVisitConfirmedAffiliate(to, locale, clientName, petName, when, practiceName, ctaURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{
		"fullName":     clientName,
		"petName":      petName,
		"when":         when,
		"practiceName": practiceName,
	}
	if vars["when"] == "" {
		vars["when"] = "—"
	}
	if vars["practiceName"] == "" {
		vars["practiceName"] = "petsFollow"
	}
	if vars["fullName"] == "" {
		vars["fullName"] = vars["petName"]
	}
	subject := mustT(locale, "emails.visit_confirmed_affiliate_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.visit_confirmed_affiliate_tagline"),
		Greeting:        mustT(locale, "emails.visit_confirmed_affiliate_greeting", vars),
		Intro:           mustT(locale, "emails.visit_confirmed_affiliate_intro", vars),
		Detail:          mustT(locale, "emails.visit_confirmed_affiliate_detail", vars),
		CTALabel:        mustT(locale, "emails.visit_confirmed_affiliate_cta"),
		CTAURL:          ctaURL,
		Disclaimer:      mustT(locale, "emails.visit_confirmed_affiliate_disclaimer"),
		Preheader:       mustT(locale, "emails.visit_confirmed_affiliate_preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendAiCrAdhesionStep sends one VetPro CR IA trial drip email (branded + i18n).
func (n *Notifier) SendAiCrAdhesionStep(to, locale, fullName, stepKey, ctaURL string, vars map[string]string) error {
	locale = i18n.NormalizeLocale(locale)
	if vars == nil {
		vars = map[string]string{}
	}
	if _, ok := vars["fullName"]; !ok {
		vars["fullName"] = fullName
	}
	if vars["fullName"] == "" {
		vars["fullName"] = mustT(locale, "emails.ai_cr_adhesion.fallback_name")
	}
	prefix := "emails.ai_cr_adhesion." + stepKey + "."
	subject := mustT(locale, prefix+"subject", vars)
	detail := mustT(locale, prefix+"detail", vars)
	if detail == prefix+"detail" {
		detail = ""
	}
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow Pro",
		Tagline:         mustT(locale, prefix+"tagline", vars),
		Greeting:        mustT(locale, prefix+"greeting", vars),
		Intro:           mustT(locale, prefix+"intro", vars),
		Detail:          detail,
		CTALabel:        mustT(locale, prefix+"cta", vars),
		CTAURL:          ctaURL,
		Disclaimer:      mustT(locale, prefix+"disclaimer", vars),
		Preheader:       mustT(locale, prefix+"preheader", vars),
		Brand:           n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendJourneyStep sends one client discovery/loyalty drip email.
// vars may include:
//
//	"_omitDetail=1" — suppress soft-upsell detail block
//	"_introNear=1"  — use intro_near when present (d330 annual near renewal)
func (n *Notifier) SendJourneyStep(to, locale, fullName, stepKey, ctaURL, unsubscribeURL string, vars map[string]string) error {
	locale = i18n.NormalizeLocale(locale)
	if vars == nil {
		vars = map[string]string{}
	}
	if _, ok := vars["fullName"]; !ok {
		vars["fullName"] = fullName
	}
	omitDetail := vars["_omitDetail"] == "1"
	introNear := vars["_introNear"] == "1"
	delete(vars, "_omitDetail")
	delete(vars, "_introNear")
	prefix := "emails.journey." + stepKey + "."
	subject := mustT(locale, prefix+"subject", vars)
	detailKey := prefix + "detail"
	detail := mustT(locale, detailKey, vars)
	if detail == detailKey || omitDetail {
		detail = ""
	}
	introKey := prefix + "intro"
	if introNear {
		nearKey := prefix + "intro_near"
		if near := mustT(locale, nearKey, vars); near != nearKey {
			introKey = nearKey
		}
	}
	body := renderBrandedEmail(brandedEmailContent{
		Lang:             locale,
		ProductLabel:     "petsFollow",
		Tagline:          mustT(locale, prefix+"tagline", vars),
		Greeting:         mustT(locale, prefix+"greeting", vars),
		Intro:            mustT(locale, introKey, vars),
		Detail:           detail,
		CTALabel:         mustT(locale, prefix+"cta", vars),
		CTAURL:           ctaURL,
		Disclaimer:       mustT(locale, prefix+"disclaimer", vars),
		Preheader:        mustT(locale, prefix+"preheader", vars),
		Brand:            n.brandURLs(),
		FooterPoweredBy:  mustT(locale, "emails.footer_powered_by"),
		FooterVisit:      mustT(locale, "emails.footer_visit_llit"),
		UnsubscribeLabel: mustT(locale, "emails.journey.unsubscribe"),
		UnsubscribeURL:   unsubscribeURL,
	})
	return n.SendVetAlert(to, subject, body)
}

// SendSupportTicketOps notifies the support inbox of a new bug-report ticket.
func (n *Notifier) SendSupportTicketOps(to, locale, ticketID, subjectLine, fullName, emailAddr, role, source, message, adminURL string) error {
	if strings.TrimSpace(to) == "" {
		return nil
	}
	vars := map[string]string{
		"ticketId": ticketID,
		"subject":  subjectLine,
		"fullName": fullName,
		"email":    emailAddr,
		"role":     role,
		"source":   source,
		"message":  message,
	}
	subject := mustT(locale, "emails.support_ticket_ops_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.support_ticket_ops_tagline"),
		Preheader:       mustT(locale, "emails.support_ticket_ops_preheader", vars),
		Greeting:        mustT(locale, "emails.support_ticket_ops_greeting"),
		Intro:           mustT(locale, "emails.support_ticket_ops_intro", vars),
		Detail:          mustT(locale, "emails.support_ticket_ops_detail", vars),
		CTALabel:        mustT(locale, "emails.support_ticket_ops_cta"),
		CTAURL:          adminURL,
		Disclaimer:      mustT(locale, "emails.support_ticket_ops_disclaimer"),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
		Brand:           n.brandURLs(),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendSupportTicketReply notifies the ticket creator of an admin reply.
func (n *Notifier) SendSupportTicketReply(to, locale, fullName, subjectLine, replyBody, siteURL string) error {
	if strings.TrimSpace(to) == "" {
		return nil
	}
	vars := map[string]string{
		"fullName":  fullName,
		"subject":   subjectLine,
		"replyBody": replyBody,
	}
	subject := mustT(locale, "emails.support_ticket_reply_subject", vars)
	body := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		ProductLabel:    "petsFollow",
		Tagline:         mustT(locale, "emails.support_ticket_reply_tagline"),
		Preheader:       mustT(locale, "emails.support_ticket_reply_preheader", vars),
		Greeting:        mustT(locale, "emails.support_ticket_reply_greeting", vars),
		Intro:           mustT(locale, "emails.support_ticket_reply_intro", vars),
		Detail:          mustT(locale, "emails.support_ticket_reply_detail", vars),
		CTALabel:        mustT(locale, "emails.support_ticket_reply_cta"),
		CTAURL:          siteURL,
		Disclaimer:      mustT(locale, "emails.support_ticket_reply_disclaimer"),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
		Brand:           n.brandURLs(),
	})
	return n.SendVetAlert(to, subject, body)
}

// SendVetLeadNotify alerts ops/commercial that a client suggested a vet not on the platform.
func (n *Notifier) SendVetLeadNotify(to, clientName, clientEmail, vetEmail, vetPhone, vetName, practiceName string) error {
	if strings.TrimSpace(to) == "" {
		return nil
	}
	subject := "Nouveau véto suggéré (app client)"
	esc := func(s string) string {
		s = strings.ReplaceAll(s, "&", "&amp;")
		s = strings.ReplaceAll(s, "<", "&lt;")
		s = strings.ReplaceAll(s, ">", "&gt;")
		return s
	}
	if practiceName == "" {
		practiceName = "—"
	}
	if vetName == "" {
		vetName = "—"
	}
	body := fmt.Sprintf(`<p>Un client a suggéré un vétérinaire absent de petsFollow.</p>
<ul>
<li><strong>Client</strong> : %s (%s)</li>
<li><strong>Véto</strong> : %s</li>
<li><strong>Cabinet</strong> : %s</li>
<li><strong>Email</strong> : %s</li>
<li><strong>Téléphone</strong> : %s</li>
</ul>`,
		esc(clientName), esc(clientEmail), esc(vetName), esc(practiceName), esc(vetEmail), esc(vetPhone))
	return n.SendVetAlert(to, subject, body)
}

// mustT returns the translation or a clear fallback that still identifies the key in tests/logs.
func mustT(locale, key string, varsList ...map[string]string) string {
	var vars map[string]string
	if len(varsList) > 0 {
		vars = varsList[0]
	}
	msg := i18n.T(locale, key, vars)
	if msg == key {
		// Fall back to French catalog explicitly once more (Normalize already did);
		// keep visible signal only if FR also missing.
		if fr := i18n.T("fr", key, vars); fr != key {
			return fr
		}
	}
	return msg
}
