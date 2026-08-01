# 28 — Multi-profils pro & santé

Roadmap produit : self-inscription client, rôle `care_pro` + specialties, ACL partage, Flutter pro light, agenda GPS, CR visite + IA.

Statut multi-profil compte : `identity.profiles` + switch Flutter/Web ; inscription pro → profil `client` auto ; attach Admin/Commercial hors soi.

Équipe cabinet : `practice.team_members` + `/team` (véto de référence).

**Poste partagé (PC bureau)** — distinct du multi-profil même compte :
- Header VetPro : avatars de l’équipe (`GET /vet/team`), clic → re-auth mot de passe (+ 2FA si actif).
- Idle configurable (défaut **2 min**) → veille : cookies httpOnly purgés via BFF ; roster + `lastPath` en localStorage uniquement (pas de JWT). **Désactivé si l’équipe n’a qu’un seul compte** (pas de poste partagé). Paramétrable par le véto de référence dans `/settings` (`deskIdleMinutes` : 1|2|5|10|15|30).
- Switch profil : même purge immédiate des cookies (pas de session active derrière le modal / autre onglet) ; « Annuler » → veille (y compris si la purge logout est encore en cours).
- Header : avatars équipe uniquement si **≥ 2** membres.
- Au déverrouillage / switch : restauration de la dernière route de l’utilisateur cible.
- Voir UC-EQ-02.

## Surfaces

| Rôle / specialty | Surface principale | Notes |
|------------------|--------------------|-------|
| `client` | Flutter (shell owner) | Self-signup `POST /auth/register-client` |
| `vet` | Nuxt Pro (full) | Flutter : shell pro light (terrain — agenda via `GET /vet/calendar`) |
| `care_pro` + specialty | Flutter (shell pro light) | Terrain : agenda, clients, fiche, CR, docs, **Messages** |
| `admin` / commercial* | Nuxt Pro | **Seed démo** multi-switch (`EnsureDemoMultiSwitchProfiles`) : voir matrice ci-dessous (pas d’auto-profil client pour tout admin via `IsProRole`) |
| `dev` | Nuxt Admin (ops léger) | Support IT : users / tickets / flags — pas billing/sales/seed · UC-AD-02 |
| `research` | Nuxt `/research` | Observatoire épidémio anonymisé (tag `dev`) — associable à `vet` / `admin` · doc [42](42-RESEARCH.md) · seed `research.demo` **+** profil `research` sur `vet.demo` / `admin.demo` (switch) |

### Matrice switch (profil home)

`POST /me/profiles/switch` + UI topbar (`CanActivateProfile`) — basée sur le profil **home** (plus ancien hors `client`), pas le rôle actif :

| Home | Peut activer |
|------|----------------|
| `admin` | tout profil possédé |
| `commercial_manager` | tout sauf `admin` |
| `commercial` | tout sauf `admin` et `commercial_manager` |
| autre | ownership seul |

Seed démo Pro (hors `client`, masqué sur Nuxt) :

| Compte | Profils |
|--------|---------|
| `admin.demo` | admin, dev, research, vet, vet_assistant, secretary, commercial, commercial_manager (+ client) |
| `commercial.manager` | commercial_manager, commercial, dev, research, vet, vet_assistant, secretary (+ client) |
| `commercial.demo` / `demo2` | commercial, dev, research, vet, vet_assistant, secretary (+ client) |

Staff cabinet : une ligne `team_members` VetPlus ; au switch staff, `team_role` est aligné sur le profil actif (pas de demote `reference_vet`).

Specialties supportées : `vet_light`, `farrier`, `physio`, `behaviorist`, `groomer`, `breeder` (labels Flutter 6 langues). Pharmacie : track [27](27-PHARMACIE-BELGIQUE.md).

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

Shell **identique** pour `care_pro` (toutes specialties) et staff cabinet (`vet` / assistant / secretary) :

Tabs : Agenda · Clients · Animaux · **Messages** · Settings.

AppBar : logo petsFollow uniquement (pas de sous-titre specialty / « Pro terrain »).

Messagerie care_pro : threads person-scoped (`practice_id` NULL) via ACL ; le care_pro initie (compose client) ; le client répond dans le même fil.

## Agenda GPS

Colonnes visite : `address_text`, `lat`, `lng` — ouverture Maps côté mobile / lien calendrier web.
`PATCH /visits/{id}/location` : `clearCoords` pour invalider lat/lng si l’adresse change sans nouveau GPS.
`PATCH /visits/{id}` (confirm / cancel / done / notes) : seuil **`write_notes`** pour client grantee et `care_pro` (aligné CR / GPS). `confirmDirect` à la création exige `full` hors véto cabinet.
Tournées (Vague O) : agenda Flutter pro light — filtres **Aujourd’hui** / **7 jours** / **Tout** (tri ASC sur les fenêtres courtes ; date = `proposedScheduledAt || scheduledAt || createdAt` ; hors `done`/`cancelled`) ; bouton **Fait** si `confirmed` + `write_notes` ; Nuxt calendrier — badge « Aujourd’hui » via fetch dédié (indépendant de la plage affichée).

## Statut

Plan multi-profils **A→O clos** (care_pro terrain, ACL, GPS/`clearCoords`, tournées, polish notifs/silent-load). Shell Flutter partagé `vet`+`care_pro` : agenda véto via `GET /vet/calendar` (plage), pas le pending-only de `/vet/visits`. Messagerie Pro Light : **staff + care_pro** (threads care_pro person-scoped). Hors scope : Places, register public care_pro, P2, monétisation, pharmacie, GCS privé PHI, compose client→care_pro depuis zéro (le care_pro initie).

## CR visite + IA

Table `visits.visit_reports` (texte, statut draft/final, audio URL optionnelle, `client_audio_consent_at`).

Flux Flutter Pro Light : **accord oral client (checkbox)** → dictée (bandeau micro + Arrêter) / fichier audio → transcription Gemini → édition → **Enregistrer** / **Finaliser**. L’action « Améliorer (IA) » reste sur **Web Pro** uniquement.

**Web Pro** (`ProVisitReportPanel`) — layout split friendly :

| Zone | Rôle |
|------|------|
| Gauche | Notes / dictée (`transcript_text`) — écrire, dicter, upload audio |
| Droite | Compte-rendu (`body_text`) — Améliorer (IA) depuis la gauche, édition manuelle, preview markdown |
| Bas | Historique replié : transcription d’origine → proposition IA → dernière version enregistrée |
| Footer sticky | Annuler (discard dirty) · Finaliser · Enregistrer |

`PUT /visits/{id}/report` accepte `bodyText` et optionnellement `transcriptText` (notes gauche persistées). `POST …/report/improve` accepte optionnellement `sourceText` (source IA sans écraser le body au préalable).

Flux Web : édition notes → « améliorer » (sections structurées) → **finalisation = validation exclusive du pro**.

**Entitlement module** (add-on VetPro, essai 90 j, 39 € HT/mois ou 390 € HT/an) : `transcribe` / `improve` gated par `practice.ai_cr_modules` — voir `documentation/32-MODULE-IA-CR.md`. CR manuel sans IA reste possible.

Sections CR vétérinaire (improve) :
- Anamnèse / motif · Examen clinique · Observations · **Diagnostic proposé** · **Médication proposée** · Plan / suivi
- Pays d’exercice : `practice.practices.country_code` (défaut `BE`) injecté dans le prompt (DCI / dénominations locales ; pas de prescription auto)
- Care_pro : templates specialty (farrier/physio/…) sans section médication véto

Champs conservés : `transcript_text` (original), `improved_text` (version IA), `body_text` (version éditée / enregistrée), `is_reference` (consultation de référence pour amélioration continue, CR final uniquement) — **historique visualisable** côté Web Pro (`/calendar`, modal consultation, dossier) et Flutter Pro Light.
Web Pro liste aussi tous les CR d’une visite (`GET /visits/{id}/reports`) pour lire le CR d’un auteur terrain (lecture seule) tout en éditant le sien.
Échec Gemini / transcription vide → `502 gemini_error` / `transcription_failed` (pas de faux succès).
`POST .../report/transcribe` exige `clientAudioConsent=true` sinon `400 audio_consent_required`.
Audio CR : **pas** servi via `/media/` public (`visit-reports/` bloqué en local via `DenySensitivePrefixes`) ; stream auth `GET /visits/{id}/report/audio` (refusé si CR `final`) ; **suppression à la finalisation** (Clear DB seulement après Delete média OK). Sur GCS, pas d’URL publique retournée pour ce prefix (GCP refuse une condition IAM sur `allUsers`).
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
`000062_cr_country_audio_consent` — `practices.country_code` + `visit_reports.client_audio_consent_at`.  
`000064_ai_cr_module` — entitlement CR IA (trial/usage/feedback/friction).
