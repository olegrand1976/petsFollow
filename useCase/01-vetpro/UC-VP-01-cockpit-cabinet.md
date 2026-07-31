# UC-VP-01 — Cockpit cabinet VetPro

| | |
|--|--|
| **ID** | `UC-VP-01` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Web VetPro |

## Objectif

Vérifier que le vétérinaire arrive dans son espace cabinet, voit le tableau de bord, ouvre un client et consulte le dossier d’un animal.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web : [https://petsfollow.ll-it-sc.be](https://petsfollow.ll-it-sc.be)
- Navigateur desktop (Chrome ou Safari)

## Étapes

1. Ouvrir le site staging → page de connexion.
2. Se connecter avec `vet.demo@petsfollow.test` / `VetDemo123!`.
3. Vérifier l’arrivée sur le **tableau de bord** (chiffres et aperçu du cabinet visibles).
4. Aller dans **Clients** → ouvrir un client (ex. client démo lié au cabinet).
5. Ouvrir un **animal** du client.
6. Parcourir rapidement : historique / relevés / care / infos générales.

## Résultat attendu

- Connexion sans erreur.
- menu du site (barre latérale et en-tête) visible.
- Fiche client et dossier animal lisibles, données seed présentes.

## Checklist

| | Résultat |
|--|----------|
| Login → dashboard | OK / KO / N/A |
| Liste clients | OK / KO / N/A |
| Dossier animal | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ Remonter via [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md).

---

*Réf. QA : A1–A2, C2*
