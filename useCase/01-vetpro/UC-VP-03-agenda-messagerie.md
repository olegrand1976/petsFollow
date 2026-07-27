# UC-VP-03 — Agenda et messagerie (solo VetPro)

| | |
|--|--|
| **ID** | `UC-VP-03` |
| **Durée** | ~8 min |
| **Priorité** | Démo |
| **Surface** | Web VetPro |

> **Skip** si [`UC-X-01`](../10-interactions/UC-X-01-messagerie-vet-client.md) déjà passé (couvre messagerie bilatérale).

## Objectif

Vérifier côté cabinet seul que l’agenda charge et que la messagerie liste les conversations clients.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |

## Prérequis

- Staging Web.
- Pour le test bilatéral (message reçu côté client) : voir `UC-X-01`.

## Étapes

1. Se connecter en `vet.demo`.
2. Ouvrir **Agenda / Calendrier**.
3. Vérifier qu’au moins une visite seed est visible (ou que l’agenda charge sans erreur).
4. Ouvrir **Messages**.
5. Ouvrir un fil de discussion existant.
6. (Optionnel) Envoyer un message texte de test.

## Résultat attendu

- Agenda charge sans page blanche / erreur.
- Liste des threads visible ; un thread s’ouvre avec historique.

## Checklist

| | Résultat |
|--|----------|
| Agenda | OK / KO / N/A |
| Liste messages | OK / KO / N/A |
| Ouverture thread | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : A3–A4, C3.1, C4*
