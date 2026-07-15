# Vision produit — MVP petsFollow

## Personas

- **Dr Martin** — véto libéral, dashboard Pro web
- **Sophie** — propriétaire chien senior, app mobile pets

## Périmètre MVP (inclus)

1. Création animal (client — Flutter)
2. Suivi clients + animaux (véto — Nuxt Pro)
3. Messagerie interne + mode indisponible véto
4. Relevé cardiaque 60s — Valider envoi véto / Recommencer (client)
5. Timeline historique (messages, relevés validés, événements)

## Livré au-delà du MVP initial (web Pro)

| Extension | Statut |
|-----------|--------|
| Inscription véto + confirmation email + onboarding profil cabinet | Livré |
| i18n FR / NL / EN (UI + erreurs API) | Livré |
| Google OAuth + 2FA TOTP (optionnel) | Livré |
| Admin plateforme (métriques, users, payments) | Livré |
| Stripe billing par animal (client Flutter + smoke API) | Livré |
| Préférences email véto, durées FC configurables, changement MDP | Livré |
| Reset mot de passe email (forgot/reset) | Livré |

## Hors MVP / post-MVP

FCM push, documents GCS, rappels récurrents, WebSocket temps réel, refresh token silencieux, CRUD admin.

## Comptes seed

Voir [AGENTS.md](../AGENTS.md) pour la liste complète (6 vétos, 7 clients, admin).

Mots de passe : `VetDemo123!` · `ClientDemo123!` · `AdminDemo123!`

Parcours spéciaux : `vet.onboarding@` (profil incomplet), `vet.unverified@` + token `demo-confirm-email`, `vet.reset@` + token `demo-reset-password`, `client.vide@` (sans animal).

## Tests web Pro

Playwright (`nuxtjs/tests/e2e/specs/`) : auth, locale, clients, messagerie, onboarding, admin. Exécution complète sur staging après deploy ; validation des specs en CI PR (`playwright test --list`).
