# Plan de phases — petsFollow

Synthèse historique / état — alignée [01-VISION-PRODUIT.md](01-VISION-PRODUIT.md). Pas de dates inventées.

## Phase 1 — Cœur MVP (livré)

- Auth véto/client, onboarding cabinet
- Clients / pets / timeline
- Relevé cardiaque
- Messagerie + indisponibilité
- i18n FR / NL / EN (+ ES ensuite)

## Phase 2 — Plateforme & monétisation (livré)

- Stripe billing par animal + mock local
- Admin métriques / users / payments
- Google OAuth + 2FA
- Reset MDP, médias, locales ES
- Addons Family / Care+ / Horse *(livrés puis **retirés de la vente** — features incluses)*
- Commissions véto + commercial + UI fiches
- Espace commercial (prospects, encode, pitch)
- Link-requests, care reminders, horse pack
- Plan monthly + steer triennial (offre actuelle : 3,50 € / 35 € / 95 €)

## Phase 3 — Engagement & backlog

| Item | Doc / note |
|------|------------|
| Push FCM opérationnel | **Livré** — messages véto + confirmation RDV (voir [08](08-MESSAGERIE-NOTIFICATIONS.md)) |
| Parcours email client 12 mois | **Livré** — [23](23-PARCOURS-EMAIL-CLIENT.md) (parallèle Discovery in-app) |
| Import clients admin CSV/XLS + Gemini | **Livré** — [24](24-IMPORT-CLIENTS-ADMIN.md) |
| WebSocket temps réel | — |
| Refresh silencieux clients | — |
| Simulation admin 10 ans | [16](16-ADMIN-SIMULATION-10ANS.md) **non livré** |
| Export / emails Care avancés | Backlog features incluses (**obsolète** comme roadmap « addon Care+ ») |
| Pharmacie BE (CNK, FEFO, DAF, VAMReg, invoices.connect) | [27](27-PHARMACIE-BELGIQUE.md) **spec / non livré** |

## Phase 4 — Multi-profils & pro santé

Voir [28-MULTI-PROFILS-PRO.md](28-MULTI-PROFILS-PRO.md).

| Sous-phase | Contenu | État |
|------------|---------|------|
| A | Self-inscription client Flutter (`/auth/register-client`) | **Livré** (+ Google create-if-absent) |
| B | Rôle `care_pro` + specialty + ACL `client_access` / `pet_access` | **Livré** |
| C1 | Création client : 409 enrichi + rattachement cabinet | **Livré** |
| C2–C3 | Partage dossier animal / fiche client (Nuxt) | **Livré** |
| D | Shell Flutter pro light (`vet_light`, `farrier` labels) | **Livré** (shell commun ; farrier différencié labels/CR) |
| E | Agenda GPS (adresse + lat/lng visites) | **Livré** (Flutter GPS ; Nuxt adresse + badge lieu) |
| F | CR visite + transcription / amélioration Gemini | **Livré** (dictée Flutter, audio fichier Nuxt, purge audio au finalize) |
| Suite | Auth médias visit-reports, specialties P1, listes pet_access client | **Livré** |
| M | Polish : dead code Google, tests shares/ACL/media, smoke register-client | **Livré** |
| N | Seed care_pro, labels shared permission, DEFAULT ACL `write_notes` | **Livré** |
| O | Tournées du jour (filtre agenda Flutter + compteur Nuxt) | **Livré** |

Ops staging : `make gcp-setup-media` (IAM SA + public avatars). PHI `visit-reports/` : pas d’URL publique app-side (GCP interdit condition allUsers).

## Go-to-market

Playbook commercial : [21-GTM-COMMERCIAL.md](21-GTM-COMMERCIAL.md) · politique tarifaire : [17](17-POLITIQUE-TARIFAIRE.md).
