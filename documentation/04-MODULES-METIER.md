# Modules métier — petsFollow

## Auth & compte

Login email/MDP, register + confirm email, forgot/reset, refresh JWT, Google OAuth, 2FA TOTP, avatar, locale, delete account.

## Practice (véto)

Profil cabinet (onboarding), durées FC, préférences email, disponibilité messagerie, overview dashboard, care overdue.

**Calendrier RDV** : plages horaires + vacances (`/vet/schedule`, `/vet/vacations`), agenda Pro `/calendar`, booking client optionnel (`client_booking_enabled`), replanification bilatérale, e-mail alerte demande, **pré-consultation** client à la confirmation ([31](31-PRECONSULTATION.md)).

## Clients & animaux

Liste clients Pro (invitations link-requests dans l’en-tête), dossier animal, photo, timeline, primary practice, envoi lien app.

## Relevé cardiaque

Sessions 15/30/60 s (config cabinet) — détail [09-RELEVE-CARDIAQUE.md](09-RELEVE-CARDIAQUE.md).

## Messagerie

Threads client↔véto, messages texte + media, read/read-all, mode indisponible — [08-MESSAGERIE-NOTIFICATIONS.md](08-MESSAGERIE-NOTIFICATIONS.md).

## Billing

Plans animal vendables : **monthly 3,50 €** (`subscription` only) · **annual 35 €** · **triennial 95 €** (recommandé) ; Checkout Stripe (ou mock), entitlements + `past_due`, Customer Portal, webhooks. Quinquennial + addons Family / Kennel / Care+ / Horse = **hors vente / legacy** (features Care/Horse/foyer/kennel incluses dès entitlement animal actif) — [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md), [17-POLITIQUE-TARIFAIRE.md](17-POLITIQUE-TARIFAIRE.md).

## Commissions

Ledger véto (progressif × facteur plan) + ledger commercial (taux par plan ; addons legacy si encore en base), accrual à l’activation (pas au renew), périodes close/mark-paid admin (véto + commercial). SPIFF commercial mix only (`commercial_bonus_awards` : sync auto + mark-paid). UI `ProCommissionSheet` + `/admin/commercial-bonuses`.

## Commercial / sales

Overview, inscriptions (`/commercial/vets` : véto · client lié · client sans liaison), list vets assignés, CRM prospects (**premier encodage gagne** : lookup + claim atomique ; pastille inactif **30 j** ; libération manager), commissions, payout profile (+ zone de base GPS/CP pour découvrir le **code** d’un commercial), page pitch. Pool libre = `commercial_user_id` NULL.

**Code Parrain (juge cabinet)** : inscription véto `/register` avec `inviteCode` (`practice.app_invite_codes` commercial/manager) → `assigned_commercial_id` définitif. **Sans code** → pool admin `/admin/vet-pool` (suggestions zone+activité + notes Gemini). Pas de sélection « près de chez vous » à l’inscription. Encode commercial : 409 si déjà assigné à un autre.

**QR client (parrainage)** : un client peut émettre `GET /me/app-invite`. Claim filleul → `practice.client_referrals` (first-wins) ; héritage commercial (referral du parrain ou `assigned_commercial` du véto référent) en **fallback** seulement ; rattachement cabinet du parrain **uniquement** si filleul libre (pas de `practice_clients`). Pas de commission au client promoteur.

**Commission** : `assigned_commercial_id` véto prioritaire, sinon fallback `commercial_referrals` client. Chaîne Comm→Véto→Client : le client joint au cabinet hérite le commercial du véto pour l’accrual (`ResolveVetCommercial`) ; une row `commercial_referrals` antérieure n’est **jamais** écrasée (first-wins), même si elle n’est plus le payé effectif.

**Vue filiation** : tables Pro `/commercial/filiation`, `/commercial-manager/filiation`, `/admin/filiation` (API `GET …/filiation`) — commercial → véto → client + badge Effectif (= même règle que Resolve).
Client sans liaison : `practice_id` NULL — pets créables sans cabinet ; liaison véto demandée ensuite (messagerie / visites). Commission commerciale à l’activation si cabinet lié.
**Responsable commercial** (`commercial_manager`) : dashboard équipe + suivi + prospects équipe (`/commercial-manager/*`) ; release single/bulk inactifs 30 j ; production manager privée (hors tableaux équipe).
Admin : CRUD commercials / managers, assign véto, pool non assignés + suggestions IA, `manager_user_id`, prospects globaux, payouts commissions, SPIFF bonuses.

## Care & Horse

Rappels care (+ seed horse pack), contacts professionnels, compétitions ; foyer / kennel (`litter_tag`, batch) — **inclus** avec entitlement animal actif (plus d’upsell addon).

## Pharmacie cabinet (Belgique) — spec

Dictionnaire CNK/AFMPS, stocks multi-dépôts FEFO, DAF + PDF, workers VAMReg / invoices.connect — **spécification** [27-PHARMACIE-BELGIQUE.md](27-PHARMACIE-BELGIQUE.md) (**non livré**). Distinct des rappels Care côté client.

## Admin plateforme

Métriques, users, payments, commissions véto & commercial, SPIFF commercial, imports, brand assets, catalogue Stripe, formation IA, modules CR. Simulation 10 ans = **backlog** ([16](16-ADMIN-SIMULATION-10ANS.md)).

### Support bug-report

- Bouton **Support** dans `ProTopbar` (tous rôles Pro) + entrées Flutter (Settings / AppBar).
- `POST /api/v1/support/tickets` avec diagnostics (console, HAR-lite, session, config, fenêtre 15 min).
- Inbox admin `/admin/support` : liste, détail, statut (`open` / `in_progress` / `resolved` / `closed`), réponse → email utilisateur.
- Email ops à `SUPPORT_INBOX_EMAIL` (défaut `support@petsfollow.app`) à la création.
- **RGPD** : export `supportTickets` (subject, message, **diagnostics**, replies, métadonnées) ; anonymisation sur `DELETE /me` (client/pro) — donc aussi via le job rétention 3 ans (`POST /internal/retention/run`), qui réutilise les mêmes chemins purge/tombstone.
- Inbox admin : filtre statut + recherche texte (`q`) sur sujet / email / nom.
