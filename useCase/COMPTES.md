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
| Admin | `AdminDemo123!` |

## Comptes utilisés dans les UC

| Profil | Email | Surface | UC typiques |
|--------|-------|---------|-------------|
| Véto VetPlus | `vet.demo@petsfollow.test` | Web VetPro | VP-*, X-01…06, EQ-01 |
| Onboarding | `vet.onboarding@petsfollow.test` | Web VetPro | VP-02 (**Destructif**) |
| Collègue | `vet.colleague@petsfollow.test` | Web VetPro | EQ-01 |
| Assistante | `vet.assist@petsfollow.test` | Web VetPro | EQ-01 |
| Secrétaire | `secretary.demo@petsfollow.test` | Web VetPro | EQ-01 |
| Care pro farrier | `farrier.demo@petsfollow.test` | Flutter Pro Light | PL-*, X-05 |
| Care pro vet_light | `vetlight.demo@petsfollow.test` | Flutter Pro Light | PL-01 (alt.) |
| Commercial Camille | `commercial.demo@petsfollow.test` | Web Commercial | CO-*, X-07 |
| Manager | `commercial.manager@petsfollow.test` | Web Manager | CM-01 |
| Admin | `admin.demo@petsfollow.test` | Web Admin | AD-01 |
| Client riche | `client.demo@petsfollow.test` | Flutter Client | CL-01/03, X-01…03, X-06 (Spirit seed), X-08 |
| Client vide | `client.vide@petsfollow.test` | Flutter Client | **CL-02 uniquement** (**Destructif**) |

> **Ne pas** réutiliser `client.vide` pour le lien cabinet (`UC-X-04`) : utiliser un **email jetable** (voir X-04).

## Téléphones commerciaux seedés

Affichés sur la page de dossier partagé (`UC-X-08`) et modifiables dans `/commercial/settings`. Un commercial sans numéro est bloqué sur `/complete-contact-phone` jusqu'à la saisie.

| Commercial | Téléphone seedé |
|------------|-----------------|
| Camille (`commercial.demo@`) | `0470 12 34 56` |
| Alex (`commercial.demo2@`) | `0471 98 76 54` |
| Bérénice manager (`commercial.manager@`) | `0472 11 22 33` |

## Autres comptes seed (hors UC V1)

Disponibles pour tests ad hoc — détail dans [`AGENTS.md`](../AGENTS.md) :

| Email | Usage possible |
|-------|----------------|
| `vet.parc@` · `vet.lyon@` | Autres cabinets |
| `vet.unverified@` · `vet.reset@` | Auth confirm / reset MDP |
| `commercial.demo2@` (Alex) | 2ᵉ commercial / Parc |
| `client.marie@` (NL) · `client.paul@` · `client.julie@` · `client.thomas@` | Multi-cabinet / i18n |

## Tokens démo (Web)

| Action | URL / token |
|--------|-------------|
| Confirm email | `/confirm-email?token=demo-confirm-email` |
| Reset password | `/reset-password?token=demo-reset-password` |

## Offre client (rappel démo)

Plans TTC animal : **3,50 € / mois** · **35 € / an** · **95 € / 3 ans** (steer triennial).  
Care / Horse / foyer / relevés / messagerie inclus dès entitlement actif (pas d’addons vendus).
