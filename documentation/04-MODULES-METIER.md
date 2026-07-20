# Modules métier — petsFollow

## Auth & compte

Login email/MDP, register + confirm email, forgot/reset, refresh JWT, Google OAuth, 2FA TOTP, avatar, locale, delete account.

## Practice (véto)

Profil cabinet (onboarding), durées FC, préférences email, disponibilité messagerie, overview dashboard, care overdue.

**Calendrier RDV** : plages horaires + vacances (`/vet/schedule`, `/vet/vacations`), agenda Pro `/calendar`, booking client optionnel (`client_booking_enabled`), replanification bilatérale, e-mail alerte demande.

## Clients & animaux

Liste clients Pro (invitations link-requests dans l’en-tête), dossier animal, photo, timeline, primary practice, envoi lien app.

## Relevé cardiaque

Sessions 15/30/60 s (config cabinet) — détail [09-RELEVE-CARDIAQUE.md](09-RELEVE-CARDIAQUE.md).

## Messagerie

Threads client↔véto, messages texte + media, read/read-all, mode indisponible — [08-MESSAGERIE-NOTIFICATIONS.md](08-MESSAGERIE-NOTIFICATIONS.md).

## Billing

Plans animal (one_time / subscription) + **addons foyer en paiement unique à vie** (Family / Kennel / Care+ / Horse), Checkout Stripe (ou mock), entitlements + `past_due` (legacy sub), Customer Portal, webhooks — [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md), [17-POLITIQUE-TARIFAIRE.md](17-POLITIQUE-TARIFAIRE.md).

## Commissions

Ledger véto (progressif × facteur plan) + ledger commercial (taux par plan/addon), accrual à l’activation (pas au renew), périodes close/mark-paid admin (véto + commercial). SPIFF commercial (`commercial_bonus_awards` : sync auto + mark-paid). UI `ProCommissionSheet` + `/admin/commercial-bonuses`.

## Commercial / sales

Overview, encode/list vets assignés, CRM prospects (contact / RDV / résultat), commissions, payout profile, page pitch. Annuaire partagé `source=directory`.  
**Responsable commercial** (`commercial_manager`) : dashboard équipe + suivi + prospects équipe (`/commercial-manager/*`) ; production manager privée (hors tableaux équipe).  
Admin : CRUD commercials / managers, assign véto, `manager_user_id`, prospects globaux, payouts commissions, SPIFF bonuses.

## Care & Horse

Rappels care (+ seed horse pack), contacts professionnels, compétitions ; household Family (≥2, pas de plafond) / Kennel (≥6, `litter_tag`, batch).

## Admin plateforme

Métriques, users, payments, commissions véto & commercial, SPIFF commercial. Simulation 10 ans = **backlog** ([16](16-ADMIN-SIMULATION-10ANS.md)).
