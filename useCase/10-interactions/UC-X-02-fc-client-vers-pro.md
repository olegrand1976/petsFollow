# UC-X-02 — Relevé respiratoire (FR) client → VetPro

| | |
|--|--|
| **ID** | `UC-X-02` |
| **Durée** | ~12 min |
| **Priorité** | Démo |
| **Surfaces** | Flutter Client → Web VetPro |

## Objectif

Valider un relevé côté propriétaire et le voir apparaître dans le dossier animal côté cabinet (BPM + commentaire).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Animal seed actif côté client.demo / VetPlus.
- Enchaînement logique après `UC-CL-03`.

## Étapes

1. **App** — login `client.demo` → animal → relevé respiratoire → taps → **valider** + commentaire (ex. `UC-X-02 repos`).
2. Noter l’heure / BPM approximatif.
3. **Web** — login `vet.demo` → Clients → même animal.
4. Vérifier le relevé dans le tableau / chart / historique.
5. Vérifier le **commentaire**.

## Résultat attendu

- Relevé **validé** visible côté Pro avec BPM et date.
- Commentaire visible.
- Un relevé abandonné (non validé) ne doit pas apparaître côté véto.

## Checklist

| | Résultat |
|--|----------|
| Validate client | OK / KO / N/A |
| Visible Pro + BPM | OK / KO / N/A |
| Commentaire Pro | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H2, F2, C5*
