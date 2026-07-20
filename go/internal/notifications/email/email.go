package email

import (
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

const defaultLLITWebsiteURL = "https://ll-it-sc.be"

type Notifier struct {
	host           string
	port           int
	from           string
	publicSiteURL  string
	llitWebsiteURL string
}

func NewNotifier(host string, port int, from, publicSiteURL, llitWebsiteURL string) *Notifier {
	if strings.TrimSpace(llitWebsiteURL) == "" {
		llitWebsiteURL = defaultLLITWebsiteURL
	}
	return &Notifier{
		host:           host,
		port:           port,
		from:           from,
		publicSiteURL:  strings.TrimRight(publicSiteURL, "/"),
		llitWebsiteURL: strings.TrimRight(llitWebsiteURL, "/"),
	}
}

func (n *Notifier) SendVetAlert(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		n.from, to, encodedSubject, body)
	if err := smtp.SendMail(addr, nil, n.from, []string{to}, []byte(msg)); err != nil {
		log.Printf("email send (dev may use mailhog): %v", err)
		// Soft-fail for local MailHog / SMTP outages — callers treat email as best-effort.
		return nil
	}
	return nil
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
		Lang:       locale,
		Tagline:    mustT(locale, "emails.confirm_registration_tagline"),
		Greeting:   mustT(locale, "emails.confirm_registration_greeting", vars),
		Intro:      mustT(locale, "emails.confirm_registration_intro"),
		CTALabel:   mustT(locale, "emails.confirm_registration_cta"),
		CTAURL:     confirmURL,
		Expiry:     mustT(locale, "emails.confirm_registration_expiry"),
		Disclaimer: mustT(locale, "emails.confirm_registration_disclaimer"),
		Preheader:  mustT(locale, "emails.confirm_registration_preheader"),
		Brand:      n.brandURLs(),
		FooterPoweredBy: mustT(locale, "emails.footer_powered_by"),
		FooterVisit:     mustT(locale, "emails.footer_visit_llit"),
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendPasswordReset(to, locale, fullName, resetURL string) error {
	locale = i18n.NormalizeLocale(locale)
	vars := map[string]string{"fullName": fullName}
	subject := mustT(locale, "emails.password_reset_subject")
	body := renderBrandedEmail(brandedEmailContent{
		Lang:       locale,
		Tagline:    mustT(locale, "emails.password_reset_tagline"),
		Greeting:   mustT(locale, "emails.password_reset_greeting", vars),
		Intro:      mustT(locale, "emails.password_reset_intro"),
		CTALabel:   mustT(locale, "emails.password_reset_cta"),
		CTAURL:     resetURL,
		Expiry:     mustT(locale, "emails.password_reset_expiry"),
		Disclaimer: mustT(locale, "emails.password_reset_disclaimer"),
		Preheader:  mustT(locale, "emails.password_reset_preheader"),
		Brand:      n.brandURLs(),
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
		Lang:       locale,
		Tagline:    mustT(locale, "emails.heartrate_validated_tagline"),
		Greeting:   mustT(locale, "emails.heartrate_validated_greeting"),
		Intro:      mustT(locale, "emails.heartrate_validated_intro", vars),
		Disclaimer: mustT(locale, "emails.heartrate_validated_disclaimer"),
		Preheader:  mustT(locale, "emails.heartrate_validated_preheader", vars),
		Brand:      n.brandURLs(),
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
