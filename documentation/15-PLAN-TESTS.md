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
| B3.5 | P2 | Spot-check i18n | FR/NL/EN/ES/ET sur login + dashboard | Pas de clés brutes `xxx.yyy` |

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
| C2.2 | P0 | Liste clients | `/clients` recherche / filtre | Résultats cohérents |
| C2.3 | P0 | Fiche client | Ouvrir client | Pets, invite app, actions |
| C2.4 | P0 | Dossier pet | Chart FC, relevés, care, RDV, timeline | Données seed visibles |
| C2.5 | P1 | Liste pets | `/pets` | Animaux transverses cabinet |
| C2.6 | P1 | Créer / rattacher client | Nouveau client ; client existant → link | 409 enrichi + link OK |
| C2.7 | P1 | Photo animal | Upload photo pet | Affichée Pro + Flutter |
| C2.8 | P1 | Invite app | Depuis client | Lien / QR / email selon UI |
| C2.9 | P1 | Link-requests | `/clients?invitations=1` accepter/refuser | Statut mis à jour ; client lié |
| C2.10 | P2 | Parrainage | `/recommend` | Flux confrère |
| C2.11 | P2 | Produits | `/produits` | Plans 3,50 / 35 / 95 ; pas d’addons vendus |
| C2.12 | P2 | Commissions véto | `/commissions` | Ledger lisible |

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
| C6.7 | P2 | CR visite Nuxt | Rapport + finalize (si UI) | Draft → final ; audio purgé |

---

## D — Web Pro — Admin

Compte : `admin.demo@petsfollow.test`

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| D1 | P0 | Dashboard | `/admin` | Métriques chargent |
| D2 | P0 | Users | `/admin/users` liste + filtres | Rôles visibles |
| D3 | P1 | Créer user | Client / véto / care_pro / commercial / manager | Création OK ; **pas** de création admin UI |
| D4 | P1 | Care_pro admin | Créer care_pro + specialty | Login Flutter pro light OK |
| D5 | P1 | Commercials | `/admin/commercials` CRUD + assign véto + manager | Assign persist |
| D6 | P1 | Prospects globaux | `/admin/prospects` | Liste |
| D7 | P1 | Payments | `/admin/payments` | Entitlements / paiements |
| D8 | P1 | Commissions véto | Close période + mark-paid | `/admin/commissions` |
| D9 | P1 | Commissions commercial | Idem commercial | `/admin/commercial-commissions` |
| D10 | P2 | SPIFF bonuses | `/admin/commercial-bonuses` | Sync / mark-paid |
| D11 | P2 | Import clients | `/admin/client-imports` upload CSV/XLS | Job + détail `[id]` |
| D12 | P2 | Training admin | `/admin/training` | UI analyse pitch (Gemini si clé) |
| D13 | P2 | Isolation rôles | Véto tente `/admin` | Refus / redirect |
| D14 | P2 | Catalogue Stripe | Admin catalogue Stripe | ACL : véto refusé |

---

## E — Web Pro — Commercial & Manager

### E1 — Commercial

Compte : `commercial.demo@petsfollow.test`

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| E1.1 | P0 | Overview | `/commercial` | Portfolio |
| E1.2 | P0 | Prospects CRM | Contact → RDV → résultat | Transitions statut |
| E1.3 | P1 | Encode véto | `/commercial/vets` inscription véto | Compte créé / assigné |
| E1.4 | P1 | Client lié | Encode client lié cabinet | Pets / commission possibles |
| E1.5 | P1 | Client libre | Client sans liaison | `vet_link_required` à la création pet |
| E1.6 | P1 | Activer pet payant | Checkout / mock activation | Commission ledger |
| E1.7 | P1 | Commissions | `/commercial/commissions` + payout profile | Montants + profil |
| E1.8 | P2 | Pitch | `/commercial/pitch` | Contenu offre à jour |
| E1.9 | P2 | Training IA | `/commercial/training` | Session Gemini (si clé) |
| E1.10 | P2 | Settings | `/commercial/settings` | Locale / prefs |
| E1.11 | P2 | Directory | Prospect `source=directory` | Annuaire partagé |

### E2 — Commercial manager

Compte : `commercial.manager@petsfollow.test`

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| E2.1 | P0 | Dashboard équipe | `/commercial-manager` | KPI équipe |
| E2.2 | P1 | Suivi | `/commercial-manager/suivi` | RDV / relances |
| E2.3 | P1 | Prospects équipe | `/commercial-manager/prospects` | Scope équipe |
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
| F1.9 | P2 | Kennel encode | Quick encode batch | Animaux créés (entitlement) |
| F1.10 | P2 | Horse panel | Contacts / compétitions | CRUD OK |
| F1.11 | P2 | Foyer / household | Home / pets liés | Affichage foyer |
| F1.12 | P2 | Discovery cards | Home J0/J2/… | Cartes ; dismiss / CTA |
| F1.13 | P2 | Pets shared | Grant reçu | Label permission ; read-only si read |
| F1.14 | P1 | Sans liaison véto | Client libre crée pet | Erreur `vet_link_required` |

### F2 — Relevé cardiaque

| ID | Pri | Cas | Étapes | Attendu |
|----|-----|-----|--------|---------|
| F2.1 | P0 | Durées cabinet | Ouvrir FC | Durées = cabinet ; défaut = plus longue |
| F2.2 | P0 | Session complète | Taps → résultat BPM | Calcul `(taps×60)/durée` |
| F2.3 | P0 | Valider + comment | Commentaire ≤500 → validate | Visible Pro (C5) |
| F2.4 | P1 | Recommencer | Cancel / restart | Pas de session fantôme côté véto |
| F2.5 | P1 | Alerte seuil | BPM hors 60–140 | Warning UI |
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
| F3.6 | P2 | Depuis détail pet | CTA message | Bon thread |

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
| G1 | P0 | Shell | Login farrier | Agenda · Clients · Pets · Settings |
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
| G12 | P2 | vet_light | Login `vetlight.demo` | Même shell ; prompts CR specialty |

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
| H6 | P1 | Billing → features | Checkout pet | Entitlement → FC + messaging + Care/Horse ; commission activation |
| H7 | P1 | Care overdue | Pro crée → client postpone/done | Dashboard véto sync |
| H8 | P2 | Indispo messagerie | Vet unavailable → client | État côté app |
| H9 | P2 | Multi-cabinet | marie (Parc) vs demo (VetPlus) | Isolation données |
| H10 | P2 | Locale emails | Changer locale → trigger email | Email dans la bonne langue |
| H11 | P2 | Invite claim | Code invite → claim mobile | Practice + attribution commission |
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

**Hors scope manuel (non livré)** : pharmacie BE ([27](27-PHARMACIE-BELGIQUE.md)), simulation 10 ans ([16](16-ADMIN-SIMULATION-10ANS.md)).

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

### API (smoke)

`make smoke` — health, auth véto/client/admin, clients, billing mock, messagerie, heartrate validate **avec comment**, timeline.

### Web Pro (Playwright)

Répertoire : `nuxtjs/tests/e2e/specs/`

| Spec | Scénario |
|------|----------|
| `01-auth` | Login véto → dashboard → clients |
| `02-locale` | Changement langue EN dans settings |
| `03-clients` | Recherche client |
| `04-messaging` | Page messagerie + deep-link thread |
| `05-onboarding` | Redirection véto profil incomplet |
| `06-admin` | Login admin → dashboard |
| `07-commercial` | Login commercial → overview / prospects |
| `08-commercial-manager` | Dashboard équipe |
| `08-requests` | Calendrier + invitations clients |
| `09-pet-detail` | Fiche animal, shares, commentaire relevé HR |
| `10-products` | `/produits` plans TTC 3,50 / 35 / 95 |
| `11-admin-stripe-catalog` | Catalogue Stripe admin + ACL véto |

Local : `cd nuxtjs && npm run test:e2e` (API + Nuxt sur 8291/3002).

CI PR : `npx playwright test --list` (validation des specs).

Staging : workflow `deploy-gcp-staging.yml` exécute Playwright contre `petsfollow.ll-it-sc.be`.

### Go / Nuxt unit / Flutter

- Go unit + intégration : `make test-go` (heartrate comment, stripe catalog, care_pro ACL, multi-profils)
- Nuxt unit : `make test-nuxt` (Vitest)
- Flutter unit/widget : `make test-flutter` (palette, comment HR payload, pro light nav)
