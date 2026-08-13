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
| Google Gemini | Amélioration CR Pro (audio temps réel) + Client AI tag `dev` (vulgarisation CR finalisés + triage conversationnel — voir privacy i18n / [`43-CLIENT-AI.md`](43-CLIENT-AI.md)) |
| Billit (si module facturation activé) | Peppol / factures cabinet — voir `33-BILLIT-INTEGRATION.md` |
| Orthanc PACS (si `PACS_ENABLED`) | Index Cloud SQL `orthanc` + DICOM GCS — voir `40-PACS.md` |

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
| Research (agrégats anonymisés) | Observatoire épidémio / stats | Opt-in cabinet + finalité distincte (à valider juridiquement) — [42](42-RESEARCH.md) | Purge ciblée à l’opt-out (`practice_id_hash`) | Rôle `research` ; **hors** `GET /me/export` |

## 4. Droits (chemins produit)

| Droit | Couverture | Chemin |
|-------|------------|--------|
| Transparence | Oui | `/legal/privacy`, Flutter legal in-app |
| Accès / portabilité | Oui (JSON) | `GET /api/v1/me/export` — inclut `imagingStudies` (métadonnées PACS) + `eidReadings` (audit eID BE hashé, [46](46-EID-BELGIQUE.md)) ; Pro `/settings` + `/commercial/settings` ; Flutter profil |
| Effacement | Oui | `DELETE /api/v1/me` — client = purge (pets CASCADE `imaging.pet_studies` + delete Orthanc studies best-effort) ; pro = tombstone |
| Rectification | Partiel | `PATCH /me` (nom, téléphone commercial), locale, mot de passe |
| Restriction / opposition | Support | `support@ll-it-sc.be` / tickets — pas d’UI dédiée |
| Consentement inscription | Oui | `"consent": true` → `terms_accepted_at` |
| Consentement activation (provisionné) | Oui | `POST /api/v1/me/accept-terms` + gate Flutter si `termsAcceptedAt` null ; **filet API** clients (`consent_required` hors allowlist me/password/export/delete) |

### Décisions volontaires

- **Tombstone Pro** : email `deleted+{id}@deleted.petsfollow.invalid`, secrets wipe ; CR / visites cabinet **conservés** (intégrité dossier).
- **Dual profil** (`identity.profiles`) : avant tombstone pro, purge des données **client-owned** (pets, threads, etc.) pour éviter les orphelins art. 17.
- **PHI streams** (`visit-reports/`, `health-books/`, …) : pas d’URL publique `/media/` — stream auth uniquement.

### Stockage média — allowlist fail-closed

Le bucket `petsfollow-media` est en lecture publique globale (binding `allUsers`, `infra/gcp/setup-gcs-media.sh`). La séparation PHI / public repose donc entièrement sur `IsSensitiveObjectKey` (`go/internal/platform/media/media.go`), qui fonctionne en **allowlist** : seuls `avatars/`, `pets/`, `messages/` et `brand/` reçoivent une URL publique. Tout autre namespace — connu ou futur — est traité comme PHI et n’est servi que par un stream authentifié. Une nouvelle catégorie d’upload est donc privée par défaut, jamais exposée par omission.

| Namespace | Accès |
|---|---|
| `documents/` (analyses, radios, courriers) | `GET /pets/{petID}/documents/{documentID}/download` (auth + ACL fiche animal) |
| `visit-reports/`, `health-books/`, `consultation-shares-v2/`, `pitch-sims/`, … | stream authentifié dédié |
| `avatars/`, `pets/`, `messages/`, `brand/` | URL publique assumée |

`pets.documents.file_url` n’est plus alimenté. Les objets écrits avant ce correctif ont eu leur URL publiée : leur clé est considérée compromise et doit être tournée une fois par environnement via `make rotate-pet-documents` (recopie sous une nouvelle clé UUID, purge de l’ancienne, vidage de `file_url` ; `--dry-run` disponible, ré-exécution sans effet).

**Reste à faire** : retirer le binding `allUsers` et passer les namespaces publics en URLs signées. C’est le correctif de fond ; il casse avatars / photos / pièces jointes tant que les URLs signées ne sont pas en place, donc hors de ce lot.

## 5. Auth & cookies (rappel)

- Nuxt : JWT **httpOnly** uniquement (`pf_token`, `pf_refresh`) ; marqueur `pf_session` lisible. Exception documentée : JWT court pour WebSocket via `GET /api/auth/ws-token`.
- Flutter : `flutter_secure_storage` uniquement.
- Endpoints auth publics : rate limit `authRL.Middleware` — `POST /auth/refresh` compris.
- Secrets jobs internes : `secretHeaderOK` (temps constant).

### Révocation des tokens (`token_version`)

`identity.users.token_version` (migration `000160`) est porté par chaque JWT dans le claim `tv` et comparé **au refresh uniquement** — aucun accès base ajouté par requête. Il est incrémenté par :

- `POST /api/v1/auth/logout` (BFF `logout.post.ts` / `logout-redirect.get.ts`, `ApiClient.logout()` Flutter) ;
- le reset de mot de passe, dans la même transaction que le nouveau hash.

Deux conséquences assumées : un access token déjà émis reste valide jusqu’à son expiration (~15 min), et le bump vaut pour **tout le compte** — se déconnecter du Web Pro ferme aussi la session Flutter du même utilisateur. Si cette UX doit changer, l’alternative est de réserver le bump au reset de mot de passe et à la suppression de compte, et d’exposer une action « Déconnecter tous mes appareils » dans `/settings`.

**Au déploiement** : les tokens émis avant la migration ne portent pas de claim `tv` (lu à 0) alors que la colonne vaut 1 par défaut. Toutes les sessions actives sont donc refusées au premier refresh et chaque utilisateur doit se reconnecter une fois — web comme Flutter. Le seul moyen de l’éviter serait d’accepter `tv == 0` comme legacy pendant la durée de vie des refresh tokens (30 jours), au prix d’une fenêtre où les anciens tokens restent irrévocables.

`PATCH /me/password` bumpe aussi : changer son mot de passe ferme les sessions des **autres** appareils, ce qui est le geste attendu quand on soupçonne une compromission. L’appareil qui en fait la demande est préservé — le handler lui réémet une paire de tokens, absorbée en cookies httpOnly par la BFF (`me/password.patch.ts`) et persistée côté Flutter (`ApiClient.changePassword`). Sans cette réémission, changer son mot de passe reviendrait à se déconnecter soi-même.

Si la réémission échoue (base ou clé de signature indisponibles), le mot de passe a **quand même** changé : la réponse porte `reauthRequired: true` et l’échec est loggé côté API. La BFF purge alors les cookies et Flutter coupe la session par `_invalidateSessionFromUnauthorized`, pour renvoyer au login immédiatement plutôt que de laisser tomber l’utilisateur au refresh suivant, sans rapport apparent avec ce qu’il vient de faire.

## 6. Checklist ops

- [ ] `RETENTION_PURGE_SECRET` câblé + `make gcp-retention-scheduler` + redeploy API
- [ ] Lifecycle GCS `dossier-shares/**` (et consultation-shares) 2 jours — `make gcp-setup-media`
- [ ] Privacy / CGU à jour (6 locales Nuxt + Flutter arb)
- [ ] Smoke / tests : `go/internal/handlers/rgpd_integration_test.go`, `security_hardening_integration_test.go`
- [ ] `make rotate-pet-documents` passé une fois par environnement (staging + prod)
- [ ] Pas de Google Fonts CDN (CSP + fonts auto-hébergées)

## 7. Références code

| Sujet | Fichier |
|-------|---------|
| Export | `go/internal/store/user_export.go` |
| Purge / tombstone | `go/internal/store/user_profile.go`, `handlers/me_profile.go`, `handlers/retention.go` |
| Accept terms | `POST /api/v1/me/accept-terms` |
| Allowlist média | `go/internal/platform/media/media.go` (`IsSensitiveObjectKey`) |
| Rotation clés documents | `go/internal/app/rotate_pet_documents.go` — `make rotate-pet-documents` |
| Révocation JWT | `go/internal/platform/authx/authx.go`, `store/auth.go` (`BumpTokenVersion`), `handlers/auth.go` |
| Invariants agents | `.cursor/rules/securite-auth-rgpd.mdc` |
| Plan tests | `documentation/15-PLAN-TESTS.md` (F12, F15, F16, S1–S9…) |

Dernière actualisation : août 2026.
