# Comptes démo — use cases

Source de vérité complète : [`AGENTS.md`](../AGENTS.md).  
Environnement : **staging** (ou seed local).

## Mots de passe

| Rôle | Mot de passe |
|------|----------------|
| Véto / équipe cabinet | `VetDemo123!` |
| Client | `ClientDemo123!` |
| Care pro (Pro Light) | `CareProDemo123!` |
| Commercial / manager | `CommercialDemo123!` |
| Admin / DEV | `AdminDemo123!` |
| Research | `ResearchDemo123!` |

## Comptes utilisés dans les UC

| Profil | Email | Surface | UC typiques |
|--------|-------|---------|-------------|
| Research (tag `dev`) | `research.demo@petsfollow.test` | Web `/research` | *(pas d’UC commercial tant que tag `dev`)* |
| Research via multi-profil | `vet.demo@petsfollow.test` (switch profil → `research`) | Web `/research` | Même MDP véto — seed attache le profil chercheur |
| Research via multi-profil | `admin.demo@petsfollow.test` (switch → `research`) | Web `/research` | MDP admin |
| Véto VetPlus | `vet.demo@petsfollow.test` | Web VetPro | VP-*, X-01…06, EQ-01 |
| Onboarding | `vet.onboarding@petsfollow.test` | Web VetPro | VP-02 (**Destructif**) |
| Collègue | `vet.colleague@petsfollow.test` | Web VetPro | EQ-01 |
| Assistante | `vet.assist@petsfollow.test` | Web VetPro | EQ-01 |
| Secrétaire | `secretary.demo@petsfollow.test` | Web VetPro | EQ-01 |
| Care pro farrier | `farrier.demo@petsfollow.test` | Flutter Pro Light | PL-*, X-05 |
| Care pro vet_light | `vetlight.demo@petsfollow.test` | Flutter Pro Light | PL-01 (alt.) |
| Commercial Camille | `commercial.demo@petsfollow.test` | Web Commercial | CO-*, X-07 — switch profils Pro (sauf admin / manager / dev) |
| Manager | `commercial.manager@petsfollow.test` | Web Manager | CM-01 — switch profils Pro (sauf admin / dev) |
| Admin | `admin.demo@petsfollow.test` | Web Admin (ops + switch tous profils Pro) | AD-01 |
| DEV support IT | `dev.demo@petsfollow.test` | Web Admin (ops) | AD-02 |
| Client riche | `client.demo@petsfollow.test` | Flutter Client | CL-01/03, X-01…03, X-06 (Spirit seed), X-08 |
| Client vide | `client.vide@petsfollow.test` | Flutter Client | **CL-02 uniquement** (**Destructif**) |
| Client nouveau | `client.nouveau@petsfollow.test` | Flutter Client | Lyon — Buddy · pending `/requests` VetPlus |

> **Ne pas** réutiliser `client.vide` pour le lien cabinet (`UC-X-04`) : utiliser un **email jetable** (voir X-04).

## Téléphones commerciaux seedés

Affichés sur la page de dossier partagé (`UC-X-08`) et modifiables dans `/commercial/settings`. Un commercial sans numéro est bloqué sur `/complete-contact-phone` jusqu'à la saisie.

| Commercial | Téléphone seedé |
|------------|-----------------|
| Camille (`commercial.demo@`) | `0470 12 34 56` |
| Alex (`commercial.demo2@`) | `0471 98 76 54` |
| Bérénice manager (`commercial.manager@`) | `0472 11 22 33` |

## Téléphones clients seedés

Visibles / filtrables sur `/clients` (VetPro) et éditables sur la fiche Identité.

| Client | Téléphone seedé |
|--------|-----------------|
| Sophie (`client.demo@`) | `0470 00 00 01` |
| Luc (`client.vide@`) | `0470 00 00 02` |
| Marie (`client.marie@`) | `0470 00 00 03` |
| Paul (`client.paul@`) | `0470 00 00 04` |
| Julie (`client.julie@`) | `0470 00 00 05` |
| Thomas (`client.thomas@`) | `0470 00 00 06` |
| Nina (`client.nouveau@`) | `0470 00 00 07` |

## Autres comptes seed (hors UC V1)

Disponibles pour tests ad hoc — détail dans [`AGENTS.md`](../AGENTS.md) :

| Email | Usage possible |
|-------|----------------|
| `vet.parc@` · `vet.lyon@` | Autres cabinets |
| `vet.unverified@` · `vet.reset@` | Auth confirm / reset MDP |
| `commercial.demo2@` (Alex) | 2ᵉ commercial / Parc |
| `client.nouveau@` (Lyon / Buddy, pending VetPlus) · `client.marie@` (NL, pending VetPlus) · `client.paul@` · `client.julie@` · `client.thomas@` | Multi-cabinet / i18n /requests |

## Tokens démo (Web)

| Action | URL / token |
|--------|-------------|
| Confirm email | `/confirm-email?token=demo-confirm-email` |
| Reset password | `/reset-password?token=demo-reset-password` |

## Offre client (rappel démo)

Plans TTC animal : **3,50 € / mois** · **35 € / an** · **95 € / 3 ans** (steer triennial).  
Care / Horse / foyer / relevés / messagerie inclus dès entitlement actif (pas d’addons vendus).
