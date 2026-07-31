# UC-CO-01 — Prospects CRM commercial

| | |
|--|--|
| **ID** | `UC-CO-01` |
| **Durée** | ~10 min |
| **Priorité** | Démo |
| **Surface** | Web Commercial |

## Objectif

Parcourir le CRM prospects du commercial (Camille) : overview, pipeline, transitions de statut.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Commercial | `commercial.demo@petsfollow.test` | `CommercialDemo123!` |

## Prérequis

- Staging Web.
- Seed Camille : prospects Bruxelles + vet.demo assigné.

## Étapes

1. Se connecter avec `commercial.demo`.
2. Arriver sur l’espace **Commercial** (overview / portfolio).
3. Ouvrir la liste des **prospects**.
4. Ouvrir un prospect → noter les champs (contact, RDV, statut).
5. Faire une transition simple si possible (ex. contact → RDV) **sans** écraser des données critiques si environnement partagé — sinon lire seulement et noter N/A sur l’écriture.
6. Vérifier qu’un prospect déjà « owned » par un collègue affiche un message adapté (si cas seed).

## Résultat attendu

- Overview et liste prospects chargent.
- Fiche prospect lisible ; transitions cohérentes si testées.

## Checklist

| | Résultat |
|--|----------|
| Overview | OK / KO / N/A |
| Liste prospects | OK / KO / N/A |
| Fiche / transition | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : E1.1–E1.2*
