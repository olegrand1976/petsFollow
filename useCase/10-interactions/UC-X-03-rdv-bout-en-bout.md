# UC-X-03 — RDV bout-en-bout

| | |
|--|--|
| **ID** | `UC-X-03` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surfaces** | Flutter Client + Web VetPro |

## Objectif

Demander ou proposer un rendez-vous et le confirmer jusqu’à cohérence agenda des deux côtés.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Booking client activé côté cabinet (sinon : véto propose le RDV depuis la fiche pet).
- Staging Web + app.

## Étapes

**Variante A — demande client**

1. App `client.demo` → animal → demander / réserver une visite.
2. Web `vet.demo` → Agenda → trouver la demande → **confirmer**.
3. App : vérifier le statut confirmé (et push si dispo).

**Variante B — proposition véto**

1. Web → fiche pet → proposer un RDV.
2. App : vérifier la visite proposée / confirmée.

## Résultat attendu

- RDV visible agenda Pro et côté client.
- Statuts cohérents (demandé / confirmé / annulé si testé).

## Checklist

| | Résultat |
|--|----------|
| Création / demande | OK / KO / N/A |
| Confirmation | OK / KO / N/A |
| Visibilité bilatérale | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H3 · RDV Pro C3 · booking / Care&RDV client F4.4–F4.5*
