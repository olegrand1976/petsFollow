# UC-X-01 — Messagerie véto ↔ client

| | |
|--|--|
| **ID** | `UC-X-01` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surfaces** | Web VetPro + Flutter Client + Flutter Pro Light (staff) |

## Objectif

Envoyer un message depuis le cabinet (Web **ou** app Pro Light) et le retrouver dans l’app propriétaire (et inversement si temps).

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
2. Envoyer un message texte unique (ex. `UC-X-01 web 14:32`).
3. **App client** — login `client.demo` → onglet **Messages** → ouvrir le même fil → vérifier le message.
4. (Optionnel) Répondre depuis l’app client → rafraîchir le thread côté Web.
5. **App Pro Light** — login `vet.demo` → onglet **Messages** → composer/texte utilisable → envoyer un second message (ex. `UC-X-01 mobile 14:35`) → vérifier côté client.
6. (N/A) `vetlight.demo` / `care_pro` : **pas** d’onglet Messages en V1.

## Résultat attendu

- Message visible Web ↔ Client ↔ Pro Light staff (après refresh si pas de push).
- Push FCM : bonus si reçu ; sinon N/A (ne pas bloquer le test).

## Checklist

| | Résultat |
|--|----------|
| Envoi Web → Client | OK / KO / N/A |
| Réponse Client → Web | OK / KO / N/A |
| Envoi Pro Light → Client | OK / KO / N/A |
| Push (si applicable) | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H1, A10, C4, F3*
