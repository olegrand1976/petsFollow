# UC-AD-02 — DEV support IT (ops léger)

| | |
|--|--|
| **ID** | `UC-AD-02` |
| **Durée** | ~10 min |
| **Priorité** | Secondaire |
| **Surface** | Web Admin (rôle `dev`) |

## Objectif

Montrer le rôle **DEV support IT** : accès ops utile (users, support, flags) **sans** billing, sales force, brand ni AI.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| DEV | `dev.demo@petsfollow.test` | `AdminDemo123!` |

## Prérequis

- Staging Web (ou seed local + migration `000114`).
- Compte seed multi-profils (`dev` + `client` + `vet` VetPlus) — bascule profil possible.

## Étapes

1. Se connecter avec `dev.demo`.
2. Vérifier le **dashboard admin** (vue allégée / pas de KPI billing).
3. Ouvrir **Utilisateurs** : liste lisible.
4. Ouvrir **Support** : liste tickets (+ recherche si présente) ; ouvrir un ticket.
5. Ouvrir **Runtime flags** : lecture des flags.
6. Vérifier la **nav** : pas d’entrées Commercials / Paiements / Imports / Brand / AI.
7. (Optionnel) Tenter `/admin/payments` ou `/admin/commercials` via URL → refus / redirect.
8. (Optionnel) Bascule profil vers client ou véto → shell adapté, retour profil `dev` OK.

## Résultat attendu

- Accès support IT sans erreur.
- Billing / sales / brand / AI **inaccessibles** (nav absente + API 403).
- Un admin garde la surface complète (`UC-AD-01`).

## Checklist

| | Résultat |
|--|----------|
| Login DEV | OK / KO / N/A |
| Users + Support + Flags | OK / KO / N/A |
| Nav limitée | OK / KO / N/A |
| Billing/sales refusés | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : D17 · Go `TestDevRoleSupportAndBillingGate`*
