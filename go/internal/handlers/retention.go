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

// secretHeaderOK compare un secret de header interne en temps constant.
func secretHeaderOK(r *http.Request, header, secret string) bool {
	return secret != "" &&
		subtle.ConstantTimeCompare([]byte(r.Header.Get(header)), []byte(secret)) == 1
}

// internalRunRetentionPurge — job cron (RGPD) : purge les comptes inactifs depuis 3 ans.
// Clients : effacement complet (DB + médias + abonnements). Pros : anonymisation.
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
		case string(kernel.RoleVet), string(kernel.RoleCommercial), string(kernel.RoleCommercialManager), string(kernel.RoleCarePro):
			if err := a.anonymizeProAccount(r.Context(), acct.ID); err != nil {
				failed++
				fmt.Printf("retention purge: pro %s failed: %v\n", acct.ID, err)
				continue
			}
			anonymizedPros++
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"candidates":     len(accounts),
		"purgedClients":  purgedClients,
		"anonymizedPros": anonymizedPros,
		"failed":         failed,
		"cutoff":         cutoff,
	})
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
	return nil
}

// anonymizeProAccount — même anonymisation que DELETE /me côté Pro.
func (a *API) anonymizeProAccount(ctx context.Context, userID string) error {
	var avatarURL string
	if u, err := a.store.GetUserByID(ctx, userID); err == nil {
		avatarURL = u.AvatarURL
	}
	if err := a.store.DeleteProAccount(ctx, userID); err != nil {
		return err
	}
	if k := media.ObjectKeyFromURL(a.cfg, avatarURL); k != "" {
		a.purgeMediaObjects(ctx, []string{k})
	}
	return nil
}
