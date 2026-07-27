# UC-X-05 — Partage animal → care pro (Pro Light)

| | |
|--|--|
| **ID** | `UC-X-05` |
| **Durée** | ~12 min |
| **Priorité** | Démo |
| **Surfaces** | Web VetPro → Flutter Pro Light |

## Objectif

Partager un animal depuis le cabinet avec un care pro ; vérifier la visibilité et les notes côté Pro Light.

## Acteurs

| Rôle | Compte | MDP |
|------|--------|-----|
| Véto | `vet.demo@petsfollow.test` | `VetDemo123!` |
| Farrier | `farrier.demo@petsfollow.test` | `CareProDemo123!` |

## Prérequis

- Seed : Spirit souvent déjà partagé avec farrier — vérifier ou créer un share.
- App staging + Web.

## Étapes

1. **Web** — login `vet.demo` → dossier animal (ex. **Spirit**, chez `client.demo`) → **Partager**.
2. Sélectionner le care pro maréchal (`farrier`) avec le droit d’**écrire des notes** si proposé.
3. **App** — login `farrier.demo` → Clients / Animaux → ouvrir l’animal partagé.
4. Vérifier la fiche (lecture).
5. (Optionnel) Ajouter une **note** terrain.

## Résultat attendu

- Animal visible en Pro Light après partage.
- Notes possibles si le droit d’écriture a été accordé ; lecture seule sinon.

## Checklist

| | Résultat |
|--|----------|
| Share depuis Pro | OK / KO / N/A |
| Visible Pro Light | OK / KO / N/A |
| Note (si ACL) | OK / KO / N/A |

## Zone retour

- Bug :
- Friction UX :
- Idée :

→ [`TEMPLATE-RETOUR.md`](../TEMPLATE-RETOUR.md)

---

*Réf. QA : H5 · partage Pro C6.3 · Pro Light G*
