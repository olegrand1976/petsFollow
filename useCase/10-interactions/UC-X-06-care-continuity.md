# UC-X-06 — Continuité Care (rappel Pro ↔ client)

| | |
|--|--|
| **ID** | `UC-X-06` |
| **Durée** | ~12 min |
| **Priorité** | Important |
| **Surfaces** | Web VetPro + Flutter Client |

## Objectif

Créer un rappel Care côté cabinet et le voir / le cocher côté propriétaire ; synchroniser l’état (dont en retard).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- Animal seed avec Care activé (client.demo).
- Staging Web + app.

## Étapes

1. **Web** — fiche animal → créer un **rappel Care** (libellé unique, échéance proche).
2. **App** — login `client.demo` → onglet **Care** → trouver le rappel.
3. Marquer le rappel comme **fait** / done côté client.
4. **Web** — rafraîchir : statut à jour ; vérifier le signal en retard dashboard si un rappel est en retard (seed ou test volontaire).

## Résultat attendu

- Rappel créé Pro visible onglet Care client.
- Done client reflété côté Pro.
- Overdue signalé sur le dashboard si applicable.

## Checklist

| | Résultat |
|--|----------|
| Création Pro | OK / KO / N/A |
| Visible onglet Care | OK / KO / N/A |
| Sync done / en retard | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H7 · rappels Care Pro C6 · onglet Care client F4.1–F4.2*
