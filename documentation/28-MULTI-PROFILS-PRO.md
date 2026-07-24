# 28 — Multi-profils pro & santé

Roadmap produit : self-inscription client, rôle `care_pro` + specialties, ACL partage, Flutter pro light, agenda GPS, CR visite + IA.

## Surfaces

| Rôle / specialty | Surface principale | Notes |
|------------------|--------------------|-------|
| `client` | Flutter (shell owner) | Self-signup `POST /auth/register-client` |
| `vet` | Nuxt Pro (full) | Flutter refuse / message « utilisez Pro web » |
| `care_pro` + specialty | Flutter (shell pro light) | Terrain : agenda, clients, fiche, CR, docs |
| `admin` / commercial* | Nuxt Pro | Inchangé |

Specialties P0 : `vet_light`, `farrier`. P1 : `physio`, `behaviorist` (labels + prompts CR). P2 : `groomer`, `breeder`. Pharmacie : track [27](27-PHARMACIE-BELGIQUE.md).

## ACL

Tables `practice.client_access` et `pets.pet_access` :

- `grantee_user_id`, `permission` (`read` | `write_notes` | `full`), `granted_by_user_id`, `expires_at` optionnel
- Les liens `practice_clients` restent un grant métier (cabinet) ; l’ACL couvre collègue / pro externe / co-owner
- Family billing ≠ multi-comptes (foyer tarifaire distinct de `client_access`)
- Timeline (`GET /pets/{id}/timeline`) : sans `write_notes`, notes de visite masquées ; messages messagerie seulement avec `full` (owner / véto cabinet inclus via ACL)
- `GET /pets` (client) : animaux **owned** + grants `pet_access` / `client_access` (champ `permission`)

## Auth

| Endpoint | Rôle |
|----------|------|
| `POST /auth/register` | Véto (existant) |
| `POST /auth/register-client` | Client self-signup + email confirm |
| `POST /auth/register-care-pro` | Care pro + `specialty` (off par défaut ; `CARE_PRO_PUBLIC_REGISTER=true`) |
| `POST /admin/care-pros` | Admin : créer care_pro vérifié + specialty |
| `POST /vet/clients` + 409 enrichi | Création ; si existe → proposer rattachement |
| `POST /vet/clients/{id}/link` | Rattacher client existant au cabinet |
| `GET/POST/DELETE /pets/{id}/shares` | Partage dossier animal |
| `GET/POST/DELETE /clients/{id}/shares` | Partage fiche client |
| `PATCH /visits/{id}/location` | Adresse / GPS visite |
| `GET/PUT /visits/{id}/report` (+ improve / transcribe / finalize) | Compte rendu + IA |

Admin `/admin/users` : création **client**, **véto**, **care_pro** (spécialités), **commercial**, **commercial_manager**. Pas de création d’`admin` depuis l’UI (compte seed / ops).

## Partage (Nuxt)

- Fiche animal : partager avec collègue (même cabinet puis email) → `pet_access`
- Fiche client : co-accès contacts / liste pets → `client_access`

## Pro light Flutter

Tabs : Agenda · Clients · (drill-down animal / docs / CR) · Settings.

## Agenda GPS

Colonnes visite : `address_text`, `lat`, `lng` — ouverture Maps côté mobile / lien calendrier web.
`PATCH /visits/{id}/location` : `clearCoords` pour invalider lat/lng si l’adresse change sans nouveau GPS.
`PATCH /visits/{id}` (confirm / cancel / done / notes) : seuil **`write_notes`** pour client grantee et `care_pro` (aligné CR / GPS). `confirmDirect` à la création exige `full` hors véto cabinet.
Tournées (Vague O) : agenda Flutter pro light — filtres **Aujourd’hui** / **7 jours** / **Tout** (tri ASC sur les fenêtres courtes ; date = `proposedScheduledAt || scheduledAt || createdAt` ; hors `done`/`cancelled`) ; bouton **Fait** si `confirmed` + `write_notes` ; Nuxt calendrier — badge « Aujourd’hui » via fetch dédié (indépendant de la plage affichée).

## Statut

Plan multi-profils **A→O clos** (care_pro terrain, ACL, GPS/`clearCoords`, tournées, polish notifs/silent-load). Shell Flutter partagé `vet`+`care_pro` : agenda véto via `GET /vet/calendar` (plage), pas le pending-only de `/vet/visits`. Hors scope : messagerie care_pro, Places, register public, P2, monétisation, pharmacie, GCS privé PHI.

## CR visite + IA

Table `visits.visit_reports` (texte, statut draft/final, audio URL optionnelle).

Flux : dictée → upload → transcription Gemini → édition → « améliorer » (sections type SOAP / specialty).
Champs conservés : `transcript_text` (original), `improved_text` (version IA), `body_text` (version éditée / enregistrée) — **historique visualisable** côté Web Pro (`/calendar`) et Flutter Pro Light.
Échec Gemini / transcription vide → `502 gemini_error` / `transcription_failed` (pas de faux succès).
Audio CR : **pas** servi via `/media/` public (`visit-reports/` bloqué en local via `DenySensitivePrefixes`) ; stream auth `GET /visits/{id}/report/audio` (refusé si CR `final`) ; **suppression à la finalisation** (Clear DB seulement après Delete média OK). Sur GCS, pas d’URL publique retournée pour ce prefix (GCP refuse une condition IAM sur `allUsers`).
Consentement UI avant upload micro/fichier (Flutter + Nuxt).
Ops : `make gcp-setup-media` — IAM SA + allUsers pour avatars. Protection PHI = pas d’URL publique à l’upload + stream auth + purge finalize (+ UUID dans la clé).

## Billing

B2B2C inchangé (client paie l’animal). Pros light gratuits au MVP. Partage d’un dossier : animal avec entitlement actif côté owner.

## Polish (vague M / N)

- `GET /pets` inclut grants `pet_access` / `client_access` (UI Flutter labels selon permission)
- Google client : create-if-absent ; plus d’erreur `google_client_not_found`
- Race email unique Google : récupération via `23505` + link
- DEFAULT SQL ACL = `write_notes` (aligné Grant API) — migr. `000042`
- Seed : `farrier.demo@` / `vetlight.demo@` (`CareProDemo123!`) + grant Spirit
- Tests : permissions ACL, `IsSensitiveObjectKey`, shares intégration, smoke register-client + media 403

## Migrations

`000039_multi_profiles_pro` — rôle `care_pro`, specialty, ACL, GPS visites, `visit_reports`.  
`000042_access_permission_default` — DEFAULT permission `write_notes`.
