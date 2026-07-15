package email

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

type Notifier struct {
	host string
	port int
	from string
}

func NewNotifier(host string, port int, from string) *Notifier {
	return &Notifier{host: host, port: port, from: from}
}

func (n *Notifier) SendVetAlert(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", n.from, to, subject, body)
	err := smtp.SendMail(addr, nil, n.from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("email send (dev may use mailhog): %v", err)
	}
	return nil
}

func (n *Notifier) SendConfirmRegistration(to, locale, fullName, confirmURL string) error {
	subject := i18n.T(locale, "emails.confirm_registration_subject", nil)
	vars := map[string]string{"fullName": fullName}
	body := renderConfirmRegistration(confirmRegistrationContent{
		Lang:       locale,
		Tagline:    i18n.T(locale, "emails.confirm_registration_tagline", nil),
		Greeting:   i18n.T(locale, "emails.confirm_registration_greeting", vars),
		Intro:      i18n.T(locale, "emails.confirm_registration_intro", nil),
		CTALabel:   i18n.T(locale, "emails.confirm_registration_cta", nil),
		Expiry:     i18n.T(locale, "emails.confirm_registration_expiry", nil),
		Disclaimer: i18n.T(locale, "emails.confirm_registration_disclaimer", nil),
		Preheader:  i18n.T(locale, "emails.confirm_registration_preheader", nil),
		ConfirmURL: confirmURL,
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendPasswordReset(to, locale, fullName, resetURL string) error {
	subject := i18n.T(locale, "emails.password_reset_subject", nil)
	vars := map[string]string{"fullName": fullName}
	body := renderConfirmRegistration(confirmRegistrationContent{
		Lang:       locale,
		Tagline:    i18n.T(locale, "emails.password_reset_tagline", nil),
		Greeting:   i18n.T(locale, "emails.password_reset_greeting", vars),
		Intro:      i18n.T(locale, "emails.password_reset_intro", nil),
		CTALabel:   i18n.T(locale, "emails.password_reset_cta", nil),
		Expiry:     i18n.T(locale, "emails.password_reset_expiry", nil),
		Disclaimer: i18n.T(locale, "emails.password_reset_disclaimer", nil),
		Preheader:  i18n.T(locale, "emails.password_reset_preheader", nil),
		ConfirmURL: resetURL,
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendHeartrateValidated(to, locale string, bpm int) error {
	subject := i18n.T(locale, "emails.heartrate_validated_subject", nil)
	body := i18n.T(locale, "emails.heartrate_validated_body", map[string]string{
		"bpm": fmt.Sprintf("%d", bpm),
	})
	return n.SendVetAlert(to, subject, body)
}

func (n *Notifier) SendNewMessage(to, locale, messageBody string) error {
	subject := i18n.T(locale, "emails.new_message_subject", nil)
	body := i18n.T(locale, "emails.new_message_body", map[string]string{
		"body": messageBody,
	})
	return n.SendVetAlert(to, subject, body)
}
