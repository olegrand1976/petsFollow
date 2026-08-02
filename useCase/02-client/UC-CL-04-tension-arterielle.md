# UC-CL-04 — Tension artérielle (côté client)

| | |
|--|--|
| **ID** | `UC-CL-04` |
| **Durée** | ~8 min |
| **Priorité** | Démo |
| **Surface** | Flutter Client |

> **Skip** si [`UC-X-10`](../10-interactions/UC-X-10-tension-labos.md) déjà passé (couvre aussi le côté VetPro + labos).

## Objectif

Saisir une tension artérielle (SYS/DIA) depuis la fiche animal et la voir dans l’historique.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Client | `client.demo@petsfollow.test` | `ClientDemo123!` |

## Prérequis

- Animal avec suivi actif (seed `client.demo`).

## Étapes

1. Se connecter en `client.demo`.
2. Ouvrir un animal → action **Tension**.
3. Saisir SYS / DIA (ex. 140 / 90), méthode Doppler, enregistrer.
4. Vérifier le snackbar de confirmation et la courbe / timeline.

## Résultat attendu

- Tension enregistrée immédiatement (pas de validation type FR).
- Visible dans l’historique client.

## Checklist

| | Résultat |
|--|----------|
| Saisie tension | OK / KO / N/A |
| Historique client | OK / KO / N/A |

## Zone retour

- Bug :
- Amélioration :
