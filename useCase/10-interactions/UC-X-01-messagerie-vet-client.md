# UC-X-01 — Messagerie véto ↔ client

| | |
|--|--|
| **ID** | `UC-X-01` |
| **Durée** | ~12 min |
| **Priorité** | Démo |
| **Surfaces** | Web VetPro + Flutter Client |

## Objectif

Envoyer un message depuis le cabinet et le retrouver dans l’app propriétaire (et inversement si temps).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- Staging Web + app mobile staging.
- Client lié au cabinet VetPlus (seed).

## Étapes

1. **Web** — login `vet.demo` → **Messages** → ouvrir le thread avec `client.demo` (ou le client seed correspondant).
2. Envoyer un message texte unique (ex. `UC-X-01 test 14:32`).
3. **App** — login `client.demo` → onglet **Messages** → ouvrir le même fil.
4. Vérifier le message du véto.
5. (Optionnel) Répondre depuis l’app → rafraîchir le thread côté Web.

## Résultat attendu

- Message visible des deux côtés (après refresh si pas de push).
- Push FCM : bonus si reçu ; sinon N/A (ne pas bloquer le test).

## Checklist

| | Résultat |
|--|----------|
| Envoi Web → Client | OK / KO / N/A |
| Réponse Client → Web | OK / KO / N/A |
| Push (si applicable) | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H1, A10, C4, F3*
