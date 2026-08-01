# Plan de tests — petsFollow

Ce document couvre :

1. **Tests manuels** (web Pro Nuxt + Flutter client + Flutter care_pro) — sections A–I
2. **Tests automatisés** (smoke / unit / e2e) — section Z

Références : [06-FLUX-UTILISATEURS](06-FLUX-UTILISATEURS.md) · [04-MODULES-METIER](04-MODULES-METIER.md) · [28-MULTI-PROFILS-PRO](28-MULTI-PROFILS-PRO.md) · [AGENTS.md](../AGENTS.md)

---

## Légende

| Symbole | Signification |
|---------|---------------|
| **P0** | Bloquant release — à passer à chaque build candidat |
| **P1** | Critique métier — avant staging → prod |
| **P2** | Couverture complète / régression périodique |
| OK / KO / N/A | Résultat session de test |

**Format résultat** : `ID | OK/KO | testeur | date | notes`

---

## Prérequis environnement

### Local

```bash
make up-infra && make migrate && make seed
make api-dev          # :8291
make nuxtjs-dev       # :3002
# Flutter : flutter run (API = http://10.0.2.2:8291 ou IP LAN)
```

### Staging

- Web : `https://petsfollow.ll-it-sc.be`
- API : même host `/api/v1`
- Flutter : build pointant staging (`--dart-define` API URL)

### Données

- Relancer seed si état corrompu : `make seed`
- Billing : mock local OK ; Stripe test keys si parcours Checkout réel
- Push FCM : ADC + `FCM_ENABLED` (sinon no-op — noter N/A pour push)

### Navigateurs / devices

| Surface | Cibles mini |
|---------|-------------|
| Web Pro | Chrome desktop + Safari ou Firefox ; viewport 1280 + 768 |
| Flutter | 1 Android physique ou émulateur ; iOS si dispo |
| Locales | FR (défaut) + au moins 1 parmi NL / EN / ES / ET |

---

## Matrice comptes démo

Mots de passe : véto `VetDemo123!` · client `ClientDemo123!` · admin `AdminDemo123!` · commercial `CommercialDemo123!` · care_pro `CareProDemo123!`

| Rôle | Email | Usage principal |
|------|-------|-----------------|
| Véto | `vet.demo@petsfollow.test` | Parcours cabinet complet (VetPlus) |
| Véto | `vet.parc@petsfollow.test` | Multi-cabinet / Marie NL |
| Véto | `vet.onboarding@petsfollow.test` | Gate onboarding |
| Véto | `vet.unverified@petsfollow.test` | Email non confirmé |
| Véto | `vet.reset@petsfollow.test` | Reset MDP (`demo-reset-password`) |
| Client | `client.demo@petsfollow.test` | Pets + Care/Kennel/Horse |
| Client | `client.vide@petsfollow.test` | Compte vide / création pet |
| Client | `client.marie@petsfollow.test` | Locale NL |
| Care_pro | `farrier.demo@petsfollow.test` | Pro light (Spirit, write_notes) |
| Care_pro | `vetlight.demo@petsfollow.test` | Pro light vet_light |
| Commercial | `commercial.demo@petsfollow.test` | CRM + encode |
| Commercial mgr | `commercial.manager@petsfollow.test` | Dashboard équipe |
| Admin | `admin.demo@petsfollow.test` | Ops plateforme |
| DEV | `dev.demo@petsfollow.test` | Support IT léger (sans billing/sales) |

Tokens démo : confirm email `demo-confirm-email` · reset `demo-reset-password`.

---

## A — Smoke P0 (30–45 min)

Parcours minimum avant toute dist / staging.

| ID | Surface | Étapes | Attendu |
|----|---------|--------|---------|
| A1 | Web | Login `vet.demo` → `/dashboard` | KPI + shell Pro visibles |
| A2 | Web | `/clients` → ouvrir un client → dossier pet | Fiche + timeline / FC / care |
| A3 | Web | `/messages` | Liste threads ; ouvrir un thread |
| A4 | Web | `/calendar` | Agenda charge ; visite seed visible |
| A5 | Web | Logout → login `admin.demo` → `/admin` | Métriques admin |
| A5b | Web | Logout → login `dev.demo` → `/admin` | Shell ops léger (users/support/flags) ; pas billing |
| A6 | Flutter | Login `client.demo` | Shell 5 tabs (Home / Pets / Care / Messages / Settings) |
| A7 | Flutter | Ouvrir un pet → démarrer FC (sans valider) | Timer + taps OK |
| A8 | Flutter | Messagerie : ouvrir un thread | Historique messages |
| A9 | Flutter | Login `farrier.demo` | Shell pro light (Agenda / Clients / …) |
| A10 | Croisé | Véto envoie message → client rafraîchit Messages | Message visible côté Flutter |

---

## B — Auth & compte (Web + Flutter)

### B1 — Public Web

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| B1.1 | P0 | Login véto | Email/MDP valides | Redirect rôle (`/dashboard` ou `/welcome`) |
| B1.2 | P0 | Login mauvais MDP | MDP incorrect | Erreur claire, pas de session |
| B1.3 | P1 | Register véto | `/register` → email | Page `/register/sent` |
| B1.4 | P1 | Confirm email | `/confirm-email?token=demo-confirm-email` | Compte confirmé / login OK |
| B1.5 | P1 | Forgot / reset | Forgot → `/reset-password?token=demo-reset-password` (`vet.reset`) | Nouveau MDP utilisable |
| B1.6 | P2 | Google OAuth (si configuré) | Bouton Google login | Session Pro ; bouton masqué si pas de client ID |
| B1.7 | P2 | 2FA | Settings → activer TOTP → logout → login + code | Gate 2FA ; refuse code faux |
| B1.8 | P1 | Must-change password | Compte force change | `/change-password` puis accès app |
| B1.9 | P2 | Pages légales | `/legal/mentions` `/privacy` `/terms` | Contenu + i18n |
| B1.10 | P2 | Landing | `/` sections produits / CTA | Aligné offre (3,50 / 35 / 95 €) |
| B1.11 | P2 | Invite landing | `/invite/[code]` (code seed si dispo) | Landing invite + CTA app |

### B2 — Auth Flutter

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| B2.1 | P0 | Login client | `client.demo` | Shell owner |
| B2.2 | P0 | Login care_pro | `farrier.demo` | Shell pro light (pas owner) |
| B2.3 | P0 | Login véto dans Flutter | Compte `vet.demo` | Message « utilisez Pro web » / refus |
| B2.4 | P1 | Register client | Self-signup → confirm email | Compte client créé |
| B2.5 | P1 | Forgot / reset | Flux MDP | Reset OK |
| B2.6 | P1 | Google client (si config) | Sign-In Google email inconnu | Create-if-absent client |
| B2.7 | P1 | Google email Pro | Compte véto via Google | Erreur `google_client_only` |
| B2.8 | P2 | Force change password | Compte temporaire | Écran dédié |
| B2.9 | P2 | Logout | Settings → logout | Retour login ; token invalidé |
| B2.10 | P2 | Delete account | Profile → delete | Compte inaccessible |

### B3 — Locale & profil

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| B3.1 | P1 | Locale Web | Settings → EN (ou NL) | UI + cookie `pf_locale` ; persist après reload |
| B3.2 | P1 | Locale Flutter | Settings → autre langue | UI + sync `PATCH /me/locale` |
| B3.3 | P2 | Client NL seed | Login `client.marie` | UI NL par défaut |
| B3.4 | P2 | Avatar / profil | Upload photo (Web ou Flutter) | Visible après refresh |
| B3.5 | P2 | Spot-check i18n | FR/NL/EN/ES/ET/IT sur login + dashboard | Pas de clés brutes `xxx.yyy` |

---

## C — Web Pro — Véto

Compte : `vet.demo@petsfollow.test`

### C1 — Onboarding & settings

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C1.1 | P0 | Onboarding incomplet | Login `vet.onboarding` | Redirect `/onboarding` ; pas de dashboard métier |
| C1.2 | P1 | Compléter onboarding | Remplir profil + durées FC (≥1) | Accès `/dashboard` |
| C1.3 | P1 | Durées FC settings | Settings → 15/30/60 | Persist ; au moins une durée |
| C1.4 | P1 | Dispo messagerie | Unavailable ON | Client voit indisponible (voir F3) |
| C1.5 | P1 | Plages / vacances | Schedule + vacations | Calendrier respecte indispos |
| C1.6 | P2 | Prefs email véto | Notifs message / FC / visit | Toggle persist |
| C1.7 | P2 | Client booking | Activer/désactiver booking client | Flutter BookVisit reflète l’état |

### C2 — Dashboard, clients, pets

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C2.1 | P0 | Dashboard | Ouvrir `/dashboard` | Overview + care overdue si seed |
| C2.2 | P0 | Liste clients | `/clients` recherche / filtre | Résultats cohérents ; colonne / filtre téléphone si seed (`0470 00 00 01` Sophie) |
| C2.3 | P0 | Fiche client | Ouvrir client | Pets, invite app, actions ; édition `contactPhone` si `clients.write` (`client-phone-save` — **manuel** ; auto = Go `TestClientContactPhone*`) |
| C2.4 | P0 | Dossier pet | Chart FC, relevés, care, RDV, timeline | Données seed visibles |
| C2.4b | P1 | Tension & labos | Onglet vitals : saisir tension Pro (site/commentaire) ; créer/éditer panel labo (`valueNum`/`valueText`) ; timeline `blood_pressure` / `lab_panel` ; client Flutter sheet tension + lecture panels | Go `TestBloodPressure*` / `TestLabPanel*` ; Flutter `pet_quick_actions_test` + `lab_panels_screen_test` ; Playwright `09-pet-detail` `@p1` ; `make smoke` BP/labs |
| C2.5 | P1 | Liste pets | `/pets` (+ `?unread=1` depuis KPI dashboard) | Animaux transverses ; filtre **Non lus** ; badge relevé non lu ; colonne **Type de relevé** (FC) ; Vitest `vet-pets-list.spec.ts` |
| C2.6 | P1 | Créer / rattacher client | Nouveau client (+ `contactPhone` optionnel) ; client existant → link | 409 enrichi + link OK ; téléphone visible liste/get ; `PATCH /clients/{id}` isolé cabinet non lié (`TestClientContactPhone*`) ; **account-global** last-write-wins si multi-cabinets (`TestClientContactPhoneAccountGlobalLastWriteWins`) |
| C2.7 | P1 | Photo animal | Upload photo pet | Affichée Pro + Flutter |
| C2.8 | P1 | Invite app | Depuis client | Lien / QR / email selon UI |
| C2.9 | P1 | Link-requests | `/clients?invitations=1` accepter/refuser | Statut mis à jour ; client lié |
| C2.10 | P2 | Parrainage | `/recommend` | Flux confrère |
| C2.11 | P2 | Produits | `/produits` | Plans 3,50 / 35 / 95 ; pas d’addons vendus |
| C2.12 | P2 | Commissions véto | `/commissions` | Ledger lisible |
| C2.13 | P0 | Nouvelle consultation | `/clients` → CTA → modal pet → visite `confirmDirect` + `consultationSession` → CR | Panel CR ; CTA DAF / facture / Terminer ; walk-in **hors** overlap agenda ; close → confirm Enregistrer/Annuler |
| C2.14 | P1 | Consultation anti-orphelins | Fermeture modal / sheet sans save CR | Visite `cancelled` (Nuxt + Flutter) ; pendant PUT CR → Cancel/X désactivés ; 409 `consultation_has_report` = garder ; **finalize CR** → auto-`done` ; retention `cancelledStaleConsultations` (âge min **6 h** sans CR, appliqué au **cron quotidien** retention ≈ 03:30) |
| C2.15 | P1 | Walk-in hors vacation/lock | `consultationSession` + `scheduledAt≈now` | Pas de 400 `on_vacation` ; pas de lock agenda ; `source=care_pro` pour terrain ; care_pro **ne peut pas** cancel/reschedule un RDV cabinet (`403 care_pro_visit_only`), `done` OK |
| C2.16 | P1 | CTA post-CR DAF / facture | Après save CR → CTA | `/daf/nouveau?visitId=` · finalize inline · `/invoicing?visitId=&mode=direct|fromDaf` + contextes ; badge draft DAF oublié `/consultations` |
| C2.17 | P1 | Historique consultations | `/consultations` liste walk-in date DESC + filtres + **soft-delete** | Client + animal + date ; lien Écouter si `hasAudio` (draft) + **durée `audioDurationSec`** ; ouvrir CR ; supprimer → hors liste (`deleted_at`) ; player modal affiche durée persistée |
| C2.18 | P1 | Reprise consultation après veille | Desk lock/switch mid-consultation | Autosave CR avant purge JWT ; reprise modal pour **le même** email ; pas de fuite vers un autre profil |
| C2.19 | P1 | CR split notes / CR | Modal ou calendrier → panel CR | Panes gauche (`visit-report-transcript`) / droite (`visit-report-body` + TipTap `visit-report-prose`) ; select langue `visit-report-target-locale` (défaut auto = langue transcription) ; Améliorer disabled si notes **et** CR vides ; footer Annuler / Enregistrer / Finaliser ; Aide rétractable + image process ; feedback qualité post-IA ; référence si final (`visit-report-reference` — Go + `03e`) |
| C2.20 | P1 | CR transcription → versions | hint-transcribe / edit notes → improve → restore | v0 transcript · v1 IA · restore cible pane ; re-improve écrase v1 |
| C2.21 | P1 | CR caractères / escape | PUT/GET + preview markdown | C0/NUL strip ; XSS preview sanitized ; guillemets/backslash round-trip |
| C2.22 | P1 | CR boutons états | idle / dirty / dictating / final | Matrice enabled/disabled Dicter · Améliorer · Annuler · Enregistrer · Finaliser · Restore |

### C3 — Calendrier & RDV

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C3.1 | P0 | Agenda | `/calendar` | Visites ; badge Aujourd’hui si applicable |
| C3.2 | P1 | Proposer RDV | Depuis fiche pet | Visite créée ; visible Flutter |
| C3.3 | P1 | Confirmer demande client | Accepter demande | Push `visit_confirmed` (si FCM) |
| C3.4 | P1 | Replanifier | Déplacer créneau | Notification / email selon prefs |
| C3.5 | P1 | Annuler / done | Actions visite | Statuts cohérents |
| C3.6 | P2 | Deep-link visite | `/calendar?visit={id}` (email) | Focus bonne visite |
| C3.7 | P2 | GPS / adresse | Saisir adresse visite | Lien Maps si présent |

### C4 — Messagerie Pro

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C4.1 | P0 | Liste threads | `/messages` | Threads clients liés |
| C4.2 | P0 | Envoyer texte | Répondre | Visible Flutter |
| C4.3 | P1 | Envoyer média | Image | Preview + côté client |
| C4.4 | P1 | Deep-link thread | URL thread | Ouverture directe |
| C4.5 | P1 | Read / read-all | Marquer lu | Compteurs à jour |
| C4.6 | P2 | Depuis dossier pet | CTA messagerie | Même thread |

### C5 — Relevés FC (côté Pro)

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C5.1 | P0 | Relevés validés | Dossier pet après validate Flutter | Ligne BPM + date |
| C5.2 | P0 | Commentaire | Session avec comment Flutter | Commentaire visible tableau + timeline |
| C5.3 | P1 | Pending invisible | Session non validée | Absente côté véto |
| C5.4 | P1 | Chart | Plusieurs points | Graphique cohérent |

### C6 — Care (Pro) & partage ACL

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| C6.1 | P1 | Créer rappel Care | Fiche pet | Visible Flutter CareTab |
| C6.2 | P1 | Overdue dashboard | Rappels en retard | Signal dashboard |
| C6.3 | P1 | Share pet | Partager animal → care_pro (write_notes) | Farrier voit Spirit |
| C6.4 | P1 | Share client | Partage fiche client | Clients listés pro light |
| C6.5 | P2 | Permissions | read vs write_notes vs full | Notes / messages masqués selon ACL |
| C6.6 | P2 | Révoquer share | DELETE share | Disparaît côté care_pro |
| C6.7 | P1 | CR visite Nuxt (IA BFF) | PUT report → POST `report-improve` → `report-finalize` (BFF Nuxt) | Routes enregistrées **et matchées** (pas 404 Nitro « Page not found ») ; 200/402/503 métier OK ; audio purgé si finalize 200 |

Auto : Playwright `@p1` [`03d-visit-report-ai-bff.spec.ts`](../nuxtjs/tests/e2e/specs/03d-visit-report-ai-bff.spec.ts).
| C6.8 | P1 | CR multi-auteurs | Visite avec CR terrain + cabinet → `/calendar` | Liste auteurs ; peer lecture seule ; « Mon CR » éditable |

### I7 — Facturation Billit (dev)

> **UI** : kill-switch [`INVOICING_UI_ENABLED`](../nuxtjs/utils/invoicing-ui.ts) (page `/invoicing` + CTA Facturer consultation/DAF). `true` = UI métier ; `false` = message WIP. API Billit mock toujours testée côté Go. (IDs **I7.*** pour éviter collision avec pharmacie C7.)

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| I7.0 | P1 | UI WIP (flag off) | `/invoicing` si `INVOICING_UI_ENABLED=false` | Message under-development |
| I7.1 | P1 | Connect mock | UI ou `POST …/connect/start` + complete | Statut `active` |
| I7.2 | P1 | Facture BE | Créer facture BE + TVA + send | draft → delivered (mock) |
| I7.3 | P1 | Contrepartie IT | Pays IT sans codice/PEC | Erreur validation |
| I7.4 | P1 | Credit note / proforma | CN + `relatedDocumentId` ; proforma send | CN OK ; proforma → `issued` |
| I7.5 | P1 | Quota mensuel | `docs_included_monthly=1` puis 2e send ; doc `sending` | 409 `docs_quota_exceeded` |
| I7.6 | P1 | Retry rejected | Invoice `rejected` → send | Rejeu → delivered (mock) |
| I7.7 | P1 | Admin mark-partner | `/admin/invoicing` + confirm + mark | `partnerListedAt` ; liste avec `practiceName` |
| I7.8 | P1 | Admin alertes ops | Connexions active sans partner >7 j ; usage ≥80 % | KPI + CSV pending PartyID |
| I7.9 | P1 | Flux A SaaS draft+send | Opt-in + draft puis send ; cron C1 boucle batches (Brussels) | Doc `saas_master` → delivered ; usage inchangé ; `POST /internal/saas-invoices/run` |

Auto UI : Playwright `@p1` `@invoicing` [`18-invoicing.spec.ts`](../nuxtjs/tests/e2e/specs/18-invoicing.spec.ts) (I7.1–I7.4) + [`06-admin.spec.ts`](../nuxtjs/tests/e2e/specs/06-admin.spec.ts) I7.9. Auto API : Go `TestInvoicing*` / `TestInvoicingAdminSaasDraft` / `TestInvoicingAdminSaasTargetsAndCron`.

---

## D — Web Pro — Admin

Comptes : `admin.demo@petsfollow.test` · DEV `dev.demo@petsfollow.test` (D17)

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| D1 | P0 | Dashboard | `/admin` | Métriques chargent |
| D2 | P0 | Users | `/admin/users` liste + filtres | Rôles visibles |
| D3 | P1 | Créer user | Client / véto / care_pro / commercial / manager | Création OK ; **pas** de création admin UI |
| D4 | P1 | Care_pro admin | Créer care_pro + specialty | Login Flutter pro light OK |
| D5 | P1 | Commercials | `/admin/commercials` CRUD + assign véto + manager | Assign persist |
| D5b | P1 | Pool cabinets | `/admin/vet-pool` + suggestions | Assign depuis suggestion |
| D5c | P1 | Branches commerciales | `/admin/sales-branches` liste + pending + run now | Auto `DUPONTD` ; job 10h/18h ; Go `TestSalesBranchesAutoCreateAndSkipPeerSponsored` |
| D6 | P1 | Prospects globaux | `/admin/prospects` | Liste |
| D7 | P1 | Payments | `/admin/payments` | Entitlements / paiements |
| D8 | P1 | Commissions véto | Close période + mark-paid | `/admin/commissions` |
| D9 | P1 | Commissions commercial | Idem commercial | `/admin/commercial-commissions` |
| D10 | P2 | SPIFF mix bonuses | `/admin/commercial-bonuses` | Sync / mark-paid |
| D11 | P2 | Import clients | `/admin/client-imports` upload CSV/XLS | Job + détail `[id]` |
| D11b | P1 | Compendium PDF | Nav admin → `/admin/compendium-imports` (flag pharmacy) | Page liste + badge `dev` ; extract + **matching CNK AFMPS** (PDF sans CNK) + pending→confirm→commit ; DELETE job · e2e `21-compendium-admin` · Go `TestCompendiumImportFlow` / `Delete` / `CNKMatch` |
| D11c | P1 | Import AFMPS CSV | Nav admin → `/admin/afmps-imports` (flag pharmacy) | Page liste + badge `dev` ; triple contrôle validate→reviewed→commit ; filtre collisions ; meta JSON merge · e2e `23-afmps-admin` · Go `TestAFMPSImportTripleGate` / `Gate1Blocked` / `MetaMerge` · unit `pharmacy/afmps_csv_test` |
| D12 | P2 | Training admin | `/admin/training` | UI analyse pitch (Gemini si clé) |
| D13 | P2 | Isolation rôles | Véto tente `/admin` | Refus / redirect |
| D14 | P1 | Support inbox | Topbar Support → ticket ; `/admin/support` liste (+ `q`) + détail + **réponse** | Ticket visible ; reply listée ; email soft-fail OK · e2e `14-support.spec.ts` `@p1` · Go search/export/anonymize |
| D16 | P1 | Alertes auth ALERT/URGENT | SMTP confirm fail / stuck unverified | Ticket `source=system` + email `OPS_NOTIFY_EMAIL` · Go `TestSMTPConfirmFailCreatesSystemAlertTicket` · job `POST /internal/auth-health/run` |
| D15 | P2 | Catalogue Stripe | Admin catalogue Stripe | ACL : véto refusé |
| D17 | P0 | Rôle DEV support IT | Login `dev.demo` → `/admin/users` + `/admin/support` + `/admin/runtime-flags` | 200 ; nav sans billing/sales/brand/AI · Go `TestDevRole*` · Playwright `19-dev-support` · UC-AD-02 · **skip si seed absent** (staging sans RESET) |
| D17b | P0 | DEV billing/sales gate | DEV → `/admin/payments` + `/admin/commercials` | redirect home `/admin` (middleware `admin-only`) · API 403 · Playwright `19-dev-support` |

---

## E — Web Pro — Commercial & Manager

### E1 — Commercial

Compte : `commercial.demo@petsfollow.test`

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| E1.1 | P0 | Overview | `/commercial` | Portfolio |
| E1.2 | P0 | Prospects CRM | Contact → RDV → résultat + lookup/claim premier encodé | Transitions ; owned → message collègue |
| E1.2b | P1 | Pastille inactif | Pastille rouge ≥ 30 j | Visible sur lignes stale |
| E1.3 | P1 | Encode véto | `/commercial/vets` inscription véto | Compte créé / assigné ; 409 si déjà assigné ailleurs |
| E1.4 | P1 | Client lié | Encode client lié cabinet | Pets / commission possibles |
| E1.5 | P1 | Client libre | Client sans liaison crée pet | 201, `practiceId` vide ; prompt liaison ensuite |
| E1.6 | P1 | Activer pet payant | Checkout / mock activation | Commission ledger |
| E1.7 | P1 | Commissions | `/commercial/commissions` + payout profile | Montants + profil |
| E1.8 | P2 | Pitch + mémo ASV | `/commercial/pitch` · `/commercial/asv-memo` | Contenu offre + leave-behind ASV (Imprimer / PDF) |
| E1.8b | P2 | Flux produit | `/flux` (nav Offre) | Profils + parcours + liens croisés ; admin / commercial / manager · Playwright `07-commercial` |
| E1.9 | P2 | Training IA | `/commercial/training` | Session Gemini (si clé) |
| E1.10 | P2 | Settings | `/commercial/settings` | Locale / prefs + champ téléphone de contact |
| E1.12 | P1 | Porte téléphone | Commercial sans `contactPhone` → toute page Pro | Redirection `/complete-contact-phone` ; après saisie, retour au tableau de bord ; pages publiques (`/dossier/**`, `/login`) non impactées |
| E1.11 | P2 | Directory | Prospect `source=directory` | Annuaire partagé |

### E2 — Commercial manager

Compte : `commercial.manager@petsfollow.test`

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| E2.1 | P0 | Dashboard équipe | `/commercial-manager` | KPI équipe |
| E2.2 | P1 | Suivi | `/commercial-manager/suivi` | RDV / relances |
| E2.3 | P1 | Prospects équipe | `/commercial-manager/prospects` | Scope équipe + libérer / bulk inactifs 30 j |
| E2.4 | P1 | Production perso | Accès `/commercial/*` | Hors tableaux équipe |
| E2.5 | P2 | Training | `/commercial-manager/training` | UI OK |
| E2.6 | P2 | Isolation | Commercial simple → URLs manager | Refus |

---

## F — Flutter — Client (owner)

Compte principal : `client.demo@petsfollow.test` · compte vide : `client.vide@…`

### F1 — Shell & pets

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F1.1 | P0 | 5 tabs | Navigation Home→Pets→Care→Messages→Settings | Pas de crash |
| F1.2 | P0 | Liste pets | PetsTab | Owner + shared labels |
| F1.3 | P0 | Détail pet | Infos, photo, actions | Sections cohérentes |
| F1.4 | P1 | Créer pet | `client.vide` → form → checkout | Plan monthly/annual/triennial |
| F1.5 | P1 | Checkout mock/Stripe | Payer | Entitlement actif ; Care/Horse inclus |
| F1.6 | P1 | Portal / resume | Billing depuis détail | Portal ou reprise session |
| F1.7 | P1 | Deep link paiement | Retour success Stripe | État actif |
| F1.8 | P1 | Éditer pet | Form édition | Persist |
| F1.8b | P1 | Puce + carnet | Create/edit + multi-photos → PDF | N° optionnels ; PDF unique compressé (API) ; affiché Pro |
| F1.9 | P2 | Kennel encode | Quick encode batch | Animaux créés (entitlement) |
| F1.10 | P2 | Horse panel | Contacts / compétitions | CRUD OK |
| F1.11 | P2 | Foyer / household | Home / pets liés | Affichage foyer |
| F1.12 | P2 | Discovery cards | Home J0/J2/… | Cartes ; dismiss / CTA |
| F1.13 | P2 | Pets shared | Grant reçu | Label permission ; read-only si read |
| F1.14 | P1 | Sans liaison véto | Client libre crée pet | 201 + dialog « Lier un vétérinaire » ; badge détail |

### F2 — Relevé cardiaque

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F2.1 | P0 | Durées cabinet | Ouvrir FC | Durées = cabinet ; défaut = plus longue |
| F2.2 | P0 | Session complète | Taps → résultat BPM | Calcul `(taps×60)/durée` |
| F2.3 | P0 | Valider + comment | Commentaire ≤500 → validate | Visible Pro (C5) |
| F2.4 | P1 | Recommencer | Cancel / restart | Pas de session fantôme côté véto |
| F2.5 | P1 | Alerte hausse BPM | Hausse ≥ delta espèce (seed 30) vs dernier validé | `isAlert` + email véto ; 1er relevé sans alerte |
| F2.5b | P1 | Espèce `other` | Pet other | CTA FC masqué ; start API `heartrate_not_supported` |
| F2.6 | P1 | How-to measure | Settings / éducation | Contenu |
| F2.7 | P1 | Premium gate | Pet sans entitlement | FC bloquée / CTA paywall |
| F2.8 | P2 | Commentaire max | >500 car. | Truncate / erreur validation |

### F3 — Messagerie client

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F3.1 | P0 | Lire / envoyer | Texte | Visible Pro |
| F3.2 | P1 | Média | Photo | Upload + affichage |
| F3.3 | P1 | Indisponible véto | Après C1.4 | Banner / état indispo |
| F3.4 | P1 | Push message | Véto écrit (FCM) | Notif + tap → Messages |
| F3.5 | P2 | Prefs notif | Désactiver `messages` | Pas de push message |
| F3.6 | P2 | Depuis détail pet | CTA message | Bon thread (practice×pet) |
| F3.7 | P1 | Ensure thread | Compose pro+pet | `POST /messaging/threads` ; lock sans véto |

### F4 — Care & RDV

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F4.1 | P0 | CareTab | Liste rappels | Done / postpone |
| F4.2 | P1 | Sync Pro | Done côté client | Overdue Pro mis à jour |
| F4.3 | P1 | Shared care read-only | Pet partagé read | Pas d’action write |
| F4.4 | P1 | Book visit | Demande créneau | Visible `/calendar` Pro |
| F4.5 | P1 | Confirm push | Véto confirme | Push `visit_confirmed` |
| F4.6 | P1 | Replanif client | Proposer déplacement | Email/push véto selon prefs |
| F4.7 | P2 | Booking disabled | Cabinet off | CTA masqué / erreur claire |
| F4.8 | P2 | Reminder settings | Prefs rappels | Persist |

### F5 — Vétos, settings, legal

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F5.1 | P1 | My vets | Liste liens | Cabinets liés |
| F5.2 | P1 | Invite véto email | Link-request | Apparaît invitations Pro |
| F5.3 | P1 | Claim invite | Deep link `petsfollow://` / web invite | Rattachement |
| F5.4 | P2 | Notif prefs | hr/care/visits/messages/discovery/billing | PATCH OK |
| F5.5 | P2 | Legal in-app | CGU / privacy | Contenu |
| F5.6 | P2 | Profile | Édition nom / etc. | Persist |

### F6 — Timeline

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F6.1 | P1 | Timeline pet | Après FC + message + care | Entrées chronologiques |
| F6.2 | P1 | Commentaire FC | Entrée session | Corps contient comment |
| F6.3 | P2 | ACL notes | Sans write_notes | Notes visite masquées |
| F6.4 | P2 | ACL messages | Sans full | Messages absents timeline |

---

## G — Flutter — Care pro (pro light)

Comptes : `farrier.demo` / `vetlight.demo` · pet seed Spirit (write_notes)

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| G1 | P0 | Shell | Login farrier | Agenda · Clients · Pets · **Messages** · Settings |
| G1b | P0 | Shell staff | Login `vet.demo` | Agenda · Clients · Pets · **Messages** · Settings |
| G2 | P0 | Agenda filtres | Aujourd’hui / 7j / Tout | Tri ASC fenêtres courtes ; hors done/cancelled |
| G3 | P1 | Marquer Fait | Bouton Fait (write_notes) | Statut done |
| G4 | P1 | Maps / GPS | Ouvrir adresse visite | App Maps |
| G5 | P1 | Clients partagés | Liste | Client grant visible |
| G6 | P1 | Fiche pet | ProLightPetScreen | Infos, timeline, docs, care |
| G7 | P1 | CR visite | Dictée / upload → transcribe → improve → finalize | Draft→final ; échec Gemini = erreur claire |
| G8 | P1 | Consentement audio | Avant micro/fichier | Consent UI obligatoire |
| G9 | P1 | Audio PHI | Après finalize | Audio inaccessible ; pas d’URL `/media/` publique |
| G10 | P2 | Permission read | Share read-only | Pas de Fait / CR write |
| G11 | P2 | Settings | Specialty, locale, logout | OK |
| G12 | P1 | vet_light | Login `vetlight.demo` | Même shell 5 tabs (Messages inclus) ; CR sans Améliorer IA |
| G13 | P1 | Messagerie staff | `vet.demo` → Messages → compose / envoi | Thread client ; saisie OK ; pas de lock « lier un véto » |
| G14 | P0 | Messagerie care_pro | `vetlight.demo` → Messages → compose Spirit / envoi | Thread person-scoped ; visible client ; distinct du fil cabinet |

---

## H — Flux croisés (dual-face) — checklist intégration

À enchaîner dans l’ordre ; chaque ligne = un scénario bout-en-bout.

| ID | Pri | Flux | Acteurs | Vérifications |
|----|-----|------|---------|---------------|
| H1 | P0 | Message véto → client | vet.demo ↔ client.demo | Texte Pro → Flutter ; inverse ; push si FCM |
| H2 | P0 | FC validate → Pro | client → vet | Pending invisible ; validated + comment + chart + timeline |
| H3 | P1 | RDV bilatéral | client book → vet confirm | Calendar + push + prefs email |
| H4 | P1 | Link-request | Flutter invite → Pro accept | Relation active ; pets possibles |
| H5 | P1 | Share → care_pro | Vet share → farrier | Agenda/fiche ; notes selon permission |
| H13 | P1 | Envoi dossier animal → pro | Client Flutter → e-mail pro → `/dossier/{token}` | ZIP (PDF + docs + carnet) ; marketing + tél. commercial ; expiry 24 h ; UC-X-08 |
| H14 | P1 | Consultation client + partage PDF | Historique Flutter → CTA disponible (final) / en attente (draft) → CR final → e-mail → `/consultation/{token}` | PDF CR brandé ; multi-auteurs finaux ; expiry 24 h ; UC-X-09 |
| H15 | P1 | Client AI (tag `dev`) | CR final → « Comprendre mon CR » ; Home/Settings → triage 24/7 Vert/Orange/Rouge | Explication additive + disclaimer ; escalade Rouge = tel cabinet + messagerie + RDV ; flag `CLIENT_AI_ENABLED` |
| H6 | P1 | Billing → features | Checkout pet | Entitlement → FC + messaging + Care/Horse ; commission activation |
| H7 | P1 | Care overdue | Pro crée → client postpone/done | Dashboard véto sync |
| H8 | P2 | Indispo messagerie | Vet unavailable → client | État côté app |
| H9 | P2 | Multi-cabinet | marie (Parc) vs demo (VetPlus) | Isolation données |
| H10 | P2 | Locale emails | Changer locale → trigger email | Email dans la bonne langue |
| H11 | P0 | Invite claim | Code invite → claim mobile / register | Practice / care_pro / commercial ; first-wins ; pas d’overwrite |
| H12 | P2 | Past_due | Simuler impayé (staging Stripe) | Gate features / portal |

---

## I — Non-régression UX / technique (P2)

| ID | Cas | Attendu |
|----|-----|---------|
| I1 | Responsive Pro 768px | Sidebar / listes utilisables |
| I2 | Table ↔ kanban (listes Pro) | Bascule `useListView` OK |
| I3 | Refresh token | Laisser session expirer / idle | Re-auth propre |
| I4 | Offline Flutter | Mode avion court | Message erreur ; reprise |
| I5 | Rotation / kill app mid-FC | Reprise ou cancel propre |
| I6 | Médias avatars | Upload puis reload cold | Toujours affichés |
| I7 | Isolation rôles URLs | Accès croisés `/admin` `/commercial-manager` | 403 / redirect |
| I8 | Charte Pro | Pas de thème dark Flutter dans Nuxt | Tokens `--pf-vet-*` |
| I9 | Offre sync | Landing `#produits` + `/produits` | Même prix / inclus |

**Pharmacie BE (dev)** : stock + DAF/PDF + VAMReg dry-run + prix/seuils/commandes/BL/inventaire + chaîne alimentaire sous flag `PHARMACY_ENABLED` — tests Go `TestPharmacy*` (dont `TestPharmacyPlanCoverage`) + Playwright `@p0`/`@p1` `@pharmacy` [`17-pharmacy-stock-daf.spec.ts`](../nuxtjs/tests/e2e/specs/17-pharmacy-stock-daf.spec.ts). Roadmap : [37](37-ROADMAP-STOCK-FACTURATION.md). DAF→Billit (BIL-9) **gelé** (gate `EnqueueInvoicesConnect`) jusqu’accès reseller. Simulation 10 ans ([16](16-ADMIN-SIMULATION-10ANS.md)) hors scope.

| ID | Prio | Cas | Attendu |
|----|------|-----|---------|
| C7.1 | P0 | Receipt lot + DLC | `POST /api/vet/pharmacy/batches` 201 ; lot non vide |
| C7.2 | P0 | Finalize DAF | Sortie FEFO ; mouvements `reason=daf` + `dafId`/`dafItemId` |
| C7.3 | P0 | Liste mouvements filtrée | `GET /api/vet/pharmacy/movements?dafId=` |
| C7.4 | P1 | DAF depuis consultation | `/daf/nouveau?clientUserId&petId&visitId` + panneau traitements | Upsert `PUT …/daf/for-visit` ; banner « depuis consultation » ; CTA libellé DAF (≠ prescriptions) ; finalize FEFO explicite **inline** (ProModal) ; protocoles 1 clic ; AMM catalogue ; rupture → réception express |
| C7.5 | P1 | Facture liée à la visite (API) | `POST …/documents` + `visitId` | `documents.visit_id` persisté ; mismatch → 400 |
| C7.6 | P1 | VAMReg dry-run à finalize antibiotique | `TestPharmacyDAFVAMRegDryRun` | `vamregStatus=sent` ; lignes `job_audit` |
| C7.7 | P1 | Prix + alerte réassort | `PUT …/prices` · `PUT …/reorder-thresholds` · `GET …/reorder-alerts` | Seuil &lt; stock → alerte |
| C7.8 | P1 | Commande réassort + BL | `TestPharmacyOrdersAndDeliveryNote` | Order `fromAlerts` → send e-mail ; `POST …/delivery-notes` 201 |
| C7.9 | P1 | Inventaire annuel | `TestPharmacyInventorySession` + e2e `@pharmacy` | Session open → count → close applique adjust |
| C7.10 | P1 | Chaîne alimentaire + temps d’attente | `TestPharmacyFoodChainAndWithdrawal` | `food_producing` sans V/L/O → 400 ; snapshot après finalize |
| C7.11 | P1 | VAMReg dry-run e2e + BL/réassort | `17-pharmacy-stock-daf.spec.ts` (@p1 @pharmacy) | Antibio → `vamregStatus=sent` ; delivery-note ; reorder alert |
| C7.12 | P1 | Couverture plan ✅ (guards) | `TestPharmacyPlanCoverage` + `TestEnqueueInvoicesConnectResellerGate` | Prix GET ; send sans notifier → 502 ; suppliers ; BL `MANUAL-*` ; dépôt étranger 400 ; inventaire CSV ; retry VAMReg `pending` stuck → OK (re-enqueue) ; cancel `pending` → `daf_vamreg_in_flight` ; ban food-chain ; FK users→pharmacy ≠ CASCADE ; gate Billit reseller |
| C7.13 | P1 | UI `/stock` + prix/BL/CSV e2e | `17-pharmacy-stock-daf.spec.ts` (@p1 @pharmacy) | Shell bands/receipt/inventory/dev-badge ; PUT/GET prices ; BL vide → `MANUAL-*` ; export CSV inventaire |
| C7.14 | P1 | Lots wasted non ressuscités | `TestPharmacyStockReceiptAndWaste` | Receipt même clé après waste → `batch_wasted` |
| C7.15 | P1 | Overlay withdrawal practice | `TestPharmacyFoodChainAndWithdrawal` | PATCH withdrawal n’altère pas `ref_medications` national |
| C7.16 | P1 | Protocoles cliniques | `GET …/protocols` + seed VetPlus | `TestPharmacyClinicalProtocolsList` ; inject UI 1 clic `03b` (`consultation-treatments-protocols`) |
| C7.17 | P1 | Dispenses pet + drafts oubliés | `GET …/pets/{id}/daf-dispenses` · `GET …/consultations/daf-drafts` | Timeline fiche animal ; badge `/consultations` + `/daf` si draft >1 h (`TestPharmacyDAFPetDispensesAndDrafts` stale>1h + e2e `03c` badge mock + `09` `pet-daf-dispenses`) ; événement dossier client post-finalize |

**Prescriptions (tag `dev`)** : brouillons + preview PDF sous flag `PRESCRIPTIONS_ENABLED` — UI `/prescriptions` + badge `nav.tagDev` ; tests Go `TestPrescriptions*` ([35](35-PRESCRIPTIONS.md)). Signature / partage dossier / chat hors scope V1. Pas de useCase commercial tant que tag `dev`.

**PACS Orthanc (tag `dev`)** : status/wake + upload `.dcm` (magic `DICM`) + viewer fiche animal + admin `/admin/pacs` (wake inclus) sous `PACS_ENABLED` — Go `TestPacs*` (upload · preview/file via `ParentSeries` · isolation cross-cabinet · `UNIQUE(orthanc_study_id)` 409 · purge Orthanc rétention · admin wake 202 · frame OOR 404 · commentaires étude POST/GET + isolation) · Vitest `pacsPoll` + `pacs-measure` (calibration mm + spacing scalé + coords image) · Playwright `@p0` [`20-pacs-admin.spec.ts`](../nuxtjs/tests/e2e/specs/20-pacs-admin.spec.ts) + [`20b-pacs-imaging.spec.ts`](../nuxtjs/tests/e2e/specs/20b-pacs-imaging.spec.ts) + [`20c-pacs-ga-net.spec.ts`](../nuxtjs/tests/e2e/specs/20c-pacs-ga-net.spec.ts) (download + EOF) · doc [40](40-PACS.md) / [40-P2](40-PACS-P2.md). Staging : Orthanc dans Cloud Build + `PACS_ORTHANC_URL` auto. Pas de useCase commercial tant que tag `dev`.

**Research observatoire (tag `dev`)** : rôle `research` + opt-in cabinet + ETL anonymisé + UI `/research` + admin `/admin/research` + V2 groups/Data room sous `RESEARCH_ENABLED` — Go `TestResearch*` / `TestResearchGroupsAndDataRoom` / `TestValidateResearch` · Playwright `@p0` [`22-research.spec.ts`](../nuxtjs/tests/e2e/specs/22-research.spec.ts) · doc [42](42-RESEARCH.md). Staging : flags on + secrets ETL/salt + `make gcp-research-etl-scheduler`. Pas de useCase commercial tant que tag `dev`.

| ID | Prio | Cas | Attendu |
|----|------|-----|---------|
| R1 | P0 | Login research → overview | KPI visibles + badge `dev` |
| R2 | P0 | Timeseries + heatmap | tables chargées (flag on) ; heatmap k≥5 cabinets distincts |
| R3 | P0 | Admin opt-ins | `/admin/research` liste non vide après seed |
| R4 | P1 | Flag off | API 404 `research_disabled` |
| R5 | P0 | Groups + Data room | créer groupe ≠ unlock ; admin `dataroom_enabled` ; events k≥5 cabinets distincts sans city/PII/`payload` (scope réseau) |

| ID | Prio | Cas | Attendu |
|----|------|-----|---------|
| C8.1 | P1 | Créer draft (prescriptions) | `POST /api/v1/vet/prescriptions` 201 ; `status=draft` |
| C8.2 | P1 | Preview PDF | `GET …/prescriptions/{id}/pdf` → `%PDF` |
| C8.3 | P1 | Flag off | `PRESCRIPTIONS_ENABLED=false` → 404 `prescriptions_disabled` |
| C8.5 | P1 | Prescription → DAF | `POST …/daf/from-prescription` | Lignes avec `ref_medication_id` → draft DAF ; sans lien catalogue → `daf_empty` |

---

## Suites recommandées par type de release

| Release | Suites |
|---------|--------|
| Hotfix / patch | **A** + scénarios touchés |
| Feature Flutter | **A** + **F**/**G** concernés + **H** lié |
| Feature Pro | **A** + **C**/**D**/**E** + **H** lié |
| Candidate staging | **A** + tous **P0/P1** (B–H) |
| Release prod / store | Staging P0/P1 + **I** spot + locales FR+EN+NL + device Android réel |

---

## J — Modules optionnels (ex-addons)

| ID | Pri | Cas | Attendu |
|----|-----|-----|---------|
| J1 | P0 | Client nouveau / vide | Modules OFF — pas Kennel / foyer / Horse panel |
| J2 | P0 | `client.demo` seed | Modules ON — Kennel + foyer + Horse Spirit |
| J3 | P1 | Settings → Options | Toggles sync `PATCH /me/feature-modules` |
| J4 | P1 | Catalogue options | Cartes Activer sans paiement |
| J5 | P2 | Désactiver module | UI se retire après reload |

## K — Multi-profil

| ID | Pri | Cas | Attendu |
|----|-----|-----|---------|
| K1 | P0 | Register / seed pro | Profil `client` auto + profil pro |
| K2 | P0 | Flutter switch | Settings → Changer de profil → shell adapté + nouveaux tokens |
| K3 | P1 | Admin attach profil | `/admin/users` — pas sur soi-même |
| K4 | P1 | Commercial attach | `POST /commercial/users/{id}/profiles` |
| K5 | P2 | farrier pro ↔ personnel | Dual shell |

## L — Équipe cabinet

| ID | Pri | Cas | Attendu |
|----|-----|-----|---------|
| L1 | P0 | `/team` référence | Liste + invite (vet.demo) |
| L2 | P1 | Seed équipe | colleague / assist / secretary |
| L3 | P1 | Override droits | PATCH permissions |
| L4 | P1 | Révoquer | Statut revoked |
| L5 | P2 | Non-référence | Invite refusé |
| L6 | P0 | Labels ACL lisibles | Droits i18n + tooltips (pas de clés brutes `clients.read`) |
| L7 | P0 | Switch poste + veille | Header avatars si équipe ≥2 ; idle (défaut 2 min, `deskIdleMinutes` cabinet — roster serveur prime sur localStorage ; Vitest `deskIdleMinutes.spec.ts`) / force lock → overlay MDP ; Annuler switch → veille (pas de restore) ; solo = pas d’idle ; restore lastPath |
| L8 | P0 | `shares.read` vs manage | Secrétaire / assist : onglet partages visible, pas de create/revoke ; GET OK / POST 403 |
| L9 | P1 | `pharmacy.*` vs clinique | Stock/DAF sur `pharmacy.write` ; override explicite indépendant ; override legacy seul `pets.write_clinical` miroite encore la pharma |

## M — Concurrence commerciale

| ID | Pri | Cas | Attendu |
|----|-----|-----|---------|
| M1 | P0 | Login commercial → `/commercial/competition` | Synthèse + cards |
| M2 | P0 | Tabs FR / BE / ES | Contenu change |
| M3 | P1 | Manager accès | Même page |
| M4 | P2 | Véto accès | Pas dans nav / refus |

### Démo commerciale 15 min

1. `commercial.demo` → `/commercial/competition` (BE) + pitch  
2. Flutter `client.demo` modules ON → FC + commentaire  
3. Switch profil `farrier.demo` → Spirit agenda GPS  
4. Pro `vet.demo` → `/team` + relevé validé  

---

## Feuille de session (copier-coller)

```text
Date :
Environnement : local / staging
Build / commit :
Testeur :
Devices : Chrome … | Android …

Smoke A :   A1□ A2□ A3□ A4□ A5□ A6□ A7□ A8□ A9□ A10□
Bloquants :
Notes :
```

---

## Z — Tests automatisés (référence)

### Philosophie

Toute mutation métier doit renforcer le filet (règle Cursor `anti-regression-quality.mdc`). Niveau le plus bas : Go intégration > Playwright `@p0` > widget Flutter. Suites auto = billing mock, comptes `*.petsfollow.test`, users jetables pour `DELETE /me`.

### API (smoke)

`make smoke` — health, auth véto/client/admin, clients, billing mock, messagerie **H1 croisé** (véto → client), heartrate validate **avec comment**, timeline, tension client + panel labo véto (`valueText` + trend crea), **H13** `GET /public/pet-dossier/{token}` inconnu → 404.

### Envoi dossier animal (Go intégration — H13)

`go test ./internal/handlers/ -run 'PetDossier|PublicPetDossier|TestUpdateMeContactPhone' -count=1`
`go test ./internal/platform/petdossier/ -count=1`

| Cas | Attendu |
|-----|---------|
| Create + meta + download ZIP | `dossier.pdf` dans le ZIP ; `commercialPhone` présent |
| Expiry | GET meta/download → 410 |
| Non-owner | POST → 403 |
| Email destinataire malformé (dont CR/LF) | POST → 400 `invalid_email` |
| Quota du jour saturé | POST → 429 `dossier_share_limit` |
| Partage périmé + nouveau partage | l'ancienne ligne (email du tiers) est supprimée |
| Consultation publique en rafale | 429 (anti-énumération de tokens) |
| Nom de document piégé (`../`) | entrée écartée du ZIP (zip slip) |

Bornes anti-DoS de la construction du ZIP (route publique, tout en mémoire) : 25 pièces jointes max et 40 Mio cumulés — au-delà, les pièces sont écartées, jamais tronquées.

Flutter widget : `pet_send_dossier_test` (envoi bloqué tant que le consentement PHI n'est pas coché ; fiche animal + dialogue sans débordement en 360 dp clavier ouvert) · Nuxt gate : `completeContactPhoneGate.spec.ts` · Playwright mocké : `16-dossier-public.spec.ts` · UC : `UC-X-08`.

Surface publique `/dossier/{token}` : `Referrer-Policy: no-referrer` et `X-Robots-Tag: noindex, nofollow` (`routeRules`) — sans ça un clic vers `/register` fuiterait le token dans le `Referer`. Purge des partages périmés : job quotidien `POST /internal/retention/run` (`make gcp-retention-scheduler`) + lifecycle GCS 2 jours sur `dossier-shares/`.

### Consultation client + partage PDF (Go intégration — H14)

`go test ./internal/handlers/ -run 'ClientConsultation' -count=1`
`go test ./internal/platform/consultationpdf/ -count=1`

| Cas | Attendu |
|-----|---------|
| Owner GET client-consultation | reports `final` uniquement ; multi-auteurs |
| Draft only | 404 `consultation_not_found` |
| Non-owner / non-client | 403 |
| Create + meta + download PDF | magic `%PDF-` ; marketing |
| Expiry | GET meta → 410 |
| ListVisits | `hasFinalReport: true` + `reportStatus: final` (owner) ; draft → `reportStatus: draft` sans `hasFinalReport` |
| Timeline client | `meta.hasReport` final-only ; `meta.reportStatus` draft\|final owner ; strip non-owner |

Flutter widget : `consultation_view_test` (CTA disponible/en attente · ExpansionTile · dédup · filet meta · share) · Playwright mocké : `17-consultation-public.spec.ts` · UC : `UC-X-09`.
Go timeline : `TestPetTimelineIncludesVisitReportForVetNotClient` (draft → `reportStatus`) · `TestPetTimelineClientHasReportOnFinalCR` · `TestPetTimelineNonOwnerStripsHasReport` · `TestClearTimelineMeta` / `stripClientConsultationFlags*`.

Surface publique `/consultation/{token}` : mêmes headers noindex / no-referrer que `/dossier/**`. Purge : retention job + préfixe media `consultation-shares/`.

### Client AI — vulgarisation CR + triage (Go + Flutter — H15)

`go test ./internal/handlers/ -run 'TestClientAI|TestClientConsultationExplain' -count=1`
`go test ./internal/platform/gemini/ -run 'TestParseClient' -count=1`
`cd flutter && flutter test test/features/client_ai/`

| Cas | Attendu |
|-----|---------|
| Flag off | 404 `client_ai_disabled` |
| Explain owner + CR final | cards + disclaimer ; 2ᵉ GET = cache |
| Explain non-owner | 403 |
| Triage session + message chocolat | level `red` + escalation `practicePhone` / canBook / canMessage |
| Flutter explain CTA | `consultation_explain_test` |
| Flutter triage CTAs | `triage_chat_test` (call / book / message) |

Doc : [`43-CLIENT-AI.md`](43-CLIENT-AI.md). Pas de useCase commercial tant que tag `dev`.

### Parrainage / QR (Go intégration — anti-régression)

`go test ./internal/handlers/ -run 'TestParrainageControl_|TestFiliationChain_' -count=1`

| ID | Cas | Attendu |
|----|-----|---------|
| P1 | Code durable commercial | Même code à chaque `Ensure` + `GET /me/app-invite` + `vetRegisterUrl` |
| P2 | QR véto → register-client | `practice_clients.vet_user_id` = promoteur ; casse normalisée |
| P3 | QR care_pro → register-client | `client_access` write_notes |
| P4 | Invite commercial bat nearby | Referral = A malgré `commercialUserId=B` ; `invite_code` persisté |
| P5 | Second code commercial | `already_linked` + `inviterId` = premier ; pas d’overwrite |
| P6 | Nearby puis même invite | Backfill `invite_code` (code non perdu) |
| P7 | Register véto | Code sales OK ; code véto → 400 ; `assignedCommercialId` seul → pool |
| P8 | Client émet QR | `GET /me/app-invite` → 200 `role=client` ; pas de `vetRegisterUrl` ; self-claim → 400 `self_referral` |
| P9 | Invite invalide + nearby | Pas de referral silencieux au mauvais commercial |
| P10 | QR client → filleul | `client_referrals` + héritage commercial + cabinet si libre ; first-wins parrain |
| P11 | QR client no-steal | Filleul déjà rattaché → lien parrain OK, pas de 2e cabinet |
| P12 | Héritage referral parrain | Filleul hérite `commercial_referrals` du sponsor |
| P13 | Commercial A conservé | Filleul déjà referral A + QR client → A inchangé ; lien parrain OK |
| P14 | Soft self-referral | `TryClaimInvite` own code → `ignored` ; 0 `client_referrals` |
| P15 | Concurrence 2 QR client | Exactement 1 `practice_clients` + 1 sponsor (FOR UPDATE) |

Fichier : `go/internal/handlers/parrainage_invite_control_integration_test.go`.

#### Playwright — invite / QR

| Spec | Scénario |
|------|----------|
| `15-app-invite` | Landing `role=client` sans CTA cabinet ; modal commercial `vetRegisterUrl` |

#### Filiation — 4 chaînes anti-perte (F1–F14)

Fichier : `go/internal/handlers/filiation_chain_integration_test.go`.

| ID | Chaîne | Cas | Attendu |
|----|--------|-----|---------|
| F1 | Comm→véto | Encode `POST /commercial/vets` | `assigned_commercial_id` = commercial |
| F2 | Comm→véto | 2e commercial encode même email | **409** `already_assigned` ; assign inchangé |
| F3 | Comm→véto | Admin unassign | assign NULL + présent dans pool `/admin/vets/unassigned` |
| F4 | Véto→client | Client existant + `claim-invite` QR véto | `practice_clients` + `practice_id` ; `status=linked` |
| F5 | Véto→client | Déjà lié A + reclaim / claim B / collègue même cabinet | `already_linked` reclaim A ; multi-cabinet B lié + primary = A ; collègue même practice → `already_linked` + `vet_user_id` reste A |
| F6 | Comm→client | `POST /commercial/clients` | standalone → `commercial_referrals` ; lié → `practice_clients` + Resolve + **0** `commercial_referrals` |
| F7 | Comm→véto→client | Invite sales → register véto → QR véto → client | `practice_clients` + `ResolveVetCommercial` = comm |
| F8 | Comm→véto→client | Referred A + claim véto assigné A | referral **reste A** ; Resolve = A |
| F9 | Comm→véto→client | Referred A + claim véto assigné C | referral **reste A** ; Resolve = **C** (priorité assign ≠ perte de row) |
| F10 | Comm→véto→client | Unassign après F8 | Resolve bascule sur fallback `commercial_referrals` (= A) |
| F10b | Comm→véto→client | Unassign après F9 (était C) | Resolve bascule sur fallback A (pas silent 0) |
| F11 | Comm→véto→client | Multi-cabinet CommA + CommB | `Resolve(P_A)=A`, `Resolve(P_B)=B`, `Resolve("")=B` (dernier lien) ; List Effectif aligné par practice |
| F12 | RGPD | Client avec `commercial_referrals` | `GET /me/export` → `commercialReferrals`, `clientReferrals`, **`filiationEvents` (contenu)**, **`petDocuments`**, **`deviceTokens`** ; `DELETE /me` purge referrals **et** events |
| F13 | Commission | Multi-cabinet Accrue 2 pets | ledger `subscription_pct` → CommA sur P_A, CommB sur P_B |
| F14 | Audit | Encode / referral / claim / unassign / **re-assign** / accept-link | `filiation_events` : `vet_assigned`, `client_referral`, **`practice_client_linked`**, `vet_unassigned` (re-assign émet unassign+assign) |
| F15 | RGPD pro | `DELETE /me` commercial | purge `filiation_events` + clear `assigned_commercial_id` + `commercial_referrals` |
| F16 | RGPD staff | `DELETE /me` `vet_assistant` / `secretary` | tombstone (pas 404) ; rétention inclut ces rôles |
| F17 | RGPD dual | care_pro + pet `owner_user_id` | `DELETE /me` purge pets puis tombstone |
| F18 | RGPD consent | client `terms_accepted_at` NULL | `POST /me/accept-terms` `{consent:true}` → `termsAcceptedAt` ; **API 403** hors allowlist (`/pets`…) jusqu’à accept |

#### Vue filiation (API + Pro UI)

`go test ./internal/handlers/ -run 'TestFiliationList_' -count=1`

| Surface | Endpoint | Page |
|---------|----------|------|
| Commercial | `GET /commercial/filiation?q=&limit=&offset=` | `/commercial/filiation` |
| Manager | `GET /commercial-manager/filiation?q=&commercialId=&limit=&offset=` (équipe + soi) | `/commercial-manager/filiation` |
| Admin | `GET /admin/filiation?q=&branchId=&commercialId=&limit=&offset=` | `/admin/filiation` |
| Events | `GET …/filiation/events?eventType=&limit=&offset=` (+ admin filters) | section historique UI |

Réponse list / events : `{ items, limit, offset, truncated }` (list défaut 2000 ; events 100). Rétrocompat `?format=items` → tableau nu. UI : pagination offset, filtres history, CSV rows+events, filtre commercial manager.

Colonnes : commercial → véto/cabinet → client + invite + parrain + badge **Effectif**. Accrue = `Resolve(practiceID)` du pet.

Tests list : scope A≠B, admin, manager, F9 effective=C, unassign → `client_referral`, UUID invalides, `limit=1` → truncated, manager events scoped, `format=items`.

**V1 livrée** ; polish restant = E2E profond / audit transactionnel (non bloquant).

Fichiers : `go/internal/store/filiation.go`, `filiation_events.go`, `filiation_*_integration_test.go`, `nuxtjs/components/pro/ProFiliationTable.vue`.

E2E Playwright `@p1` : `07-commercial`, `08-commercial-manager`, `06-admin` (smoke pages filiation + history + export).
### Web Pro (Playwright)

Répertoire : `nuxtjs/tests/e2e/specs/`

| Spec | Scénario | Tag |
|------|----------|-----|
| `01-auth` | Login / register / forgot-reset | `@p0` |
| `02-locale` | Changement langue EN dans settings | |
| `03-clients` | Recherche client | `@p0` |
| `03b-consultation` | Nouvelle consultation : CR→Terminer · close sans save · CTA DAF/facture · traitements→FEFO→finalize · protocole 1 clic | `@p0` |
| `03c-consultations-history` | Historique `/consultations` : liste walk-in + filtre + ouvrir CR + soft-delete + **durée audio / player** + badge draft DAF | `@p1` |
| `03d-visit-report-ai-bff` | BFF CR IA : POST `/api/visits/:id/report-improve` (+ finalize, `me/ai-module/roi`) ≠ 404 Nitro | `@p1` |
| `03e-visit-report-versions` | CR split : panes · cancel dirty · restore versions · escape save/reload · boutons · toggle référence post-finalize | `@p1` |
| `04-messaging` | Page messagerie + deep-link + PJ | `@p0` |
| `05-onboarding` | Redirection véto profil incomplet | |
| `06-admin` | Admin dashboard / users / commercials / filiation | `@p1` (filiation) |
| `07-commercial` | Login commercial → overview / prospects / pitch / mémo ASV / `/flux` / filiation | `@p1` (filiation) |
| `08-commercial-manager` | Dashboard manager / suivi / prospects / mémo ASV / filiation | `@p1` (filiation) |
| `08-requests` | Calendrier + invitations clients | |
| `09-pet-detail` | Fiche animal, CTA consultation, shares, commentaire relevé HR | `@p0` (parcours chart/HR) + `@p1` CTA |
| `10-products` | `/produits` plans TTC 3,50 / 35 / 95 | |
| `11-admin-stripe-catalog` | Catalogue Stripe admin + ACL véto | |
| `12-competition` | Concurrence commerciale FR/BE/ES | |
| `13-team-staff-smoke` | Assist / secretary ACL + shares.read + pharmacy caps + factu readonly | `@p0` |
| `13b-desk-switch` | Switch poste partagé + veille (lock overlay) + reprise consultation mid-veille (scénario G) | `@p0` |
| `14-support` | Ticket support | `@p1` |
| `15-app-invite` | Landing QR client sans CTA cabinet ; modal commercial dual lien | |
| `16-dossier-public` | Page `/dossier/{token}` meta + expiry + CTA register (mock API) | `@p0` |
| `19-dev-support` | DEV ops léger users/support/flags | `@p0` |
| `20-pacs-admin` | Admin PACS metrics/logs/playground + **wake clic** (flag on) | `@p0` |
| `20b-pacs-imaging` | Fiche animal onglet Imagerie + upload (canvas défaut ; Cornerstone si `NUXT_PUBLIC_PACS_VIEWER_ENGINE=cornerstone`) | `@p0` |
| `20c-pacs-ga-net` | P2.3 : download `.dcm` (magic DICM) + preview frame OOR → 404 (+ clamp canvas) | `@p0` |
| `22-research` | Observatoire Research + admin opt-ins (flag on) | `@p0` |

Local :

```bash
make test-e2e-p0   # --grep @p0
make test-e2e      # suite complète
```

Prérequis : API `:8291` + Nuxt `:3002` + seed (`AUTH_RATE_LIMIT_PER_MIN=1000` via `make api-dev`).

CI PR : job `playwright` — stack Postgres + API + Nuxt preview, exécute `--grep @p0`.

Staging (`deploy-gcp-staging.yml`) : smoke API postdeploy + Playwright suite complète (`--grep-invert @flaky`) contre Cloud Run Nuxt.

**Rollback staging** si Playwright / smoke post-deploy rouge : workflow en échec (pas de rollback auto). Revenir à la révision Cloud Run précédente :

```bash
gcloud run services update-traffic petsfollow-api --to-revisions=PREV=100 --region=europe-west9
gcloud run services update-traffic petsfollow-nuxtjs --to-revisions=PREV=100 --region=europe-west9
# ou redeploy du commit known-good sur branche staging
```

### Go / Nuxt unit / Flutter

- Go unit + intégration : `make test-go` — CI backend avec Postgres + migrate/seed (plus de skip DB) ; alertes auth : `TestSMTPConfirmFailCreatesSystemAlertTicket` ; reset staging admin : `TestAdminStagingSeed*` ; seed preserve : `TestSeedPreservesSupportTickets` / `TestSeedPreservesClientGraph` / `TestSeedPreservesProtectedRoles` ; densification démo : `TestSeedMass` (`make seed-mass` après `make seed`, emails `mass.*@petsfollow.test`)
- Billit / invoicing : `go test ./internal/invoicing/...` + intégration `TestInvoicing*` / `TestInvoicingWebhook*` / `TestInvoicingAdminMarkPartner` ; Playwright `@p1` `@invoicing` `18-invoicing.spec.ts` (UI métier si `INVOICING_UI_ENABLED`) ; admin `/admin/invoicing` (`06-admin.spec.ts`) — checklist **I7**
- Nuxt unit : `make test-nuxt` (Vitest) — inclus dans `make test`
- Flutter unit/widget : `make test-flutter` ; smoke API : `make test-flutter-smoke` (opt-in, hors CI PR)
- Dist Android / Play : `flutter test` obligatoire avant build (`SKIP_TESTS=1` pour override conscient) — pas le smoke API
