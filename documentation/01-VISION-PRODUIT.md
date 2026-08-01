# Vision produit — petsFollow

**Promesse** : continuité de soins prescrite, matérialisée en **passeport digital** de l’animal — Web cabinet · mobile ProLight / care pro · mobile particulier.  
Le relevé cardiaque est un **module** différenciant, pas l’identité produit. Positionnement → [14](14-POSITIONNEMENT-MARKETING.md).

## Personas

- **Dr Martin** — véto libéral, dashboard Pro web (+ mode terrain *vet_light* mobile)
- **Sophie** — propriétaire chien senior, app mobile pets (self-inscription possible)
- **Léa** — commerciale / apporteuse, espace Pro commercial
- **Marc** — maréchal-ferrant (`care_pro` / `farrier`), agenda terrain + CR ferrage
- **Dr Nora** — expert Research (`research`), Observatoire épidémio anonymisé (opt-in cabinets) — tag `dev` — voir [42](42-RESEARCH.md)
- Autres care pro : physio, comportementaliste, toiletteur, éleveur — voir [28](28-MULTI-PROFILS-PRO.md)

## Périmètre cœur (continuité + passeport)

1. Création animal (client — Flutter)
2. Suivi clients + animaux (véto — Nuxt Pro)
3. Messagerie interne + mode indisponible véto (véto ↔ propriétaire — **pas** care_pro)
4. Timeline historique (messages, relevés validés, événements)
5. Partage multi-acteurs (ACL `pet_access` / `client_access`) — collègue / care pro / notes / CR / docs
6. Relevé cardiaque (durée 15/30/60 s selon paramètres du cabinet) — Valider envoi véto / Recommencer (client) — **feature**

## Livré au-delà du MVP initial

| Extension | Statut |
|-----------|--------|
| Inscription véto + confirmation email + onboarding profil cabinet | Livré |
| i18n FR / NL / EN / ES / ET / IT (UI + erreurs API) | Livré |
| Google OAuth + 2FA TOTP (optionnel) | Livré |
| Admin plateforme (métriques, users, payments, commercials) | Livré |
| Stripe billing par animal (monthly / annual / triennial ; quinquennial + addons = legacy hors vente) | Livré |
| Commissions véto + commercial (ledger, fiches UI `ProCommissionSheet`) | Livré |
| Espace commercial (overview, vets, prospects, commissions, pitch) | Livré |
| Link-requests client → véto (`/requests`) | Livré |
| Care reminders + Horse pack (inclus entitlement animal actif ; plus vendus en addon) | Livré |
| Médias (avatars / photos / messages) local + GCS staging | Livré |
| Préférences email véto, durées FC configurables, changement MDP | Livré |
| Reset mot de passe email (forgot/reset) | Livré |

## Post-MVP / backlog

- FCM push (device tokens déjà en base)
- WebSocket temps réel
- Refresh token silencieux côté clients
- Simulation prospection admin 10 ans → [16](16-ADMIN-SIMULATION-10ANS.md) (**non livré**)
- Export / emails Care avancés (features incluses — plus de roadmap « addon Care+ »)
- Multi-profils pro / partage / CR IA → [28](28-MULTI-PROFILS-PRO.md) (**livré** — axes passeport ; messagerie care_pro Pro Light livrée)
- petsFollow Research (observatoire anonymisé, rôle `research`) → [42](42-RESEARCH.md) (**tag `dev`**)

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
