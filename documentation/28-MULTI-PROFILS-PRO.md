# 28 — Multi-profils pro & santé

Roadmap produit : self-inscription client, rôle `care_pro` + specialties, ACL partage, Flutter pro light, agenda GPS, CR visite + IA.

## Surfaces

| Rôle / specialty | Surface principale | Notes |
|------------------|--------------------|-------|
| `client` | Flutter (shell owner) | Self-signup `POST /auth/register-client` |
| `vet` | Nuxt Pro (full) | Flutter refuse / message « utilisez Pro web » |
| `care_pro` + specialty | Flutter (shell pro light) | Terrain : agenda, clients, fiche, CR, docs |
| `admin` / commercial* | Nuxt Pro | Inchangé |

Specialties P0 : `vet_light`, `farrier`. P1 : `physio`, `behaviorist`. P2 : `groomer`, `breeder`. Pharmacie : track [27](27-PHARMACIE-BELGIQUE.md).

## ACL

Tables `practice.client_access` et `pets.pet_access` :

- `grantee_user_id`, `permission` (`read` | `write_notes` | `full`), `granted_by_user_id`, `expires_at` optionnel
- Les liens `practice_clients` restent un grant métier (cabinet) ; l’ACL couvre collègue / pro externe / co-owner
- Family billing ≠ multi-comptes (foyer tarifaire distinct de `client_access`)
- Timeline (`GET /pets/{id}/timeline`) : sans `write_notes`, pas de messages messagerie ; notes de visite masquées (meta statut conservée)

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

## CR visite + IA

Table `visits.visit_reports` (texte, statut draft/final, audio URL optionnelle).

Flux : dictée → upload → transcription Gemini → édition → « améliorer » (sections type SOAP).
Échec Gemini / transcription vide → `502 gemini_error` / `transcription_failed` (pas de faux succès).

RGPD : consentement + rétention audio à définir avant prod ; durée de rétention audio recommandée ≤ 30 jours après finalisation du CR.

## Billing

B2B2C inchangé (client paie l’animal). Pros light gratuits au MVP. Partage d’un dossier : animal avec entitlement actif côté owner.

## Migrations

`000039_multi_profiles_pro` — rôle `care_pro`, specialty, ACL, GPS visites, `visit_reports`.
