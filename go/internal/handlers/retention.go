package handlers

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const retentionInactivity = 3 * 365 * 24 * time.Hour // « 3 ans d'inactivité » des textes légaux
// Min age before a walk-in without CR can be cancelled. Effective delay is this age
// plus the next daily retention cron (03:30 — infra/gcp/setup-retention-scheduler.sh).
const consultationOrphanMaxAge = 6 * time.Hour

// secretHeaderOK compare un secret de header interne en temps constant.
func secretHeaderOK(r *http.Request, header, secret string) bool {
	return secret != "" &&
		subtle.ConstantTimeCompare([]byte(r.Header.Get(header)), []byte(secret)) == 1
}

// internalRunRetentionPurge — job cron (RGPD) : purge les comptes inactifs depuis 3 ans.
// Clients : effacement complet (DB + médias + abonnements). Pros : anonymisation.
// Les tables pharmacy.* (dont stock_movements) ne sont PAS purgées ici — conservation
// registres typique 5 ans (UE 2019/6) ; preuve cabinet : GET …/movements/retention-stats.
// UPDATE/DELETE stock_movements révoqués pour petsfollow_app (migration 000172) ; seul
// pharmacy.rgpd_null_stock_movement_created_by nullifie created_by à l’anonymisation pro.
// Protégé par le header X-Retention-Secret (env RETENTION_PURGE_SECRET).
func (a *API) internalRunRetentionPurge(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Retention-Secret", a.cfg.RetentionPurgeSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	cutoff := time.Now().Add(-retentionInactivity)
	accounts, err := a.store.ListInactiveAccounts(r.Context(), cutoff, 100)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	purgedClients, anonymizedPros, failed := 0, 0, 0
	for _, acct := range accounts {
		switch acct.Role {
		case string(kernel.RoleClient):
			if err := a.purgeClientAccount(r.Context(), acct.ID); err != nil {
				failed++
				fmt.Printf("retention purge: client %s failed: %v\n", acct.ID, err)
				continue
			}
			purgedClients++
		case string(kernel.RoleVet), string(kernel.RoleVetAssistant), string(kernel.RoleSecretary),
			string(kernel.RoleCommercial), string(kernel.RoleCommercialManager), string(kernel.RoleCarePro):
			if err := a.anonymizeProAccount(r.Context(), acct.ID); err != nil {
				failed++
				fmt.Printf("retention purge: pro %s failed: %v\n", acct.ID, err)
				continue
			}
			anonymizedPros++
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"candidates":                  len(accounts),
		"purgedClients":               purgedClients,
		"anonymizedPros":              anonymizedPros,
		"failed":                      failed,
		"cutoff":                      cutoff,
		"purgedDossierShares":         a.purgeExpiredDossierShares(r.Context(), ""),
		"purgedConsultationShares":    a.purgeExpiredConsultationShares(r.Context(), ""),
		"purgedWebhookEvents":         a.purgeOldInvoicingWebhooks(r.Context()),
		"rejectedStaleSending":        a.rejectStaleInvoicingSending(r.Context()),
		"cancelledStaleConsultations": a.cancelStaleConsultationOrphans(r.Context()),
		"purgedSmsLogs":               a.purgeOldSmsLogs(r.Context()),
		"purgedSmsInbound":            a.purgeOldSmsInbound(r.Context()),
	})
}

const invoicingWebhookRetention = 90 * 24 * time.Hour
const invoicingSendingStale = 7 * 24 * time.Hour

// smsLogRetention — minimisation : le journal SMS (numéros) ne sert qu'à l'audit
// d'envoi et à l'idempotence des rappels, purgé bien avant les 3 ans du compte.
const smsLogRetention = 365 * 24 * time.Hour

func (a *API) purgeOldSmsLogs(ctx context.Context) int {
	n, err := a.store.PurgeOldSmsLogs(ctx, time.Now().Add(-smsLogRetention), 500)
	if err != nil {
		fmt.Printf("retention purge: sms logs failed: %v\n", err)
		return 0
	}
	return n
}

func (a *API) purgeOldSmsInbound(ctx context.Context) int {
	n, err := a.store.PurgeOldSmsInbound(ctx, time.Now().Add(-smsLogRetention), 500)
	if err != nil {
		fmt.Printf("retention purge: sms inbound failed: %v\n", err)
		return 0
	}
	return n
}

func (a *API) purgeOldInvoicingWebhooks(ctx context.Context) int {
	n, err := a.store.PurgeOldWebhookEvents(ctx, time.Now().Add(-invoicingWebhookRetention), 500)
	if err != nil {
		fmt.Printf("retention purge: invoicing webhooks failed: %v\n", err)
		return 0
	}
	return n
}

func (a *API) rejectStaleInvoicingSending(ctx context.Context) int {
	n, err := a.store.MarkStaleSendingDocuments(ctx, time.Now().Add(-invoicingSendingStale), 200)
	if err != nil {
		fmt.Printf("retention purge: stale invoicing sending failed: %v\n", err)
		return 0
	}
	return n
}

func (a *API) cancelStaleConsultationOrphans(ctx context.Context) int {
	n, err := a.store.CancelStaleConsultationOrphans(ctx, time.Now().Add(-consultationOrphanMaxAge), 200)
	if err != nil {
		fmt.Printf("retention purge: stale consultations failed: %v\n", err)
		return 0
	}
	return n
}

// purgeExpiredDossierShares supprime les partages de dossier périmés (lignes + ZIP
// en bucket) et renvoie le nombre de lignes traitées. ownerUserID vide = tous.
func (a *API) purgeExpiredDossierShares(ctx context.Context, ownerUserID string) int {
	keys, deleted, err := a.store.PurgeExpiredDossierShares(ctx, ownerUserID, 200)
	if err != nil {
		fmt.Printf("retention purge: dossier shares failed: %v\n", err)
		return 0
	}
	a.purgeMediaObjects(ctx, keys)
	return deleted
}

// purgeClientAccount — même effacement que DELETE /me côté client.
func (a *API) purgeClientAccount(ctx context.Context, userID string) error {
	artifacts, artErr := a.store.CollectClientAccountArtifacts(ctx, userID)
	if artErr != nil {
		fmt.Printf("retention purge: collect artifacts for %s failed: %v\n", userID, artErr)
	}
	if err := a.store.DeleteClientAccount(ctx, userID); err != nil {
		return err
	}
	a.billing.CancelUserSubscriptions(ctx, artifacts.SubscriptionIDs)
	keys := artifacts.MediaObjectKeys
	for _, u := range artifacts.MediaURLs {
		if k := media.ObjectKeyFromURL(a.cfg, u); k != "" {
			keys = append(keys, k)
		}
	}
	a.purgeMediaObjects(ctx, keys)
	a.purgeOrthancStudies(ctx, artifacts.OrthancStudyIDs)
	return nil
}

// anonymizeProAccount — même anonymisation que DELETE /me côté Pro
// (y compris purge client-owned dual profil + médias / Stripe best-effort).
func (a *API) anonymizeProAccount(ctx context.Context, userID string) error {
	artifacts, artErr := a.store.CollectClientAccountArtifacts(ctx, userID)
	if artErr != nil {
		fmt.Printf("retention purge: collect pro artifacts for %s failed: %v\n", userID, artErr)
	}
	if err := a.store.DeleteProAccount(ctx, userID); err != nil {
		return err
	}
	a.billing.CancelUserSubscriptions(ctx, artifacts.SubscriptionIDs)
	keys := artifacts.MediaObjectKeys
	for _, u := range artifacts.MediaURLs {
		if k := media.ObjectKeyFromURL(a.cfg, u); k != "" {
			keys = append(keys, k)
		}
	}
	a.purgeMediaObjects(ctx, keys)
	a.purgeOrthancStudies(ctx, artifacts.OrthancStudyIDs)
	return nil
}

func (a *API) purgeOrthancStudies(ctx context.Context, studyIDs []string) {
	if len(studyIDs) == 0 {
		return
	}
	// Purge PHI Orthanc even when the produit flag is off (UI gated separately).
	client := a.orthanc()
	if client == nil {
		return
	}
	for _, id := range studyIDs {
		if err := client.deleteStudy(ctx, id); err != nil {
			fmt.Printf("retention purge: orthanc study %s: %v\n", id, err)
		}
	}
}
