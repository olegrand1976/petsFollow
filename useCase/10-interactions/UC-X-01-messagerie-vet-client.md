# UC-X-01 — Messagerie véto ↔ client

| | |
|--|--|
| **ID** | `UC-X-01` |
| **Durée** | ~15 min |
| **Priorité** | Démo |
| **Surfaces** | Web VetPro + Flutter Client + Flutter Pro Light |

## Objectif

Envoyer un message depuis le cabinet (Web **ou** app Pro Light) et le retrouver dans l’app propriétaire (et inversement si temps). Couvrir aussi le fil **care_pro ↔ client** (thread distinct du cabinet).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Care pro (vet_light) | `vetlight.demo@petsfollow.test` | `CareProDemo123!` |
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- Staging Web + app mobile staging.
- Client lié au cabinet VetPlus (seed) ; Spirit partagé write_notes avec vetlight.

## Étapes

1. **Web** — login `vet.demo` → **Messages** → ouvrir le thread avec `client.demo` (ou le client seed correspondant).
2. Envoyer un message texte unique (ex. `UC-X-01 web 14:32`).
3. **App client** — login `client.demo` → onglet **Messages** → ouvrir le même fil → vérifier le message.
4. (Optionnel) Répondre depuis l’app client → rafraîchir le thread côté Web.
5. **App Pro Light staff** — login `vet.demo` → onglet **Messages** → composer/texte utilisable → envoyer un second message (ex. `UC-X-01 mobile 14:35`) → vérifier côté client.
6. **App Pro Light care_pro** — login `vetlight.demo` → onglet **Messages** → composer vers `client.demo` / Spirit → envoyer (ex. `UC-X-01 care 14:40`) → vérifier côté client (fil distinct du cabinet) → (optionnel) réponse client.

## Résultat attendu

- Message visible Web ↔ Client ↔ Pro Light staff (après refresh si pas de push).
- Message care_pro ↔ Client sur un **autre** thread (pas mélangé au fil VetPlus).
- Push FCM : bonus si reçu ; sinon N/A (ne pas bloquer le test).

## Checklist

| | Résultat |
|--|----------|
| Envoi Web → Client | OK / KO / N/A |
| Réponse Client → Web | OK / KO / N/A |
| Envoi Pro Light staff → Client | OK / KO / N/A |
| Envoi care_pro → Client | OK / KO / N/A |
| Push (si applicable) | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H1, A10, C4, F3, G1, G13*
