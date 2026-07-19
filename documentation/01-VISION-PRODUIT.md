# Vision produit — petsFollow

## Personas

- **Dr Martin** — véto libéral, dashboard Pro web
- **Sophie** — propriétaire chien senior, app mobile pets
- **Léa** — commerciale / apporteuse, espace Pro commercial

## Périmètre cœur (MVP)

1. Création animal (client — Flutter)
2. Suivi clients + animaux (véto — Nuxt Pro)
3. Messagerie interne + mode indisponible véto
4. Relevé cardiaque (durée 15/30/60 s selon paramètres du cabinet) — Valider envoi véto / Recommencer (client)
5. Timeline historique (messages, relevés validés, événements)

## Livré au-delà du MVP initial

| Extension | Statut |
|-----------|--------|
| Inscription véto + confirmation email + onboarding profil cabinet | Livré |
| i18n FR / NL / EN / ES (UI + erreurs API) | Livré |
| Google OAuth + 2FA TOTP (optionnel) | Livré |
| Admin plateforme (métriques, users, payments, commercials) | Livré |
| Stripe billing par animal + addons Family / Care+ / Horse | Livré |
| Commissions véto + commercial (ledger, fiches UI `ProCommissionSheet`) | Livré |
| Espace commercial (overview, vets, prospects, commissions, pitch) | Livré |
| Link-requests client → véto (`/requests`) | Livré |
| Horse pack (contacts, compétitions) + care reminders | Livré |
| Médias (avatars / photos / messages) local + GCS staging | Livré |
| Préférences email véto, durées FC configurables, changement MDP | Livré |
| Reset mot de passe email (forgot/reset) | Livré |

## Post-MVP / backlog

- FCM push (device tokens déjà en base)
- WebSocket temps réel
- Refresh token silencieux côté clients
- Simulation prospection admin 10 ans → [16](16-ADMIN-SIMULATION-10ANS.md) (**non livré**)
- Export / emails Care+ avancés

## Comptes seed

Voir [AGENTS.md](../AGENTS.md) pour la liste complète.

Mots de passe : `VetDemo123!` · `ClientDemo123!` · `AdminDemo123!` · `CommercialDemo123!`

| Rôle | Email utile |
|------|-------------|
| Véto | `vet.demo@petsfollow.test` |
| Client | `client.demo@petsfollow.test` |
| Commercial | `commercial.demo@petsfollow.test` |
| Admin | `admin.demo@petsfollow.test` |

Parcours spéciaux : `vet.onboarding@` (profil incomplet), `vet.unverified@` + token `demo-confirm-email`, `vet.reset@` + token `demo-reset-password`, `client.vide@` (sans animal).

## Tests web Pro

Playwright (`nuxtjs/tests/e2e/specs/`) : auth, locale, clients, messagerie, onboarding, admin, **commercial**, **requests**. Exécution complète sur staging après deploy ; validation des specs en CI PR (`playwright test --list`).
