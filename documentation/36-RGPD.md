# 36 — RGPD (technique & conformité produit)

Référence technique petsFollow pour les données personnelles, les droits, et le partage Web Pro (Nuxt) ↔ app Flutter. **Pas** un registre DPO juridique formel — à croiser avec le conseil légal pour la version contractuelle.

## 1. Architecture des données

```
Flutter ── JWT (flutter_secure_storage) ──┐
                                          ├──► API Go (:8291) ──► Postgres
Nuxt BFF ── cookies httpOnly (pf_token /  ┘         │
            pf_refresh) + pf_session                └──► GCS (médias / PHI)
```

- **Une seule source de vérité** : API Go + Postgres (et objets GCS).
- **Pas de sync de session** ni de magasin partagé navigateur ↔ app.
- Chaque face authentifie indépendamment contre la même `identity.users`.

### Canaux hors API (pas une réplication de base)

| Canal | Contenu | TTL / garde-fous |
|-------|---------|------------------|
| Email confirm / reset | liens vers pages Nuxt | tokens one-shot |
| Partage dossier / consultation | lien public + ZIP PHI | 24 h, rate-limit, quota, purge rétention + lifecycle GCS 2 j |
| Deep links `petsfollow://` | confirm, invite, preconsult, payment | pas de PHI dans l’URL |
| FCM | titres / navigation | tokens appareil ; pré-dialogue consent Flutter |

## 2. Responsable & sous-traitants (synthèse)

| Acteur | Rôle |
|--------|------|
| LL-IT-SC / petsFollow | Responsable de traitement (offre SaaS + app client) |
| Google Cloud (Cloud Run, Cloud SQL, GCS) | Hébergement UE (config projet) |
| Stripe | Paiements abonnements client / SaaS |
| Firebase Cloud Messaging | Push mobile |
| Google Gemini | Amélioration CR (audio temps réel — voir privacy i18n) |
| Billit (si module facturation activé) | Peppol / factures cabinet — voir `33-BILLIT-INTEGRATION.md` |

Aligner périodiquement les pages légales (`nuxtjs/locales/*/legal.privacy`, Flutter l10n) avec cette liste.

## 3. Registre synthétique des traitements

| Traitement | Finalité | Base (produit) | Conservation | Faces |
|------------|----------|----------------|--------------|-------|
| Compte identité | Auth, profil | Contrat / intérêt légitime ; CGU à l’inscription ou 1re activation | Jusqu’à suppression ; purge 3 ans d’inactivité | Nuxt, Flutter, API |
| Dossier animal / relevés / messagerie | Continuité de soins | Contrat | Idem compte client (purge) ; CR cabinet conservés après tombstone pro | Flutter, Pro |
| Facturation / entitlements | Abonnement | Contrat | Stripe + tables billing ; cancel à l’effacement client | Flutter, API |
| Support tickets | Assistance | Intérêt légitime | Anonymisés sur `DELETE /me` | Pro, Flutter |
| Partage dossier PHI | Envoi ponctuel à un pro | **Consentement explicite** UI | 24 h + purge | Flutter → email → Nuxt public |
| Audio CR | Compte rendu | Consentement oral horodaté | Tant que le CR ; purge audio avant finalize si échec | Pro, Pro Light |
| FCM tokens | Notifications | Consentement éclairé (pré-dialogue) | Jusqu’à logout / delete device / tombstone | Flutter |
| Clients provisionnés / import | Onboarding cabinet | Exécution contrat cabinet ; **CGU acceptées à l’activation** (`POST /me/accept-terms`) | Idem compte | Pro crée → Flutter active |

## 4. Droits (chemins produit)

| Droit | Couverture | Chemin |
|-------|------------|--------|
| Transparence | Oui | `/legal/privacy`, Flutter legal in-app |
| Accès / portabilité | Oui (JSON) | `GET /api/v1/me/export` — Pro `/settings` + `/commercial/settings` ; Flutter profil |
| Effacement | Oui | `DELETE /api/v1/me` — client = purge ; pro = tombstone (données cliniques cabinet conservées) |
| Rectification | Partiel | `PATCH /me` (nom, téléphone commercial), locale, mot de passe |
| Restriction / opposition | Support | `support@ll-it-sc.be` / tickets — pas d’UI dédiée |
| Consentement inscription | Oui | `"consent": true` → `terms_accepted_at` |
| Consentement activation (provisionné) | Oui | `POST /api/v1/me/accept-terms` + gate Flutter si `termsAcceptedAt` null ; **filet API** clients (`consent_required` hors allowlist me/password/export/delete) |

### Décisions volontaires

- **Tombstone Pro** : email `deleted+{id}@deleted.petsfollow.invalid`, secrets wipe ; CR / visites cabinet **conservés** (intégrité dossier).
- **Dual profil** (`identity.profiles`) : avant tombstone pro, purge des données **client-owned** (pets, threads, etc.) pour éviter les orphelins art. 17.
- **PHI streams** (`visit-reports/`, `health-books/`, …) : pas d’URL publique `/media/` — stream auth uniquement.

## 5. Auth & cookies (rappel)

- Nuxt : JWT **httpOnly** uniquement (`pf_token`, `pf_refresh`) ; marqueur `pf_session` lisible. Exception documentée : JWT court pour WebSocket via `GET /api/auth/ws-token`.
- Flutter : `flutter_secure_storage` uniquement.
- Endpoints auth publics : rate limit `authRL.Middleware`.
- Secrets jobs internes : `secretHeaderOK` (temps constant).

## 6. Checklist ops

- [ ] `RETENTION_PURGE_SECRET` câblé + `make gcp-retention-scheduler` + redeploy API
- [ ] Lifecycle GCS `dossier-shares/**` (et consultation-shares) 2 jours — `make gcp-setup-media`
- [ ] Privacy / CGU à jour (6 locales Nuxt + Flutter arb)
- [ ] Smoke / tests : `go/internal/handlers/rgpd_integration_test.go`
- [ ] Pas de Google Fonts CDN (CSP + fonts auto-hébergées)

## 7. Références code

| Sujet | Fichier |
|-------|---------|
| Export | `go/internal/store/user_export.go` |
| Purge / tombstone | `go/internal/store/user_profile.go`, `handlers/me_profile.go`, `handlers/retention.go` |
| Accept terms | `POST /api/v1/me/accept-terms` |
| Invariants agents | `.cursor/rules/securite-auth-rgpd.mdc` |
| Plan tests | `documentation/15-PLAN-TESTS.md` (F12, F15, F16…) |

Dernière actualisation : juillet 2026.
