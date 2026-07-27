# UC-CL-03 — Relevé cardiaque (côté client)

| | |
|--|--|
| **ID** | `UC-CL-03` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Flutter Client |

> **Skip** si [`UC-X-02`](../10-interactions/UC-X-02-fc-client-vers-pro.md) déjà passé (couvre validate + côté VetPro).

## Objectif

Démarrer un relevé cardiaque sur un animal, saisir les taps, puis valider (avec commentaire optionnel).

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- Animal avec suivi actif (seed `client.demo`).
- Pour voir le résultat côté VetPro : enchaîner avec `UC-X-02`.

## Étapes

1. Se connecter en `client.demo`.
2. Ouvrir un animal → lancer un **relevé cardiaque**.
3. Effectuer les taps pendant le timer (15 / 30 / 60 s selon options).
4. **Valider** le relevé.
5. Ajouter un **commentaire** si proposé (ex. « repos »).
6. Vérifier que le relevé apparaît dans l’historique animal côté client.

## Résultat attendu

- Timer et taps fonctionnels.
- Après validation : relevé enregistré côté client (BPM + date).
- Un relevé non validé ne doit pas être considéré comme envoyé au véto.

## Checklist

| | Résultat |
|--|----------|
| Démarrage FC | OK / KO / N/A |
| Validation | OK / KO / N/A |
| Historique client | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : A7, F2*
