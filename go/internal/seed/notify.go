package seed

import (
	"context"
	"log"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// NotifyStaff emails admin/commercial/commercial_manager (+ ops) after a staging seed.
// Skips *.petsfollow.test. Returns the number of successful sends and the first SMTP error (if any).
func NotifyStaff(ctx context.Context, st *store.Store, notifier *email.Notifier, siteURL, opsEmail string) (int, error) {
	if st == nil || notifier == nil {
		return 0, nil
	}
	recipients, err := st.ListDigestRecipients(ctx)
	if err != nil {
		return 0, err
	}
	siteURL = strings.TrimRight(siteURL, "/")
	sent := map[string]struct{}{}
	var firstErr error
	sendOne := func(to, locale, fullName string) {
		to = strings.TrimSpace(to)
		if to == "" || strings.HasSuffix(strings.ToLower(to), "@petsfollow.test") {
			return
		}
		key := strings.ToLower(to)
		if _, ok := sent[key]; ok {
			return
		}
		if err := notifier.SendStagingSeedNotice(to, locale, fullName, siteURL); err != nil {
			log.Printf("seed-notify: %s: %v", to, err)
			if firstErr == nil {
				firstErr = err
			}
			return
		}
		sent[key] = struct{}{}
	}
	for _, r := range recipients {
		sendOne(r.Email, i18n.NormalizeLocale(r.PreferredLocale), r.FullName)
	}
	if ops := strings.TrimSpace(opsEmail); ops != "" {
		sendOne(ops, "fr", "")
	}
	log.Printf("seed-notify: %d recipient(s)", len(sent))
	return len(sent), firstErr
}
