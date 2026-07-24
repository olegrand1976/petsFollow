# Parcours email découverte & fidélisation client

Parcours drip **12 mois** envoyé aux propriétaires (clients), **en parallèle** du Discovery in-app (cartes J0/J2/J4/J6). Contenu éducatif + upsells contextualisés (Family, Kennel, Care+, Horse, plans).

## Architecture

| Élément | Détail |
|---------|--------|
| Scheduler | Ticker horaire in-process dans l’API (`go/internal/engagement/journey`) |
| Lock | `pg_try_advisory_lock` sur **connexion dédiée** (`WithAdvisoryLock`, clé `824719001`) |
| Tables | `discovery.email_journey`, `discovery.email_sends` (migration `000027`) |
| Prefs | `notifications.client_preferences.discovery` (éducatif/upsell) · `.billing` (paiement) |
| Opt-out | Lien footer → `GET/POST /api/v1/public/journey/unsubscribe?token=…` (coupe **discovery** seulement ; billing inchangé) |
| CTA | `PETS_APP_DOWNLOAD_URL` + UTM `utm_campaign=client_journey` |
| i18n | `emails.journey.*` (FR / NL / EN / ES / ET) |
| Hook Stripe | `billing.Hooks.OnOwnerPastDue` → `Runner.TriggerPastDue` (immédiat) + scan horaire (filet) |

Env :

| Variable | Défaut | Rôle |
|----------|--------|------|
| `JOURNEY_EMAIL_ENABLED` | `true` | Active le runner |
| `JOURNEY_EMAIL_INTERVAL` | `1h` | Intervalle du ticker |
| `PETSFOLLOW_API_PUBLIC_URL` | `http://localhost:8291` | Base URL liens unsubscribe |
| `PETS_APP_DOWNLOAD_URL` | (Firebase App Dist) | CTA app |

## Enrôlement

1. `CreateClientForVet` → `EnrollEmailJourney` (anchor = maintenant)
2. Boot API → `BackfillEmailJourneys` (clients existants, **anchor = NOW()** pour éviter un rattrapage massif d’emails dus)
3. Seed → backfill après enrichment
4. `SendAppDownloadInvite` → marque `d0_welcome` en `skipped` (évite doublon)

## Calendrier (offset depuis `anchor_at`)

| step_key | J+ | Notes |
|----------|----|-------|
| `d0_welcome` | 0 | Bienvenue / télécharger (skip si invite app) |
| `d1_activate` | 1 | Créer animal + plan (skip si déjà un pet) |
| `d2_first_measure` | 2 | Première mesure (skip si HR validé) |
| `d4_routine` | 4 | Routine ; detail Care+ omis si Care+ déjà actif |
| `d6_vet_link` | 6 | Messagerie |
| `d10_visits` | 10 | Visites |
| `d14_checkpoint` | 14 | Bilan inclus |
| `d30_habit` | 30 | Habitude ; detail Family si ≥1 pet sans Family/Kennel |
| `d45_care_plus` | 45 | Upsell Care+ (skip si actif) |
| `d60_horse` | 60 | Upsell Horse (si cheval) |
| `d75_kennel` | 75 | Upsell Kennel (≥6 pets) |
| `d90_quarter` | 90 | Trimestre ; detail Family si 2–5 pets sans pack |
| `d120_seasonal` | 120 | Soins saisonniers |
| `d180_midyear` | 180 | Mi-parcours |
| `d270_reengage` | 270 | Relance si inactif ≥60 j |
| `d330_prerenew` | 330 | Tip long terme ; `intro_near` si annual ≤75 j |
| `d365_anniversary` | 365 | Anniversaire → journey `completed` |

Événements :

| step_key | Condition | Pref | Déclencheur |
|----------|-----------|------|-------------|
| `evt_pending_payment` | pet `pending_payment` ≥3 j | billing | scan horaire |
| `evt_past_due` | entitlement/addon `past_due` | billing | webhook Stripe **+** scan |
| `evt_inactive_hr` | 0 HR validé ≥21 j, après J+14 ; cooldown 90 j | discovery | scan horaire |

## Idempotence & skip

- `(user_id, step_key)` unique dans `email_sends`
- Statut `sent` ou `skipped` (raison dans `meta.reason` : `pref_off`, `has_care_plus`, …)
- Soft upsells en `detail` : omis via `_omitDetail` (Care+ d4, Family d30/d90)
- `evt_inactive_hr` : ré-envoi possible après 90 j (upsert `sent_at`)

## Ops local

```bash
make up-infra && make migrate && make seed && make api-dev
# MailHog UI : http://localhost:8026
# Forcer un step : UPDATE discovery.email_journey SET anchor_at = NOW() - INTERVAL '2 days' WHERE user_id = '...';
# Relancer le runner : redémarrer l’API (RunOnce au boot) ou attendre JOURNEY_EMAIL_INTERVAL
# Test intégration : cd go && go test ./internal/engagement/journey/ -count=1
```

## Hors scope

- Push FCM discovery
- Deep links Flutter natifs
- Emails Care+ J-3/J0 (`ListCarePlusEmailCandidates`)
