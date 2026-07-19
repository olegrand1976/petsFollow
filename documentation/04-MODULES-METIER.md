# Modules métier — petsFollow

## Auth & compte

Login email/MDP, register + confirm email, forgot/reset, refresh JWT, Google OAuth, 2FA TOTP, avatar, locale, delete account.

## Practice (véto)

Profil cabinet (onboarding), durées FC, préférences email, disponibilité messagerie, overview dashboard, visits / care overdue.

## Clients & animaux

Liste clients Pro, dossier animal, photo, timeline, primary practice, invitations / link-requests client↔véto, envoi lien app.

## Relevé cardiaque

Sessions 15/30/60 s (config cabinet) — détail [09-RELEVE-CARDIAQUE.md](09-RELEVE-CARDIAQUE.md).

## Messagerie

Threads client↔véto, messages texte + media, read/read-all, mode indisponible — [08-MESSAGERIE-NOTIFICATIONS.md](08-MESSAGERIE-NOTIFICATIONS.md).

## Billing

Plans + Checkout Stripe (ou mock), entitlements, Customer Portal, addons Family / Care+ / Horse — [07-STRIPE-BILLING.md](07-STRIPE-BILLING.md), [17-POLITIQUE-TARIFAIRE.md](17-POLITIQUE-TARIFAIRE.md).

## Commissions

Ledger véto (progressif × facteur plan) + ledger commercial (taux par plan/addon), périodes close/mark-paid admin, UI `ProCommissionSheet`.

## Commercial / sales

Overview, encode/list vets assignés, CRM prospects, commissions, payout profile, page pitch. Admin : CRUD commercials, assign, prospects globaux.

## Care & Horse

Rappels care (+ seed horse pack), contacts professionnels, compétitions ; household / family limit 2–3 pets.

## Admin plateforme

Métriques, users, payments, commissions véto & commercial. Simulation 10 ans = **backlog** ([16](16-ADMIN-SIMULATION-10ANS.md)).
