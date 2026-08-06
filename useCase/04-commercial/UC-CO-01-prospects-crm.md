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
4. Ouvrir un prospect (lien fiche `/commercial/prospects/[id]`) → timeline, tâches, statut/RDV.
5. Ajouter une note / appel et créer une tâche due J+1, puis la cocher.
6. Ouvrir **Agenda** (`/commercial/agenda`) → RDV + tâches de la semaine.
7. Faire une transition simple si possible (ex. contact → RDV) **sans** écraser des données critiques si environnement partagé — sinon lire seulement et noter N/A sur l’écriture.
8. Vérifier qu’un prospect déjà « owned » par un collègue affiche un message adapté (si cas seed).
9. Ouvrir **Emails / Templates** → parcourir le catalogue (intro, RDV, nurture J+1/J+3/J+7…).
10. Depuis une fiche prospect avec e-mail : **Envoyer un e-mail** → choisir un template → envoyer → voir l’historique (statut / ouvertures / clics) sur `/commercial/emails`.

## Résultat attendu

- Overview et liste prospects chargent.
- Fiche unifiée : timeline chronologique (notes, appels, mails, statuts) + tâches.
- Agenda semaine + mes tâches visibles.
- Templates éditables ; envoi SMTP traqué (opens/clics) visible dans l’historique.

## Checklist

| | Résultat |
|--|----------|
| Overview | OK / KO / N/A |
| Liste prospects | OK / KO / N/A |
| Fiche / timeline / tâches | OK / KO / N/A |
| Agenda | OK / KO / N/A |
| Templates / envoi mail | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : E1.1–E1.2 / E1.2c*
